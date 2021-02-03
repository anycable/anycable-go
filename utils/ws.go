package utils

import (
	"net"

	"github.com/gobwas/ws"
)

// CloseWS closes WebSocket connection with the specified close code and reason
func CloseWS(conn net.Conn, code ws.StatusCode, reason string) {
	ws.WriteFrame(conn, ws.NewCloseFrame(ws.NewCloseFrameBody(
		code,
		reason,
	)))
	conn.Write(ws.CompiledClose)
}
