package graphql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"berty.tech/core/api/node"
	"berty.tech/core/api/node/graphql/graph/generated"
	"berty.tech/core/api/node/graphql/models"
	gql "berty.tech/core/api/protobuf/graphql"
	"berty.tech/core/entity"
	network_metric "berty.tech/core/network/metric"
	"berty.tech/core/pkg/deviceinfo"
	"berty.tech/core/push"
	"berty.tech/core/sql"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"go.uber.org/zap"
)

type Resolver struct {
	client node.ServiceClient
}

func New(client node.ServiceClient) generated.Config {
	return generated.Config{
		Resolvers: &Resolver{client},
	}
}
func (r *Resolver) GqlNode() generated.GqlNodeResolver {
	return &gqlNodeResolver{r}
}
func (r *Resolver) BertyEntityContact() generated.BertyEntityContactResolver {
	return &bertyEntityContactResolver{r}
}
func (r *Resolver) BertyEntityConversation() generated.BertyEntityConversationResolver {
	return &bertyEntityConversationResolver{r}
}
func (r *Resolver) BertyEntityConversationMember() generated.BertyEntityConversationMemberResolver {
	return &bertyEntityConversationMemberResolver{r}
}
func (r *Resolver) BertyEntityDevice() generated.BertyEntityDeviceResolver {
	return &bertyEntityDeviceResolver{r}
}

func (r *Resolver) BertyEntityDevicePushIdentifier() generated.BertyEntityDevicePushIdentifierResolver {
	return &bertyEntityDevicePushIdentifierResolver{r}
}

func (r *Resolver) BertyEntityDevicePushConfig() generated.BertyEntityDevicePushConfigResolver {
	return &bertyEntityDevicePushConfigResolver{r}
}

func (r *Resolver) BertyEntityEvent() generated.BertyEntityEventResolver {
	return &bertyEntityEventResolver{r}
}

// func (r *Resolver) BertyP2pPeer() generated.BertyP2pPeerResolver {
// 	return &bertyP2pPeerResolver{r}
// }
func (r *Resolver) GoogleProtobufFieldDescriptorProto() generated.GoogleProtobufFieldDescriptorProtoResolver {
	return &googleProtobufFieldDescriptorProtoResolver{r}
}
func (r *Resolver) GoogleProtobufFieldOptions() generated.GoogleProtobufFieldOptionsResolver {
	return &googleProtobufFieldOptionsResolver{r}
}
func (r *Resolver) GoogleProtobufFileOptions() generated.GoogleProtobufFileOptionsResolver {
	return &googleProtobufFileOptionsResolver{r}
}
func (r *Resolver) GoogleProtobufMethodOptions() generated.GoogleProtobufMethodOptionsResolver {
	return &googleProtobufMethodOptionsResolver{r}
}

func (r *Resolver) Mutation() generated.MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() generated.QueryResolver {
	return &queryResolver{r}
}
func (r *Resolver) Subscription() generated.SubscriptionResolver {
	return &subscriptionResolver{r}
}

type gqlNodeResolver struct{ *Resolver }

func (r *gqlNodeResolver) ID(ctx context.Context, obj *gql.Node) (string, error) {
	// TODO: update entity id in db to have unique IDs in whole database
	return "unknown:" + obj.ID, nil
}

type bertyEntityContactResolver struct{ *Resolver }

func (r *bertyEntityContactResolver) ID(ctx context.Context, obj *entity.Contact) (string, error) {
	return "contact:" + obj.ID, nil
}

type bertyEntityConversationResolver struct{ *Resolver }

func (r *bertyEntityConversationResolver) ID(ctx context.Context, obj *entity.Conversation) (string, error) {
	return "conversation:" + obj.ID, nil
}

type bertyEntityConversationMemberResolver struct{ *Resolver }

func (r *bertyEntityConversationMemberResolver) ID(ctx context.Context, obj *entity.ConversationMember) (string, error) {
	return "conversation_member:" + obj.ID, nil
}
func (r *bertyEntityConversationMemberResolver) ContactID(ctx context.Context, obj *entity.ConversationMember) (string, error) {
	return "contact:" + obj.ContactID, nil
}
func (r *bertyEntityConversationMemberResolver) ConversationID(ctx context.Context, obj *entity.ConversationMember) (string, error) {
	return "conversation:" + obj.ConversationID, nil
}

type bertyEntityDevicePushIdentifierResolver struct{ *Resolver }

func (r *bertyEntityDevicePushIdentifierResolver) ID(ctx context.Context, obj *entity.DevicePushIdentifier) (string, error) {
	return "device_push_identifier:" + obj.ID, nil
}

type bertyEntityDevicePushConfigResolver struct{ *Resolver }

func (r *bertyEntityDevicePushConfigResolver) ID(ctx context.Context, obj *entity.DevicePushConfig) (string, error) {
	return "device_push_config:" + obj.ID, nil
}

type bertyEntityDeviceResolver struct{ *Resolver }

func (r *bertyEntityDeviceResolver) ID(ctx context.Context, obj *entity.Device) (string, error) {
	return "device:" + obj.ID, nil
}

type bertyEntityEventResolver struct{ *Resolver }

func (r *bertyEntityEventResolver) ID(ctx context.Context, obj *entity.Event) (string, error) {
	return "event:" + obj.ID, nil
}

func (r *bertyEntityEventResolver) ConversationID(ctx context.Context, obj *entity.Event) (string, error) {
	return "conversation:" + obj.ConversationID, nil
}

func (r *bertyEntityEventResolver) Attributes(ctx context.Context, obj *entity.Event) ([]byte, error) {
	attrs, err := obj.GetAttrs()
	if err != nil {
		return nil, err
	}
	return json.Marshal(attrs)
}

// type bertyP2pPeerResolver struct{ *Resolver }

// func (r *bertyP2pPeerResolver) Addrs(ctx context.Context, obj *network_metric.Peer) ([]byte, error) {
// 	return json.Marshal(obj.GetAddrs())
// }

type googleProtobufFieldDescriptorProtoResolver struct{ *Resolver }

func (r *googleProtobufFieldDescriptorProtoResolver) Label(ctx context.Context, obj *descriptor.FieldDescriptorProto) (*int32, error) {
	ret := int32(*obj.Label)
	return &ret, nil
}
func (r *googleProtobufFieldDescriptorProtoResolver) Type(ctx context.Context, obj *descriptor.FieldDescriptorProto) (*int32, error) {
	ret := int32(*obj.Type)
	return &ret, nil
}

type googleProtobufFieldOptionsResolver struct{ *Resolver }

func (r *googleProtobufFieldOptionsResolver) Ctype(ctx context.Context, obj *descriptor.FieldOptions) (*int32, error) {
	ret := int32(*obj.Ctype)
	return &ret, nil
}
func (r *googleProtobufFieldOptionsResolver) Jstype(ctx context.Context, obj *descriptor.FieldOptions) (*int32, error) {
	ret := int32(*obj.Jstype)
	return &ret, nil
}

type googleProtobufFileOptionsResolver struct{ *Resolver }

func (r *googleProtobufFileOptionsResolver) PhpMetadataNamespace(ctx context.Context, obj *descriptor.FileOptions) (string, error) {
	return "", nil
}
func (r *googleProtobufFileOptionsResolver) RubyPackage(ctx context.Context, obj *descriptor.FileOptions) (string, error) {
	return "", nil
}

func (r *googleProtobufFileOptionsResolver) OptimizeFor(ctx context.Context, obj *descriptor.FileOptions) (*int32, error) {
	ret := int32(*obj.OptimizeFor)
	return &ret, nil
}

type googleProtobufMethodOptionsResolver struct{ *Resolver }

func (r *googleProtobufMethodOptionsResolver) IdempotencyLevel(ctx context.Context, obj *descriptor.MethodOptions) (*int32, error) {
	ret := int32(*obj.IdempotencyLevel)
	return &ret, nil
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) ConfigUpdate(ctx context.Context, id string, createdAt *time.Time, updatedAt *time.Time, myself *entity.Contact, myselfID string, currentDevice *entity.Device, currentDeviceID string, cryptoParams []byte, pushRelayPubkeyAPNS string, pushRelayPubkeyFCM string, notificationsEnabled bool, notificationsPreviews bool, notificationsDebugLevel *int32) (*entity.Config, error) {
	debugNotificationsVerbosity := entity.DebugVerbosity_VERBOSITY_LEVEL_NONE
	if notificationsDebugLevel != nil {
		debugNotificationsVerbosity = entity.DebugVerbosity(*notificationsDebugLevel)
	}

	return r.client.ConfigUpdate(ctx, &entity.Config{
		PushRelayPubkeyAPNS:        pushRelayPubkeyAPNS,
		PushRelayPubkeyFCM:         pushRelayPubkeyFCM,
		NotificationsEnabled:       notificationsEnabled,
		NotificationsPreviews:      notificationsPreviews,
		DebugNotificationVerbosity: debugNotificationsVerbosity,
	})
}

func (r *mutationResolver) DevicePushConfigNativeRegister(ctx context.Context, T bool) (*node.Void, error) {
	return r.client.DevicePushConfigNativeRegister(ctx, &node.Void{})
}

func (r *mutationResolver) DevicePushConfigNativeUnregister(ctx context.Context, T bool) (*node.Void, error) {
	return r.client.DevicePushConfigNativeUnregister(ctx, &node.Void{})
}

func (r *mutationResolver) DevicePushConfigCreate(ctx context.Context, relayPubkey *string, pushID []byte, pushType *int32) (*entity.DevicePushConfig, error) {
	var pushTypeEnum push.DevicePushType
	if pushType != nil {
		pushTypeEnum = push.DevicePushType(*pushType)
	}

	return r.client.DevicePushConfigCreate(ctx, &node.DevicePushConfigCreateInput{
		PushID:      pushID,
		PushType:    pushTypeEnum,
		RelayPubkey: *relayPubkey,
	})
}

func (r *mutationResolver) DevicePushConfigRemove(ctx context.Context, id string) (*entity.DevicePushConfig, error) {
	id = strings.SplitN(id, ":", 2)[1]

	return r.client.DevicePushConfigRemove(ctx, &entity.DevicePushConfig{
		ID: id,
	})
}

func (r *mutationResolver) DevicePushConfigUpdate(ctx context.Context, id string, createdAt *time.Time, updatedAt *time.Time, deviceID string, pushType *int32, pushID []byte, relayPubkey string) (*entity.DevicePushConfig, error) {
	var pushTypeEnum push.DevicePushType
	if pushType != nil {
		pushTypeEnum = push.DevicePushType(*pushType)
	}

	return r.client.DevicePushConfigUpdate(ctx, &entity.DevicePushConfig{
		ID:          id,
		PushID:      pushID,
		PushType:    pushTypeEnum,
		RelayPubkey: relayPubkey,
	})
}

func (r *mutationResolver) RunIntegrationTests(ctx context.Context, name string) (*node.IntegrationTestOutput, error) {
	return r.client.RunIntegrationTests(ctx, &node.IntegrationTestInput{
		Name: name,
	})
}

func (r *mutationResolver) ContactRequest(ctx context.Context, contactID, contactOverrideDisplayName, introText string) (*entity.Contact, error) {
	contactID = strings.SplitN(contactID, ":", 2)[1]
	return r.client.ContactRequest(ctx, &node.ContactRequestInput{
		ContactID:                  contactID,
		ContactOverrideDisplayName: contactOverrideDisplayName,
		IntroText:                  introText,
	})
}
func (r *mutationResolver) ContactAcceptRequest(ctx context.Context, contactID string) (*entity.Contact, error) {
	return r.client.ContactAcceptRequest(ctx, &node.ContactAcceptRequestInput{
		ContactID: strings.SplitN(contactID, ":", 2)[1],
	})
}
func (r *mutationResolver) ContactRemove(ctx context.Context, id string, createdAt *time.Time, updatedAt *time.Time, sigchain []byte, status *int32, devices []*entity.Device, displayName string, displayStatus string, overrideDisplayName string, overrideDisplayStatus string) (*entity.Contact, error) {
	return r.client.ContactRemove(ctx, &entity.Contact{
		ID: strings.SplitN(id, ":", 2)[1],
	})
}
func (r *mutationResolver) ContactUpdate(ctx context.Context, id string, createdAt *time.Time, updatedAt *time.Time, sigchain []byte, status *int32, devices []*entity.Device, displayName string, displayStatus string, overrideDisplayName string, overrideDisplayStatus string) (*entity.Contact, error) {
	return r.client.ContactUpdate(ctx, &entity.Contact{
		ID:                    strings.SplitN(id, ":", 2)[1],
		DisplayName:           displayName,
		OverrideDisplayName:   overrideDisplayName,
		DisplayStatus:         displayStatus,
		OverrideDisplayStatus: overrideDisplayStatus,
	})
}
func (r *mutationResolver) ConversationCreate(ctx context.Context, contacts []*entity.Contact, title string, topic string) (*entity.Conversation, error) {
	if contacts != nil {
		for i, contact := range contacts {
			contacts[i].ID = strings.SplitN(contact.ID, ":", 2)[1]
		}
	}
	return r.client.ConversationCreate(ctx, &node.ConversationCreateInput{
		Contacts: contacts,
		Title:    title,
		Topic:    topic,
	})
}
func (r *mutationResolver) ConversationInvite(ctx context.Context, conversation *entity.Conversation, members []*entity.ConversationMember) (*entity.Conversation, error) {

	return r.client.ConversationInvite(ctx, &node.ConversationManageMembersInput{Conversation: conversation, Members: members})
}
func (r *mutationResolver) ConversationExclude(ctx context.Context, conversation *entity.Conversation, members []*entity.ConversationMember) (*entity.Conversation, error) {
	return r.client.ConversationExclude(ctx, &node.ConversationManageMembersInput{Conversation: conversation, Members: members})
}
func (r *mutationResolver) ConversationAddMessage(ctx context.Context, conversation *entity.Conversation, message *entity.Message) (*entity.Event, error) {
	if conversation != nil {
		if conversation.ID != "" {
			conversation.ID = strings.SplitN(conversation.ID, ":", 2)[1]
		}
	}
	return r.client.ConversationAddMessage(ctx, &node.ConversationAddMessageInput{Conversation: conversation, Message: message})
}
func (r *mutationResolver) GenerateFakeData(ctx context.Context, T bool) (*node.Void, error) {
	return r.client.GenerateFakeData(ctx, &node.Void{T: true})
}
func (r *mutationResolver) DebugRequeueEvent(ctx context.Context, eventID string) (*entity.Event, error) {
	eventID = strings.SplitN(eventID, ":", 2)[1]

	return r.client.DebugRequeueEvent(ctx, &node.EventIDInput{
		EventID: eventID,
	})
}

func (r *mutationResolver) DebugRequeueAll(ctx context.Context, T bool) (*node.Void, error) {
	return r.client.DebugRequeueAll(ctx, &node.Void{})
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Node(ctx context.Context, id string) (models.Node, error) {
	gID := strings.SplitN(id, ":", 2)
	switch gID[0] {
	case "contact":
		return r.Contact(ctx, &entity.Contact{ID: id})
	case "conversation":
		return r.client.Conversation(ctx, &entity.Conversation{ID: id})
	case "conversation_member":
		return r.client.ConversationMember(ctx, &entity.ConversationMember{ID: id})
	case "event":
		return r.GetEvent(ctx, id)
	default:
		logger().Warn("unknown node type", zap.String("node_type", gID[0]))
		return nil, nil
	}
}

func (r *queryResolver) LogfileList(ctx context.Context, T bool) ([]*node.LogfileEntry, error) {
	stream, err := r.client.LogfileList(ctx, &node.Void{})
	if err != nil {
		return nil, err
	}

	entries := []*node.LogfileEntry{}
	for {
		entry, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *queryResolver) ID(ctx context.Context, T bool) (*network_metric.Peer, error) {
	return r.client.ID(ctx, &node.Void{T: T})
}

func (r *queryResolver) Protocols(ctx context.Context, id string, _ []string, _ *int32) (*node.ProtocolsOutput, error) {
	return r.client.Protocols(ctx, &network_metric.Peer{
		ID: id,
	})
}

func (r *queryResolver) EventList(ctx context.Context, filter *entity.Event, rawOnlyWithoutAckedAt *int32, rawOnlyWithoutSeenAt *int32, orderBy string, orderDesc bool, first *int32, after *string, last *int32, before *string) (*node.EventListConnection, error) {
	onlyWithoutAckedAt := node.NullableTrueFalse_Null
	if rawOnlyWithoutAckedAt != nil {
		onlyWithoutAckedAt = node.NullableTrueFalse(*rawOnlyWithoutAckedAt)
		logger().Info(fmt.Sprintf("raw value %+v parsed value %+v", rawOnlyWithoutAckedAt, onlyWithoutAckedAt))
	}

	onlyWithoutSeenAt := node.NullableTrueFalse_Null
	if rawOnlyWithoutSeenAt != nil {
		onlyWithoutSeenAt = node.NullableTrueFalse(*rawOnlyWithoutSeenAt)
		logger().Info(fmt.Sprintf("raw value %+v parsed value %+v", rawOnlyWithoutSeenAt, onlyWithoutSeenAt))
	}

	if filter != nil {
		if filter.ID != "" {
			filter.ID = strings.SplitN(filter.ID, ":", 2)[1]
		}
		if filter.ConversationID != "" {
			filter.ConversationID = strings.SplitN(filter.ConversationID, ":", 2)[1]
		}
	}

	input := &node.EventListInput{
		Filter:             filter,
		Paginate:           getPagination(first, after, last, before),
		OnlyWithoutAckedAt: onlyWithoutAckedAt,
		OnlyWithoutSeenAt:  onlyWithoutSeenAt,
	}

	input.Paginate.OrderBy = orderBy
	input.Paginate.OrderDesc = orderDesc

	if input.Paginate.First > 0 || input.Paginate.Last == 0 {
		input.Paginate.First++ // querying one more field to fullfil HasNextPage, FIXME: optimize this
	} else if input.Paginate.Last > 0 {
		input.Paginate.Last++
	}

	stream, err := r.client.EventList(ctx, input)

	if err != nil {
		return nil, err
	}

	output := &node.EventListConnection{
		Edges: []*node.EventEdge{},
	}
	count := int32(0)
	hasNextPage := false
	for {
		n, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if (input.Paginate.First > 0 && count >= input.Paginate.First-1) || (input.Paginate.Last > 0 && count >= input.Paginate.Last-1) { // related to input.Paginate.First++
			hasNextPage = true
			break
		}
		count++

		var cursor string
		switch orderBy {
		case "", "id":
			cursor = n.ID
		case "created_at":
			cursor = n.CreatedAt.Format(sql.TimestampFormat)
		case "updated_at":
			cursor = n.UpdatedAt.Format(sql.TimestampFormat)
		}

		output.Edges = append(output.Edges, &node.EventEdge{
			Node:   n,
			Cursor: cursor,
		})
	}

	output.PageInfo = &node.PageInfo{}
	if len(output.Edges) == 0 {
		output.PageInfo.StartCursor = ""
		output.PageInfo.EndCursor = ""
	} else {
		output.PageInfo.StartCursor = output.Edges[0].Cursor
		output.PageInfo.EndCursor = output.Edges[len(output.Edges)-1].Cursor
	}

	if input.Paginate.First > 0 || input.Paginate.Last == 0 {
		output.PageInfo.HasPreviousPage = input.Paginate.After != ""
	} else {
		output.PageInfo.HasPreviousPage = input.Paginate.Before != ""
	}
	output.PageInfo.HasNextPage = hasNextPage
	return output, nil
}

func (r *queryResolver) EventUnseen(ctx context.Context, filter *entity.Event, rawOnlyWithoutAckedAt *int32, rawOnlyWithoutSeenAt *int32, orderBy string, orderDesc bool, first *int32, after *string, last *int32, before *string) ([]*entity.Event, error) {
	onlyWithoutAckedAt := node.NullableTrueFalse_Null
	if rawOnlyWithoutAckedAt != nil {
		onlyWithoutAckedAt = node.NullableTrueFalse(*rawOnlyWithoutAckedAt)
		logger().Info(fmt.Sprintf("raw value %+v parsed value %+v", rawOnlyWithoutAckedAt, onlyWithoutAckedAt))
	}

	onlyWithoutSeenAt := node.NullableTrueFalse_Null
	if rawOnlyWithoutSeenAt != nil {
		onlyWithoutSeenAt = node.NullableTrueFalse(*rawOnlyWithoutSeenAt)
		logger().Info(fmt.Sprintf("raw value %+v parsed value %+v", rawOnlyWithoutSeenAt, onlyWithoutSeenAt))
	}

	if filter != nil {
		if filter.ID != "" {
			filter.ID = strings.SplitN(filter.ID, ":", 2)[1]
		}
		if filter.ConversationID != "" {
			filter.ConversationID = strings.SplitN(filter.ConversationID, ":", 2)[1]
		}
	}

	input := &node.EventListInput{
		Filter:             filter,
		OnlyWithoutAckedAt: onlyWithoutAckedAt,
		OnlyWithoutSeenAt:  onlyWithoutSeenAt,
	}

	stream, err := r.client.EventUnseen(ctx, input)
	if err != nil {
		return nil, err
	}

	entries := []*entity.Event{}
	for {
		entry, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func (r *queryResolver) GetEvent(ctx context.Context, id string) (*entity.Event, error) {
	id = strings.SplitN(id, ":", 2)[1]

	return r.client.GetEvent(ctx, &entity.Event{
		ID: strings.SplitN(id, ":", 2)[1],
	})
}

func (r *mutationResolver) EventSeen(ctx context.Context, id string) (*entity.Event, error) {
	id = strings.SplitN(id, ":", 2)[1]

	return r.client.EventSeen(ctx, &entity.Event{
		ID: id,
	})
}

func (r *queryResolver) ConfigPublic(ctx context.Context, void bool) (*entity.Config, error) {
	return r.client.ConfigPublic(ctx, &node.Void{})
}

func (r *queryResolver) ContactList(ctx context.Context, filter *entity.Contact, orderBy string, orderDesc bool, first *int32, after *string, last *int32, before *string) (*node.ContactListConnection, error) {
	if filter != nil && filter.ID != "" {
		filter.ID = strings.SplitN(filter.ID, ":", 2)[1]
	}

	input := &node.ContactListInput{
		Filter:   filter,
		Paginate: getPagination(first, after, last, before),
	}

	input.Paginate.OrderBy = orderBy
	input.Paginate.OrderDesc = orderDesc

	input.Paginate.First++ // querying one more field to fullfil HasNextPage, FIXME: optimize this

	stream, err := r.client.ContactList(ctx, input)
	if err != nil {
		return nil, err
	}

	output := &node.ContactListConnection{
		Edges: []*node.ContactEdge{},
	}
	count := int32(0)
	hasNextPage := false
	for {
		n, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if count >= input.Paginate.First-1 { // related to input.Paginate.First++
			hasNextPage = true
			break
		}
		count++

		var cursor string
		switch orderBy {
		case "", "id":
			cursor = n.ID
		case "created_at":
			cursor = n.CreatedAt.Format(sql.TimestampFormat)
		case "updated_at":
			cursor = n.UpdatedAt.Format(sql.TimestampFormat)
		}

		output.Edges = append(output.Edges, &node.ContactEdge{
			Node:   n,
			Cursor: cursor,
		})
	}

	output.PageInfo = &node.PageInfo{}
	if len(output.Edges) == 0 {
		output.PageInfo.StartCursor = ""
		output.PageInfo.EndCursor = ""
	} else {
		output.PageInfo.StartCursor = output.Edges[0].Cursor
		output.PageInfo.EndCursor = output.Edges[len(output.Edges)-1].Cursor
	}

	output.PageInfo.HasPreviousPage = input.Paginate.After != ""
	output.PageInfo.HasNextPage = hasNextPage
	return output, nil
}
func (r *queryResolver) Contact(ctx context.Context, filter *entity.Contact) (*entity.Contact, error) {
	if filter.ID != "" {
		filter.ID = strings.SplitN(filter.ID, ":", 2)[1]
	}
	if filter.Devices != nil && len(filter.Devices) != 0 {
		for i := range filter.Devices {
			filter.Devices[i].ID = strings.SplitN(filter.Devices[i].ID, ":", 2)[1]
		}
	}
	return r.client.Contact(ctx, &node.ContactInput{Filter: filter})
}
func (r *queryResolver) ConversationList(ctx context.Context, filter *entity.Conversation, orderBy string, orderDesc bool, first *int32, after *string, last *int32, before *string) (*node.ConversationListConnection, error) {
	if filter != nil && filter.ID != "" {
		filter.ID = strings.SplitN(filter.ID, ":", 2)[1]
	}

	input := &node.ConversationListInput{
		Filter:   filter,
		Paginate: getPagination(first, after, last, before),
	}

	input.Paginate.OrderBy = orderBy
	input.Paginate.OrderDesc = orderDesc

	input.Paginate.First++ // querying one more field to fullfil HasNextPage, FIXME: optimize this

	stream, err := r.client.ConversationList(ctx, input)
	if err != nil {
		return nil, err
	}

	output := &node.ConversationListConnection{
		Edges: []*node.ConversationEdge{},
	}
	count := int32(0)
	hasNextPage := false
	for {
		n, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if count >= input.Paginate.First-1 { // related to input.Paginate.First++
			hasNextPage = true
			break
		}
		count++

		var cursor string
		switch orderBy {
		case "", "id":
			cursor = n.ID
		case "created_at":
			cursor = n.CreatedAt.Format(sql.TimestampFormat)
		case "updated_at":
			cursor = n.UpdatedAt.Format(sql.TimestampFormat)
		}

		output.Edges = append(output.Edges, &node.ConversationEdge{
			Node:   n,
			Cursor: cursor,
		})
	}

	output.PageInfo = &node.PageInfo{}
	if len(output.Edges) == 0 {
		output.PageInfo.StartCursor = ""
		output.PageInfo.EndCursor = ""
	} else {
		output.PageInfo.StartCursor = output.Edges[0].Cursor
		output.PageInfo.EndCursor = output.Edges[len(output.Edges)-1].Cursor
	}

	output.PageInfo.HasPreviousPage = input.Paginate.After != ""
	output.PageInfo.HasNextPage = hasNextPage
	return output, nil
}
func (r *queryResolver) ContactCheckPublicKey(ctx context.Context, contact *entity.Contact) (*node.Bool, error) {
	contact.ID = strings.SplitN(contact.ID, ":", 2)[1]

	return r.client.ContactCheckPublicKey(ctx, &node.ContactInput{
		Filter: contact,
	})
}

func (r *mutationResolver) ConversationUpdate(ctx context.Context, id string, createdAt, updatedAt, readAt *time.Time, title, topic string, infos string, members []*entity.ConversationMember) (*entity.Conversation, error) {
	if id == "" {
		return nil, errors.New("no id supplied")
	}
	id = strings.SplitN(id, ":", 2)[1]
	if members != nil && len(members) > 0 {
		for i := range members {
			if members[i] == nil || members[i].ID == "" {
				continue
			}
			members[i].ID = strings.SplitN(members[i].ID, ":", 2)[1]
		}
	}

	return r.client.ConversationUpdate(ctx, &entity.Conversation{
		ID:        id,
		CreatedAt: *createdAt,
		UpdatedAt: *updatedAt,
		ReadAt:    *readAt,
		Title:     title,
		Topic:     topic,
		Infos:     infos,
		Members:   members,
	})
}

func (r *mutationResolver) ConversationRead(ctx context.Context, id string) (*entity.Conversation, error) {
	id = strings.SplitN(id, ":", 2)[1]

	return r.client.ConversationRead(ctx, &entity.Conversation{
		ID: id,
	})
}

func (r *mutationResolver) ConversationRemove(ctx context.Context, id string) (*entity.Conversation, error) {
	id = strings.SplitN(id, ":", 2)[1]

	return r.client.ConversationRemove(ctx, &entity.Conversation{
		ID: id,
	})
}

func (r *queryResolver) Conversation(ctx context.Context, id string, createdAt, updatedAt, readAt *time.Time, title, topic string, infos string, members []*entity.ConversationMember) (*entity.Conversation, error) {
	if id != "" {
		id = strings.SplitN(id, ":", 2)[1]
	}
	if members != nil && len(members) > 0 {
		for i := range members {
			if members[i] == nil || members[i].ID == "" {
				continue
			}
			members[i].ID = strings.SplitN(members[i].ID, ":", 2)[1]
		}
	}

	return r.client.Conversation(ctx, &entity.Conversation{
		ID: id,
	})
}
func (r *queryResolver) ConversationMember(ctx context.Context, id string, createAt, updatedAt *time.Time, status *int32, contact *entity.Contact, conversationID, contactID string) (*entity.ConversationMember, error) {
	if id != "" {
		id = strings.SplitN(id, ":", 2)[1]
	}
	if contact.ID != "" {
		contact.ID = strings.SplitN(contact.ID, ":", 2)[1]
	}
	if contact.Devices != nil && len(contact.Devices) != 0 {
		for i := range contact.Devices {
			contact.Devices[i].ID = strings.SplitN(contact.Devices[i].ID, ":", 2)[1]
		}
	}
	// if conversationID != "" {
	// 	conversationID = strings.SplitN(conversationID, ":", 2)[1]
	// }
	// if contactID != "" {
	// 	contactID = strings.SplitN(contactID, ":", 2)[1]
	// }
	return r.client.ConversationMember(ctx, &entity.ConversationMember{
		ID: id,
	})
}
func (r *queryResolver) ConversationLastEvent(ctx context.Context, id string) (*entity.Event, error) {
	if id != "" {
		id = strings.SplitN(id, ":", 2)[1]
	}
	// if conversationID != "" {
	// 	conversationID = strings.SplitN(conversationID, ":", 2)[1]
	// }
	// if contactID != "" {
	// 	contactID = strings.SplitN(contactID, ":", 2)[1]
	// }
	return r.client.ConversationLastEvent(ctx, &entity.Conversation{
		ID: id,
	})
}
func (r *queryResolver) DeviceInfos(ctx context.Context, T bool) (*deviceinfo.DeviceInfos, error) {
	return r.client.DeviceInfos(ctx, &node.Void{T: true})
}
func (r *queryResolver) AppVersion(ctx context.Context, T bool) (*node.AppVersionOutput, error) {
	return r.client.AppVersion(ctx, &node.Void{T: true})
}
func (r *queryResolver) TestPanic(ctx context.Context, T bool) (*node.Void, error) {
	return r.client.TestPanic(ctx, &node.Void{})
}
func (r *queryResolver) TestError(ctx context.Context, kind string) (*node.Void, error) {
	return r.client.TestError(ctx, &node.TestErrorInput{Kind: kind})
}
func (r *queryResolver) TestLogBackgroundError(ctx context.Context, T bool) (*node.Void, error) {
	return r.client.TestLogBackgroundError(ctx, &node.Void{})
}
func (r *queryResolver) TestLogBackgroundWarn(ctx context.Context, T bool) (*node.Void, error) {
	return r.client.TestLogBackgroundWarn(ctx, &node.Void{})
}
func (r *queryResolver) TestLogBackgroundDebug(ctx context.Context, T bool) (*node.Void, error) {
	return r.client.TestLogBackgroundDebug(ctx, &node.Void{})
}

func (r *queryResolver) Peers(ctx context.Context, _ bool) (*network_metric.Peers, error) {
	return r.client.Peers(ctx, &node.Void{})
}

func (r *queryResolver) Libp2PPing(ctx context.Context, str string) (*node.Bool, error) {
	return r.client.Libp2PPing(ctx, &network_metric.PingReq{Str: str})
}

func (r *queryResolver) GetListenAddrs(ctx context.Context, _ bool) (*network_metric.ListAddrs, error) {
	return r.client.GetListenAddrs(ctx, &node.Void{})
}

func (r *queryResolver) GetListenInterfaceAddrs(ctx context.Context, T bool) (*network_metric.ListAddrs, error) {
	return r.client.GetListenInterfaceAddrs(ctx, &node.Void{})
}

type subscriptionResolver struct{ *Resolver }

func (r *subscriptionResolver) CommitLogStream(ctx context.Context, t bool) (<-chan *node.CommitLog, error) {
	stream, err := r.client.CommitLogStream(ctx, &node.Void{T: t})
	channel := make(chan *node.CommitLog, 1)

	if err != nil {
		return nil, err
	}
	go func() {
		for {
			elem, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger().Error(err.Error())
				break
			}
			channel <- elem
		}
	}()
	return channel, nil
}

func (r *subscriptionResolver) EventStream(ctx context.Context, filter *entity.Event) (<-chan *entity.Event, error) {
	stream, err := r.client.EventStream(ctx, &node.EventStreamInput{Filter: filter})
	channel := make(chan *entity.Event, 1)

	if err != nil {
		return nil, err
	}
	go func() {
		for {
			elem, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger().Error(err.Error())
				break
			}
			channel <- elem
		}
	}()
	return channel, nil
}

func (r *subscriptionResolver) LogfileRead(ctx context.Context, path string) (<-chan *node.LogEntry, error) {
	stream, err := r.client.LogfileRead(ctx, &node.LogfileReadInput{
		Path: path,
	})
	if err != nil {
		return nil, err
	}

	channel := make(chan *node.LogEntry, 1)
	go func() {
		for {
			elem, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger().Error(err.Error())
				break
			}
			channel <- elem
		}
	}()
	return channel, nil
}

func (r *subscriptionResolver) LogStream(ctx context.Context, continues bool, logLevel, namespaces string, last int32) (<-chan *node.LogEntry, error) {
	stream, err := r.client.LogStream(ctx, &node.LogStreamInput{
		Continues:  continues,
		LogLevel:   logLevel,
		Namespaces: namespaces,
		Last:       last,
	})
	if err != nil {
		return nil, err
	}
	channel := make(chan *node.LogEntry, 1)
	go func() {
		for {
			elem, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger().Error(err.Error())
				break
			}
			channel <- elem
		}
	}()
	return channel, nil
}

func (r *subscriptionResolver) MonitorPeers(ctx context.Context, _ bool) (<-chan *network_metric.Peer, error) {
	stream, err := r.client.MonitorPeers(ctx, &node.Void{})
	if err != nil {
		return nil, err
	}

	channel := make(chan *network_metric.Peer, 10)
	go func() {
		for {
			elem, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger().Error(err.Error())
				break
			}
			channel <- elem
		}
	}()
	return channel, nil
}

func (r *queryResolver) DevicePushConfigList(ctx context.Context, void bool) (*node.DevicePushConfigListOutput, error) {
	return r.client.DevicePushConfigList(ctx, &node.Void{})
}

func (r *subscriptionResolver) MonitorBandwidth(ctx context.Context, id *string, _ *int64, _ *int64, _ *float64, _ *float64, mtype *int32) (<-chan *network_metric.BandwidthStats, error) {
	if mtype == nil {
		_mtype := int32(network_metric.MetricsType_GLOBAL)
		mtype = &_mtype
	}

	if id == nil {
		var _id string
		id = &_id
	}

	stream, err := r.client.MonitorBandwidth(ctx, &network_metric.BandwidthStats{
		ID:   *id,
		Type: network_metric.MetricsType(*mtype),
	})

	if err != nil {

		return nil, err
	}

	channel := make(chan *network_metric.BandwidthStats, 10)
	go func() {
		for {
			elem, err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger().Error(err.Error())
				break
			}
			channel <- elem
		}
	}()

	return channel, nil
}

// Helpers
func getPagination(first *int32, after *string, last *int32, before *string) *node.Pagination {
	pagination := &node.Pagination{}
	if first != nil {
		pagination.First = *first
	} else if last == nil {
		pagination.First = 10
	}
	if last != nil {
		pagination.Last = *last
	} else if first == nil {
		pagination.Last = 10
	}
	if after != nil {
		pagination.After = *after
	}
	if before != nil {
		pagination.Before = *before
	}
	return pagination
}
