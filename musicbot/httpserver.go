package musicbot

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

type HttpServer struct {
	server *http.Server
}

func NewHttpServer(handler func(http.ResponseWriter, *http.Request), host string, path string) *HttpServer {
	mux := http.NewServeMux()
	mux.HandleFunc(fmt.Sprintf("/%s", path), handler)

	s := &HttpServer{
		server: &http.Server{
			Addr:    host,
			Handler: mux,
		},
	}
	return s
}

func (s *HttpServer) Start() {
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("failed to start http server", slog.Any("err", err))
	}
}

func (s *HttpServer) Close(ctx context.Context) {
	if err := s.server.Shutdown(ctx); err != nil {
		slog.Error("failed to shut down http server", slog.Any("err", err))
	}
}
