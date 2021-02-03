package node

import (
	"net"
	"sync"
	"time"

	"github.com/anycable/anycable-go/common"
	"github.com/anycable/anycable-go/utils"
	"github.com/apex/log"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

const (
	// CloseNormalClosure indicates normal closure
	CloseNormalClosure = ws.StatusNormalClosure

	// CloseInternalServerErr indicates closure because of internal error
	CloseInternalServerErr = ws.StatusInternalServerError

	// CloseAbnormalClosure indicates abnormal close
	CloseAbnormalClosure = ws.StatusAbnormalClosure

	// CloseGoingAway indicates closing because of server shuts down or client disconnects
	CloseGoingAway = ws.StatusGoingAway

	// CloseNoStatusReceived indicates no status close
	CloseNoStatusReceived = ws.StatusNoStatusRcvd

	writeWait    = 10 * time.Second
	pingInterval = 3 * time.Second
)

const (
	textFrame  ws.OpCode = ws.OpText
	closeFrame ws.OpCode = ws.OpClose
)

type sentFrame struct {
	frameType   ws.OpCode
	payload     []byte
	closeCode   ws.StatusCode
	closeReason string
}

// Session represents active client
type Session struct {
	node          *Node
	conn          net.Conn
	env           *common.SessionEnv
	subscriptions map[string]bool
	send          chan sentFrame
	closed        bool
	connected     bool
	mu            sync.Mutex
	pingTimer     *time.Timer

	UID         string
	Identifiers string
	Log         *log.Entry
}

// NewSession build a new Session struct from ws connetion and http request
func NewSession(node *Node, conn net.Conn, url string, headers map[string]string, uid string) (*Session, error) {
	session := &Session{
		node:          node,
		conn:          conn,
		env:           common.NewSessionEnv(url, &headers),
		subscriptions: make(map[string]bool),
		send:          make(chan sentFrame, 256),
		closed:        false,
		connected:     false,
	}

	session.UID = uid

	ctx := node.log.WithFields(log.Fields{
		"sid": session.UID,
	})

	session.Log = ctx

	err := node.Authenticate(session)

	if err != nil {
		defer session.Close("Auth Error", CloseInternalServerErr)
	}

	go session.SendMessages()

	session.addPing()

	return session, err
}

// SendMessages waits for incoming messages and send them to the client connection
func (s *Session) SendMessages() {
	defer s.Disconnect("Write Failed", CloseAbnormalClosure)
	for message := range s.send {
		switch message.frameType {
		case textFrame:
			err := s.write(message.payload, time.Now().Add(writeWait))

			if err != nil {
				return
			}
		case closeFrame:
			utils.CloseWS(s.conn, message.closeCode, message.closeReason)
			return
		default:
			s.Log.Errorf("Unknown frame type: %v", message)
			return
		}
	}
}

// Send data to client connection
func (s *Session) Send(msg []byte) {
	s.sendFrame(&sentFrame{frameType: textFrame, payload: msg})
}

func (s *Session) sendClose(reason string, code ws.StatusCode) {
	s.sendFrame(&sentFrame{
		frameType:   closeFrame,
		closeReason: reason,
		closeCode:   code,
	})
}

func (s *Session) sendFrame(frame *sentFrame) {
	s.mu.Lock()

	if s.send == nil {
		s.mu.Unlock()
		return
	}

	select {
	case s.send <- *frame:
	default:
		if s.send != nil {
			close(s.send)
			defer s.Disconnect("Write failed", CloseAbnormalClosure)
		}

		s.send = nil
	}

	s.mu.Unlock()
}

func (s *Session) write(message []byte, deadline time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := wsutil.WriteServerMessage(s.conn, textFrame, message)

	return err
}

// ReadMessages reads messages from ws connection and send them to node
func (s *Session) ReadMessages() {
	for {
		message, op, err := wsutil.ReadClientData(s.conn)

		if err != nil {
			s.Log.Debugf("Websocket close error: %v", err)
			s.Disconnect("Read failed", CloseAbnormalClosure)
			break
		}

		if op == closeFrame {
			s.Log.Debugf("Websocket closed: %v", err)
			s.Disconnect("Read closed", CloseNormalClosure)
		}

		if err := s.node.HandleCommand(s, message); err != nil {
			s.Log.Warnf("Failed to handle incoming message '%s' with error: %v", message, err)
		}
	}
}

// Disconnect enqueues RPC disconnect request and closes the connection
func (s *Session) Disconnect(reason string, code ws.StatusCode) {
	s.mu.Lock()
	if s.connected {
		defer s.node.Disconnect(s) // nolint:errcheck
	}
	s.connected = false
	s.mu.Unlock()

	s.Close(reason, code)
}

// Close websocket connection with the specified reason
func (s *Session) Close(reason string, code ws.StatusCode) {
	s.mu.Lock()

	if s.closed {
		s.mu.Unlock()
		return
	}

	s.closed = true
	s.mu.Unlock()

	s.sendClose(reason, code)

	if s.pingTimer != nil {
		s.pingTimer.Stop()
	}
}

func (s *Session) sendPing() {
	deadline := time.Now().Add(pingInterval / 2)
	err := s.write(newPingMessage(), deadline)

	if err == nil {
		s.addPing()
	} else {
		s.Disconnect("Ping failed", CloseAbnormalClosure)
	}
}

func (s *Session) addPing() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	s.pingTimer = time.AfterFunc(pingInterval, s.sendPing)
}

func newPingMessage() []byte {
	return (&common.PingMessage{Type: "ping", Message: time.Now().Unix()}).ToJSON()
}
