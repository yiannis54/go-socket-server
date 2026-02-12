package sockets

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/yiannis54/go-socket-server/internal/middleware"
)

const (
	channelBytes      = 256
	defaultBufferSize = 1024
)

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  defaultBufferSize,
		WriteBufferSize: defaultBufferSize,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	// should not defer here conn.Close(), moved to goroutines
	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, channelBytes),
	}

	client.hub.register <- client
	if id, ok := middleware.UserIDFromRequest(r.Context()); ok {
		client.ID = id
	}

	// Allow collection of memory referenced by the caller by doing all work in new goroutines.
	go client.writePump()
	go client.readPump()
}
