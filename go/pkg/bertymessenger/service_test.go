package bertymessenger

import (
	"context"
	"testing"
	"time"

	"berty.tech/berty/v2/go/internal/testutil"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestServiceStream(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	node, cleanup := testingNode(ctx, t)
	defer cleanup()

	// first event is account update
	{
		event := node.NextEvent(t)
		require.Equal(t, event.Type, StreamEvent_TypeAccountUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		account := payload.(*StreamEvent_AccountUpdated).Account
		require.Equal(t, account, node.GetAccount())
		require.NotEmpty(t, account.Link)
		require.NotEmpty(t, account.PublicKey)
		require.Equal(t, account.State, Account_NotReady)
		require.Empty(t, account.DisplayName)
	}

	// second event is list end
	{
		event := node.NextEvent(t)
		require.Equal(t, event.Type, StreamEvent_TypeListEnd)
		require.Empty(t, event.Payload)
	}

	// no more event
	{
		event := node.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)
	}
}

func TestServiceSetName(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	node, cleanup := testingNode(ctx, t)
	defer cleanup()

	// set name before opening the stream
	node.SetName(t, "foo")

	// first event is account update
	{
		event := node.NextEvent(t)
		require.Equal(t, event.Type, StreamEvent_TypeAccountUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		account := payload.(*StreamEvent_AccountUpdated).Account
		require.Equal(t, account, node.GetAccount())
		require.NotEmpty(t, account.Link)
		require.NotEmpty(t, account.PublicKey)
		require.Equal(t, account.State, Account_Ready)
		require.Equal(t, account.DisplayName, "foo")
	}

	// second event is list end
	{
		event := node.NextEvent(t)
		require.Equal(t, event.Type, StreamEvent_TypeListEnd)
		require.Empty(t, event.Payload)
	}

	// no more event
	{
		event := node.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)
	}
}

func TestServiceSetNameAsync(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	node, cleanup := testingNode(ctx, t)
	defer cleanup()

	// first event is account update
	{
		event := node.NextEvent(t)
		require.Equal(t, event.Type, StreamEvent_TypeAccountUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		account := payload.(*StreamEvent_AccountUpdated).Account
		require.Equal(t, account, node.GetAccount())
		require.NotEmpty(t, account.Link)
		require.NotEmpty(t, account.PublicKey)
		require.Equal(t, account.State, Account_NotReady)
		require.Empty(t, account.DisplayName)
	}

	// second event is list end
	{
		event := node.NextEvent(t)
		require.Equal(t, event.Type, StreamEvent_TypeListEnd)
		require.Empty(t, event.Payload)
	}

	// set name after opening the stream
	previousAccount := node.GetAccount()
	node.SetName(t, "foo")

	// new account update event
	{
		event := node.NextEvent(t)
		require.Equal(t, event.Type, StreamEvent_TypeAccountUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		account := payload.(*StreamEvent_AccountUpdated).Account
		require.Equal(t, account, node.GetAccount())
		require.NotEmpty(t, account.Link)
		require.NotEmpty(t, account.PublicKey)
		require.Equal(t, account.State, Account_Ready)
		require.Equal(t, account.DisplayName, "foo")
		require.Equal(t, account.PublicKey, previousAccount.PublicKey)
		require.NotEqual(t, account.Link, previousAccount.Link)
	}

	// no more event
	{
		event := node.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)
	}
}

func TestServiceStreamCancel(t *testing.T) {
	ctx, ctxCancel := context.WithTimeout(context.Background(), 5*time.Second)
	node, cleanup := testingNode(ctx, t)
	defer cleanup()

	// first event is account update
	{
		event := node.NextEvent(t)
		require.Equal(t, event.Type, StreamEvent_TypeAccountUpdated)
	}

	// cancel
	ctxCancel()

	// second event fails
	{
		var err error
		for err == nil {
			_, err = node.GetStream(t).Recv()
		}
		require.True(t, isGRPCCanceledError(err))
	}
}

func TestServiceContactRequest(t *testing.T) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	node, cleanup := testingNode(ctx, t)
	defer cleanup()

	// drain init events
	node.DrainInitEvents(t)

	// send contact request
	const contactName = "zxxma-iphone"
	{
		link := "https://berty.tech/id#key=CiDXcXUOl1rpm2FcbOf3TFtn-FYkl_sOwA5run1LGXHOPRIg4xCLGP-BWzgIWRH0Vz9D8aGAq1kyno5Oqv6ysAljZmA&name=" + contactName
		ownMetadata := []byte("bar")
		metadata, err := proto.Marshal(&ContactMetadata{DisplayName: contactName})
		require.NoError(t, err)
		deeplinkReply, err := node.GetClient().ParseDeepLink(ctx, &ParseDeepLink_Request{Link: link})
		require.NoError(t, err)
		req := &SendContactRequest_Request{
			BertyID:     deeplinkReply.BertyID,
			Metadata:    metadata,
			OwnMetadata: ownMetadata,
		}
		_, err = node.GetClient().SendContactRequest(ctx, req)
		require.NoError(t, err)
		require.Len(t, node.contacts, 0)
		require.Len(t, node.conversations, 0)
	}

	// check for ContactUpdated event
	{
		event := node.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeContactUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		contact := payload.(*StreamEvent_ContactUpdated).Contact
		require.NotNil(t, contact)
		require.Equal(t, contact.GetDisplayName(), contactName)
		require.Equal(t, contact.GetState(), Contact_OutgoingRequestEnqueued)
		require.Len(t, node.contacts, 1)
		require.Len(t, node.conversations, 0)
	}

	// no more event
	{
		event := node.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)
	}
}

func TestServiceConversationCreateLive(t *testing.T) {
	testutil.SkipUnstable(t)

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	node, cleanup := testingNode(ctx, t)
	defer cleanup()

	// drain init events
	node.DrainInitEvents(t)

	// create conversation
	const conversationName = "Tasty"
	var createdConversationPK string
	{
		reply, err := node.GetClient().ConversationCreate(ctx, &ConversationCreate_Request{DisplayName: conversationName})
		require.NoError(t, err)
		require.NotEmpty(t, reply.GetPublicKey())
		createdConversationPK = reply.GetPublicKey()
	}

	// check for the first ConversationUpdated event (triggered on the join protocol event)
	{
		event := node.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeConversationUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		conversation := payload.(*StreamEvent_ConversationUpdated).Conversation
		require.NotNil(t, conversation)
		require.Equal(t, conversation.GetPublicKey(), createdConversationPK)
		require.Empty(t, conversation.GetDisplayName())
		require.Empty(t, conversation.GetLink())
	}

	// check for the second ConversationUpdated event (triggered when receiving the metadata with display name)
	{
		event := node.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeConversationUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		conversation := payload.(*StreamEvent_ConversationUpdated).Conversation
		require.NotNil(t, conversation)
		require.Equal(t, conversation.GetPublicKey(), createdConversationPK)
		require.Equal(t, conversation.GetDisplayName(), conversationName)
		require.NotEmpty(t, conversation.GetLink())
	}

	// no more event
	{
		event := node.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)
	}
}

func TestServiceConversationCreateAsync(t *testing.T) {
	testutil.SkipUnstable(t)

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	node, cleanup := testingNode(ctx, t)
	defer cleanup()

	// create conversation
	const conversationName = "Tasty"
	var createdConversationPK string
	{
		reply, err := node.GetClient().ConversationCreate(ctx, &ConversationCreate_Request{DisplayName: conversationName})
		require.NoError(t, err)
		require.NotEmpty(t, reply.GetPublicKey())
		createdConversationPK = reply.GetPublicKey()
	}

	// first event is account
	{
		event := node.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeAccountUpdated)
	}

	// second event is the conversation, with display name
	{
		event := node.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeConversationUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		conversation := payload.(*StreamEvent_ConversationUpdated).Conversation
		require.NotNil(t, conversation)
		require.Equal(t, conversation.GetPublicKey(), createdConversationPK)
		require.Equal(t, conversation.GetDisplayName(), conversationName)
		require.NotEmpty(t, conversation.GetLink())
	}

	// then, the list end event
	{
		event := node.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeListEnd)
	}

	// no more event
	{
		event := node.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)
	}
}

func Test1To1AddContact(t *testing.T) {
	testutil.SkipSlow(t)
	testutil.SkipUnstable(t)

	logger := testutil.Logger(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	clients, cleanup := testingInfra(ctx, t, 2, logger)
	defer cleanup()

	// Init accounts
	var (
		alice = NewTestingAccount(ctx, t, clients[0], logger)
		bob   = NewTestingAccount(ctx, t, clients[1], logger)
	)
	{
		defer alice.Close()
		alice.SetName(t, "Alice")
		alice.DrainInitEvents(t)
		require.NotEmpty(t, alice.GetAccount().GetLink())

		defer bob.Close()
		bob.SetName(t, "Bob")
		bob.DrainInitEvents(t)
		require.NotEmpty(t, bob.GetAccount().GetLink())
	}

	// Bob add Alice as contact (and she accepts)
	{
		testAddContact(ctx, t, bob, alice)
		require.Len(t, alice.contacts, 1)
		require.Len(t, bob.contacts, 1)
		// FIXME: should have 1 conversation
		require.Len(t, alice.conversations, 0)
		require.Len(t, bob.conversations, 0)
	}

	// no more event
	{
		event := alice.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)

		event = bob.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)
	}
}

func Test1To1Exchange(t *testing.T) {
	testutil.SkipSlow(t)
	testutil.SkipUnstable(t)

	logger := testutil.Logger(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	clients, cleanup := testingInfra(ctx, t, 2, logger)
	defer cleanup()

	// Init accounts
	var (
		alice = NewTestingAccount(ctx, t, clients[0], logger)
		bob   = NewTestingAccount(ctx, t, clients[1], logger)
	)
	{
		defer alice.Close()
		alice.SetName(t, "Alice")
		alice.DrainInitEvents(t)
		require.NotEmpty(t, alice.GetAccount().GetLink())
		require.Len(t, alice.contacts, 0)
		require.Len(t, alice.conversations, 0)

		defer bob.Close()
		bob.SetName(t, "Bob")
		bob.DrainInitEvents(t)
		require.NotEmpty(t, bob.GetAccount().GetLink())
		require.Len(t, bob.contacts, 0)
		require.Len(t, bob.conversations, 0)
	}

	// Bob add Alice as contact (and she accepts)
	var groupPK string
	{
		aliceContact := testAddContact(ctx, t, bob, alice)
		groupPK = aliceContact.GetConversationPublicKey()
	}

	// Exchange messages
	{
		testSendGroupMessage(ctx, t, groupPK, alice, []*TestingAccount{bob}, "Hello Bob!", logger)
		testSendGroupMessage(ctx, t, groupPK, bob, []*TestingAccount{alice}, "Hello Alice!", logger)
	}

	// no more event
	{
		event := alice.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)

		event = bob.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)
	}
}

func Test3PeersCreateJoinConversation(t *testing.T) {
	testutil.SkipSlow(t)
	testutil.SkipUnstable(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	logger := testutil.Logger(t)
	clients, cleanup := testingInfra(ctx, t, 3, logger)
	defer cleanup()

	// create nodes
	var creator *TestingAccount
	{
		creator = NewTestingAccount(ctx, t, clients[0], logger)
		defer creator.Close()
		creator.DrainInitEvents(t)
	}
	var joiners = make([]*TestingAccount, 2)
	{
		for i := 0; i < 2; i++ {
			joiners[i] = NewTestingAccount(ctx, t, clients[i+1], logger)
			defer joiners[i].Close()
			joiners[i].DrainInitEvents(t)
		}
	}

	// creator creates a new conversation
	createdConv := testCreateConversation(ctx, t, creator, "My Group", nil, logger)

	// joiners join the conversation
	for _, joiner := range joiners {
		testJoinConversation(ctx, t, joiner, createdConv, logger)
	}

	// FIXME: check that member names are propagated

	// no more event
	{
		event := creator.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)
		for _, joiner := range joiners {
			event = joiner.TryNextEvent(t, 100*time.Millisecond)
			require.Nil(t, event)
		}
	}
}

func Test3PeersExchange(t *testing.T) {
	testutil.SkipSlow(t)
	testutil.SkipUnstable(t)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	logger := testutil.Logger(t)
	clients, cleanup := testingInfra(ctx, t, 3, logger)
	defer cleanup()

	// create nodes
	var creator *TestingAccount
	{
		creator = NewTestingAccount(ctx, t, clients[0], logger)
		defer creator.Close()
		creator.DrainInitEvents(t)
	}
	var joiners = make([]*TestingAccount, 2)
	{
		for i := 0; i < 2; i++ {
			joiners[i] = NewTestingAccount(ctx, t, clients[i+1], logger)
			defer joiners[i].Close()
			joiners[i].DrainInitEvents(t)
		}
	}

	// creator creates a new conversation
	createdConv := testCreateConversation(ctx, t, creator, "My Group", nil, logger)

	// joiners join the conversation
	for _, joiner := range joiners {
		testJoinConversation(ctx, t, joiner, createdConv, logger)
	}

	// FIXME: replace by a check
	time.Sleep(5 * time.Second)

	// interact
	{
		testSendGroupMessage(ctx, t, createdConv.GetPublicKey(), creator, []*TestingAccount{joiners[0], joiners[1]}, "Hello Group! (creator)", logger)
		testSendGroupMessage(ctx, t, createdConv.GetPublicKey(), joiners[0], []*TestingAccount{creator, joiners[1]}, "Hello Group! (joiner1)", logger)
		testSendGroupMessage(ctx, t, createdConv.GetPublicKey(), joiners[1], []*TestingAccount{creator, joiners[0]}, "Hello Group! (joiner2)", logger)
	}

	// no more event
	{
		event := creator.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)
		for _, joiner := range joiners {
			event = joiner.TryNextEvent(t, 100*time.Millisecond)
			require.Nil(t, event)
		}
	}
}

func TestConversationInvitation(t *testing.T) {
	testutil.SkipUnstable(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	logger := testutil.Logger(t)
	clients, cleanup := testingInfra(ctx, t, 3, logger)
	defer cleanup()

	// create nodes
	var alice, bob, john *TestingAccount
	{
		alice = NewTestingAccount(ctx, t, clients[0], logger)
		defer alice.Close()
		alice.SetName(t, "Alice")
		alice.DrainInitEvents(t)

		bob = NewTestingAccount(ctx, t, clients[1], logger)
		defer bob.Close()
		bob.SetName(t, "Bob")
		bob.DrainInitEvents(t)

		john = NewTestingAccount(ctx, t, clients[2], logger)
		defer john.Close()
		john.SetName(t, "John")
		john.DrainInitEvents(t)
	}

	// contact requests
	{
		testAddContact(ctx, t, alice, bob)
		testAddContact(ctx, t, alice, john)
		testAddContact(ctx, t, bob, john)
		require.Len(t, alice.contacts, 2)
		require.Len(t, bob.contacts, 2)
		require.Len(t, john.contacts, 2)
		require.Len(t, alice.conversations, 0)
		require.Len(t, bob.conversations, 0)
		require.Len(t, john.conversations, 0)
	}

	// create group
	{
		testCreateConversation(ctx, t, alice, "Alice & Friends", []*TestingAccount{bob, john}, logger)
		require.Len(t, alice.contacts, 2)
		require.Len(t, bob.contacts, 2)
		require.Len(t, john.contacts, 2)
		require.Len(t, alice.conversations, 1)
		require.Len(t, bob.conversations, 1)
		require.Len(t, john.conversations, 1)
	}

	// no more event
	{
		event := alice.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)

		event = bob.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)

		event = john.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)
	}
}

func TestConversationInvitationAndExchange(t *testing.T) {
	testutil.SkipSlow(t)
	testutil.SkipUnstable(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	logger := testutil.Logger(t)
	clients, cleanup := testingInfra(ctx, t, 3, logger)
	defer cleanup()

	// create nodes
	var alice, bob, john *TestingAccount
	{
		alice = NewTestingAccount(ctx, t, clients[0], logger)
		defer alice.Close()
		alice.SetName(t, "Alice")
		alice.DrainInitEvents(t)

		bob = NewTestingAccount(ctx, t, clients[1], logger)
		defer bob.Close()
		bob.SetName(t, "Bob")
		bob.DrainInitEvents(t)

		john = NewTestingAccount(ctx, t, clients[2], logger)
		defer john.Close()
		john.SetName(t, "John")
		john.DrainInitEvents(t)
	}

	// contact requests
	{
		testAddContact(ctx, t, alice, bob)
		testAddContact(ctx, t, alice, john)
		testAddContact(ctx, t, bob, john)
		require.Len(t, alice.contacts, 2)
		require.Len(t, bob.contacts, 2)
		require.Len(t, john.contacts, 2)
		require.Len(t, alice.conversations, 0)
		require.Len(t, bob.conversations, 0)
		require.Len(t, john.conversations, 0)
	}

	// create group
	var createdConv *Conversation
	{
		createdConv = testCreateConversation(ctx, t, alice, "Alice & Friends", []*TestingAccount{bob, john}, logger)
		require.Len(t, alice.contacts, 2)
		require.Len(t, bob.contacts, 2)
		require.Len(t, john.contacts, 2)
		require.Len(t, alice.conversations, 1)
		require.Len(t, bob.conversations, 1)
		require.Len(t, john.conversations, 1)
	}

	// FIXME: replace by a check
	time.Sleep(5 * time.Second)

	// interact
	{
		testSendGroupMessage(ctx, t, createdConv.GetPublicKey(), alice, []*TestingAccount{bob, john}, "Hello Group! (alice)", logger)
		testSendGroupMessage(ctx, t, createdConv.GetPublicKey(), bob, []*TestingAccount{alice, john}, "Hello Group! (bob)", logger)
		testSendGroupMessage(ctx, t, createdConv.GetPublicKey(), john, []*TestingAccount{alice, bob}, "Hello Group! (john)", logger)
	}

	// no more event
	{
		event := alice.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)

		event = bob.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)

		event = john.TryNextEvent(t, 100*time.Millisecond)
		require.Nil(t, event)
	}
}

func testJoinConversation(ctx context.Context, t *testing.T, joiner *TestingAccount, existingConv *Conversation, logger *zap.Logger) {
	t.Helper()

	// joiner joins the conversation
	{
		ret, err := joiner.GetClient().ConversationJoin(ctx, &ConversationJoin_Request{Link: existingConv.GetLink()})
		require.NoError(t, err)
		require.Empty(t, ret)
		logger.Debug("testJoinConversation: conversation joined")
	}

	// joiner has ConversationUpdated event
	{
		event := joiner.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeConversationUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		conversation := payload.(*StreamEvent_ConversationUpdated).GetConversation()
		if existingConv.GetPublicKey() != "" {
			require.Equal(t, conversation.GetPublicKey(), existingConv.GetPublicKey())
		}
		if existingConv.GetDisplayName() != "" {
			require.Equal(t, conversation.GetDisplayName(), existingConv.GetDisplayName())
		}
		require.Equal(t, conversation.GetLink(), existingConv.GetLink())
		logger.Debug("testJoinConversation: conversation joined confirmation received")
	}
}

func testAddContact(ctx context.Context, t *testing.T, requester, requested *TestingAccount) *Contact {
	t.Helper()
	// Requester sends a contact request to requested
	{
		ret, err := requester.GetClient().ContactRequest(ctx, &ContactRequest_Request{Link: requested.GetAccount().GetLink()})
		require.NoError(t, err)
		require.Empty(t, ret)
	}

	// Requester has a contact updated event (outgoing request enqueued)
	{
		event := requester.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeContactUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		contact := payload.(*StreamEvent_ContactUpdated).Contact
		require.NotEmpty(t, contact.GetPublicKey())
		require.Equal(t, contact.GetPublicKey(), requested.GetAccount().GetPublicKey())
		require.Equal(t, contact.GetDisplayName(), requested.GetAccount().GetDisplayName())
		require.Equal(t, contact.GetState(), Contact_OutgoingRequestEnqueued)
		require.Empty(t, contact.GetConversationPublicKey())
	}

	// Requester has a contact updated event (outgoing request sent)
	{
		event := requester.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeContactUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		contact := payload.(*StreamEvent_ContactUpdated).Contact
		require.NotEmpty(t, contact.GetPublicKey())
		require.Equal(t, contact.GetPublicKey(), requested.GetAccount().GetPublicKey())
		require.Equal(t, contact.GetDisplayName(), requested.GetAccount().GetDisplayName())
		require.Equal(t, contact.GetState(), Contact_OutgoingRequestSent)
		require.Empty(t, contact.GetConversationPublicKey())
	}

	// Requested receives the contact request
	{
		event := requested.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeContactUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		contact := payload.(*StreamEvent_ContactUpdated).Contact
		require.NotEmpty(t, contact.GetPublicKey())
		require.Equal(t, contact.GetPublicKey(), requester.GetAccount().GetPublicKey())
		require.Equal(t, contact.GetDisplayName(), requester.GetAccount().GetDisplayName())
		require.Equal(t, contact.GetState(), Contact_IncomingRequest)
	}

	// Requested accepts the contact request
	{
		ret, err := requested.GetClient().ContactAccept(ctx, &ContactAccept_Request{PublicKey: requester.GetAccount().GetPublicKey()})
		require.NoError(t, err)
		require.Empty(t, ret)
	}

	// Requested receives the contact request
	var groupPK string
	{
		event := requested.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeContactUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		contact := payload.(*StreamEvent_ContactUpdated).Contact
		require.NotEmpty(t, contact.GetPublicKey())
		require.Equal(t, contact.GetPublicKey(), requester.GetAccount().GetPublicKey())
		require.Equal(t, contact.GetDisplayName(), requester.GetAccount().GetDisplayName())
		require.Equal(t, contact.GetState(), Contact_Established)
		groupPK = contact.GetConversationPublicKey()
	}

	// FIXME: should also have a conversation created event

	// Requester has a contact updated event (Established)
	{
		event := requester.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeContactUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		contact := payload.(*StreamEvent_ContactUpdated).Contact
		require.NotEmpty(t, contact.GetPublicKey())
		require.Equal(t, contact.GetPublicKey(), requested.GetAccount().GetPublicKey())
		require.Equal(t, contact.GetDisplayName(), requested.GetAccount().GetDisplayName())
		require.Equal(t, contact.GetState(), Contact_Established)
		require.Equal(t, contact.GetConversationPublicKey(), groupPK)
		return contact
	}
}

func testSendGroupMessage(ctx context.Context, t *testing.T, groupPK string, sender *TestingAccount, receivers []*TestingAccount, msg string, logger *zap.Logger) {
	t.Helper()

	// sender interacts
	var beforeSend, afterSend int64
	{
		beforeSend = timestampMs(time.Now())
		userMessage, err := proto.Marshal(&AppMessage_UserMessage{Body: msg})
		require.NoError(t, err)
		interactionRequest := Interact_Request{Type: AppMessage_TypeUserMessage, Payload: userMessage, ConversationPublicKey: groupPK}
		_, err = sender.GetClient().Interact(ctx, &interactionRequest)
		require.NoError(t, err)
		afterSend = timestampMs(time.Now())
		logger.Debug("testSendGroupMessage: message sent")
	}

	// sender has own interact event
	var messageCid string
	{
		event := sender.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeInteractionUpdated)
		eventPayload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		interaction := eventPayload.(*StreamEvent_InteractionUpdated).Interaction
		require.NotEmpty(t, interaction.GetCID())
		messageCid = interaction.GetCID()
		require.Equal(t, interaction.GetType(), AppMessage_TypeUserMessage)
		require.Equal(t, interaction.GetConversationPublicKey(), groupPK)
		require.True(t, interaction.GetIsMe())
		require.Equal(t, interaction.GetCID(), messageCid)
		interactionPayload, err := interaction.UnmarshalPayload()
		require.NoError(t, err)
		userMessage := interactionPayload.(*AppMessage_UserMessage)
		require.Equal(t, userMessage.GetBody(), msg)
		require.LessOrEqual(t, beforeSend, userMessage.GetSentDate())
		require.LessOrEqual(t, userMessage.GetSentDate(), afterSend)
		logger.Debug("testSendGroupMessage: message received by creator")
	}

	// sender has the ack event too
	for i := 0; i < len(receivers); i++ {
		event := sender.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeInteractionUpdated)
		eventPayload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		interaction := eventPayload.(*StreamEvent_InteractionUpdated).Interaction
		require.NotEmpty(t, interaction.GetCID())
		require.Equal(t, interaction.GetType(), AppMessage_TypeAcknowledge)
		require.Equal(t, interaction.GetConversationPublicKey(), groupPK)
		require.False(t, interaction.GetIsMe())
		interactionPayload, err := interaction.UnmarshalPayload()
		require.NoError(t, err)
		ack := interactionPayload.(*AppMessage_Acknowledge)
		require.Equal(t, ack.GetTarget(), messageCid)
		logger.Debug("testSendGroupMessage: message ack received by creator")
		// FIXME: check if the ack is from the good receiver, or useless?
	}

	for _, receiver := range receivers {
		gotOwnAck := false
		gotOthersAcks := 0
		gotMsg := false
		// we should receive one message + one ack per receiver
		for i := 0; i < len(receivers)+1; i++ {
			event := receiver.NextEvent(t)
			require.Equal(t, event.GetType(), StreamEvent_TypeInteractionUpdated)
			eventPayload, err := event.UnmarshalPayload()
			require.NoError(t, err)
			interaction := eventPayload.(*StreamEvent_InteractionUpdated).Interaction
			require.NotEmpty(t, interaction.GetCID())
			require.Equal(t, interaction.GetConversationPublicKey(), groupPK)
			interactionPayload, err := interaction.UnmarshalPayload()
			require.NoError(t, err)
			switch {
			case interaction.GetType() == AppMessage_TypeAcknowledge && interaction.GetIsMe():
				require.False(t, gotOwnAck)
				gotOwnAck = true
				ack := interactionPayload.(*AppMessage_Acknowledge)
				require.Equal(t, ack.GetTarget(), messageCid)
			case interaction.GetType() == AppMessage_TypeAcknowledge && !interaction.GetIsMe():
				ack := interactionPayload.(*AppMessage_Acknowledge)
				require.Equal(t, ack.GetTarget(), messageCid)
				gotOthersAcks++
			case interaction.GetType() == AppMessage_TypeUserMessage:
				require.False(t, gotMsg)
				gotMsg = true
				require.Equal(t, interaction.GetCID(), messageCid)
				userMessage := interactionPayload.(*AppMessage_UserMessage)
				require.Equal(t, userMessage.GetBody(), msg)
				require.LessOrEqual(t, beforeSend, userMessage.GetSentDate())
				require.LessOrEqual(t, userMessage.GetSentDate(), afterSend)
			}
		}
		require.True(t, gotOwnAck)
		require.True(t, gotMsg)
		require.Equal(t, gotOthersAcks, len(receivers)-1)
	}
}

func testCreateConversation(ctx context.Context, t *testing.T, creator *TestingAccount, convName string, invitees []*TestingAccount, logger *zap.Logger) *Conversation {
	t.Helper()

	// creator creates a conversation
	var convPK string
	{
		contactsToInvite := make([]string, len(invitees))
		for idx, invitee := range invitees {
			contactsToInvite[idx] = invitee.GetAccount().GetPublicKey()
		}
		createdConv, err := creator.GetClient().ConversationCreate(ctx, &ConversationCreate_Request{DisplayName: convName, ContactsToInvite: contactsToInvite})
		require.NoError(t, err)
		require.NotEmpty(t, createdConv.GetPublicKey())
		convPK = createdConv.GetPublicKey()
	}

	// creator has a ConversationUpdated event for the conversation creation
	{
		event := creator.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeConversationUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		conversation := payload.(*StreamEvent_ConversationUpdated).GetConversation()
		require.Equal(t, convPK, conversation.GetPublicKey())
		require.Equal(t, conversation.GetDisplayName(), "")
	}

	// creator has a ConversationUpdated event for the display name
	var createdConv *Conversation
	{
		event := creator.NextEvent(t)
		require.Equal(t, event.GetType(), StreamEvent_TypeConversationUpdated)
		payload, err := event.UnmarshalPayload()
		require.NoError(t, err)
		conversation := payload.(*StreamEvent_ConversationUpdated).GetConversation()
		require.Equal(t, convPK, conversation.GetPublicKey())
		require.Equal(t, conversation.GetDisplayName(), convName)
		createdConv = conversation
	}

	for _, invitee := range invitees {
		// creator see the invitation in 1-1 conv
		{
			event := creator.NextEvent(t)
			require.Equal(t, event.GetType(), StreamEvent_TypeInteractionUpdated)
			eventPayload, err := event.UnmarshalPayload()
			require.NoError(t, err)
			interaction := eventPayload.(*StreamEvent_InteractionUpdated).GetInteraction()
			require.Equal(t, interaction.GetType(), AppMessage_TypeGroupInvitation)
			require.NotEmpty(t, interaction.GetCID())
			require.NotEqual(t, convPK, interaction.GetConversationPublicKey())
			require.True(t, interaction.GetIsMe())
			// FIXME: require.Equal, 1to1conv.pk
			interactionPayload, err := interaction.UnmarshalPayload()
			require.NoError(t, err)
			inviteLink := interactionPayload.(*AppMessage_GroupInvitation).GetLink()
			require.NotEmpty(t, inviteLink)

		}

		// invitee receive the invitation
		var inviteLink string
		{
			event := invitee.NextEvent(t)
			require.Equal(t, event.GetType(), StreamEvent_TypeInteractionUpdated)
			eventPayload, err := event.UnmarshalPayload()
			require.NoError(t, err)
			interaction := eventPayload.(*StreamEvent_InteractionUpdated).GetInteraction()
			require.Equal(t, interaction.GetType(), AppMessage_TypeGroupInvitation)
			require.NotEmpty(t, interaction.GetCID())
			require.NotEqual(t, convPK, interaction.GetConversationPublicKey())
			require.False(t, interaction.GetIsMe())
			// FIXME: require.Equal, 1to1conv.pk
			interactionPayload, err := interaction.UnmarshalPayload()
			require.NoError(t, err)
			inviteLink = interactionPayload.(*AppMessage_GroupInvitation).GetLink()
			require.NotEmpty(t, inviteLink)
		}

		// invitee accepts the invitation
		{
			conversation := &Conversation{
				Link: inviteLink,
				// FIXME: parse the link to get the name and public key (bonus)
			}
			testJoinConversation(ctx, t, invitee, conversation, logger)
		}
	}

	return createdConv
}
