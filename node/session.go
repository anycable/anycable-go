package node

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/anycable/anycable-go/utils"
	"github.com/apex/log"
	"github.com/gorilla/websocket"
	nanoid "github.com/matoous/go-nanoid"
)

const (
	// DefaultCloseStatus is what it states)
	DefaultCloseStatus = 3000

	writeWait      = 10 * time.Second
	maxMessageSize = 512
	pingInterval   = 3 * time.Second
)

// Session represents active client
type Session struct {
	node          *Node
	ws            *websocket.Conn
	path          string
	headers       map[string]string
	subscriptions map[string]bool
	send          chan []byte
	closed        bool
	connected     bool
	mu            sync.Mutex
	pingTimer     *time.Timer

	UID         string
	Identifiers string
	Log         *log.Entry
}

type pingMessage struct {
	Type    string      `json:"type"`
	Message interface{} `json:"message"`
}

func (p *pingMessage) toJSON() []byte {
	jsonStr, err := json.Marshal(&p)
	if err != nil {
		panic("Failed to build ping JSON 😲")
	}
	return jsonStr
}

// NewSession build a new Session struct from ws connetion and http request
func NewSession(node *Node, ws *websocket.Conn, request *http.Request) (*Session, error) {
	path := request.URL.String()
	headers := utils.FetchHeaders(request, node.Config.Headers)

	session := &Session{
		node:          node,
		ws:            ws,
		path:          path,
		headers:       headers,
		subscriptions: make(map[string]bool),
		send:          make(chan []byte, 256),
		closed:        false,
		connected:     false,
	}

	uid, err := nanoid.Nanoid()

	if err != nil {
		defer session.Close("Nanoid Error")
		return nil, err
	}

	session.UID = uid

	ctx := node.log.WithFields(log.Fields{
		"sid": session.UID,
	})

	session.Log = ctx

	err = node.Authenticate(session, path, &headers)

	if err != nil {
		defer session.Close("Auth Error")
		return nil, err
	}

	go session.SendMessages()

	session.addPing()

	return session, nil
}

// SendMessages waits for incoming messages and send them to the client connection
func (s *Session) SendMessages() {
	defer s.Disconnect("Write Failed")
	for {
		select {
		case message, ok := <-s.send:
			if !ok {
				return
			}

			err := s.write(message, time.Now().Add(writeWait))

			if err != nil {
				return
			}
		}
	}
}

func (s *Session) write(message []byte, deadline time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ws.SetWriteDeadline(deadline)

	w, err := s.ws.NextWriter(websocket.TextMessage)

	if err != nil {
		return err
	}

	w.Write(message)

	return w.Close()
}

// Send data to client connection
func (s *Session) Send(msg []byte) {
	select {
	case s.send <- msg:
	default:
		s.mu.Lock()

		if s.send != nil {
			close(s.send)
			defer s.Disconnect("Write failed")
		}

		defer s.mu.Unlock()
		s.send = nil
	}
}

// ReadMessages reads messages from ws connection and send them to node
func (s *Session) ReadMessages() {
	// s.ws.SetReadLimit(MaxMessageSize)

	defer s.Disconnect("")

	for {
		_, message, err := s.ws.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				s.Log.Debugf("Websocket read error: %v", err)
			}
			break
		}

		s.node.HandleCommand(s, message)
	}
}

// Disconnect enqueues RPC disconnect request and closes the connection
func (s *Session) Disconnect(reason string) {
	s.mu.Lock()
	if !s.connected {
		s.node.Disconnect(s)
	}
	s.connected = false
	s.mu.Unlock()

	s.Close(reason)
}

// Close websocket connection with the specified reason
func (s *Session) Close(reason string) {
	s.mu.Lock()
	if s.closed {
		return
	}
	s.closed = true
	s.mu.Unlock()

	if s.pingTimer != nil {
		s.pingTimer.Stop()
	}

	// TODO: make deadline and status code configurable
	deadline := time.Now().Add(time.Second)
	msg := websocket.FormatCloseMessage(DefaultCloseStatus, reason)
	s.ws.WriteControl(websocket.CloseMessage, msg, deadline)
	s.ws.Close()
}

func (s *Session) sendPing() {
	deadline := time.Now().Add(pingInterval / 2)
	err := s.write(newPingMessage(), deadline)

	if err == nil {
		s.addPing()
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
	return (&pingMessage{Type: "ping", Message: time.Now().Unix()}).toJSON()
}
