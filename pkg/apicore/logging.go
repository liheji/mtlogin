package apicore

import (
	"time"
)

type LogEvent struct {
	Server          string
	Method          string
	URL             string
	Status          int
	Headers         map[string]string
	ResponseHeaders map[string]string
	Input           []byte
	Output          []byte
	Cost            time.Duration
	Err             error
}

func RequestBody(req *Request) []byte {
	if req == nil {
		return nil
	}
	if len(req.BodyBytes) > 0 {
		return req.BodyBytes
	}
	return []byte(req.Body)
}
