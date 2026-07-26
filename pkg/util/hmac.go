package util

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

// HmacSha1Base64 计算 HMAC-SHA1 并返回 Base64 字符串。
func HmacSha1Base64(message string, secret string) string {
	h := hmac.New(sha1.New, []byte(secret))
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
