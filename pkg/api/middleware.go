package api

import (
	"bufio"
	"net"
	"net/http"
	"time"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/sebest/xff"

	"github.com/zbiljic/authzy/pkg/config"
	"github.com/zbiljic/authzy/pkg/logger"
)

func withConfigMiddleware(c *config.Config) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = withConfig(ctx, c)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func userIPMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			ip := r.Header.Get("X-Real-IP")
			if ip == "" {
				ip = xff.GetRemoteAddr(r)
			}

			userIP := net.ParseIP(ip)
			if userIP != nil {
				ctx = withUserIP(ctx, userIP)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func requestIDMiddleware(c *config.APIConfig, log logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := ""
			if c.RequestIDHeader != "" {
				id = r.Header.Get(c.RequestIDHeader)
			}
			if id == "" {
				uid, err := uuid.NewV4()
				if err != nil {
					handleError(w, r, log, err)
					return
				}
				id = uid.String()
			}

			ctx := r.Context()
			ctx = withRequestID(ctx, id)

			if c.RequestIDHeader != "" {
				w.Header().Set(c.RequestIDHeader, id)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type responseLogger struct {
	w      http.ResponseWriter
	status int
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

func (l *responseLogger) Flush() {
	l.w.(http.Flusher).Flush()
}

func (l *responseLogger) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return l.w.(http.Hijacker).Hijack()
}

func (l *responseLogger) CloseNotify() <-chan bool {
	// staticcheck SA1019 CloseNotifier interface is required by gorilla compress handler
	// nolint:staticcheck
	return l.w.(http.CloseNotifier).CloseNotify() // skipcq: SCC-SA1019
}

func (l *responseLogger) Push(target string, opts *http.PushOptions) error {
	return l.w.(http.Pusher).Push(target, opts)
}

func (l *responseLogger) Write(b []byte) (int, error) {
	return l.w.Write(b)
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	if l.status == 0 {
		l.status = s
	}
}

func loggingContextMiddleware(log logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			ctx := r.Context()

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}
			logFields := logger.Fields{
				"method":      r.Method,
				"path":        r.URL.Path,
				"remote_addr": ip,
			}

			requestID := getRequestID(r.Context())
			if requestID != "" {
				logFields["request_id"] = requestID
				ctx = log.NewContext(ctx, logger.Fields{"request_id": requestID})
			}

			if v := getRequestID(r.Context()); v != "" {
				logFields["request_id"] = v
				ctx = log.NewContext(ctx, logger.Fields{"request_id": v})
			}

			if v := r.Referer(); v != "" {
				logFields["referrer"] = v
			}
			if v := r.UserAgent(); v != "" {
				logFields["user-agent"] = v
			}
			if v := r.Header.Get("X-Forwarded-For"); v != "" {
				logFields["x-forwarded-for"] = v
			}
			if v := r.Header.Get("X-Real-IP"); v != "" {
				logFields["x-real-ip"] = v
			}

			log.WithFields(logFields).Info("START")

			rlw := &responseLogger{w, 0}

			next.ServeHTTP(rlw, r.WithContext(ctx))

			status := rlw.status
			if status == 0 {
				status = http.StatusOK
			}
			logFieldsEnd := logger.Fields{
				"status":   status,
				"duration": time.Since(startTime).Seconds(),
			}

			if requestID != "" {
				logFieldsEnd["request_id"] = requestID
			}

			log.WithFields(logFieldsEnd).Info("END")
		})
	}
}
