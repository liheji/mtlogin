package apicore

import (
	"context"
	"mtlogin/pkg/log"
	"strings"
	"time"
)

func Logging() Tripper {
	return func(next RoundTripper) RoundTripper {
		return RoundTripFunc(func(ctx context.Context, req *Request) (*Response, error) {
			start := time.Now()
			resp, err := next.RoundTrip(ctx, req)
			cost := time.Since(start)

			event := LogEvent{Cost: cost, Err: err}
			if req != nil {
				event.Server = req.Server
				if event.Server == "" {
					event.Server = ServerFromContext(ctx)
				}
				event.Method = req.Method
				event.URL = req.URL
				event.Headers = copyMap(req.Headers)
				event.Input = append([]byte(nil), RequestBody(req)...)
			}
			if event.Server == "" {
				event.Server = "unknown"
			}
			if resp != nil {
				event.Status = resp.Status
				event.ResponseHeaders = copyMap(resp.Headers)
				event.Output = []byte(resp.Body)
			}
			msg := &log.Message{}
			if logCtx, ok := ctx.(*log.Context); ok && logCtx != nil && logCtx.LogID() != "" {
				msg.AddField("logid", logCtx.LogID())
			}
			msg.AddField("did", log.MaskSecret(HeaderValue(event.Headers, "Did")))
			msg.AddField("visitor_id", log.MaskSecret(FirstHeaderValue(event.Headers, "Visitor-ID", "visitorid")))
			msg.AddField("HTTP_"+strings.Title(event.Server), map[string]any{
				"server":  event.Server,
				"path":    log.SanitizeURL(event.URL),
				"method":  event.Method,
				"status":  event.Status,
				"IHeader": log.SanitizeHeaders(ToHTTPHeader(event.Headers)),
				"OHeader": log.SanitizeHeaders(ToHTTPHeader(event.ResponseHeaders)),
				"input":   string(log.SanitizeBody(HeaderValue(event.Headers, "Content-Type"), event.Input)),
				"output":  string(log.SanitizeBody(HeaderValue(event.ResponseHeaders, "Content-Type"), event.Output)),
				"cost": map[string]any{
					"millisecond": event.Cost / time.Millisecond,
					"duration":    event.Cost,
				},
				"error": errorString(err),
			})
			log.GetHttpLog().Info(msg.String())
			return resp, err
		})
	}
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return log.SanitizeText(err.Error())
}

func copyMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
