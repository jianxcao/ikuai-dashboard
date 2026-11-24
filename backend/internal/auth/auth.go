// Package auth 提供 Dashboard 用户名密码认证和 Session Token 管理。
//
// 凭据通过环境变量配置：
//   - USER_NAME：登录用户名，默认 "admin"
//   - PASSWORD：登录密码，默认 "admin"
//   - SESSION_TTL_DAYS：Session 有效期（天），默认 365
//   - SESSION_SECRET：可选会话签名主密钥（未设置时自动使用进程级随机盐）
//   - DISABLE_AUTH：设为 "true" 跳过所有认证（仅开发用途）
//
// Session Token 格式（无状态，HMAC-SHA256 签名）：
//
//	base64url(payload) + "." + base64url(HMAC-SHA256(secret, base64url(payload)))
//
// payload = username + ":" + unix_expiry_timestamp
package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"ikuai-dashboard/backend/internal/config"
)

var (
	once        sync.Once
	sessionTTL  time.Duration
	processSalt []byte
)

func init() {
	once.Do(initAuth)
}

func initAuth() {
	day := 365
	if v := os.Getenv("SESSION_TTL_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			day = n
		}
	}
	sessionTTL = time.Duration(day*24) * time.Hour

	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err == nil {
		processSalt = salt
	} else {
		fallback := sha256.Sum256([]byte(fmt.Sprintf("%d:%d", os.Getpid(), time.Now().UnixNano())))
		processSalt = fallback[:]
	}
}

func effectiveCreds() (string, string, []byte) {
	user := strings.TrimSpace(os.Getenv("USER_NAME"))
	pass := os.Getenv("PASSWORD")
	if cfg := config.Snapshot(); cfg != nil {
		if user == "" && strings.TrimSpace(cfg.Server.Username) != "" {
			user = strings.TrimSpace(cfg.Server.Username)
		}
		if pass == "" && cfg.Server.Password != "" {
			pass = cfg.Server.Password
		}
	}
	if user == "" {
		user = "admin"
	}
	if pass == "" {
		pass = "admin"
	}

	var hash [32]byte
	if customSecret := strings.TrimSpace(os.Getenv("SESSION_SECRET")); customSecret != "" {
		hash = sha256.Sum256([]byte("ikuai-v4:" + customSecret + ":" + user + ":" + pass))
	} else {
		buf := make([]byte, 0, len(processSalt)+len(user)+len(pass)+16)
		buf = append(buf, processSalt...)
		buf = append(buf, ':')
		buf = append(buf, user...)
		buf = append(buf, ':')
		buf = append(buf, pass...)
		hash = sha256.Sum256(buf)
	}
	return user, pass, hash[:]
}

// IsAuthEnabled 返回是否启用了登录认证。
// DISABLE_AUTH=true 时返回 false（跳过所有认证）。
func IsAuthEnabled() bool {
	return !strings.EqualFold(strings.TrimSpace(os.Getenv("DISABLE_AUTH")), "true")
}

// ValidateCredentials 验证用户名和密码是否正确。
func ValidateCredentials(username, password string) bool {
	u, p, _ := effectiveCreds()
	userOK := subtle.ConstantTimeCompare([]byte(username), []byte(u)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(password), []byte(p)) == 1
	return userOK && passOK
}

// GenerateSessionToken 为指定用户名生成 Session Token。
func GenerateSessionToken(username string) string {
	expiry := time.Now().Add(sessionTTL).Unix()
	payload := fmt.Sprintf("%s:%d", username, expiry)
	encoded := base64.RawURLEncoding.EncodeToString([]byte(payload))
	_, _, secret := effectiveCreds()
	sig := sign(encoded, secret)
	return encoded + "." + sig
}

// CurrentUsername 从合法的 Session Token 中解析用户名，失败时返回空字符串。
func CurrentUsername(token string) string {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return ""
	}
	encoded, sig := parts[0], parts[1]
	_, _, secret := effectiveCreds()
	if !hmac.Equal([]byte(sig), []byte(sign(encoded, secret))) {
		return ""
	}
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return ""
	}
	payload := string(raw)
	idx := strings.LastIndex(payload, ":")
	if idx <= 0 {
		return ""
	}
	expiry, err := strconv.ParseInt(payload[idx+1:], 10, 64)
	if err != nil || time.Now().Unix() >= expiry {
		return ""
	}
	return payload[:idx]
}

// ValidateSessionToken 验证 Session Token 是否有效（未篡改且未过期）。
func ValidateSessionToken(token string) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	encoded, sig := parts[0], parts[1]

	// 验证签名
	_, _, secret := effectiveCreds()
	if !hmac.Equal([]byte(sig), []byte(sign(encoded, secret))) {
		return false
	}

	// 解码 payload，验证有效期
	raw, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return false
	}
	payload := string(raw)
	idx := strings.LastIndex(payload, ":")
	if idx < 0 {
		return false
	}
	expiry, err := strconv.ParseInt(payload[idx+1:], 10, 64)
	if err != nil {
		return false
	}
	return time.Now().Unix() < expiry
}

// SessionTTLSeconds 返回 Session 有效期秒数。
func SessionTTLSeconds() int64 {
	return int64(sessionTTL.Seconds())
}

// sign 使用内存 secret 对 data 进行 HMAC-SHA256 签名，返回 base64url 编码结果。
func sign(data string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
