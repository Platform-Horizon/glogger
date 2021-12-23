package glogger

import (
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	correlationIDKey = "X-Request-Id"
	contentTypeKey   = "Content-Type"
	userAgentKey     = "user-agent"
	forwardedHostKey = "X-Forwarded-Host"
	forwardedForKey  = "X-Forwarded-For"
)

// Request struct contains items of request info log.
type Request struct {
	Path        string `json:"path,omitempty"`
	Method      string `json:"method,omitempty"`
	Query       string `json:"query,omitempty"`
	ContentType string `json:"content-type,omitempty"`
	Scheme      string `json:"scheme,omitempty"`
	Protocol    string `json:"protocol,omitempty"`
	UserAgent   string `json:"userAgent,omitempty"`
}

// Response struct contains items of response info log.
type Response struct {
	StatusCode   int     `json:"statusCode,omitempty"`
	ResponseTime float64 `json:"responseTime,omitempty"`
}

// Host struct contains items of host info log.
type Host struct {
	Hostname          string `json:"hostname,omitempty"`
	ForwardedHostname string `json:"forwardedHostname,omitempty"`
	IP                string `json:"ip,omitempty"`
}

// HTTP is the struct of the log formatter
type HTTP struct {
	Request  *Request  `json:"request,omitempty"`
	Response *Response `json:"response,omitempty"`
}

func getCorrelationID(header http.Header) string {
	if correlationID := header.Get(correlationIDKey); correlationID != "" {
		return correlationID
	}

	correlationID, err := uuid.NewRandom()

	if err != nil {
		return ""
	}

	return correlationID.String()
}

func removePort(host string) string {
	return strings.Split(host, ":")[0]
}

func getIP(request *http.Request) string {
	result := request.Header.Get(forwardedForKey)

	if result == "" {
		result = request.RemoteAddr
	}

	return result
}

// LoggingMiddleware is a gorilla/mux middleware to log all requests
// It logs the incoming request and when request is completed.
func LoggingMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			start := time.Now()

			correlationID := getCorrelationID(r.Header)
			ctx := WithLogger(r.Context(), logrus.NewEntry(logger).WithFields(logrus.Fields{
				"correlationId": correlationID,
			}))

			writer := readableResponseWriter{writer: rw, statusCode: http.StatusOK}

			Get(ctx).WithFields(logrus.Fields{
				"http": HTTP{
					Request: &Request{
						Path:        r.URL.RequestURI(),
						Method:      r.Method,
						ContentType: r.Header.Get(contentTypeKey),
						UserAgent:   r.Header.Get(userAgentKey),
						Query:       r.URL.RawQuery,
						Scheme:      r.URL.Scheme,
						Protocol:    r.Proto,
					},
				},
				"host": Host{
					Hostname:          removePort(r.Host),
					ForwardedHostname: r.Header.Get(forwardedHostKey),
					IP:                getIP(r),
				},
			}).Trace("Incoming Request")

			next.ServeHTTP(&writer, r.WithContext(ctx))

			Get(ctx).WithFields(logrus.Fields{
				"http": HTTP{
					Request: &Request{
						Path:        r.URL.RequestURI(),
						Method:      r.Method,
						ContentType: r.Header.Get(contentTypeKey),
						UserAgent:   r.Header.Get(userAgentKey),
						Scheme:      r.URL.Scheme,
						Query:       r.URL.RawQuery,
						Protocol:    r.Proto,
					},
					Response: &Response{
						StatusCode:   writer.statusCode,
						ResponseTime: float64(time.Since(start).Seconds()),
					},
				},
				"host": Host{
					Hostname:          removePort(r.Host),
					ForwardedHostname: r.Header.Get(forwardedHostKey),
					IP:                getIP(r),
				},
			}).Info("Completed Request")

		})
	}
}
