package node

import (
	"fmt"
	"net/http"

	"github.com/anycable/anycable-go/utils"
	"github.com/apex/log"
	"github.com/gobwas/ws"
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

		rheader := map[string][]string{"X-AnyCable-Version": {utils.Version()}}

		upgrader := ws.HTTPUpgrader{
			Header:   rheader,
			Protocol: func(proto string) bool { return proto == "actioncable-v1-json" },
		}

		conn, _, _, err := upgrader.Upgrade(r, w)
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
			utils.CloseWS(conn, ws.StatusAbnormalClosure, "UID Retrieval Error")
			return
		}

		// Separate goroutine for better GC of caller's data.
		go func() {
			session, err := NewSession(app, conn, url, headers, uid)

			if err != nil {
				ctx.Errorf("Websocket session initialization failed: %v", err)
				return
			}

			session.Log.Debug("websocket session established")

			session.ReadMessages()

			session.Log.Debug("websocket session completed")
		}()
	})
}
