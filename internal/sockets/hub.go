package sockets

import (
	"context"
	"encoding/json"
	"log"
)

// Hub is a struct that holds all the clients and the messages that are sent to them.
type Hub struct {
	// Registered clients.
	clients map[*Client]struct{}

	// map of ws client per user id.
	users map[string]*Client

	// map of rooms for events notifications.
	rooms map[string]map[*Client]struct{}

	// Inbound messages from broadcasting to clients.
	Broadcast chan *MessageWithRoom

	// Inbound messages from private messaging to clients.
	Private chan *MessageWithUser

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// channels for incoming register/unregister room subscriptions.
	registerRoom   chan *Subscription
	unregisterRoom chan *Subscription
}

// NewHub returns a new socket Hub.
func NewHub() *Hub {
	return &Hub{
		Broadcast:      make(chan *MessageWithRoom),
		Private:        make(chan *MessageWithUser),
		register:       make(chan *Client),
		unregister:     make(chan *Client),
		registerRoom:   make(chan *Subscription),
		unregisterRoom: make(chan *Subscription),
		clients:        make(map[*Client]struct{}),
		users:          make(map[string]*Client),
		rooms:          make(map[string]map[*Client]struct{}),
	}
}

// Run starts a hub that listens for incoming messages.
// It blocks until the given context is cancelled, then closes the hub and returns.
//
//nolint:cyclop // TODO: reduce cyclomatic complexity.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
			h.users[client.ID] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				h.unRegisterClient(client)
			}
		case subscription := <-h.registerRoom:
			if h.rooms[subscription.Room] == nil {
				h.rooms[subscription.Room] = make(map[*Client]struct{})
			}
			h.rooms[subscription.Room][subscription.client] = struct{}{}
		case subscription := <-h.unregisterRoom:
			if h.rooms[subscription.Room] != nil {
				delete(h.rooms[subscription.Room], subscription.client)
				if len(h.rooms[subscription.Room]) == 0 {
					delete(h.rooms, subscription.Room)
				}
			}
		case messageWithRoom := <-h.Broadcast:
			h.handleBroadcastMessage(messageWithRoom)
		case messageWithUser := <-h.Private:
			h.handlePrivateMessage(messageWithUser)
		case <-ctx.Done():
			h.Close()
			return
		}
	}
}

// GetUser returns the hub client data from user id.
func (h *Hub) GetUser(userID string) *Client {
	return h.users[userID]
}

func (h *Hub) unRegisterClient(client *Client) {
	for roomName := range h.rooms {
		delete(h.rooms[roomName], client)
		if len(h.rooms[roomName]) == 0 {
			delete(h.rooms, roomName)
		}
	}
	delete(h.users, client.ID)
	delete(h.clients, client)
	close(client.send)
}

func (h *Hub) handlePrivateMessage(messageWithUser *MessageWithUser) {
	client, ok := h.users[messageWithUser.UserID]
	if !ok {
		log.Println("sockets: no client to send private message")
		return
	}
	message, err := json.Marshal(messageWithUser.Message)
	if err != nil {
		log.Printf("sockets: could not marshal private message: %v", err)
		return
	}
	select {
	case client.send <- message:
	default:
		h.unRegisterClient(client)
	}
}

//nolint:cyclop // TODO: reduce cyclomatic complexity.
func (h *Hub) handleBroadcastMessage(messageWithRoom *MessageWithRoom) {
	message, err := json.Marshal(messageWithRoom.Message)
	if err != nil {
		log.Printf("sockets: could not marshal broadcast message: %v", err)
		return
	}

	// if room not passed, send to all subscribers.
	if messageWithRoom.RoomName == nil {
		for client := range h.clients {
			select {
			case client.send <- message:
			default:
				h.unRegisterClient(client)
			}
		}
		return
	}

	room, ok := h.rooms[*messageWithRoom.RoomName]
	if !ok {
		log.Printf("sockets: room not found or noone in room: %v", messageWithRoom.RoomName)
		return
	}

	if len(room) > 0 {
		for client := range room {
			select {
			case client.send <- message:
			default:
				h.unRegisterClient(client)
			}
		}
	}
}

// Close removes all map elements and closes hub channels.
func (h *Hub) Close() {
	for room, clients := range h.rooms {
		for client := range clients {
			delete(clients, client)
		}

		delete(h.rooms, room)
	}

	for user := range h.users {
		delete(h.users, user)
	}

	for client := range h.clients {
		delete(h.clients, client)
	}

	close(h.unregisterRoom)
	close(h.registerRoom)
	close(h.unregister)
	close(h.register)
	close(h.Private)
	close(h.Broadcast)
}
