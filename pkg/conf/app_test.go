package conf

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadAppIncludesMTeamSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.toml")
	content := `[system]
proxy = "http://127.0.0.1:8080"

[app]
crontab = "0 * * * *"
min_delay = 1
max_delay = 2

[refresh]
timeout = 30
retry_on_auth_failure = true

[mt]
proxy = "socks5://127.0.0.1:1080"
api_host = "api.m-team.io"
referer = "https://kp.m-team.cc/"
timeout = 60

[mt.password_auth]
username = "user"
password = "password"
totp_secret = "secret"

[mt.token_auth]
auth = "token"
did = "did"
visitor_id = "visitor"

[mt.header]
version = "1.1.4"
web_version = "1140"
user_agent = "test-agent"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadApp(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MT.APIHost != "api.m-team.io" || cfg.MT.Header.UserAgent != "test-agent" {
		t.Fatalf("unexpected mt snapshot: %+v", cfg.MT)
	}
	if cfg.MT.PasswordAuth.Username != "user" || cfg.MT.TokenAuth.VisitorID != "visitor" {
		t.Fatalf("unexpected auth snapshot: %+v", cfg.MT)
	}
}

func TestLoadTOMLClassifiesReadFailure(t *testing.T) {
	var dst map[string]any
	err := LoadTOML(filepath.Join(t.TempDir(), "missing.toml"), &dst)
	if !errors.Is(err, ErrRead) {
		t.Fatalf("expected ErrRead, got %v", err)
	}
}

func TestLoadTOMLRejectsUnknownKeys(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("known = \"value\"\nunknwon = true\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var dst struct {
		Known string `toml:"known"`
	}
	if err := LoadTOML(path, &dst); !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), "unknwon") {
		t.Fatalf("expected unknown key rejection, got %v", err)
	}
}
