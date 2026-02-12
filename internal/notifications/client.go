package notifications

import (
	"context"

	"github.com/yiannis54/go-socket-server/internal/sockets"
)

// Client is the notification service consumed for sending messages to server.
type Client struct {
	hub *sockets.Hub
}

// NewClient returns a client notification service.
func NewClient(hub *sockets.Hub) *Client {
	return &Client{
		hub: hub,
	}
}

// Broadcast is called from clients for broadcasting a message.
func (c *Client) Broadcast(ctx context.Context, message *sockets.Message) {
	if c == nil || c.hub == nil || message == nil {
		return
	}

	withRoom := &sockets.MessageWithRoom{
		Message:  *message,
		RoomName: nil,
	}
	c.hub.Broadcast <- withRoom
}

// PrivateNotify is called from clients to return a notification back to caller.
func (c *Client) PrivateNotify(ctx context.Context, message *sockets.MessageWithUser) {
	if c == nil || c.hub == nil {
		return
	}

	c.hub.Private <- message
}

// NotifyRoom is called for broadcasting to a specific room.
func (c *Client) NotifyRoom(ctx context.Context, message *sockets.MessageWithRoom) {
	if c == nil || c.hub == nil {
		return
	}

	c.hub.Broadcast <- message
}
