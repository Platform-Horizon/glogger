package glogger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"gotest.tools/assert"
)

type ExpectedLogFields struct {
	Level   logrus.Level
	Message string
	Time    int64
	Http    HTTP
	Host    Host
}

const hostname = "localhost"
const port = "3000"
const userAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36"
const clientHost = "client-host"
const contenType = "application/json; charset=utf-8"

var ip string
var defaultRequestPath = fmt.Sprintf("http://%s:%s/my-req", hostname, port)

func testMiddlewareInvocation(next http.HandlerFunc, requestID string, logger *logrus.Logger, requestPath string) *test.Hook {

	if requestPath == "" {
		requestPath = defaultRequestPath
	}

	request := httptest.NewRequest(http.MethodGet, requestPath, nil)
	request.Header.Add("Content-Type", contenType)
	request.Header.Add("x-request-id", requestID)
	request.Header.Add("user-agent", userAgent)
	request.Header.Add("x-forwarded-for", ip)
	request.Header.Add("x-forwarded-host", clientHost)
	ip = removePort(request.RemoteAddr)

	var hook *test.Hook

	if logger == nil {
		logger, hook = test.NewNullLogger()
		logger.SetLevel(logrus.TraceLevel)
	}

	if logger != nil {
		hook = test.NewLocal(logger)
	}

	handler := LoggingMiddleware(logger)
	server := handler(next)
	writer := httptest.NewRecorder()
	server.ServeHTTP(writer, request)

	return hook
}

func assertJSON(t *testing.T, str string) error {
	var fields logrus.Fields

	err := json.Unmarshal([]byte(str), &fields)
	return err
}

func TestLoggingMiddleware(t *testing.T) {
	t.Run("", func(t *testing.T) {
		var buffer bytes.Buffer
		logger, _ := Init(InitOptions{Level: "trace"})
		logger.Out = &buffer
		const logMessage = "New log message"
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Get(r.Context()).Info(logMessage)
		})
		hook := testMiddlewareInvocation(handler, "", logger, "")

		assert.Equal(t, len(hook.AllEntries()), 3, "Number of logs is not 3")
		str := buffer.String()

		for i, value := range strings.Split(strings.TrimSpace(str), "\n") {
			err := assertJSON(t, value)
			assert.Equal(t, err, nil, "log %d is not a JSON", i)
		}
	})

	t.Run("Test GET Route without query paramaters", func(t *testing.T) {
		statusCode := 200
		handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(statusCode)
		})
		hook := testMiddlewareInvocation(handler, "", nil, "http://localhost:3000/api/v1/users")
		entries := hook.AllEntries()

		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		actualIncomingRequest := entries[0]
		incomingRequestAssertions(t, actualIncomingRequest, ExpectedLogFields{
			Level:   logrus.TraceLevel,
			Message: "Incoming Request",
			Time:    time.Now().Unix(),
			Http: HTTP{
				Request: &Request{
					Path:        "/api/v1/users",
					Method:      "GET",
					ContentType: contenType,
					Scheme:      "http",
					Protocol:    "HTTP/1.1",
					UserAgent:   userAgent,
				},
				Response: nil,
			},
			Host: Host{
				Hostname:          hostname,
				IP:                ip,
				ForwardedHostname: clientHost,
			},
		})

		actualCompletedRequest := entries[1]
		completedRequestAssertions(t, actualCompletedRequest, ExpectedLogFields{
			Level:   logrus.TraceLevel,
			Message: "Incoming Request",
			Time:    time.Now().Unix(),
			Http: HTTP{
				Request: &Request{
					Path:        "/api/v1/users",
					Method:      "GET",
					ContentType: contenType,
					Scheme:      "http",
					Protocol:    "HTTP/1.1",
					UserAgent:   userAgent,
				},
				Response: &Response{
					StatusCode:   statusCode,
					ResponseTime: 0,
				},
			},
			Host: Host{
				Hostname:          hostname,
				IP:                ip,
				ForwardedHostname: clientHost,
			},
		})
	})

	t.Run("Test GET Route with query paramaters", func(t *testing.T) {
		statusCode := 200
		handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(statusCode)
		})

		hook := testMiddlewareInvocation(handler, "", nil, "http://localhost:3000/api/v1/users?name=Test")
		entries := hook.AllEntries()

		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		actualIncomingRequest := entries[0]
		incomingRequestAssertions(t, actualIncomingRequest, ExpectedLogFields{
			Level:   logrus.TraceLevel,
			Message: "Incoming Request",
			Time:    time.Now().Unix(),
			Http: HTTP{
				Request: &Request{
					Path:        "/api/v1/users?name=Test",
					Method:      "GET",
					ContentType: contenType,
					Query:       "name=Test",
					Scheme:      "http",
					Protocol:    "HTTP/1.1",
					UserAgent:   userAgent,
				},
				Response: nil,
			},
			Host: Host{
				Hostname:          hostname,
				IP:                ip,
				ForwardedHostname: clientHost,
			},
		})

		actualCompletedRequest := entries[1]
		completedRequestAssertions(t, actualCompletedRequest, ExpectedLogFields{
			Level:   logrus.TraceLevel,
			Message: "Incoming Request",
			Time:    time.Now().Unix(),
			Http: HTTP{
				Request: &Request{
					Path:        "/api/v1/users?name=Test",
					Method:      "GET",
					ContentType: contenType,
					Scheme:      "http",
					Protocol:    "HTTP/1.1",
					Query:       "name=Test",
					UserAgent:   userAgent,
				},
				Response: &Response{
					StatusCode:   statusCode,
					ResponseTime: 0,
				},
			},
			Host: Host{
				Hostname:          hostname,
				IP:                ip,
				ForwardedHostname: clientHost,
			},
		})
	})

	t.Run("Test POST Route", func(t *testing.T) {
		statusCode := 200
		handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(statusCode)
		})

		hook := testMiddlewareInvocation(handler, "", nil, "http://localhost:3000/api/v1/users?name=Test")
		entries := hook.AllEntries()

		assert.Equal(t, len(entries), 2, "Unexpected entries length.")

		actualIncomingRequest := entries[0]
		incomingRequestAssertions(t, actualIncomingRequest, ExpectedLogFields{
			Level:   logrus.TraceLevel,
			Message: "Incoming Request",
			Time:    time.Now().Unix(),
			Http: HTTP{
				Request: &Request{
					Path:        "/api/v1/users?name=Test",
					Method:      "POST",
					ContentType: contenType,
					Query:       "name=Test",
					Scheme:      "http",
					Protocol:    "HTTP/1.1",
					UserAgent:   userAgent,
				},
				Response: nil,
			},
			Host: Host{
				Hostname:          hostname,
				IP:                ip,
				ForwardedHostname: clientHost,
			},
		})

		actualCompletedRequest := entries[1]
		completedRequestAssertions(t, actualCompletedRequest, ExpectedLogFields{
			Level:   logrus.TraceLevel,
			Message: "Incoming Request",
			Time:    time.Now().Unix(),
			Http: HTTP{
				Request: &Request{
					Path:        "/api/v1/users?name=Test",
					Method:      "POST",
					ContentType: contenType,
					Scheme:      "http",
					Protocol:    "HTTP/1.1",
					Query:       "name=Test",
					UserAgent:   userAgent,
				},
				Response: &Response{
					StatusCode:   statusCode,
					ResponseTime: 0,
				},
			},
			Host: Host{
				Hostname:          hostname,
				IP:                ip,
				ForwardedHostname: clientHost,
			},
		})
	})
}

func incomingRequestAssertions(t *testing.T, actual *logrus.Entry, expected ExpectedLogFields) {
	http := actual.Data["http"].(HTTP)
	host := actual.Data["host"].(Host)

	assert.Assert(t, http.Request != nil, "Unexpected http Request nil")
	assert.Assert(t, http.Response == nil, "Unexpected http Response not nil")
	assert.Equal(t, http.Request.Method, expected.Http.Request.Method, "Unexpected http method for log in incoming request")
	assert.Equal(t, http.Request.Path, expected.Http.Request.Path, "Unexpected http request path for log in incoming request")
	assert.Equal(t, http.Request.ContentType, expected.Http.Request.ContentType, "Unexpected http request content-type for log in incoming request")
	assert.Equal(t, http.Request.Scheme, expected.Http.Request.Scheme, "Unexpected http request scheme for log in incoming request")
	assert.Equal(t, http.Request.Protocol, expected.Http.Request.Protocol, "Unexpected http request protocolog for log in incoming request")
	assert.Equal(t, http.Request.UserAgent, expected.Http.Request.UserAgent, "Unexpected http user-agent for log in incoming request")
	assert.Equal(t, http.Request.Query, expected.Http.Request.Query, "Unexpected http request query for log in incoming request")
	assert.Equal(t, host.IP, expected.Host.IP, "Unexpected host IP for log in incoming request")
	assert.Equal(t, host.Hostname, expected.Host.Hostname, "Unexpected hostname for log in incoming request")
	assert.Equal(t, host.ForwardedHostname, expected.Host.ForwardedHostname, "Unexpected forwarded-hostname for log in incoming request")
}

func completedRequestAssertions(t *testing.T, actual *logrus.Entry, expected ExpectedLogFields) {
	http := actual.Data["http"].(HTTP)
	host := actual.Data["host"].(Host)

	assert.Assert(t, http.Request != nil, "Unexpected http Request nil")
	assert.Equal(t, http.Request.Method, expected.Http.Request.Method, "Unexpected http method for log in completed request")
	assert.Equal(t, http.Request.Path, expected.Http.Request.Path, "Unexpected http request path for log in completed request")
	assert.Equal(t, http.Request.ContentType, expected.Http.Request.ContentType, "Unexpected http request content-type for log in completed request")
	assert.Equal(t, http.Request.Scheme, expected.Http.Request.Scheme, "Unexpected http request scheme for log in completed request")
	assert.Equal(t, http.Request.Protocol, expected.Http.Request.Protocol, "Unexpected http request protocolog for log in completed request")
	assert.Equal(t, http.Request.UserAgent, expected.Http.Request.UserAgent, "Unexpected http user-agent for log in completed request")
	assert.Equal(t, http.Request.Query, expected.Http.Request.Query, "Unexpected http request query for log in completed request")
	assert.Assert(t, http.Response != nil, "Unexpected http Response nil")
	assert.Equal(t, http.Response.StatusCode, expected.Http.Response.StatusCode, "Unexpected http response status code in completed request")
	assert.Assert(t, http.Response.ResponseTime != 0, "Unexpected http response time equal to 0")
	assert.Equal(t, host.IP, expected.Host.IP, "Unexpected host IP for log in completed request")
	assert.Equal(t, host.Hostname, expected.Host.Hostname, "Unexpected hostname for log in completed request")
	assert.Equal(t, host.ForwardedHostname, expected.Host.ForwardedHostname, "Unexpected forwarded-hostname for log in completed request")
}
