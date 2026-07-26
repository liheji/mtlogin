package apimtauth

import (
	"testing"
	"time"
)

func TestConfigValidateAcceptsCompletePasswordAuth(t *testing.T) {
	cfg := Config{
		APIHost:      "api.m-team.io",
		Timeout:      time.Minute,
		PasswordAuth: PasswordAuthConfig{Username: "user", Password: "pass"},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
}
