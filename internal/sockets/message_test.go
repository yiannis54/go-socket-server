package sockets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMessageWithRoom(t *testing.T) {
	t.Run("should return valid message with room", func(t *testing.T) {
		room := "test123"
		entityID := "abc-xyz"

		expected := MessageWithRoom{
			Message: Message{
				Type:        TypeInfo,
				EntityID:    "abc-xyz",
				MessageBody: nil,
			},
			RoomName: &room,
		}

		assert.Equal(t, expected, *NewRoomMessage(TypeInfo, entityID, room, nil))
	})

	t.Run("should return valid message with body", func(t *testing.T) {
		room := "test123"
		entityID := "abc-xyz"

		expected := MessageWithRoom{
			Message: Message{
				Type:        TypeInfo,
				EntityID:    "abc-xyz",
				MessageBody: nil,
			},
			RoomName: &room,
		}

		assert.Equal(t, expected, *NewRoomMessage(TypeInfo, entityID, room, nil))
	})
}
