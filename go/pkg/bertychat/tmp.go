package bertychat

// this file contains temorary functions that should be deleted as soon as they are replace by a better solution.
// if you add new functions in this file, try to always put an annoying warning log message, as a reminder.

import (
	"fmt"
	"math/rand"

	"berty.tech/go/pkg/chatmodel"
	"github.com/brianvoe/gofakeit"
	"go.uber.org/zap"
)

func fakeConversation(logger *zap.Logger) *chatmodel.Conversation {
	logger.Warn("randomConversation is temporary")

	created := gofakeit.Date()
	updated := gofakeit.Date()
	return &chatmodel.Conversation{
		ID:        uint64(rand.Uint32())<<32 + uint64(rand.Uint32()),
		CreatedAt: created,
		UpdatedAt: updated,
		Title:     fmt.Sprintf("%s %s", gofakeit.HackerIngverb(), gofakeit.HackerAdjective()),
	}
}

func fakeContact(logger *zap.Logger) *chatmodel.Contact {
	logger.Warn("randomContact is temporary")

	created := gofakeit.Date()
	updated := gofakeit.Date()
	seen := gofakeit.Date()
	return &chatmodel.Contact{
		ID:        uint64(rand.Uint32())<<32 + uint64(rand.Uint32()),
		CreatedAt: created,
		UpdatedAt: updated,
		SeenAt:    &seen,
		Name:      fmt.Sprintf("%s %s", gofakeit.HackerIngverb(), gofakeit.HackerAdjective()),
		// ProtocolID
		// AvatarUri
		// StatusEmoji
		// StatusText
		// Kind
		// Blocked
		// Devices
	}
}
