package storemember

import (
	"context"

	"berty.tech/berty/go/internal/group"
	"berty.tech/berty/go/internal/orbitutil/orbitutilapi"
	"berty.tech/berty/go/internal/orbitutil/storegroup"
	"berty.tech/berty/go/pkg/errcode"
	"berty.tech/go-ipfs-log/identityprovider"
	"berty.tech/go-orbit-db/address"
	"berty.tech/go-orbit-db/iface"
	"berty.tech/go-orbit-db/stores/operation"
	coreapi "github.com/ipfs/interface-go-ipfs-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pkg/errors"
)

const StoreType = "member_store"

type memberStore struct {
	storegroup.BaseGroupStore
}

func (m *memberStore) ListMembers() ([]*group.MemberDevice, error) {
	values, ok := m.Index().Get("").([]*group.MemberDevice)
	if !ok {
		return nil, errors.New("unable to cast entries")
	}

	return values, nil
}

func (m *memberStore) RedeemInvitation(ctx context.Context, memberPrivateKey, devicePrivateKey crypto.PrivKey, invitation *group.Invitation) (operation.Operation, error) {
	payload, err := group.NewMemberEntryPayload(memberPrivateKey, devicePrivateKey, invitation)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	env, err := group.SealStorePayload(payload, m.GetGroupContext().GetGroup(), devicePrivateKey)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	op := operation.NewOperation(nil, "ADD", env)

	e, err := m.AddOperation(ctx, op, nil)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	op, err = operation.ParseOperation(e)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	return op, nil
}

func ConstructorFactory(s orbitutilapi.BertyOrbitDB) iface.StoreConstructor {
	return func(ctx context.Context, ipfs coreapi.CoreAPI, identity *identityprovider.Identity, addr address.Address, options *iface.NewStoreOptions) (iface.Store, error) {
		store := &memberStore{}
		if err := s.InitGroupStore(ctx, NewMemberStoreIndex, store, ipfs, identity, addr, options); err != nil {
			return nil, errors.Wrap(err, "unable to initialize base store")
		}

		return store, nil
	}
}

var _ orbitutilapi.MemberStore = (*memberStore)(nil)
