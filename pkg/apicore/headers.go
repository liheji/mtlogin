package apicore

import (
	"net/http"
	"strings"
)

func ToHTTPHeader(headers map[string]string) http.Header {
	out := make(http.Header, len(headers))
	for key, value := range headers {
		out.Set(key, value)
	}
	return out
}

func HeaderValue(headers map[string]string, key string) string {
	for k, v := range headers {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

func FirstHeaderValue(headers map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := HeaderValue(headers, key); value != "" {
			return value
		}
	}
	return ""
}
