package browsersync

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// to write a message to the peer
	WRITE_DEADLINE = 10 * time.Second
	// to read a pong message from the peer
	PONG_DEADLINE = 60 * time.Second
	// to send pings to the peer
	PING_PERIOD = PONG_DEADLINE * 9 / 10
	// max outbound messages
	OUTBOUND_MESSAGES = 256
)

var (
	newline = []byte{'\n'}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client stands between the socket and the hub
type Client struct {
	hub            *Hub
	conn           *websocket.Conn
	inboundMessage chan []byte
}

func communicate(hub *Hub, w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	client := &Client{hub: hub, conn: conn, inboundMessage: make(chan []byte, OUTBOUND_MESSAGES)}
	client.hub.register <- client
	go client.writeToSocket()
	//client.readToHub()

	return nil
}

// Reads messages from the socket to the hub
func (c *Client) readToHub() {
	defer c.close()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("An error happened when reading from the Websocket client: %v", err)
			}
			break
		}
	}
}

// Unregisters from the hub and closes the connection
func (c *Client) close() {
	c.hub.unregister <- c
	c.conn.Close()
}

// Writes message from the hub to the socket
func (c *Client) writeToSocket() error {
	ticker := time.NewTicker(PING_PERIOD)
	defer func() {
		ticker.Stop()
		c.close()
	}()
	for {
		select {
		case message, ok := <-c.inboundMessage:
			if !ok {
				// The hub closed the channel.
				c.write(websocket.CloseMessage, []byte{})
				return nil
			}

			if err := c.writeDeadline(); err != nil {
				return err
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return err
			}
			w.Write(message)

			n := len(c.inboundMessage)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.inboundMessage)
			}

			if err := w.Close(); err != nil {
				return err
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return err
			}
		}
	}
}

// Writes to the socket
func (c *Client) write(messageType int, data []byte) error {
	if err := c.writeDeadline(); err != nil {
		return err
	}
	return c.conn.WriteMessage(messageType, data)
}

func (c *Client) writeDeadline() error {
	return c.conn.SetWriteDeadline(time.Now().Add(WRITE_DEADLINE))
}
