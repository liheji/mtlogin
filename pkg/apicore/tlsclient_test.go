package apicore

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/bogdanfinn/tls-client/profiles"
)

type fakeHTTPClient struct {
	mu              sync.Mutex
	req             *fhttp.Request
	closeIdleCalled bool
	response        *fhttp.Response
	err             error
	do              func(*fhttp.Request) (*fhttp.Response, error)
}

func (f *fakeHTTPClient) Do(req *fhttp.Request) (*fhttp.Response, error) {
	f.mu.Lock()
	f.req = req
	f.mu.Unlock()
	if f.do != nil {
		return f.do(req)
	}
	if f.response != nil || f.err != nil {
		return f.response, f.err
	}
	return &fhttp.Response{
		StatusCode: http.StatusCreated,
		Header:     fhttp.Header{"Did": {"new-did"}, "Content-Type": {"application/json"}},
		Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
	}, nil
}

func (f *fakeHTTPClient) CloseIdleConnections() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closeIdleCalled = true
}

func TestRoundTripConvertsRequestAndResponse(t *testing.T) {
	fake := &fakeHTTPClient{}
	var captured clientConfig
	client := newTLSClient(func(cfg clientConfig) (httpClient, error) {
		captured = cfg
		return fake, nil
	})

	resp, err := client.RoundTrip(context.Background(), &Request{
		URL:             "https://api.m-team.io/api/member/profile",
		Method:          http.MethodPost,
		Timeout:         5 * time.Second,
		Proxy:           "http://127.0.0.1:8080",
		DisableRedirect: true,
		UserAgent:       "ua",
		Headers: map[string]string{
			"Authorization": "token",
			"Did":           "did",
			"Content-Type":  "application/json",
		},
		Body: `{"safe":"ok"}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != http.StatusCreated || resp.Body != `{"ok":true}` || resp.Headers["Did"] != "new-did" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if captured.proxy != "http://127.0.0.1:8080" || captured.timeout != 5*time.Second || !captured.disableRedirect {
		t.Fatalf("request config was not used: %+v", captured)
	}
	if fake.req.Method != http.MethodPost || fake.req.URL.String() != "https://api.m-team.io/api/member/profile" {
		t.Fatalf("unexpected outgoing request: method=%s url=%s", fake.req.Method, fake.req.URL.String())
	}
	if fake.req.Header.Get("Authorization") != "token" || fake.req.Header.Get("Did") != "did" || fake.req.Header.Get("User-Agent") != "ua" {
		t.Fatalf("unexpected outgoing headers: %+v", fake.req.Header)
	}
	body, err := io.ReadAll(fake.req.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"safe":"ok"}` {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestRoundTripReusesClientForSameConfig(t *testing.T) {
	created := 0
	fake := &fakeHTTPClient{}
	client := newTLSClient(func(cfg clientConfig) (httpClient, error) {
		created++
		return fake, nil
	})
	req := Request{URL: "https://api.m-team.io/api/one", Method: http.MethodGet, Timeout: 3 * time.Second}
	if _, err := client.RoundTrip(context.Background(), &req); err != nil {
		t.Fatal(err)
	}
	req.URL = "https://api.m-team.io/api/two"
	if _, err := client.RoundTrip(context.Background(), &req); err != nil {
		t.Fatal(err)
	}
	if created != 1 {
		t.Fatalf("expected one pooled client, got %d", created)
	}
	if fake.closeIdleCalled {
		t.Fatal("pooled client was closed after a request")
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if !fake.closeIdleCalled {
		t.Fatal("pooled client was not closed")
	}
}

func TestRoundTripSeparatesDifferentConfigs(t *testing.T) {
	created := 0
	client := newTLSClient(func(cfg clientConfig) (httpClient, error) {
		created++
		return &fakeHTTPClient{}, nil
	})
	first := Request{URL: "https://example.com/one", Method: http.MethodGet, Proxy: "http://127.0.0.1:8080"}
	second := Request{URL: "https://example.com/two", Method: http.MethodGet}
	if _, err := client.RoundTrip(context.Background(), &first); err != nil {
		t.Fatal(err)
	}
	if _, err := client.RoundTrip(context.Background(), &second); err != nil {
		t.Fatal(err)
	}
	if created != 2 {
		t.Fatalf("expected separate pooled clients, got %d", created)
	}
}

func TestRoundTripPropagatesContextCancellation(t *testing.T) {
	fake := &fakeHTTPClient{do: func(req *fhttp.Request) (*fhttp.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	}}
	client := newTLSClient(func(cfg clientConfig) (httpClient, error) { return fake, nil })
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := client.RoundTrip(ctx, &Request{URL: "https://example.com", Method: http.MethodGet})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context deadline, got %v", err)
	}
}

func TestCookieScopeSharesJarAndEmptyScopeIsStateless(t *testing.T) {
	var configs []clientConfig
	client := newTLSClient(func(cfg clientConfig) (httpClient, error) {
		configs = append(configs, cfg)
		return &fakeHTTPClient{}, nil
	})
	requests := []Request{
		{URL: "https://example.com/one", Method: http.MethodGet, CookieScope: "mteam", Timeout: time.Second},
		{URL: "https://example.com/two", Method: http.MethodGet, CookieScope: "mteam", Timeout: 2 * time.Second},
		{URL: "https://example.com/notify", Method: http.MethodGet, Timeout: 3 * time.Second},
	}
	for i := range requests {
		if _, err := client.RoundTrip(context.Background(), &requests[i]); err != nil {
			t.Fatal(err)
		}
	}
	if configs[0].jar == nil || configs[0].jar != configs[1].jar {
		t.Fatal("same cookie scope did not share a jar")
	}
	if configs[2].jar != nil {
		t.Fatal("empty cookie scope should not allocate a jar")
	}
}

func TestCloseRejectsNewRequests(t *testing.T) {
	client := newTLSClient(func(cfg clientConfig) (httpClient, error) { return &fakeHTTPClient{}, nil })
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	_, err := client.RoundTrip(context.Background(), &Request{URL: "https://example.com", Method: http.MethodGet})
	if !errors.Is(err, ErrTLSClientClosed) {
		t.Fatalf("expected ErrTLSClientClosed, got %v", err)
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentRoundTripsCreateOneClient(t *testing.T) {
	var mu sync.Mutex
	created := 0
	client := newTLSClient(func(cfg clientConfig) (httpClient, error) {
		mu.Lock()
		created++
		mu.Unlock()
		return &fakeHTTPClient{}, nil
	})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := Request{URL: "https://example.com", Method: http.MethodGet}
			if _, err := client.RoundTrip(context.Background(), &req); err != nil {
				t.Errorf("round trip: %v", err)
			}
		}()
	}
	wg.Wait()
	if created != 1 {
		t.Fatalf("expected one pooled client, got %d", created)
	}
}

func TestCloseWaitsForInflightRequest(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	fake := &fakeHTTPClient{do: func(req *fhttp.Request) (*fhttp.Response, error) {
		close(started)
		<-release
		return &fhttp.Response{StatusCode: http.StatusOK, Header: fhttp.Header{}, Body: io.NopCloser(strings.NewReader("ok"))}, nil
	}}
	client := newTLSClient(func(cfg clientConfig) (httpClient, error) { return fake, nil })
	requestDone := make(chan struct{})
	go func() {
		defer close(requestDone)
		req := Request{URL: "https://example.com", Method: http.MethodGet}
		_, _ = client.RoundTrip(context.Background(), &req)
	}()
	<-started
	closeDone := make(chan struct{})
	go func() {
		defer close(closeDone)
		_ = client.Close()
	}()
	select {
	case <-closeDone:
		t.Fatal("Close returned before the request completed")
	case <-time.After(10 * time.Millisecond):
	}
	close(release)
	<-requestDone
	<-closeDone
}

func TestClientProfileForUserAgentUsesOnlyBuiltInProfiles(t *testing.T) {
	chrome := clientProfileForUserAgent("Mozilla/5.0 Chrome/146.0.0.0 Safari/537.36")
	if chrome.GetClientHelloStr() != profiles.Chrome_146.GetClientHelloStr() {
		t.Fatal("expected Chrome UA to select built-in Chrome profile")
	}
	firefox := clientProfileForUserAgent("Mozilla/5.0 Firefox/148.0")
	if firefox.GetClientHelloStr() != profiles.Firefox_148.GetClientHelloStr() {
		t.Fatal("expected Firefox UA to select built-in Firefox profile")
	}
	unknown := clientProfileForUserAgent("custom-bot/1.0")
	if unknown.GetClientHelloStr() != profiles.DefaultClientProfile.GetClientHelloStr() {
		t.Fatal("expected unknown UA to use the default built-in profile")
	}
}

func TestRoundTripRejectsOversizedResponseBody(t *testing.T) {
	fake := &fakeHTTPClient{
		do: func(*fhttp.Request) (*fhttp.Response, error) {
			return &fhttp.Response{
				StatusCode: http.StatusOK,
				Header:     fhttp.Header{},
				Body:       io.NopCloser(strings.NewReader(strings.Repeat("a", maxResponseBytes+10))),
			}, nil
		},
	}
	client := newTLSClient(func(cfg clientConfig) (httpClient, error) { return fake, nil })

	_, err := client.RoundTrip(context.Background(), &Request{URL: "https://api.m-team.io/x", Method: http.MethodGet})
	if err == nil || !strings.Contains(err.Error(), "limit") {
		t.Fatalf("expected oversized body error, got %v", err)
	}
}
