package wsrpc

import (
	"log/slog"
	"net/http"

	"github.com/anycable/anycable-go/server"
	"github.com/anycable/anycable-go/version"
	"github.com/anycable/anycable-go/ws"
	"github.com/gorilla/websocket"
)

func ClientHandler(s *Server, conf *ws.Config, l *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := l.With("context", "wsrpc")

		upgrader := websocket.Upgrader{
			ReadBufferSize:    conf.ReadBufferSize,
			WriteBufferSize:   conf.WriteBufferSize,
			EnableCompression: conf.EnableCompression,
		}

		rheader := map[string][]string{"X-AnyCable-Version": {version.Version()}}

		info, err := server.NewRequestInfo(r, nil)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if !s.Authenticate(info) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		wsc, err := upgrader.Upgrade(w, r, rheader)
		if err != nil {
			ctx.Debug("WebSocket connection upgrade failed", "error", err)
			return
		}

		wsc.SetReadLimit(conf.MaxMessageSize)

		if conf.EnableCompression {
			wsc.EnableWriteCompression(true)
		}

		go func() {
			conn := ws.NewConnection(wsc)
			s.ServeClient(conn, info)
		}()
	})
}
