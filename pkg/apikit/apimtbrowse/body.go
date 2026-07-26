package apimtbrowse

import (
	"bytes"
	"mime/multipart"
	"net/url"
)

func MultipartBody(fields map[string]string) (string, string) {
	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)
	for key, value := range fields {
		_ = writer.WriteField(key, value)
	}
	_ = writer.Close()
	return body.String(), writer.FormDataContentType()
}

func QueryBody(fields map[string]string) string {
	body := url.Values{}
	for key, value := range fields {
		body.Add(key, value)
	}
	return body.Encode()
}
