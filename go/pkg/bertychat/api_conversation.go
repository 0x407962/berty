package bertychat

import (
	"context"
	"math/rand"

	"github.com/pkg/errors"
)

func (c *client) ConversationList(req *ConversationListRequest, stream Account_ConversationListServer) error {
	max := rand.Intn(11) + 5 // 5-15
	for i := 0; i < max; i++ {
		conversation := fakeConversation(c.logger)
		err := stream.Send(&ConversationListReply{Conversation: conversation})
		if err != nil {
			return errors.Wrap(err, "failed to send conversation to stream")
		}
	}
	return nil
}

func (c *client) ConversationGet(ctx context.Context, input *ConversationGetRequest) (*ConversationGetReply, error) {
	if input == nil || input.ID == "" {
		return nil, ErrMissingInput
	}
	if input.ID == "invalid" { // simulating an invalid ID (tmp)
		return nil, ErrInvalidInput
	}
	return &ConversationGetReply{
		Conversation: fakeConversation(c.logger),
	}, nil
}

func (c *client) ConversationCreate(context.Context, *ConversationCreateRequest) (*ConversationCreateReply, error) {
	return nil, ErrNotImplemented
}

func (c *client) ConversationLeave(context.Context, *ConversationLeaveRequest) (*ConversationLeaveReply, error) {
	return nil, ErrNotImplemented
}

func (c *client) ConversationErase(context.Context, *ConversationEraseRequest) (*ConversationEraseReply, error) {
	return nil, ErrNotImplemented
}

func (c *client) ConversationSetSeenPosition(context.Context, *ConversationSetSeenPositionRequest) (*ConversationSetSeenPositionReply, error) {
	return nil, ErrNotImplemented
}

func (c *client) ConversationUpdateSettings(context.Context, *ConversationUpdateSettingsRequest) (*ConversationUpdateSettingsReply, error) {
	return nil, ErrNotImplemented
}
