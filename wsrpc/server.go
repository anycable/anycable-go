package wsrpc

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"sync"

	"github.com/anycable/anycable-go/server"
	"github.com/anycable/anycable-go/ws"
)

type Server struct {
	conf *Config
	srv  *server.HTTPServer
	log  *slog.Logger

	clients    []*Client
	currentIdx int
	clientsLen int
	// Mapes client UIDs to indexes in the clients slice,
	// so we can quickly find a client by UID
	clientToIdx map[string]int
	clientsMu   sync.RWMutex
}

func NewServer(c *Config, l *slog.Logger) *Server {
	return &Server{
		conf:        c,
		log:         l,
		clients:     make([]*Client, 0),
		clientToIdx: make(map[string]int),
	}
}

func (s *Server) Start() error {
	if s.conf.Host != "" && s.conf.Host != server.Host {
		srv, err := server.NewServer(s.conf.Host, strconv.Itoa(s.conf.Port), server.SSL, 0)
		if err != nil {
			return err
		}
		s.srv = srv
	} else {
		srv, err := server.ForPort(strconv.Itoa(s.conf.Port))
		if err != nil {
			return err
		}
		s.srv = srv
	}

	s.srv.SetupHandler(s.conf.Path, ClientHandler(s, s.conf.WS, s.log))
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {

	return nil
}

func (s *Server) Authenticate(info *server.RequestInfo) bool {
	// TODO: check secret
	return true
}

func (s *Server) Invoke(ctx context.Context, command string, payload []byte, meta *map[string]string) ([]byte, int, error) {
	return nil, 503, errors.New("Not implemented")
}

func (s *Server) ServeClient(conn *ws.Connection, info *server.RequestInfo) {
	logger := s.log.With("uid", info.UID)
	logger.Debug("WS RPC connection established")

	client := NewClient(info.UID, conn, logger)

	s.checkinClient(client)

	go func() {
		if err := client.Serve(context.Background(), s.handleMessage); err != nil {
			s.checkoutClient(client)
			logger.Debug("client disconnected", "error", err)
			// TODO: re-enqueue unprocessed messages
		} else {
			s.checkoutClient(client)
		}
	}()
}

func (s *Server) handleMessage(msg []byte) {
}

func (s *Server) checkinClient(c *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	s.clients = append(s.clients, c)
	s.clientToIdx[c.ID()] = s.clientsLen
	s.clientsLen++
	// Make the new client the next one in line
	s.currentIdx = s.clientsLen - 1
}

func (s *Server) checkoutClient(c *Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	id := c.ID()

	idx, ok := s.clientToIdx[id]
	if !ok {
		s.log.Warn("trying to checkout uknown client", "uid", id)
		return
	}

	// Remove client from the slice
	s.clients = append(s.clients[:idx], s.clients[idx+1:]...)

	// Update the map
	delete(s.clientToIdx, id)

	s.clientsLen--
}
