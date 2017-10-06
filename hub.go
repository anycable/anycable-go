package main

import (
	"encoding/json"
)

type SubscriptionInfo struct {
	conn       *Conn
	stream     string
	identifier string
}

type StreamMessage struct {
	Stream string `json:"stream"`
	Data   string `json:"data"`
}

type Hub struct {
	// Registered connections.
	connections map[*Conn]bool

	// Messages for all connections.
	broadcast chan []byte

	// Messages for specified stream.
	stream_broadcast chan *StreamMessage

	// Register requests from the connections.
	register chan *Conn

	// Unregister requests from connections.
	unregister chan *Conn

	// Subscribe requests to strreams.
	subscribe chan *SubscriptionInfo

	// Unsubscribe requests from streams.
	unsubscribe chan *SubscriptionInfo

	// Maps streams to connections
	streams map[string]map[*Conn]string

	// Maps connections to identifiers to streams
	connection_streams map[*Conn]map[string][]string

	// Control channel to shutdown hub
	shutdown chan bool
}

var hub = Hub{
	broadcast:          make(chan []byte),
	stream_broadcast:   make(chan *StreamMessage),
	register:           make(chan *Conn),
	unregister:         make(chan *Conn),
	subscribe:          make(chan *SubscriptionInfo),
	unsubscribe:        make(chan *SubscriptionInfo),
	connections:        make(map[*Conn]bool),
	streams:            make(map[string]map[*Conn]string),
	connection_streams: make(map[*Conn]map[string][]string),
	shutdown:           make(chan bool),
}

func (h *Hub) run() {
	for {
		select {
		case conn := <-h.register:
			app.Logger.Debugf("Register connection %v", conn)
			h.connections[conn] = true

		case conn := <-h.unregister:
			app.Logger.Debugf("Unregister connection %v", conn)

			h.UnsubscribeConnection(conn)

			if _, ok := h.connections[conn]; ok {
				h.CloseConnection(conn)
			}

		case message := <-h.broadcast:
			app.Logger.Debugf("Broadcast message %s", message)
			for conn := range h.connections {
				select {
				case conn.send <- message:
				default:
					hub.CloseConnection(conn)
				}
			}

		case stream_message := <-h.stream_broadcast:
			app.Logger.Debugf("Broadcast to stream %s: %s", stream_message.Stream, stream_message.Data)

			if _, ok := h.streams[stream_message.Stream]; !ok {
				app.Logger.Debugf("No connections for stream %s", stream_message.Stream)
				break
			}

			buf := make(map[string][]byte)

			for conn, id := range h.streams[stream_message.Stream] {
				var bdata []byte

				if msg, ok := buf[id]; ok {
					bdata = msg
				} else {
					bdata = BuildMessage(stream_message.Data, id)
					buf[id] = bdata
				}
				select {
				case conn.send <- bdata:
				default:
					h.CloseConnection(conn)
				}
			}

		case subinfo := <-h.subscribe:
			app.Logger.Debugf("Subscribe to stream %s for %s", subinfo.stream, subinfo.conn.identifiers)

			if _, ok := h.streams[subinfo.stream]; !ok {
				h.streams[subinfo.stream] = make(map[*Conn]string)
			}

			h.streams[subinfo.stream][subinfo.conn] = subinfo.identifier

			if _, ok := h.connection_streams[subinfo.conn]; !ok {
				h.connection_streams[subinfo.conn] = make(map[string][]string)
			}

			h.connection_streams[subinfo.conn][subinfo.identifier] = append(
				h.connection_streams[subinfo.conn][subinfo.identifier],
				subinfo.stream)

		case subinfo := <-h.unsubscribe:
			h.UnsubscribeConnectionFromChannel(subinfo.conn, subinfo.identifier)

		case <-h.shutdown:
			// TODO: notify about disconnection
			return
		}
	}
}

func (h *Hub) Shutdown() {
	h.shutdown <- true
}

func (h *Hub) Size() int {
	return len(h.connections)
}

func (h *Hub) UnsubscribeConnection(conn *Conn) {
	app.Logger.Debugf("Unsubscribe from all streams: %s", conn.identifiers)

	for channel, _ := range h.connection_streams[conn] {
		h.UnsubscribeConnectionFromChannel(conn, channel)
	}

	delete(h.connection_streams, conn)
}

func (h *Hub) UnsubscribeConnectionFromChannel(conn *Conn, channel string) {
	app.Logger.Debugf("Unsubscribe from channel %s: %s", channel, conn.identifiers)

	if _, ok := h.connection_streams[conn]; !ok {
		return
	}

	for _, stream := range h.connection_streams[conn][channel] {
		delete(h.streams[stream], conn)

		if len(h.streams[stream]) == 0 {
			delete(h.streams, stream)
		}
	}
}

func (h *Hub) CloseConnection(conn *Conn) {
	if conn.send != nil {
		close(conn.send)
	}

	// Make it nil to avoid panic on writing into it
	conn.send = nil
	delete(h.connections, conn)
}

func BuildMessage(data string, identifier string) []byte {
	var msg map[string]interface{}

	json.Unmarshal([]byte(data), &msg)

	return (&Reply{Identifier: identifier, Message: msg}).toJSON()
}
