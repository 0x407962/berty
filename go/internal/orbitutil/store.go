package orbitutil

import (
	"sync"

	"berty.tech/berty/go/pkg/errcode"
	ipfslog "berty.tech/go-ipfs-log"
	"berty.tech/go-orbit-db/stores/basestore"
	"berty.tech/go-orbit-db/stores/operation"
)

type BaseGroupStore struct {
	basestore.BaseStore

	groupContext GroupContext
	lock         sync.RWMutex
}

func (b *BaseGroupStore) SetGroupContext(g GroupContext) {
	b.lock.Lock()
	b.groupContext = g
	b.lock.Unlock()
}

func (b *BaseGroupStore) GetGroupContext() GroupContext {
	b.lock.RLock()
	defer b.lock.RUnlock()

	return b.groupContext
}

func UnwrapOperation(opEntry ipfslog.Entry) ([]byte, error) {
	entry, ok := opEntry.(ipfslog.Entry)
	if !ok {
		return nil, errcode.ErrInvalidInput
	}

	op, err := operation.ParseOperation(entry)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	return op.GetValue(), nil
}

var _ GroupStore = (*BaseGroupStore)(nil)
