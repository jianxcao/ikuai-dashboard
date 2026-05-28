// Package auth 提供 Dashboard 用户名密码认证和 Session Token 管理。
//
// 凭据通过环境变量配置：
//   - USER_NAME：登录用户名，默认 "admin"
//   - PASSWORD：登录密码，默认 "admin"
//   - SESSION_TTL_HOURS：Session 有效期（小时），默认 12
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
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	once         sync.Once
	hmacSecret   []byte
	sessionTTL   time.Duration
	disableAuth  bool
	loginUser    string
	loginPassRaw string
)

func init() {
	once.Do(initAuth)
}

func initAuth() {
	// 读取环境变量
	loginUser = getEnvOrDefault("USER_NAME", "admin")
	loginPassRaw = getEnvOrDefault("PASSWORD", "admin")
	disableAuth = strings.EqualFold(os.Getenv("DISABLE_AUTH"), "true")

	day := 365
	if v := os.Getenv("SESSION_TTL_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			day = 365
		}
	}
	sessionTTL = time.Duration(day*24) * time.Hour

	// 根据用户名和密码生成固定的 HMAC secret
	// 这样只要凭据不改，重启服务后之前的 Session 依然有效
	hash := sha256.Sum256([]byte(loginUser + ":" + loginPassRaw))
	hmacSecret = hash[:]
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// IsAuthEnabled 返回是否启用了登录认证。
// DISABLE_AUTH=true 时返回 false（跳过所有认证）。
func IsAuthEnabled() bool {
	return !disableAuth
}

// ValidateCredentials 验证用户名和密码是否正确。
func ValidateCredentials(username, password string) bool {
	return username == loginUser && password == loginPassRaw
}

// GenerateSessionToken 为指定用户名生成 Session Token。
func GenerateSessionToken(username string) string {
	expiry := time.Now().Add(sessionTTL).Unix()
	payload := fmt.Sprintf("%s:%d", username, expiry)
	encoded := base64.RawURLEncoding.EncodeToString([]byte(payload))
	sig := sign(encoded)
	return encoded + "." + sig
}

// ValidateSessionToken 验证 Session Token 是否有效（未篡改且未过期）。
func ValidateSessionToken(token string) bool {
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 {
		return false
	}
	encoded, sig := parts[0], parts[1]

	// 验证签名
	if !hmac.Equal([]byte(sig), []byte(sign(encoded))) {
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
func sign(data string) string {
	mac := hmac.New(sha256.New, hmacSecret)
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
