package apinotify

import (
	"context"
	"net/http"
)

type Sender interface {
	GetJSON(ctx context.Context, server, rawURL string, header http.Header) (int, []byte, error)
	PostJSON(ctx context.Context, server, rawURL string, body []byte, header http.Header) (int, []byte, error)
}
