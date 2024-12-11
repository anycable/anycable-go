package wsrpc

import (
	"context"
	"errors"
	"log/slog"
)

type Server struct {
	conf *Config
	log  *slog.Logger
}

func NewServer(c *Config, l *slog.Logger) *Server {
	return &Server{conf: c, log: l}
}

func (s *Server) Start() error {
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

func (s *Server) Invoke(ctx context.Context, command string, payload []byte, meta *map[string]string) ([]byte, int, error) {
	return nil, 503, errors.New("Not implemented")
}
