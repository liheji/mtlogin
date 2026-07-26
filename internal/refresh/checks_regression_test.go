package refresh

import (
	"context"
	"errors"
	"testing"

	"mtlogin/internal/domain"
	"mtlogin/pkg/log"
)

func TestRunChecksPropagatesCriticalForumErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{name: "canceled", err: context.Canceled},
		{name: "deadline", err: context.DeadlineExceeded},
		{name: "authentication", err: domain.MTAuthFailed.WithCause(errors.New("session expired"))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := &orderedSession{browseForumErr: tt.err}
			_, err := runChecks(log.NewContext(context.Background()), session)
			if !errors.Is(err, tt.err) {
				t.Fatalf("expected forum error %v, got %v", tt.err, err)
			}
		})
	}
}

func TestRunChecksKeepsOrdinaryForumErrorsBestEffort(t *testing.T) {
	forumErr := errors.New("forum unavailable")
	session := &orderedSession{browseForumErr: forumErr}
	if _, err := runChecks(log.NewContext(context.Background()), session); err != nil {
		t.Fatalf("ordinary forum errors should remain best effort, got %v", err)
	}
}
