package wsrpc

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/anycable/anycable-go/ws"
)

type Client struct {
	id   string
	conn *ws.Connection
	log  *slog.Logger

	sendCh    chan []byte
	writeWait time.Duration

	closeMu sync.Mutex
	closed  bool
	closeFn context.CancelFunc
}

func NewClient(id string, conn *ws.Connection, l *slog.Logger) *Client {
	return &Client{
		id:        id,
		conn:      conn,
		log:       l,
		sendCh:    make(chan []byte, 128),
		writeWait: 5 * time.Second,
	}
}

func (c *Client) ID() string {
	return c.id
}

func (c *Client) Serve(ctx context.Context, handler func(msg []byte)) error {
	newCtx, cancel := context.WithCancel(ctx)
	c.closeMu.Lock()
	c.closeFn = cancel
	c.closeMu.Unlock()

	go c.sendMessages(newCtx)

	for {
		if c.isClosed() {
			return nil
		}

		message, err := c.conn.Read()

		if err != nil {
			if ws.IsCloseError(err) {
				c.log.Debug("WebSocket closed", "error", err)
			} else {
				c.log.Debug("WebSocket close error", "error", err)
			}

			if c.isClosed() {
				return nil
			}

			return err
		}

		handler(message)
	}
}

func (c *Client) Send(msg []byte) bool {
	if c.isClosed() {
		return false
	}

	select {
	case c.sendCh <- msg:
		return true
	default:
		// Buffer is full, let the server to fallback to another client
		return false
	}
}

func (c *Client) Close() {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	if c.closed {
		return
	}

	c.closed = true
	if c.closeFn != nil {
		c.closeFn()
	}
	c.conn.Close(ws.CloseNormalClosure, "Closed by server")
}

func (c *Client) Flush() []byte {
	buf := make([]byte, 0, len(c.sendCh))

	for {
		select {
		case msg := <-c.sendCh:
			buf = append(buf, msg...)
		default:
			close(c.sendCh)
			return buf
		}
	}
}

func (c *Client) sendMessages(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.Close()
			return
		case msg := <-c.sendCh:
			if err := c.conn.Write(msg, time.Now().Add(c.writeWait)); err != nil {
				c.log.Debug("error sending message", "err", err)
				// re-enqueue message
				if ok := c.Send(msg); !ok {
					c.log.Error("failed to re-enqueue failed message due to full buffer")
				}
				c.Close()
				return
			}
		}
	}
}

func (c *Client) isClosed() bool {
	c.closeMu.Lock()
	defer c.closeMu.Unlock()

	return c.closed
}
