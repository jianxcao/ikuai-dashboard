package ikuaiv4

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// APIBase is the fixed REST prefix for iKuai OS v4.0 local API.
	APIBase        = "/api/v4.0"
	defaultTimeout = 15 * time.Second
)

// NormalizeBaseURL trims whitespace/trailing slashes and ensures an http(s) scheme.
func NormalizeBaseURL(raw string) string {
	s := strings.TrimSpace(raw)
	s = strings.TrimRight(s, "/")
	if s == "" {
		return ""
	}
	lower := strings.ToLower(s)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		s = "https://" + s
	}
	return s
}

// SanitizeNil replaces bare `nil` tokens in JSON value positions with `null`.
// Some iKuai v4 firmware responses emit bare `nil`; quoted strings are preserved.
func SanitizeNil(body []byte) []byte {
	body = bytes.ReplaceAll(body, []byte("\r\n"), []byte("\n"))
	n := len(body)
	if n < 3 {
		return body
	}

	out := make([]byte, 0, n+8)
	inString := false
	for i := 0; i < n; i++ {
		c := body[i]
		if c == '\\' && inString && i+1 < n {
			out = append(out, c, body[i+1])
			i++
			continue
		}
		if c == '"' {
			inString = !inString
			out = append(out, c)
			continue
		}
		if inString {
			out = append(out, c)
			continue
		}
		if c == 'n' && i+2 < n && body[i+1] == 'i' && body[i+2] == 'l' {
			prevOK := i == 0 || !isIdentChar(body[i-1])
			nextOK := i+3 >= n || !isIdentChar(body[i+3])
			if prevOK && nextOK {
				out = append(out, 'n', 'u', 'l', 'l')
				i += 2
				continue
			}
		}
		out = append(out, c)
	}
	return out
}

func isIdentChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

// APIErrorDetail holds per-field validation details from the iKuai v4 API.
type APIErrorDetail struct {
	Field string `json:"field"`
	Type  string `json:"type"`
	Msg   string `json:"msg"`
}

// APIError represents an HTTP 4xx/5xx or non-zero iKuai v4 response code.
type APIError struct {
	HTTPStatus int
	Code       int
	Message    string
	Details    []APIErrorDetail
}

func (e *APIError) Error() string {
	s := fmt.Sprintf("[%d] %s", e.Code, e.Message)
	for _, d := range e.Details {
		s += fmt.Sprintf(" (%s: %s)", d.Field, d.Msg)
	}
	return s
}

var errorHints = map[int]string{
	1003:  "当前 API Token 权限不足或固件未授权此模块",
	2019:  "路由器自身无法访问外网或爱快云服务，请检查该节点的 WAN 出口与 DNS 设置",
	3001:  "参数错误，请检查必填项或字段格式",
	3007:  "API Token 无效或已过期，请在节点设置中更新令牌",
	1008:  "会话已过期，请检查路由器令牌状态",
	10014: "权限不足，请确认 iKuai 后台为该 API Token 分配了相应权限",
}

// Client is a pure-v4 HTTP client communicating with /api/v4.0 via Bearer Token.
type Client struct {
	BaseURL            string
	Token              string
	InsecureSkipVerify bool
	httpClient         *http.Client
	sem                chan struct{}
}

// NewClient constructs an iKuai v4 REST client with optional TLS skip-verify.
func NewClient(baseURL, token string, insecureSkipVerify bool) *Client {
	var transport *http.Transport
	if base, ok := http.DefaultTransport.(*http.Transport); ok {
		transport = base.Clone()
	} else {
		transport = &http.Transport{}
	}
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify} //nolint:gosec
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 24
	transport.IdleConnTimeout = 90 * time.Second
	transport.TLSHandshakeTimeout = 10 * time.Second
	transport.ForceAttemptHTTP2 = true

	return &Client{
		BaseURL:            NormalizeBaseURL(baseURL),
		Token:              strings.TrimSpace(token),
		InsecureSkipVerify: insecureSkipVerify,
		httpClient: &http.Client{
			Timeout:   defaultTimeout,
			Transport: transport,
		},
		sem: make(chan struct{}, 3),
	}
}

// NewClientWithHTTP constructs a v4 client using a custom *http.Client (for testing).
func NewClientWithHTTP(baseURL, token string, httpClient *http.Client) *Client {
	return &Client{
		BaseURL:    NormalizeBaseURL(baseURL),
		Token:      strings.TrimSpace(token),
		httpClient: httpClient,
	}
}

func (c *Client) buildURL(path string, params map[string]string) string {
	cleanPath := "/" + strings.TrimLeft(strings.TrimSpace(path), "/")
	if !strings.HasPrefix(cleanPath, APIBase) {
		cleanPath = APIBase + cleanPath
	}
	fullURL := c.BaseURL + cleanPath
	if len(params) > 0 {
		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		fullURL += "?" + q.Encode()
	}
	return fullURL
}

func (c *Client) headers() http.Header {
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	h.Set("Accept", "application/json")
	if c.Token != "" {
		h.Set("Authorization", "Bearer "+c.Token)
	}
	return h
}

// PreviewDryRun constructs a sanitized request preview without hitting the router.
func (c *Client) PreviewDryRun(method, path string, params map[string]string, body any) map[string]any {
	tokenPreview := "<empty>"
	if c.Token != "" {
		tokenPreview = "Bearer <redacted>"
	}
	preview := map[string]any{
		"dry_run": true,
		"method":  method,
		"url":     c.buildURL(path, params),
		"headers": map[string]string{
			"Accept":        "application/json",
			"Content-Type":  "application/json",
			"Authorization": tokenPreview,
		},
	}
	if body != nil {
		preview["body"] = body
	}
	return preview
}

func (c *Client) acquire(ctx context.Context) (func(), error) {
	if c.sem == nil {
		return func() {}, nil
	}
	select {
	case c.sem <- struct{}{}:
		return func() { <-c.sem }, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Get sends a GET request to the iKuai v4 API and extracts data/results payload.
func (c *Client) Get(ctx context.Context, path string, params map[string]string) (json.RawMessage, error) {
	release, err := c.acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("连接爱快 v4 接口超时: %w", err)
	}
	defer release()

	fullURL := c.buildURL(path, params)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header = c.headers()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("连接爱快 v4 接口失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return c.checkResponse(resp)
}

// Post sends a POST request with JSON body.
func (c *Client) Post(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.doJSON(ctx, http.MethodPost, path, body)
}

// Put sends a PUT request with JSON body.
func (c *Client) Put(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.doJSON(ctx, http.MethodPut, path, body)
}

// Patch sends a PATCH request with JSON body.
func (c *Client) Patch(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.doJSON(ctx, http.MethodPatch, path, body)
}

// Delete sends a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) (json.RawMessage, error) {
	return c.doJSON(ctx, http.MethodDelete, path, nil)
}

func (c *Client) doJSON(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	release, err := c.acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("连接爱快 v4 接口超时: %w", err)
	}
	defer release()

	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyBytes = b
	} else {
		bodyBytes = []byte("{}")
	}

	req, err := http.NewRequestWithContext(ctx, method, c.buildURL(path, nil), bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header = c.headers()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("连接爱快 v4 接口失败: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return c.checkResponse(resp)
}

func (c *Client) checkResponse(resp *http.Response) (json.RawMessage, error) {
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTTP %d 读取响应失败: %w", resp.StatusCode, err)
	}

	body := SanitizeNil(raw)
	trimmedBody := strings.TrimSpace(string(body))
	if trimmedBody == "" && resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return []byte(`{"message":"Success"}`), nil
	}
	if resp.StatusCode >= 400 && strings.HasPrefix(trimmedBody, "<") {
		reqPath := ""
		if resp.Request != nil && resp.Request.URL != nil {
			reqPath = resp.Request.URL.Path
		}
		return nil, &APIError{
			HTTPStatus: resp.StatusCode,
			Code:       resp.StatusCode,
			Message:    fmt.Sprintf("当前路由器固件未开放此接口 (%s, HTTP %d)", reqPath, resp.StatusCode),
		}
	}

	var envelope struct {
		Code    int              `json:"code"`
		Message string           `json:"message"`
		Errors  json.RawMessage  `json:"errors"`
		Data    json.RawMessage  `json:"data"`
		Results json.RawMessage  `json:"results"`
		RowID   json.RawMessage  `json:"rowid"`
		Details []APIErrorDetail `json:"details"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("HTTP %d 非 JSON 响应: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if resp.StatusCode >= 400 || (envelope.Code != 0 && envelope.Code != 20000) {
		msg := strings.TrimSpace(envelope.Message)
		if msg == "" {
			msg = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		if errExtra := extractErrorsString(envelope.Errors); errExtra != "" && !strings.Contains(msg, errExtra) {
			msg = fmt.Sprintf("%s（%s）", msg, errExtra)
		}
		if hint, ok := errorHints[envelope.Code]; ok {
			msg = msg + " — " + hint
		}
		return nil, &APIError{
			HTTPStatus: resp.StatusCode,
			Code:       envelope.Code,
			Message:    msg,
			Details:    envelope.Details,
		}
	}

	payload := envelope.Data
	if isEmptyJSON(payload) && !isEmptyJSON(envelope.Results) {
		payload = envelope.Results
	}

	if isEmptyJSON(payload) {
		msg := envelope.Message
		if msg == "" {
			msg = "ok"
		}
		synthetic := map[string]any{"message": msg}
		if !isEmptyJSON(envelope.RowID) {
			var rowID any
			dec := json.NewDecoder(bytes.NewReader(envelope.RowID))
			dec.UseNumber()
			if err := dec.Decode(&rowID); err == nil {
				synthetic["rowid"] = rowID
			}
		}
		b, _ := json.Marshal(synthetic)
		return b, nil
	}

	return payload, nil
}

func isEmptyJSON(raw json.RawMessage) bool {
	trimmed := strings.TrimSpace(string(raw))
	return trimmed == "" || trimmed == "null"
}

func extractErrorsString(raw json.RawMessage) string {
	if isEmptyJSON(raw) {
		return ""
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		var clean []string
		for _, item := range arr {
			if t := strings.TrimSpace(item); t != "" {
				clean = append(clean, t)
			}
		}
		return strings.Join(clean, "; ")
	}
	return ""
}
