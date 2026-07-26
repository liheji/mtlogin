package apimtbrowse

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"mtlogin/pkg/apicore"
	"mtlogin/pkg/log"
)

type stubHandler struct {
	responses map[string]apicore.Response
	request   apicore.Request
}

func (e *stubHandler) RoundTrip(ctx context.Context, req *apicore.Request) (*apicore.Response, error) {
	e.request = *req
	for path, resp := range e.responses {
		if strings.Contains(req.URL, path) {
			return &resp, nil
		}
	}
	return &apicore.Response{Status: 200, Body: `{"code":"0","message":"SUCCESS"}`, Headers: map[string]string{}}, nil
}

func TestBrowseForumHappyPath(t *testing.T) {
	topicData := `{"code":"0","message":"SUCCESS","data":{"pageNumber":"1","pageSize":"100",
"total":"1","totalPages":"1","data":[
{"id":"123","author":"456","lastPost":{"authorId":"789"}}
]}}`
	handler := &stubHandler{responses: map[string]apicore.Response{
		"/api/forum/forums": {
			Status:  200,
			Body:    `{"code":"0","message":"SUCCESS","data":{"forumsList":[{"id":"2"},{"id":"3"}],"lastPost":{"1":{"post":{"authorId":"100"}}}}}`,
			Headers: map[string]string{},
		},
		"/api/member/bases": {
			Status: 200, Body: `{"code":"0","message":"SUCCESS","data":{"100":{"uid":"100"}}}`,
			Headers: map[string]string{},
		},
		"/api/forum/topic/search": {
			Status: 200, Body: topicData,
			Headers: map[string]string{},
		},
		"/api/forum/topic/viewHits": {
			Status: 200, Body: `{"code":"0","message":"SUCCESS"}`,
			Headers: map[string]string{},
		},
		"/api/forum/topic/detail": {
			Status: 200, Body: `{"code":"0","message":"SUCCESS"}`,
			Headers: map[string]string{},
		},
		"/api/member/updateLastBrowse": {
			Status: 200, Body: `{"code":"0","message":"SUCCESS"}`,
			Headers: map[string]string{},
		},
	}}
	for _, ep := range WarmupEndpoints() {
		if _, ok := handler.responses[ep.Path]; !ok {
			handler.responses[ep.Path] = apicore.Response{
				Status:  200,
				Body:    `{"code":"0","message":"SUCCESS"}`,
				Headers: map[string]string{},
			}
		}
	}

	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	session := client.NewSession(testState())

	updated, err := session.BrowseForum(log.NewContext(context.Background()))
	if err != nil {
		t.Fatalf("BrowseForum should succeed, got %v", err)
	}
	if !updated {
		t.Fatal("BrowseForum should report lastBrowse update on happy path")
	}
}

func TestBrowseForumPropagatesContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client, err := NewClient(Config{APIHost: "api.m-team.io"}, &captureHandler{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.NewSession(testState()).BrowseForum(log.NewContext(ctx))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}

func TestBrowseForumPropagatesDeadlineExceeded(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	client, err := NewClient(Config{APIHost: "api.m-team.io"}, &captureHandler{})
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.NewSession(testState()).BrowseForum(log.NewContext(ctx))
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline exceeded, got %v", err)
	}
}

func TestBrowseForumPropagatesAuthenticationFailure(t *testing.T) {
	handler := &captureHandler{response: apicore.Response{
		Status:  200,
		Body:    `{"code":"NOT_LOGIN","message":"need login"}`,
		Headers: map[string]string{},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.NewSession(testState()).BrowseForum(log.NewContext(context.Background()))
	if !errors.Is(err, ErrAuthFailed) {
		t.Fatalf("expected authentication failure, got %v", err)
	}
}

func TestBrowseForumKeepsBusinessFailuresBestEffort(t *testing.T) {
	handler := &captureHandler{response: apicore.Response{
		Status:  200,
		Body:    `{"code":"FAIL","message":"forum unavailable"}`,
		Headers: map[string]string{},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}

	updated, err := client.NewSession(testState()).BrowseForum(log.NewContext(context.Background()))
	if err != nil {
		t.Fatalf("business forum failure should be best effort, got %v", err)
	}
	if updated {
		t.Fatal("forum business failure should not report an update")
	}
}

func TestPhaseVisitForumsRejectsMalformedData(t *testing.T) {
	handler := &captureHandler{response: apicore.Response{
		Status:  200,
		Body:    `{"code":"0","message":"SUCCESS","data":"bad"}`,
		Headers: map[string]string{},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	session := client.NewSession(testState())

	_, _, err = session.PhaseVisitForums(log.NewContext(context.Background()))
	if !errors.Is(err, ErrBusiness) {
		t.Fatalf("expected MTBusinessFailed, got %v", err)
	}
}

func TestPhaseResolveMembersRejectsMalformedData(t *testing.T) {
	handler := &captureHandler{response: apicore.Response{
		Status:  200,
		Body:    `{"code":"0","message":"SUCCESS","data":"bad"}`,
		Headers: map[string]string{},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	session := client.NewSession(testState())

	err = session.PhaseResolveMembers(log.NewContext(context.Background()), []int64{100})
	if !errors.Is(err, ErrBusiness) {
		t.Fatalf("expected MTBusinessFailed, got %v", err)
	}
}

func TestPhaseSearchTopicsRejectsMalformedData(t *testing.T) {
	handler := &captureHandler{response: apicore.Response{
		Status:  200,
		Body:    `{"code":"0","message":"SUCCESS","data":"bad"}`,
		Headers: map[string]string{},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	session := client.NewSession(testState())

	_, _, _, err = session.PhaseSearchTopics(log.NewContext(context.Background()), 2)
	if !errors.Is(err, ErrBusiness) {
		t.Fatalf("expected MTBusinessFailed, got %v", err)
	}
}

func TestPhaseViewTopicChecksBusinessEnvelope(t *testing.T) {
	handler := &stubHandler{responses: map[string]apicore.Response{
		"/api/forum/topic/viewHits": {
			Status:  200,
			Body:    `{"code":"0","message":"SUCCESS"}`,
			Headers: map[string]string{},
		},
		"/api/forum/forums": {
			Status:  200,
			Body:    `{"code":"0","message":"SUCCESS"}`,
			Headers: map[string]string{},
		},
		"/api/forum/topic/detail": {
			Status:  200,
			Body:    `{"code":"FAIL","message":"topic failed"}`,
			Headers: map[string]string{},
		},
	}}
	client, err := NewClient(Config{APIHost: "api.m-team.io"}, handler)
	if err != nil {
		t.Fatal(err)
	}
	session := client.NewSession(testState())

	err = session.PhaseViewTopic(log.NewContext(context.Background()), 123)
	if !errors.Is(err, ErrBusiness) {
		t.Fatalf("expected MTBusinessFailed, got %v", err)
	}
}
