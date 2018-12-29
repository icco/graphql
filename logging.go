package graphql

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/felixge/httpsnoop"
	"github.com/hellofresh/logging-go/context"
	stackdriver "github.com/icco/logrus-stackdriver-formatter"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

// InitLogging initializes a logger to send things to stackdriver.
func InitLogging() *logrus.Logger {
	log.Formatter = stackdriver.NewFormatter()
	log.Level = logrus.DebugLevel
	log.SetOutput(os.Stdout)

	log.Info("Logger successfully initialised!")

	return log
}

// LoggingMiddleware is a middleware for writing request logs in a stuctured
// format to stackdriver.
func LoggingMiddleware() func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.New(r.Context()))

			// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#HttpRequest
			// https://github.com/icco/logrus-stackdriver-formatter/blob/sd-v2/formatter.go#L53
			request := &stackdriver.HttpRequest{
				RequestMethod: r.Method,
				RequestUrl:    r.RequestURI,
				RemoteIp:      r.RemoteAddr,
				Referer:       r.Referer(),
				UserAgent:     r.UserAgent(),
				RequestSize:   strconv.FormatInt(r.ContentLength, 10),
			}

			m := httpsnoop.CaptureMetrics(handler, w, r)

			request.Status = strconv.Itoa(m.Code)
			request.Latency = fmt.Sprintf("%.9fs", m.Duration.Seconds())
			request.ResponseSize = strconv.FormatInt(m.Written, 10)

			fields := logrus.Fields{"httpRequest": request}

			// No idea if this works
			traceHeader := r.Header.Get("X-Cloud-Trace-Context")
			if traceHeader != "" {
				fields["trace"] = traceHeader
			}

			log.WithFields(fields).Info("Completed request")
		})
	}
}
