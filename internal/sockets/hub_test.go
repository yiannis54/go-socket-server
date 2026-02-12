package sockets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	assert.NotNil(t, hub.Broadcast)
	assert.NotNil(t, hub.Private)
	assert.NotNil(t, hub.register)
	assert.NotNil(t, hub.unregister)
	assert.NotNil(t, hub.registerRoom)
	assert.NotNil(t, hub.unregisterRoom)
	assert.NotNil(t, hub.clients)
	assert.NotNil(t, hub.users)
	assert.NotNil(t, hub.rooms)

	client := &Client{
		hub: hub,
		ID:  "abc-xyz",
	}
	hub.clients[client] = struct{}{}
	hub.users["abc-xyz"] = client
	hub.rooms["public"] = hub.clients
	hub.Close()
}
