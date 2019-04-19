package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"berty.tech/core/entity"

	networkmock "berty.tech/core/network/mock"
	"github.com/pkg/errors"
)

func setupNonAcknowledgedEventDestinations() (*AppMock, time.Time, time.Time, time.Time) {
	n, err := NewAppMock(context.Background(), &entity.Device{Name: "Alice's iPhone"}, networkmock.NewEnqueuer(context.Background()), WithUnencryptedDb())

	if err != nil {
		panic(err)
	}

	now := time.Now()
	past := now.Add(-time.Second)
	future := now.Add(time.Second)

	n.db.Save(&entity.Event{ID: "Event1", Direction: entity.Event_Outgoing, TargetType: entity.Event_ToSpecificContact, TargetAddr: "Receiver1", SentAt: &past})
	n.db.Save(&entity.EventDispatch{EventID: "Event1", ContactID: "Receiver1", DeviceID: "Receiver1", SentAt: &past})
	n.db.Save(&entity.Event{ID: "Event2", Direction: entity.Event_Outgoing, TargetType: entity.Event_ToSpecificContact, TargetAddr: "Receiver1", SentAt: &future})
	n.db.Save(&entity.EventDispatch{EventID: "Event2", ContactID: "Receiver1", DeviceID: "Receiver1", SentAt: &future})
	n.db.Save(&entity.Event{ID: "Event3", Direction: entity.Event_Outgoing, TargetType: entity.Event_ToSpecificContact, TargetAddr: "Receiver2", SentAt: &future})
	n.db.Save(&entity.EventDispatch{EventID: "Event3", ContactID: "Receiver2", DeviceID: "Receiver2", SentAt: &future})

	n.db.Save(&entity.Event{ID: "Event4", Direction: entity.Event_Incoming, TargetType: entity.Event_ToSpecificContact, TargetAddr: "Receiver1", SentAt: &past})
	n.db.Save(&entity.EventDispatch{EventID: "Event4", ContactID: "Receiver1", DeviceID: "Receiver1", SentAt: &past})
	n.db.Save(&entity.Event{ID: "Event5", Direction: entity.Event_Incoming, TargetType: entity.Event_ToSpecificContact, TargetAddr: "Receiver1", SentAt: &future})
	n.db.Save(&entity.EventDispatch{EventID: "Event5", ContactID: "Receiver1", DeviceID: "Receiver1", SentAt: &future})
	n.db.Save(&entity.Event{ID: "Event6", Direction: entity.Event_Incoming, TargetType: entity.Event_ToSpecificContact, TargetAddr: "Receiver2", SentAt: &future})
	n.db.Save(&entity.EventDispatch{EventID: "Event6", ContactID: "Receiver2", DeviceID: "Receiver2", SentAt: &future})

	n.db.Save(&entity.Event{ID: "Event7", Direction: entity.Event_Outgoing, TargetType: entity.Event_ToSpecificConversation, TargetAddr: "Conversation1", SentAt: &past})
	n.db.Save(&entity.EventDispatch{EventID: "Event7", ContactID: "Receiver1", DeviceID: "Receiver1", SentAt: &past})
	n.db.Save(&entity.Event{ID: "Event8", Direction: entity.Event_Outgoing, TargetType: entity.Event_ToSpecificConversation, TargetAddr: "Conversation1", SentAt: &future})
	n.db.Save(&entity.EventDispatch{EventID: "Event8", ContactID: "Receiver1", DeviceID: "Receiver1", SentAt: &future})
	n.db.Save(&entity.Event{ID: "Event9", Direction: entity.Event_Outgoing, TargetType: entity.Event_ToSpecificConversation, TargetAddr: "Conversation2", SentAt: &future})
	n.db.Save(&entity.EventDispatch{EventID: "Event9", ContactID: "Receiver2", DeviceID: "Receiver2", SentAt: &future})

	n.db.Save(&entity.Event{ID: "Event10", Direction: entity.Event_Incoming, TargetType: entity.Event_ToSpecificConversation, TargetAddr: "Conversation1", SentAt: &past})
	n.db.Save(&entity.EventDispatch{EventID: "Event10", ContactID: "Receiver1", DeviceID: "Receiver1", SentAt: &past})
	n.db.Save(&entity.Event{ID: "Event11", Direction: entity.Event_Incoming, TargetType: entity.Event_ToSpecificConversation, TargetAddr: "Conversation1", SentAt: &future})
	n.db.Save(&entity.EventDispatch{EventID: "Event11", ContactID: "Receiver1", DeviceID: "Receiver1", SentAt: &future})
	n.db.Save(&entity.Event{ID: "Event12", Direction: entity.Event_Incoming, TargetType: entity.Event_ToSpecificConversation, TargetAddr: "Conversation2", SentAt: &future})
	n.db.Save(&entity.EventDispatch{EventID: "Event12", ContactID: "Receiver2", DeviceID: "Receiver2", SentAt: &future})

	return n, past, now, future
}

func TestEventRetry(t *testing.T) {
	appMock, _, now, _ := setupNonAcknowledgedEventDestinations()
	defer appMock.Close()

	expected := map[string]bool{
		"Event1": false,
		"Event2": false,
		"Event7": false,
		"Event8": false,
	}

	events, err := appMock.node.OldEventsRetry(context.Background(), now)

	if err != nil {
		t.Error(err)
	}

	for _, event := range events {
		value, ok := expected[event.ID]
		if ok == false {
			t.Error(errors.New(fmt.Sprintf("%s was not suppoesed to be found", event.ID)))
		}

		if value == true {
			t.Error(errors.New(fmt.Sprintf("%s was found twice", event.ID)))
		}

		expected[event.ID] = true
	}

	for identifier, value := range expected {
		if value == false {
			t.Error(errors.New(fmt.Sprintf("%s was supposed to be found but was not", identifier)))
		}
	}

}
