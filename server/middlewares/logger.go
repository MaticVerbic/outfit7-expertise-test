package middlewares

import (
	"context"
	"expertisetest/config"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/sirupsen/logrus"
)

// LoggerMiddleware wraps each http request with some useful data for logging.
func LoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := time.Now()

		requestID := middleware.GetReqID(r.Context())
		if requestID == "" {
			return
		}

		log := logrus.WithFields(logrus.Fields{
			"request_id": requestID,
			"uri":        fmt.Sprintf("%s%s", r.Host, r.RequestURI),
		})

		fields := logrus.Fields{
			"http_proto":  r.Proto,
			"http_method": r.Method,
			"remote_addr": r.RemoteAddr,
			"request_id":  requestID,
			"user_agent":  r.UserAgent(),
		}

		// Log only initially with extra fields.
		log.WithFields(fields).Debug("http request")

		defer func(s time.Time, logger *logrus.Entry) {
			logger.WithFields(logrus.Fields{
				"request_id": requestID,
				"elapsed":    time.Since(s),
			}).Info("http request processed")
		}(s, log)

		r = r.WithContext(context.WithValue(r.Context(), config.LogKey, log))

		h.ServeHTTP(w, r)
	})
}
