package bertybot

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"moul.io/u"

	"berty.tech/berty/v2/go/pkg/bertymessenger"
)

type Bot struct {
	client          bertymessenger.MessengerServiceClient
	logger          *zap.Logger
	displayName     string
	bertyID         *bertymessenger.InstanceShareableBertyID_Reply
	skipReplay      bool
	skipAcknowledge bool
	skipMyself      bool
	handlers        map[HandlerType][]Handler
	isReplaying     bool
	handledEvents   uint
	store           struct {
		conversations map[string]*bertymessenger.Conversation
		mutex         sync.Mutex
	}
}

// New initializes a new Bot.
// The order of the passed options may have an impact.
func New(opts ...NewOption) (*Bot, error) {
	b := Bot{
		logger:   zap.NewNop(),
		handlers: make(map[HandlerType][]Handler),
	}
	b.store.conversations = make(map[string]*bertymessenger.Conversation)

	// configure bot with options
	for _, opt := range opts {
		if err := opt(&b); err != nil {
			return nil, fmt.Errorf("bot: opt failed: %w", err)
		}
	}

	// check minimal requirements
	if b.client == nil {
		return nil, fmt.Errorf("bot: missing messenger client")
	}

	// apply defaults
	if b.displayName == "" {
		b.displayName = "My Berty Bot"
	}

	// retrieve Berty ID to check if everything is well configured, and cache it for easy access
	{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req := &bertymessenger.InstanceShareableBertyID_Request{
			DisplayName: b.displayName,
		}
		ret, err := b.client.InstanceShareableBertyID(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("bot: cannot retrieve berty ID: %w", err)
		}
		b.bertyID = ret
	}

	return &b, nil
}

// BertyIDURL returns the shareable Berty ID in the form of `https://berty.tech/id#xxx`.
func (b *Bot) BertyIDURL() string {
	return b.bertyID.HTMLURL
}

// PublicKey returns the public key of the messenger node.
func (b *Bot) PublicKey() string {
	return u.B64Encode(b.bertyID.BertyID.AccountPK)
}

// Start starts the main event loop and can be stopped by canceling the passed context.
func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("connecting to the event stream")
	s, err := b.client.EventStream(ctx, &bertymessenger.EventStream_Request{})
	if err != nil {
		return fmt.Errorf("failed to listen to EventStream: %w", err)
	}

	b.isReplaying = true
	for {
		gme, err := s.Recv()
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}

		if b.isReplaying {
			if gme.Event.Type == bertymessenger.StreamEvent_TypeListEnd {
				b.logger.Info("finished replaying logs from the previous sessions", zap.Uint("count", b.handledEvents))
				b.isReplaying = false
			}
			b.handledEvents++

			if b.skipReplay {
				continue
			}
		}

		if err := b.handleEvent(ctx, gme.Event); err != nil {
			b.logger.Error("bot.handleEvent failed", zap.Error(err))
		}
	}
}
