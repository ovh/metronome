package ws

import (
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
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

	go func() {
		if err := c.readPump(); err != nil {
			log.WithError(err).Error("Could not read from the pump")
		}
	}()

	go func() {
		if err := c.writePump(); err != nil {
			log.WithError(err).Error("Could not write in the pump")
		}
	}()

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
func (c *Client) readPump() error {
	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		return err
	}

	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		mt, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Warnf("WS error: %v", err)
			}
			close(c.receive)
			return err
		}

		if mt == websocket.TextMessage {
			c.receive <- string(message)
		}
	}
}

// writePump pumps messages to the websocket connection.
func (c *Client) writePump() error {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			log.WithError(err).Error("Could not gracefully close the websocket")
		}
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return err
			}

			if !ok {
				return c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
				return err
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return err
			}

			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return err
			}
		}
	}
}
