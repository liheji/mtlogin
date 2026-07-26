package log

import (
	"net/http"
	"strings"
	"testing"
)

func TestSanitizeHeadersRedactsIdentityHeaders(t *testing.T) {
	in := http.Header{
		"Did":           []string{"ak0rhehqiiugwst3nzxtmvyb681zk2op"},
		"Visitorid":     []string{"visitor-secret"},
		"Visitor-Id":    []string{"another-secret"},
		"Authorization": []string{"Bearer top-secret-token"},
		"User-Agent":    []string{"chrome"},
		"Content-Type":  []string{"application/json"},
	}
	out := SanitizeHeaders(in)

	for _, key := range []string{"Did", "Visitorid", "Visitor-Id", "Authorization"} {
		got := out.Get(key)
		if got != redacted {
			t.Errorf("header %q = %q, want %q", key, got, redacted)
		}
	}
	// Non-sensitive headers must survive unchanged.
	if out.Get("User-Agent") != "chrome" {
		t.Errorf("User-Agent = %q, want chrome", out.Get("User-Agent"))
	}
	if out.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", out.Get("Content-Type"))
	}
	// Original header values must not leak anywhere in the output.
	for key, values := range out {
		for _, v := range values {
			if strings.Contains(v, "ak0rhehqiiugwst3nzxtmvyb681zk2op") ||
				strings.Contains(v, "visitor-secret") ||
				strings.Contains(v, "top-secret-token") {
				t.Errorf("sensitive value leaked in header %q: %q", key, v)
			}
		}
	}
}

func TestSanitizeHeadersNil(t *testing.T) {
	if SanitizeHeaders(nil) != nil {
		t.Fatal("SanitizeHeaders(nil) should return nil")
	}
}

func TestMaskSecret(t *testing.T) {
	if got := MaskSecret(""); got != "" {
		t.Errorf("MaskSecret(\"\") = %q, want empty", got)
	}
	if got := MaskSecret("ak0rhehqiiugwst3nzxtmvyb681zk2op"); got != redacted {
		t.Errorf("MaskSecret(secret) = %q, want %q", got, redacted)
	}
}

func TestSanitizeBodyJSONNested(t *testing.T) {
	body := []byte(`{"outer":{"password":"p@ss","did":"device-123","ok":"keep"},"token":"t"}`)
	out := string(SanitizeBody("application/json", body))
	for _, leak := range []string{"p@ss", "device-123", `"t"`} {
		if strings.Contains(out, leak) {
			t.Errorf("sanitized JSON leaked %q: %s", leak, out)
		}
	}
	if !strings.Contains(out, "keep") {
		t.Errorf("non-sensitive value dropped: %s", out)
	}
}

func TestSanitizeURLBotTokenAndQuery(t *testing.T) {
	out := SanitizeURL("https://api.telegram.org/bot123456:ABCDEF/sendMessage?token=leak&chat_id=1")
	if strings.Contains(out, "123456:ABCDEF") {
		t.Errorf("bot token leaked: %s", out)
	}
	if strings.Contains(out, "leak") {
		t.Errorf("query token leaked: %s", out)
	}
	if !strings.Contains(out, "chat_id=1") {
		t.Errorf("non-sensitive query dropped: %s", out)
	}
}
