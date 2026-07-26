package apicore

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type Client struct {
	rt RoundTripper
}

func NewClient(rt RoundTripper) *Client {
	return &Client{rt: rt}
}

func (c *Client) Do(ctx context.Context, req Request) (Response, error) {
	if c == nil || c.rt == nil {
		return Response{}, fmt.Errorf("apicore client is nil")
	}
	resp, err := c.rt.RoundTrip(ctx, &req)
	if err != nil {
		return Response{}, err
	}
	if resp == nil {
		return Response{}, fmt.Errorf("apicore response is nil")
	}
	return *resp, nil
}

func (c *Client) GetJSON(ctx context.Context, server, rawURL string, header http.Header) (int, []byte, error) {
	headers := headerToMap(header)
	headers["Accept"] = "application/json"
	return c.do(ctx, Request{
		Server:  server,
		URL:     rawURL,
		Method:  http.MethodGet,
		Headers: headers,
	})
}

func (c *Client) PostJSON(ctx context.Context, server, rawURL string, body []byte, header http.Header) (int, []byte, error) {
	headers := headerToMap(header)
	headers["Content-Type"] = "application/json"
	return c.do(ctx, Request{
		Server:    server,
		URL:       rawURL,
		Method:    http.MethodPost,
		Headers:   headers,
		Body:      string(body),
		BodyBytes: body,
	})
}

func (c *Client) do(ctx context.Context, req Request) (int, []byte, error) {
	resp, err := c.Do(ctx, req)
	if err != nil {
		return 0, nil, err
	}
	return resp.Status, []byte(resp.Body), nil
}

func headerToMap(header http.Header) map[string]string {
	out := make(map[string]string, len(header))
	for key, values := range header {
		if len(values) == 0 {
			continue
		}
		out[key] = strings.Join(values, ",")
	}
	return out
}
