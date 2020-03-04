package orbitutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"berty.tech/berty/go/pkg/bertyprotocol"
	"github.com/stretchr/testify/assert"
)

func countEntries(t testing.TB, out <-chan *bertyprotocol.GroupMessageEvent, expected int) {
	found := 0

	for range out {
		found++
	}

	assert.Equal(t, expected, found)
}

func Test_AddMessage_ListMessages_manually_supplying_secrets(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	memberCount := 2
	deviceCount := 1
	entriesCount := 25

	testMsg1 := []byte("first message")

	peers, _ := CreatePeersWithGroup(ctx, t, "/tmp/message_test", memberCount, deviceCount, true)
	defer DropPeers(t, peers)

	dPK0 := peers[0].GetGroupContext().GetDevicePrivKey().GetPublic()
	ds0, err := peers[0].GetGroupContext().GetMessageKeysHolder().GetOwnDeviceChainKey(ctx)
	assert.NoError(t, err)

	err = RegisterChainKeyForDevice(ctx, peers[1].GetGroupContext().GetMessageKeysHolder(), dPK0, ds0)
	assert.NoError(t, err)

	_, err = peers[0].GetGroupContext().GetMessageStore().AddMessage(ctx, testMsg1)
	assert.NoError(t, err)

	<-time.After(time.Second)

	out, err := peers[0].GetGroupContext().GetMessageStore().ListMessages(ctx)
	assert.NoError(t, err)

	countEntries(t, out, 1)

	for i := 0; i < entriesCount; i++ {
		payload := []byte(fmt.Sprintf("test message %d", i))
		_, err = peers[0].GetGroupContext().GetMessageStore().AddMessage(ctx, payload)
		assert.NoError(t, err)
	}

	<-time.After(time.Second * 2)

	out, err = peers[1].GetGroupContext().GetMessageStore().ListMessages(ctx)
	assert.NoError(t, err)

	<-time.After(time.Second)

	countEntries(t, out, entriesCount+1)

	// TODO: check that ListMessages can be called multiple times with the same output
	// TODO: check that message are correctly ordered
	// TODO: check that message are correctly decrypted
	// TODO: check that message sender is correct
	// TODO: check that message parents IDs are valid
	// TODO: check that message IDs are valid
}
