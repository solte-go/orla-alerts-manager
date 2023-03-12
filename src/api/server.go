package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

type ctxKey int8

const (
	ctxKeyRequestID ctxKey = iota
)

type Handler interface {
	Register(s *Server) (string, chi.Router)
}

type Server struct {
	router *chi.Mux
	logger *zap.Logger
}

func NewServer(logger *zap.Logger) *Server {
	r := chi.NewRouter()
	return &Server{router: r, logger: logger}
}

func (s *Server) Run(ctx context.Context, port int, handlers ...Handler) {
	for _, h := range handlers {
		path, router := h.Register(s)
		s.router.Mount(path, router)
	}

	//TODO graceful shutdown
	go func() {
		<-ctx.Done()
		s.logger.Warn("server shutdown")
	}()

	s.logger.Debug(fmt.Sprintf("Server running on port %d", port))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), s.router)
	if err != nil {
		panic(err)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) Error(w http.ResponseWriter, r *http.Request, code int, err error) {
	s.Respond(w, r, code, map[string]string{"error": err.Error()})
}

func (s *Server) Respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.WriteHeader(code)
	if data != nil {
		err := json.NewEncoder(w).Encode(data)
		if err != nil {
			s.logger.Error("Writing error", zap.Error(err))
			return
		}
	}
}
