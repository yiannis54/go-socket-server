package sockets

// Room subscription actions.
const (
	subscribeAction   string = "enter"
	unsubscribeAction string = "leave"
)

// Message holds the information of the notification message sent.
type Message struct {
	Type        MessageType `json:"type"`
	EntityID    string      `json:"entityId"`
	MessageBody any         `json:"message"`
}

// MessageWithRoom adds a room in the message information sent.
type MessageWithRoom struct {
	Message
	RoomName *string `json:"room"`
}

// MessageWithUser adds a user id in the message information sent.
type MessageWithUser struct {
	Message
	UserID string `json:"userId"`
}

// IncomingSubscription is used as incoming message for changing rooms.
type IncomingSubscription struct {
	Action string `json:"action"`
	Room   string `json:"room"`
}

// Subscription is the object sent to hub for handling the room registrations.
type Subscription struct {
	Room   string
	client *Client
}

func newSubscription(room string, c *Client) *Subscription {
	return &Subscription{
		Room:   room,
		client: c,
	}
}

// NewRoomMessage constructs and returns a message with room to be sent to subscribed users.
func NewRoomMessage(msgType MessageType, entityID, room string, body any) *MessageWithRoom {
	return &MessageWithRoom{
		Message: Message{
			Type:        msgType,
			EntityID:    entityID,
			MessageBody: body,
		},
		RoomName: &room,
	}
}
