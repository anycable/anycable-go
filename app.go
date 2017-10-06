package main

import (
	"encoding/json"

	pb "github.com/anycable/anycable-go/protos"
	"github.com/op/go-logging"
)

type App struct {
	Pinger       *Pinger
	Subscriber   *Subscriber
	Disconnector *DisconnectNotifier
	Logger       *logging.Logger
}

const (
	PING = "ping"
)

type Message struct {
	Command    string `json:"command"`
	Identifier string `json:"identifier"`
	Data       string `json:"data"`
}

type Reply struct {
	Type       string      `json:"type,omitempty"`
	Identifier string      `json:"identifier"`
	Message    interface{} `json:"message"`
}

func (r *Reply) toJSON() []byte {
	jsonStr, err := json.Marshal(&r)
	if err != nil {
		panic("Failed to build JSON")
	}
	return jsonStr
}

var app = &App{
	Logger: logging.MustGetLogger("main"),
}

func (app *App) Connected(conn *Conn, transmissions []string) {
	app.Pinger.Increment()

	hub.register <- conn

	Transmit(conn, transmissions)
}

func (app *App) Subscribe(conn *Conn, msg *Message) {
	if _, ok := conn.subscriptions[msg.Identifier]; ok {
		app.Logger.Warningf("Already Subscribed to %s", msg.Identifier)
		return
	}

	res := rpc.Subscribe(conn.identifiers, msg.Identifier)

	if res.Status == 1 {
		conn.subscriptions[msg.Identifier] = true
	}

	app.Logger.Debugf("Subscribe %s", res)

	HandleReply(conn, msg, res)
}

func (app *App) Unsubscribe(conn *Conn, msg *Message) {
	if _, ok := conn.subscriptions[msg.Identifier]; !ok {
		app.Logger.Warningf("Unknown subscription %s", msg.Identifier)
		return
	}

	res := rpc.Unsubscribe(conn.identifiers, msg.Identifier)

	if res.Status == 1 {
		delete(conn.subscriptions, msg.Identifier)
	}

	HandleReply(conn, msg, res)
}

func (app *App) Perform(conn *Conn, msg *Message) {
	if _, ok := conn.subscriptions[msg.Identifier]; !ok {
		app.Logger.Warningf("Unknown subscription %s", msg.Identifier)
		return
	}

	res := rpc.Perform(conn.identifiers, msg.Identifier, msg.Data)

	app.Logger.Debugf("Perform %s", res)

	HandleReply(conn, msg, res)
}

func (app *App) Disconnected(conn *Conn) {
	app.Pinger.Decrement()

	hub.unregister <- conn

	app.Disconnector.Notify(conn)
}

func (app *App) BroadcastAll(message []byte) {
	hub.broadcast <- message
}

func Transmit(conn *Conn, transmissions []string) {
	for _, msg := range transmissions {
		conn.send <- []byte(msg)
	}
}

func HandleReply(conn *Conn, msg *Message, reply *pb.CommandResponse) {
	if reply.Disconnect {
		defer conn.ws.Close()
	}

	if reply.StopStreams {
		hub.unsubscribe <- &SubscriptionInfo{conn: conn, identifier: msg.Identifier}
	}

	for _, s := range reply.Streams {
		hub.subscribe <- &SubscriptionInfo{conn: conn, stream: s, identifier: msg.Identifier}
	}

	Transmit(conn, reply.Transmissions)
}
