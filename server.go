package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/namsral/flag"
	"github.com/op/go-logging"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 3 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Conn is an middleman between the websocket connection and the hub.
type Conn struct {
	// The websocket connection.
	ws *websocket.Conn

	// Connection identifiers as received from RPC server
	identifiers string

	// Connection subscriptions
	subscriptions map[string]bool

	// Buffered channel of outbound messages.
	send chan []byte
}

var version string

var log = logging.MustGetLogger("main")

var rpchost = flag.String("rpc", "0.0.0.0:50051", "rpc service address")

var redishost = flag.String("redis", "redis://localhost:6379/5", "redis address")

var redischannel = flag.String("redis_channel", "anycable", "redis channel")

var addr = flag.String("addr", "localhost:8080", "http service address")

var wspath = flag.String("wspath", "/cable", "WS endpoint path")

var disconnectRate = flag.Int("disconnect_rate", 100, "the number of Disconnect calls per second")

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	Subprotocols:    []string{"actioncable-v1-json"},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Conn) readPump() {
	defer func() {
		log.Debugf("Disconnect on read error")
		app.Disconnected(c)
		c.ws.Close()
	}()
	for {
		_, message, err := c.ws.ReadMessage()

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Debugf("read error: %v", err)
			}
			break
		}

		msg := &Message{}

		if err := json.Unmarshal(message, &msg); err != nil {
			log.Debugf("Unable to parse message due to invalid JSON. \nMessage:\n  %s\nError:\n  %s", message, err)
		} else {
			log.Debugf("Client message: %s", msg)
			switch msg.Command {
			case "subscribe":
				app.Subscribe(c, msg)
			case "unsubscribe":
				app.Unsubscribe(c, msg)
			case "message":
				app.Perform(c, msg)
			default:
				log.Debugf("Unknown command: %s", msg.Command)
			}
		}
	}
}

// write writes a message with the given message type and payload.
func (c *Conn) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

func (c *Conn) writePump() {
	defer c.ws.Close()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				// The hub closed the channel.
				c.write(websocket.CloseMessage, []byte{})
				return
			}

			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := c.ws.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Critical(err)
		return
	}

	response := rpc.VerifyConnection(r)

	log.Debugf("Auth %s", response)

	if response.Status != 1 {
		log.Warningf("Auth Failed")
		ws.Close()
		return
	}

	conn := &Conn{send: make(chan []byte, 256), ws: ws, identifiers: response.Identifiers, subscriptions: make(map[string]bool)}
	app.Connected(conn, response.Transmissions)
	go conn.writePump()
	conn.readPump()
}

func main() {
	logflag := flag.Bool("log", false, "enable verbose logging")
	showVersion := flag.Bool("version", false, "show version")
	flag.Parse()

	if *showVersion {
		fmt.Println(version)
		return
	}

	backend := logging.AddModuleLevel(logging.NewLogBackend(os.Stderr, "", 0))

	if *logflag {
		backend.SetLevel(logging.DEBUG, "")
	} else {
		backend.SetLevel(logging.INFO, "")
	}

	logging.SetBackend(backend)

	go hub.run()

	app.Pinger = NewPinger(pingPeriod)
	go app.Pinger.run()

	rpc.Init(*rpchost)
	defer rpc.Close()

	app.Subscriber = &Subscriber{host: *redishost, channel: *redischannel}
	go app.Subscriber.run()

	app.Disconnector = &DisconnectNotifier{rate: *disconnectRate, disconnect: make(chan *Conn)}
	go app.Disconnector.run()

	log.Infof("Running AnyCable websocket server v%s on %s at %s", version, *addr, *wspath)
	http.HandleFunc(*wspath, serveWs)
	http.ListenAndServe(*addr, nil)
}
