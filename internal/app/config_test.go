package app

import (
	"testing"
	"time"

	"mtlogin/pkg/conf"
)

func TestBuildMTeamConfigsUsesSingleSnapshot(t *testing.T) {
	raw := conf.MTeamConfig{
		Proxy:   "http://127.0.0.1:7890",
		APIHost: "api.m-team.io",
		Referer: "https://kp.m-team.cc/",
		Timeout: 60,
		PasswordAuth: conf.PasswordAuthConfig{
			Username:   "user",
			Password:   "pass",
			TotpSecret: "secret",
		},
		TokenAuth: conf.TokenAuthConfig{
			Auth:      "token",
			DID:       "did",
			VisitorID: "visitor",
		},
		Header: conf.MTeamHeaderConfig{
			Version:    "1.1.4",
			WebVersion: "1140",
			UserAgent:  "ua",
		},
	}

	auth, browse, err := buildMTeamConfigs(raw)
	if err != nil {
		t.Fatal(err)
	}
	if auth.APIHost != browse.APIHost || auth.Timeout != time.Minute || browse.Timeout != time.Minute {
		t.Fatalf("configs do not share snapshot values: auth=%+v browse=%+v", auth, browse)
	}
	if auth.PasswordAuth.Username != "user" || auth.TokenAuth.VisitorID != "visitor" {
		t.Fatalf("auth config lost credentials: %+v", auth)
	}
	if auth.UserAgent != "ua" || browse.UserAgent != "ua" {
		t.Fatalf("configs lost headers: auth=%+v browse=%+v", auth, browse)
	}
}
