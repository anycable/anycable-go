package node

import (
	"fmt"
	"net/http"

	"github.com/anycable/anycable-go/utils"
	"github.com/apex/log"
	"github.com/gorilla/websocket"
)

// WSConfig contains WebSocket connection configuration.
type WSConfig struct {
	ReadBufferSize    int
	WriteBufferSize   int
	MaxMessageSize    int64
	EnableCompression bool
}

// NewWSConfig build a new WSConfig struct
func NewWSConfig() WSConfig {
	return WSConfig{}
}

// WebsocketHandler generate a new http handler for WebSocket connections
func WebsocketHandler(app *Node, fetchHeaders []string, config *WSConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := log.WithField("context", "ws")

		upgrader := websocket.Upgrader{
			CheckOrigin:       func(r *http.Request) bool { return true },
			Subprotocols:      []string{"actioncable-v1-json"},
			ReadBufferSize:    config.ReadBufferSize,
			WriteBufferSize:   config.WriteBufferSize,
			EnableCompression: config.EnableCompression,
		}

		rheader := map[string][]string{"X-AnyCable-Version": {utils.Version()}}
		ws, err := upgrader.Upgrade(w, r, rheader)
		if err != nil {
			ctx.Debugf("Websocket connection upgrade error: %#v", err.Error())
			return
		}

		url := r.URL.String()

		if !r.URL.IsAbs() {
			// See https://github.com/golang/go/issues/28940#issuecomment-441749380
			scheme := "http://"
			if r.TLS != nil {
				scheme = "https://"
			}
			url = fmt.Sprintf("%s%s%s", scheme, r.Host, url)
		}

		headers := utils.FetchHeaders(r, fetchHeaders)

		uid, err := utils.FetchUID(r)
		if err != nil {
			utils.CloseWS(ws, websocket.CloseAbnormalClosure, "UID Retrieval Error")
			return
		}

		ws.SetReadLimit(config.MaxMessageSize)

		if config.EnableCompression {
			ws.EnableWriteCompression(true)
		}

		// Separate goroutine for better GC of caller's data.
		go func() {
			wrappedConn := WSConnection{conn: ws}
			session := NewSession(app, wrappedConn, url, headers, uid)

			err := app.Authenticate(session)

			if err != nil {
				ctx.Errorf("Websocket session initialization failed: %v", err)
				return
			}

			session.Log.Debug("websocket session established")

			serveConnection(app, ws, session)
		}()
	})
}

func serveConnection(app *Node, ws *websocket.Conn, session *Session) {
	go func() {
		for {
			err := session.ReadMessage()

			if err != nil {
				break
			}
		}
	}()
}
