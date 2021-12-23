package glogger

import (
	"net/http"
)

type readableResponseWriter struct {
	writer     http.ResponseWriter
	statusCode int
	length     int
}

func (writer *readableResponseWriter) WriteHeader(code int) {
	writer.statusCode = code
	writer.writer.WriteHeader(code)
}

func (writer *readableResponseWriter) Write(b []byte) (int, error) {
	n, err := writer.writer.Write(b)

	if err != nil {
		return n, err
	}

	writer.length += n

	return n, err
}

func (writer *readableResponseWriter) Header() http.Header {
	return writer.writer.Header()
}

func (writer *readableResponseWriter) Length() int {
	return writer.length
}
