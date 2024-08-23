package musicbot

import (
	"context"
	"log/slog"
	"net/http"
)

type HttpServer struct {
	server *http.Server
}

func NewHttpServer(handler func(http.ResponseWriter, *http.Request)) *HttpServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/my_track", handler)

	s := &HttpServer{
		server: &http.Server{
			Addr:    "localhost:8080",
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
