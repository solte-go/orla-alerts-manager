package api

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	code int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.code = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (s *Server) SetRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.New().String()
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ctxKeyRequestID, id)))
	})
}

func (s *Server) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := s.logger.With(
			zap.String("type", "api_request"),
			zap.Any("src_addr", r.RemoteAddr),
			zap.Any("request_id", r.Context().Value(ctxKeyRequestID)),
		)

		logger.Info(fmt.Sprintf("started %s %s", r.Method, r.RequestURI))

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		level := zap.InfoLevel
		switch {
		case rw.code >= 500:
			level = zap.ErrorLevel
		case rw.code == 400 || rw.code >= 402:
			level = zap.WarnLevel
		default:
		}
		logger.Log(
			level,
			fmt.Sprintf("completed with %d %s in %v", rw.code, http.StatusText(rw.code), time.Now().Sub(start)),
		)
	})
}

func (s *Server) LogMetricsRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("Here")

		logger := s.logger.With(
			zap.String("type", "api_request"),
			zap.Any("src_addr", r.RemoteAddr),
			zap.Any("request_id", r.Context().Value(ctxKeyRequestID)),
		)

		logger.Debug(fmt.Sprintf("started %s %s", r.Method, r.RequestURI))

		start := time.Now()
		rw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(rw, r)

		level := zap.DebugLevel
		switch {
		case rw.code >= 500:
			level = zap.ErrorLevel
		case rw.code == 400 || rw.code >= 402:
			level = zap.WarnLevel
		default:
		}
		logger.Log(
			level,
			fmt.Sprintf("completed with %d %s in %v", rw.code, http.StatusText(rw.code), time.Now().Sub(start)),
		)
	})
}
