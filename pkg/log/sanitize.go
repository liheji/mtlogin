package log

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const redacted = "[REDACTED]"

var sensitiveKeys = map[string]struct{}{
	"access_token":        {},
	"agent_secret":        {},
	"agentsecret":         {},
	"auth":                {},
	"authorization":       {},
	"bot_token":           {},
	"corp_secret":         {},
	"corpsecret":          {},
	"did":                 {},
	"otp":                 {},
	"otpcode":             {},
	"password":            {},
	"proxy_authorization": {},
	"secret":              {},
	"sign":                {},
	"token":               {},
	"visitorid":           {},
	"visitor_id":          {},
	"webhook_token":       {},
	"x_api_key":           {},
	"x_auth_token":        {},
}

var telegramBotTokenPattern = regexp.MustCompile(`/bot[^/]+`)
var sensitiveTextValuePattern = regexp.MustCompile(`(?i)(access_token|agent_secret|auth|authorization|bot_token|corp_secret|did|otp|password|proxy_authorization|secret|token|visitor_?id)(\s*[:=]\s*)([^&\s,;}\]]+)`)

// SanitizeHeaders returns a copy of header with the values of any sensitive
// header (Authorization, Did, Visitor-ID, ...) replaced by a redaction marker.
// The header key itself is preserved so logs still show which headers were sent.
func SanitizeHeaders(header http.Header) http.Header {
	if header == nil {
		return nil
	}
	out := make(http.Header, len(header))
	for key, values := range header {
		if isSensitiveKey(key) {
			redactedValues := make([]string, len(values))
			for i := range redactedValues {
				redactedValues[i] = redacted
			}
			out[key] = redactedValues
			continue
		}
		copied := make([]string, len(values))
		copy(copied, values)
		out[key] = copied
	}
	return out
}

// MaskSecret returns an empty string for an empty value and the redaction
// marker otherwise. Use it for standalone secret fields (DID, visitorID) so the
// log records whether a value was present without leaking the value itself.
func MaskSecret(value string) string {
	if value == "" {
		return ""
	}
	return redacted
}

func SanitizeURL(raw string) string {
	if raw == "" {
		return raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return telegramBotTokenPattern.ReplaceAllString(raw, "/bot"+redacted)
	}
	if parsed.User != nil {
		parsed.User = url.User(redacted)
	}
	parsed.Path = telegramBotTokenPattern.ReplaceAllString(parsed.Path, "/bot"+redacted)
	query := parsed.Query()
	for key, values := range query {
		if !isSensitiveKey(key) {
			continue
		}
		for i := range values {
			values[i] = redacted
		}
		query[key] = values
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func SanitizeBody(contentType string, body []byte) []byte {
	if len(body) == 0 {
		return body
	}
	text := string(body)
	if strings.Contains(strings.ToLower(contentType), "json") {
		var value any
		if err := json.Unmarshal(body, &value); err == nil {
			sanitizeJSONValue(value)
			data, err := json.Marshal(value)
			if err == nil {
				return data
			}
		}
	}
	if values, err := url.ParseQuery(text); err == nil && len(values) > 0 {
		changed := false
		for key, items := range values {
			if !isSensitiveKey(key) {
				continue
			}
			changed = true
			for i := range items {
				items[i] = redacted
			}
			values[key] = items
		}
		if changed {
			return []byte(values.Encode())
		}
	}
	return []byte(SanitizeText(text))
}

func SanitizeText(text string) string {
	text = telegramBotTokenPattern.ReplaceAllString(text, "/bot"+redacted)
	return sensitiveTextValuePattern.ReplaceAllString(text, `${1}${2}`+redacted)
}

func sanitizeJSONValue(value any) {
	switch v := value.(type) {
	case map[string]any:
		for key, item := range v {
			if isSensitiveKey(key) {
				v[key] = redacted
				continue
			}
			sanitizeJSONValue(item)
		}
	case []any:
		for _, item := range v {
			sanitizeJSONValue(item)
		}
	}
}

func isSensitiveKey(key string) bool {
	normalized := strings.NewReplacer("-", "_", ".", "_").Replace(strings.ToLower(key))
	if _, ok := sensitiveKeys[normalized]; ok {
		return true
	}
	return strings.Contains(normalized, "password") ||
		strings.Contains(normalized, "secret") ||
		strings.Contains(normalized, "token") ||
		strings.Contains(normalized, "auth")
}
