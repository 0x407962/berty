package orbitutil

import (
	"context"
	"sync"

	ipfslog "berty.tech/go-ipfs-log"
	"berty.tech/go-orbit-db/events"
	"berty.tech/go-orbit-db/iface"
	"berty.tech/go-orbit-db/stores/operation"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/crypto"

	"berty.tech/berty/go/internal/account"
	"berty.tech/berty/go/pkg/bertyprotocol"
	"berty.tech/berty/go/pkg/errcode"
)

type metadataStoreIndex struct {
	members         map[string][]*account.MemberDevice
	devices         map[string]*account.MemberDevice
	admins          map[crypto.PubKey]struct{}
	handledEvents   map[string]struct{}
	sentSecrets     map[string]struct{}
	g               *bertyprotocol.Group
	ownMemberDevice *account.MemberDevice
	ctx             context.Context
	eventEmitter    events.EmitterInterface
	lock            sync.RWMutex
}

func (m *metadataStoreIndex) Get(key string) interface{} {
	return nil
}

func openMetadataEntry(g *bertyprotocol.Group, log ipfslog.Log, e ipfslog.Entry) (*bertyprotocol.GroupMetadataEvent, *bertyprotocol.GroupMetadata, proto.Message, error) {
	op, err := operation.ParseOperation(e)
	if err != nil {
		// TODO: log
		return nil, nil, nil, err
	}

	meta, event, err := bertyprotocol.OpenGroupEnvelope(g, op.GetValue())
	if err != nil {
		// TODO: log
		return nil, nil, nil, err
	}

	metaEvent, err := bertyprotocol.NewGroupMetadataEventFromEntry(log, e, meta, event, g)
	if err != nil {
		// TODO: log
		return nil, nil, nil, err
	}

	return metaEvent, meta, event, nil
}

func (m *metadataStoreIndex) UpdateIndex(log ipfslog.Log, entries []ipfslog.Entry) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	for _, e := range log.Values().Slice() {
		metaEvent, meta, event, err := openMetadataEntry(m.g, log, e)
		if err != nil {
			// TODO: log
			continue
		}

		if _, ok := m.handledEvents[e.GetHash().String()]; ok {
			continue
		}

		m.handledEvents[e.GetHash().String()] = struct{}{}

		switch meta.EventType {
		case bertyprotocol.EventTypeMultiMemberGroupInitialMemberAnnounced:
			err = m.handleMultiMemberInitialMember(event)
		case bertyprotocol.EventTypeGroupMemberDeviceAdded:
			err = m.handleGroupAddMemberDevice(event)
		case bertyprotocol.EventTypeGroupDeviceSecretAdded:
			err = m.handleGroupAddDeviceSecret(event)
		case bertyprotocol.EventTypeMultiMemberGroupAdminRoleGranted:
			err = m.handleMultiMemberGrantAdminRole(event)
		default:
			err = errcode.ErrNotImplemented
		}

		if err != nil {
			// TODO: log
			_ = err
			continue
		}

		m.eventEmitter.Emit(m.ctx, metaEvent)
	}

	return nil
}

func (m *metadataStoreIndex) handleMultiMemberInitialMember(event proto.Message) error {
	e, ok := event.(*bertyprotocol.MultiMemberInitialMember)
	if !ok {
		return errcode.ErrInvalidInput
	}

	pk, err := crypto.UnmarshalEd25519PublicKey(e.MemberPK)
	if err != nil {
		return errcode.ErrDeserialization.Wrap(err)
	}

	if _, ok := m.admins[pk]; ok {
		return errcode.ErrInternal
	}

	m.admins[pk] = struct{}{}

	return nil
}

func (m *metadataStoreIndex) handleGroupAddMemberDevice(event proto.Message) error {
	e, ok := event.(*bertyprotocol.GroupAddMemberDevice)
	if !ok {
		return errcode.ErrInvalidInput
	}

	member, err := crypto.UnmarshalEd25519PublicKey(e.MemberPK)
	if err != nil {
		return errcode.ErrDeserialization.Wrap(err)
	}

	device, err := crypto.UnmarshalEd25519PublicKey(e.DevicePK)
	if err != nil {
		return errcode.ErrDeserialization.Wrap(err)
	}

	if _, ok := m.devices[string(e.DevicePK)]; ok {
		return nil
	}

	if _, ok := m.devices[string(e.DevicePK)]; ok {
		return errcode.ErrInternal
	}

	m.devices[string(e.DevicePK)] = &account.MemberDevice{
		Member: member,
		Device: device,
	}

	m.members[string(e.MemberPK)] = append(m.members[string(e.MemberPK)], &account.MemberDevice{
		Member: member,
		Device: device,
	})

	return nil
}

func (m *metadataStoreIndex) handleGroupAddDeviceSecret(event proto.Message) error {
	e, ok := event.(*bertyprotocol.GroupAddDeviceSecret)
	if !ok {
		return errcode.ErrInvalidInput
	}

	destPK, err := crypto.UnmarshalEd25519PublicKey(e.DestMemberPK)
	if err != nil {
		return errcode.ErrDeserialization.Wrap(err)
	}

	senderPK, err := crypto.UnmarshalEd25519PublicKey(e.DevicePK)
	if err != nil {
		return errcode.ErrDeserialization.Wrap(err)
	}

	if m.ownMemberDevice.Device.Equals(senderPK) {
		m.sentSecrets[string(e.DestMemberPK)] = struct{}{}
	}

	if !destPK.Equals(m.ownMemberDevice.Member) {
		return errcode.ErrGroupSecretOtherDestMember
	}

	return nil
}

func (m *metadataStoreIndex) handleMultiMemberGrantAdminRole(event proto.Message) error {
	// TODO:

	return nil
}

func (m *metadataStoreIndex) GetMemberByDevice(pk crypto.PubKey) (crypto.PubKey, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	id, err := pk.Raw()
	if err != nil {
		return nil, errcode.ErrInvalidInput.Wrap(err)
	}

	device, ok := m.devices[string(id)]
	if !ok {
		return nil, errcode.ErrMissingInput
	}

	return device.Member, nil
}

func (m *metadataStoreIndex) GetDevicesForMember(pk crypto.PubKey) ([]crypto.PubKey, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	id, err := pk.Raw()
	if err != nil {
		return nil, errcode.ErrInvalidInput.Wrap(err)
	}

	mds, ok := m.members[string(id)]
	if !ok {
		return nil, errcode.ErrInvalidInput
	}

	ret := make([]crypto.PubKey, len(mds))
	for i, md := range mds {
		ret[i] = md.Device
	}

	return ret, nil
}

func (m *metadataStoreIndex) MemberCount() int {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return len(m.members)
}

func (m *metadataStoreIndex) DeviceCount() int {
	m.lock.RLock()
	defer m.lock.RUnlock()

	return len(m.devices)
}

func (m *metadataStoreIndex) ListMembers() []crypto.PubKey {
	m.lock.RLock()
	defer m.lock.RUnlock()

	members := make([]crypto.PubKey, len(m.members))
	i := 0

	for _, md := range m.members {
		members[i] = md[0].Member
		i++
	}

	return members
}

func (m *metadataStoreIndex) ListDevices() []crypto.PubKey {
	m.lock.RLock()
	defer m.lock.RUnlock()

	devices := make([]crypto.PubKey, len(m.devices))
	i := 0

	for _, md := range m.devices {
		devices[i] = md.Device
		i++
	}

	return devices
}

func (m *metadataStoreIndex) ListAdmins() []crypto.PubKey {
	m.lock.RLock()
	defer m.lock.RUnlock()

	admins := make([]crypto.PubKey, len(m.admins))
	i := 0

	for admin := range m.admins {
		admins[i] = admin
		i++
	}

	return admins
}

func (m *metadataStoreIndex) AreSecretsAlreadySent(pk crypto.PubKey) (bool, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	key, err := pk.Raw()
	if err != nil {
		return false, errcode.ErrInvalidInput.Wrap(err)
	}

	_, ok := m.sentSecrets[string(key)]
	return ok, nil
}

// NewMetadataStoreIndex returns a new index to manage the list of the group members
func NewMetadataIndex(ctx context.Context, eventEmitter events.EmitterInterface, g *bertyprotocol.Group, memberDevice *account.MemberDevice) iface.IndexConstructor {
	return func(publicKey []byte) iface.StoreIndex {
		return &metadataStoreIndex{
			members:         map[string][]*account.MemberDevice{},
			devices:         map[string]*account.MemberDevice{},
			admins:          map[crypto.PubKey]struct{}{},
			sentSecrets:     map[string]struct{}{},
			handledEvents:   map[string]struct{}{},
			g:               g,
			eventEmitter:    eventEmitter,
			ownMemberDevice: memberDevice,
			ctx:             ctx,
		}
	}
}

var _ iface.StoreIndex = &metadataStoreIndex{}
