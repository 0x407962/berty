package node

import (
	"context"
	"fmt"
	"time"

	"github.com/gogo/protobuf/proto"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"berty.tech/core/api/node"
	"berty.tech/core/api/p2p"
	"berty.tech/core/crypto/keypair"
	"berty.tech/core/pkg/tracing"
)

// EventsRetry updates SentAt and requeue an event
func (n *Node) EventRequeue(ctx context.Context, event *p2p.Event) error {
	var span opentracing.Span
	span, ctx = tracing.EnterFunc(ctx)
	defer span.Finish()

	sql := n.sql(ctx)

	now := time.Now()
	event.SentAt = &now
	if err := sql.Save(event).Error; err != nil {
		return errors.Wrap(err, "error while updating SentAt on event")
	}
	n.outgoingEvents <- event

	return nil
}

// EventsRetry sends events which lack an AckedAt value emitted before the supplied time value
func (n *Node) EventsRetry(ctx context.Context, before time.Time) ([]*p2p.Event, error) {
	var span opentracing.Span
	span, ctx = tracing.EnterFunc(ctx)
	defer span.Finish()

	sql := n.sql(ctx)
	var retriedEvents []*p2p.Event
	destinations, err := p2p.FindNonAcknowledgedEventDestinations(sql, before)

	if err != nil {
		return nil, err
	}

	for _, destination := range destinations {
		events, err := p2p.FindNonAcknowledgedEventsForDestination(sql, destination)

		if err != nil {
			n.LogBackgroundError(ctx, errors.Wrap(err, "error while retrieving events for dst"))
			continue
		}

		for _, event := range events {
			if err := n.EventRequeue(ctx, event); err != nil {
				n.LogBackgroundError(ctx, errors.Wrap(err, "error while enqueuing event"))
				continue
			}
			retriedEvents = append(retriedEvents, event)
		}
	}

	return retriedEvents, nil
}

func (n *Node) cron(ctx context.Context) {
	var span opentracing.Span
	span, ctx = tracing.EnterFunc(ctx)
	defer span.Finish()
	for true {
		before := time.Now().Add(-time.Second * 60 * 10)
		if _, err := n.EventsRetry(ctx, before); err != nil {
			n.LogBackgroundError(ctx, err)
		}

		time.Sleep(time.Second * 60)
	}
}

// Start is the node's mainloop
func (n *Node) Start(ctx context.Context, withCron, withNodeEvents bool) error {
	var span opentracing.Span
	span, ctx = tracing.EnterFunc(ctx)
	defer span.Finish()

	if withCron {
		go n.cron(ctx)
	}

	if withNodeEvents {
		// "node started" event
		go func() {
			time.Sleep(time.Second)
			n.EnqueueNodeEvent(ctx, node.Kind_NodeStarted, nil)
		}()

		// "node is alive" event
		go func() {
			for {
				n.EnqueueNodeEvent(ctx, node.Kind_NodeIsAlive, nil)
				time.Sleep(30 * time.Second)
			}
		}()
	}

	for {
		select {
		case event := <-n.outgoingEvents:
			logger().Debug("outgoing event", zap.Stringer("event", event))
			envelope := p2p.Envelope{}
			eventBytes, err := proto.Marshal(event)
			if err != nil {
				n.LogBackgroundError(ctx, errors.Wrap(err, "failed to marshal outgoing event"))
				continue
			}

			event.SenderID = n.b64pubkey

			switch {
			case event.ReceiverID != "": // ContactEvent
				envelope.Source = n.aliasEnvelopeForContact(ctx, &envelope, event)
				envelope.ChannelID = event.ReceiverID
				envelope.EncryptedEvent = eventBytes // FIXME: encrypt for receiver

			case event.ConversationID != "": //ConversationEvent
				envelope.Source = n.aliasEnvelopeForConversation(ctx, &envelope, event)
				envelope.ChannelID = event.ConversationID
				envelope.EncryptedEvent = eventBytes // FIXME: encrypt for conversation

			default:
				n.LogBackgroundError(ctx, fmt.Errorf("unhandled event type"))
			}

			if envelope.Signature, err = keypair.Sign(n.crypto, &envelope); err != nil {
				n.LogBackgroundError(ctx, errors.Wrap(err, "failed to sign envelope"))
				continue
			}

			// Async subscribe to conversation
			// wait for 1s to simulate a sync subscription,
			// if too long, the task will be done in background
			done := make(chan bool, 1)
			go func() {
				// FIXME: make something smarter, i.e., grouping events by contact or network driver
				if err := n.networkDriver.Emit(ctx, &envelope); err != nil {
					n.LogBackgroundError(ctx, errors.Wrap(err, "failed to emit envelope on network"))
				}
				done <- true
			}()
			select {
			case <-done:
			case <-time.After(1 * time.Second):
			}

			// push the outgoing event on the client stream
			n.clientEvents <- event

			// emit the outgoing event on the node event stream
		case event := <-n.clientEvents:
			logger().Debug("client event", zap.Stringer("event", event))
			n.clientEventsMutex.Lock()
			for _, sub := range n.clientEventsSubscribers {
				if sub.filter(event) {
					sub.queue <- event
				}
			}
			n.clientEventsMutex.Unlock()
		}
	}
}
