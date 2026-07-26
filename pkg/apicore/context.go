package apicore

import "context"

type contextKey string

const (
	proxyKey   contextKey = "apicore.proxy"
	serverKey  contextKey = "apicore.server"
	headersKey contextKey = "apicore.headers"
)

func WithProxy(ctx context.Context, proxy string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if proxy == "" {
		return ctx
	}
	return context.WithValue(ctx, proxyKey, proxy)
}

func ProxyFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	proxy, _ := ctx.Value(proxyKey).(string)
	return proxy
}

func WithServer(ctx context.Context, server string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if server == "" {
		return ctx
	}
	return context.WithValue(ctx, serverKey, server)
}

func ServerFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	server, _ := ctx.Value(serverKey).(string)
	return server
}

func WithHeaders(ctx context.Context, headers map[string]string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(headers) == 0 {
		return ctx
	}
	return context.WithValue(ctx, headersKey, copyHeaders(headers))
}

func HeadersFromContext(ctx context.Context) map[string]string {
	if ctx == nil {
		return nil
	}
	headers, _ := ctx.Value(headersKey).(map[string]string)
	return copyHeaders(headers)
}

func copyHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}
	out := make(map[string]string, len(headers))
	for key, value := range headers {
		out[key] = value
	}
	return out
}
