package apicore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

const defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36"

// maxResponseBytes 限制单个响应体读取上限，防止异常/恶意大响应耗尽内存。
const maxResponseBytes = 16 << 20 // 16 MiB

var ErrTLSClientClosed = errors.New("tls client is closed")

type httpClient interface {
	Do(req *fhttp.Request) (*fhttp.Response, error)
	CloseIdleConnections()
}

type clientKey struct {
	profile         string
	proxy           string
	timeoutMillis   int
	disableRedirect bool
	cookieScope     string
}

type clientConfig struct {
	profileName     string
	profile         profiles.ClientProfile
	proxy           string
	timeout         time.Duration
	disableRedirect bool
	jar             fhttp.CookieJar
}

type clientFactory func(clientConfig) (httpClient, error)

type TLSClient struct {
	mu        sync.Mutex
	newClient clientFactory
	clients   map[clientKey]httpClient
	jars      map[string]fhttp.CookieJar
	closed    bool
	closeDone chan struct{}
	inflight  sync.WaitGroup
}

func NewTLSClient() (*TLSClient, error) {
	return newTLSClient(defaultNewClient), nil
}

func newTLSClient(factory clientFactory) *TLSClient {
	return &TLSClient{
		newClient: factory,
		clients:   make(map[clientKey]httpClient),
		jars:      make(map[string]fhttp.CookieJar),
		closeDone: make(chan struct{}),
	}
}

func (c *TLSClient) RoundTrip(ctx context.Context, req *Request) (*Response, error) {
	if c == nil || c.newClient == nil {
		return nil, fmt.Errorf("tls client is nil")
	}
	if req == nil {
		return nil, fmt.Errorf("tls request is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	userAgent := effectiveUserAgent(*req)
	client, err := c.acquire(req, userAgent)
	if err != nil {
		return nil, err
	}
	defer c.inflight.Done()

	httpReq, err := buildRequest(ctx, *req, userAgent)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// 对不可信响应设上限，避免异常大响应导致内存放大 / OOM。
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > maxResponseBytes {
		return nil, fmt.Errorf("response body exceeds %d bytes limit", maxResponseBytes)
	}
	return &Response{
		Status:  resp.StatusCode,
		Body:    string(body),
		Headers: headersToMap(resp.Header),
	}, nil
}

func (c *TLSClient) acquire(req *Request, userAgent string) (httpClient, error) {
	profileName, profile := profileForUserAgent(userAgent)
	key := clientKey{
		profile:         profileName,
		proxy:           req.Proxy,
		timeoutMillis:   timeoutMilliseconds(req.Timeout),
		disableRedirect: req.DisableRedirect,
		cookieScope:     req.CookieScope,
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, ErrTLSClientClosed
	}
	if client := c.clients[key]; client != nil {
		c.inflight.Add(1)
		return client, nil
	}

	var jar fhttp.CookieJar
	if req.CookieScope != "" {
		jar = c.jars[req.CookieScope]
		if jar == nil {
			jar = tlsclient.NewCookieJar()
			c.jars[req.CookieScope] = jar
		}
	}
	client, err := c.newClient(clientConfig{
		profileName:     profileName,
		profile:         profile,
		proxy:           req.Proxy,
		timeout:         req.Timeout,
		disableRedirect: req.DisableRedirect,
		jar:             jar,
	})
	if err != nil {
		return nil, err
	}
	c.clients[key] = client
	c.inflight.Add(1)
	return client, nil
}

func (c *TLSClient) Close() error {
	if c == nil {
		return nil
	}
	c.mu.Lock()
	if c.closed {
		done := c.closeDone
		c.mu.Unlock()
		<-done
		return nil
	}
	c.closed = true
	done := c.closeDone
	c.mu.Unlock()

	c.inflight.Wait()
	c.mu.Lock()
	for _, client := range c.clients {
		client.CloseIdleConnections()
	}
	c.clients = nil
	c.jars = nil
	close(done)
	c.mu.Unlock()
	return nil
}

func defaultNewClient(cfg clientConfig) (httpClient, error) {
	options := []tlsclient.HttpClientOption{
		tlsclient.WithClientProfile(cfg.profile),
		tlsclient.WithRandomTLSExtensionOrder(),
	}
	if cfg.disableRedirect {
		options = append(options, tlsclient.WithNotFollowRedirects())
	}
	if cfg.timeout > 0 {
		options = append(options, tlsclient.WithTimeoutMilliseconds(timeoutMilliseconds(cfg.timeout)))
	}
	if cfg.proxy != "" {
		options = append(options, tlsclient.WithProxyUrl(cfg.proxy))
	}
	if cfg.jar != nil {
		options = append(options, tlsclient.WithCookieJar(cfg.jar))
	}
	return tlsclient.NewHttpClient(nil, options...)
}

func effectiveUserAgent(req Request) string {
	if req.UserAgent != "" {
		return req.UserAgent
	}
	for key, value := range req.Headers {
		if strings.EqualFold(key, "User-Agent") && value != "" {
			return value
		}
	}
	return defaultUserAgent
}

func clientProfileForUserAgent(userAgent string) profiles.ClientProfile {
	_, profile := profileForUserAgent(userAgent)
	return profile
}

func profileForUserAgent(userAgent string) (string, profiles.ClientProfile) {
	normalized := strings.ToLower(userAgent)
	switch {
	case strings.Contains(normalized, "firefox/"):
		return "firefox_148", profiles.Firefox_148
	case strings.Contains(normalized, "safari/") && !strings.Contains(normalized, "chrome/") && !strings.Contains(normalized, "chromium/"):
		return "safari_16_0", profiles.Safari_16_0
	case strings.Contains(normalized, "chrome/"), strings.Contains(normalized, "chromium/"), strings.Contains(normalized, "edg/"):
		return "chrome_146", profiles.Chrome_146
	default:
		return "default", profiles.DefaultClientProfile
	}
}

func buildRequest(ctx context.Context, req Request, userAgent string) (*fhttp.Request, error) {
	body := requestBody(req)
	httpReq, err := fhttp.NewRequestWithContext(ctx, req.Method, req.URL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}
	httpReq.Header.Set("User-Agent", userAgent)
	return httpReq, nil
}

func requestBody(req Request) []byte {
	if len(req.BodyBytes) > 0 {
		return req.BodyBytes
	}
	return []byte(req.Body)
}

func headersToMap(headers fhttp.Header) map[string]string {
	out := make(map[string]string, len(headers))
	for key, values := range headers {
		if len(values) > 0 {
			out[key] = values[0]
		}
	}
	return out
}

func timeoutMilliseconds(timeout time.Duration) int {
	if timeout <= 0 {
		return 0
	}
	return int(math.Ceil(float64(timeout) / float64(time.Millisecond)))
}
