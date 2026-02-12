package sockets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"

	"github.com/yiannis54/go-socket-server/internal/middleware"
)

type User struct {
}

type SocketsTestSuite struct {
	suite.Suite
	hub       *Hub
	userID    string
	server    *httptest.Server
	ws        *websocket.Conn
	cancelHub context.CancelFunc
}

func TestSocketsTestSuite(t *testing.T) {
	suite.Run(t, new(SocketsTestSuite))
}

func (suite *SocketsTestSuite) SetupSuite() {
	socketHub := NewHub()
	suite.hub = socketHub
	suite.userID = uuid.NewString()

	ctx, cancel := context.WithCancel(context.Background())
	suite.cancelHub = cancel

	go suite.hub.Run(ctx)
}

func (suite *SocketsTestSuite) TearDownSuite() {
	suite.cancelHub()
}

func (suite *SocketsTestSuite) SetupTest() {
	s := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(context.WithValue(r.Context(), middleware.UserIDContextKey, suite.userID))
			ServeWs(suite.hub, w, r)
		}))
	suite.server = s

	wsURL := "ws" + strings.TrimPrefix(s.URL, "http")
	// Connect to the server
	ws, res, err := websocket.DefaultDialer.Dial(wsURL, nil)
	suite.Require().NoError(err)
	time.Sleep(100 * time.Millisecond)
	suite.ws = ws
	time.Sleep(100 * time.Millisecond)

	res.Body.Close()
}

func (suite *SocketsTestSuite) TearDownTest() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.ws != nil {
		suite.ws.Close()
	}
}

func (suite *SocketsTestSuite) TestHub_RegisterUnregisterRoom() {
	suite.Assert().Len(suite.hub.clients, 1)
	suite.Assert().Len(suite.hub.rooms, 0)

	suite.Run("should fail to enter room with bad json", func() {
		suite.Require().NoError(suite.ws.WriteMessage(websocket.TextMessage, []byte(`{"action": "`)))
		time.Sleep(100 * time.Millisecond)
		suite.Assert().Len(suite.hub.rooms, 0)
	})

	suite.Run("should fail to enter room with no incoming room info", func() {
		suite.Require().NoError(suite.ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"action": %q}`, subscribeAction))))
		time.Sleep(100 * time.Millisecond)
		suite.Assert().Len(suite.hub.rooms, 0)
	})

	suite.Run("should fail to enter room with invalid action", func() {
		suite.Require().NoError(suite.ws.WriteMessage(websocket.TextMessage, []byte(`{"action": "fly", "room": "2023-06-08"}`)))
		time.Sleep(100 * time.Millisecond)
		suite.Assert().Len(suite.hub.rooms, 0)
	})

	suite.Run("should enter and leave room correctly", func() {
		suite.Require().NoError(suite.ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"action": %q, "room": "2023-06-08"}`, subscribeAction))))
		time.Sleep(100 * time.Millisecond)
		suite.Assert().Len(suite.hub.rooms, 1)

		suite.Require().NoError(suite.ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"action": %q, "room": "2023-06-08"}`, unsubscribeAction))))
		time.Sleep(100 * time.Millisecond)
		suite.Assert().Len(suite.hub.rooms, 0)
	})
}

func (suite *SocketsTestSuite) TestHub_unRegisterClient() {
	time.Sleep(200 * time.Millisecond)
	suite.Require().NotNil(suite.hub.GetUser(suite.userID))
	suite.hub.unRegisterClient(suite.hub.users[suite.userID])
	time.Sleep(200 * time.Millisecond)
	suite.Assert().Nil(suite.hub.GetUser(suite.userID))
}

func (suite *SocketsTestSuite) TestHub_handlePrivateMessage() {
	msg := &MessageWithUser{
		UserID: suite.userID,
		Message: Message{
			Type:        "info",
			MessageBody: "Hello, user!",
		},
	}

	suite.hub.handlePrivateMessage(msg)

	// Expect the server to echo the message back.
	suite.Require().NoError(suite.ws.SetReadDeadline(time.Now().Add(time.Second * 5)))
	mt, incoming, err := suite.ws.ReadMessage()
	suite.Require().NoError(err)
	suite.Assert().Equal(websocket.TextMessage, mt)

	incomingMsg := Message{}
	suite.Require().NoError(json.Unmarshal(incoming, &incomingMsg))
	suite.Assert().Equal(msg.Message, incomingMsg)
}

func (suite *SocketsTestSuite) TestHub_handleBroadcastMessage() {
	msg := &MessageWithRoom{
		RoomName: nil, // Sending to all clients
		Message: Message{
			Type:        "info",
			MessageBody: "Hello, all!",
		},
	}

	suite.hub.handleBroadcastMessage(msg)

	suite.Require().NoError(suite.ws.SetReadDeadline(time.Now().Add(time.Second * 2)))
	mt, incoming, err := suite.ws.ReadMessage()
	suite.Require().NoError(err)
	suite.Assert().Equal(websocket.TextMessage, mt)

	incomingMsg := Message{}
	suite.Require().NoError(json.Unmarshal(incoming, &incomingMsg))
	suite.Assert().Equal(msg.Message, incomingMsg)
}

func (suite *SocketsTestSuite) TestHub_handleBroadcastMessage_Room() {
	roomName := "2023-01-01"
	msg := &MessageWithRoom{
		RoomName: &roomName,
		Message: Message{
			Type:        "info",
			MessageBody: "Hello, to the room!",
		},
	}

	suite.Run("should receive message if in room", func() {
		suite.Require().NoError(suite.ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"action": %q, "room": "2023-01-01"}`, subscribeAction))))
		time.Sleep(100 * time.Millisecond)
		suite.Require().Len(suite.hub.rooms, 1)

		suite.hub.handleBroadcastMessage(msg)

		suite.Require().NoError(suite.ws.SetReadDeadline(time.Now().Add(time.Second * 2)))
		mt, incoming, err := suite.ws.ReadMessage()
		suite.Require().NoError(err)
		suite.Assert().Equal(websocket.TextMessage, mt)

		incomingMsg := Message{}
		suite.Require().NoError(json.Unmarshal(incoming, &incomingMsg))
		suite.Assert().Equal(msg.Message, incomingMsg)
	})

	suite.Run("should not receive message if not in room", func() {
		suite.Require().NoError(suite.ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf(`{"action": %q, "room": "2023-01-01"}`, unsubscribeAction))))
		time.Sleep(100 * time.Millisecond)
		suite.Require().Len(suite.hub.rooms, 0)

		suite.hub.handleBroadcastMessage(msg)

		suite.Require().NoError(suite.ws.SetReadDeadline(time.Now().Add(time.Second * 2)))
		_, _, err := suite.ws.ReadMessage()
		suite.Require().Error(err)
	})
}
