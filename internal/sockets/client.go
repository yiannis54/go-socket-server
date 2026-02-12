package sockets

import (
	"encoding/json"
	"errors"
	"log"
	"slices"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 45 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// Client is the websocket client.
type Client struct {
	hub *Hub
	ID  string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { return c.conn.SetReadDeadline(time.Now().Add(pongWait)) })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("read message error: %v\n", err)
			}
			break
		}

		// if message is FE ping to keep connection alive, discard it.
		if string(message) == "0" {
			continue
		}

		incomingMsg := IncomingSubscription{}
		err = json.Unmarshal(message, &incomingMsg)
		if err != nil {
			log.Printf("unmarshal read message error: %v\n", err)
			continue
		}

		if err := validateIncomingMessage(incomingMsg); err != nil {
			log.Printf("validate incoming message error: %v\n", err)
			continue
		}

		if incomingMsg.Action == unsubscribeAction {
			c.hub.unregisterRoom <- newSubscription(incomingMsg.Room, c)
			continue
		}

		c.hub.registerRoom <- newSubscription(incomingMsg.Room, c)
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writr, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = writr.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				_, _ = writr.Write(<-c.send)
			}

			if err := writr.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.conn.WriteControl(websocket.PingMessage, []byte("ping"), time.Now().Add(writeWait)); err != nil {
				c.conn.Close()
				return
			}
		}
	}
}

func validateIncomingMessage(msg IncomingSubscription) error {
	if msg.Action == "" || msg.Room == "" {
		return errors.New("invalid subscription message body")
	}

	validActions := []string{subscribeAction, unsubscribeAction}
	if slices.Contains(validActions, msg.Action) {
		return nil
	}

	return errors.New("invalid subscription action")
}
