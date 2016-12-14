package ws

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client handle websockets clients
type Client struct {
	conn *websocket.Conn

	// Buffered channel of inbound messages.
	receive chan string
	// Buffered channel of outbound messages.
	send chan string
}

// NewClient return a new client
func NewClient(conn *websocket.Conn) *Client {
	c := &Client{
		conn:    conn,
		receive: make(chan string, 64),
		send:    make(chan string, 256),
	}

	go c.readPump()
	go c.writePump()

	return c
}

// Close close the client connection
func (c *Client) Close() {
	close(c.send)
}

// Messages return the inbound messages channel
func (c *Client) Messages() <-chan string {
	return c.receive
}

// Send send a message to the outbound channel
func (c *Client) Send(msg string) {
	c.send <- msg
}

// readPump pumps messages from the websocket connection.
func (c *Client) readPump() {
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		mt, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Warnf("WS error: %v", err)
			}
			close(c.receive)
			break
		}

		if mt == websocket.TextMessage {
			c.receive <- string(message)
		}
	}
}

// writePump pumps messages to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.conn.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}
