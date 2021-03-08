package browsersync

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// to write a message to the peer
	writeDeadline = 10 * time.Second
	// to read a pong message from the peer
	ponDeadline = 60 * time.Second
	// to send pings to the peer
	pingPeriod = ponDeadline * 9 / 10
	// max outbound messages
	outboundMessages = 256
)

var (
	newline  = []byte{'\n'}
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// Client stands between the socket and the hub and is one directional (write only)
type Client struct {
	hub             *Hub
	conn            *websocket.Conn
	outboundMessage chan []byte
}

// Sets up the socket communication
func communicate(hub *Hub, w http.ResponseWriter, r *http.Request) error {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	client := &Client{hub: hub, conn: conn, outboundMessage: make(chan []byte, outboundMessages)}
	client.hub.register <- client

	defer func() {
		if err := client.writeToSocket(); err != nil {
			// TODO: log
			return
		}
	}()

	return nil
}

// Unregisters from the hub and closes the connection
func (c *Client) close() {
	c.conn.Close()
	close(c.outboundMessage)
}

// Writes message from the hub to the socket
func (c *Client) writeToSocket() error {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case message, ok := <-c.outboundMessage:
			if !ok {
				// The hub closed the channel.
				return c.write(websocket.CloseMessage, []byte{})
			}

			if err := c.writeDeadline(); err != nil {
				return err
			}
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return err
			}
			if _, err := w.Write(message); err != nil {
				return err
			}

			n := len(c.outboundMessage)
			for i := 0; i < n; i++ {
				if _, err := w.Write(newline); err != nil {
					return err
				}
				if _, err := w.Write(<-c.outboundMessage); err != nil {
					return err
				}
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

// Writes with a certain deadline
func (c *Client) writeDeadline() error {
	return c.conn.SetWriteDeadline(time.Now().Add(writeDeadline))
}
