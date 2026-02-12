package notifications

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yiannis54/go-socket-server/internal/sockets"
)

func TestEmptyClient(t *testing.T) {
	c := Client{}

	t.Run("should return directly with no client at broadcast", func(t *testing.T) {
		assert.NotPanics(t, func() {
			c.Broadcast(context.Background(), &sockets.Message{})
		})
	})

	t.Run("should return directly with no client at private notify", func(t *testing.T) {
		assert.NotPanics(t, func() {
			c.PrivateNotify(context.Background(), &sockets.MessageWithUser{})
		})
	})
	t.Run("should return directly with no client at event change", func(t *testing.T) {
		assert.NotPanics(t, func() {
			c.NotifyRoom(context.Background(), &sockets.MessageWithRoom{})
		})
	})
}
