package initutil

import (
	"flag"
	"fmt"
	"os"
	"path"

	badger_opts "github.com/dgraph-io/badger/options"
	datastore "github.com/ipfs/go-datastore"
	sync_ds "github.com/ipfs/go-datastore/sync"
	badger "github.com/ipfs/go-ds-badger"
	"go.uber.org/zap"

	"berty.tech/berty/v2/go/pkg/errcode"
)

const InMemoryDir = ":memory:"

func (m *Manager) SetupDatastoreFlags(fs *flag.FlagSet) {
	dir := m.Datastore.Dir
	if dir == "" {
		dir = m.Datastore.defaultDir
	}
	fs.StringVar(&m.Datastore.Dir, "store.dir", dir, "root datastore directory")
	fs.BoolVar(&m.Datastore.InMemory, "store.inmem", m.Datastore.InMemory, "disable datastore persistence")
	fs.BoolVar(&m.Datastore.FileIO, "store.fileio", m.Datastore.FileIO, "enable FileIO Option, files will be loaded using standard I/O")
}

func (m *Manager) GetDatastoreDir() (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.getDatastoreDir()
}

func (m *Manager) getDatastoreDir() (string, error) {
	m.applyDefaults()

	if m.Datastore.dir != "" {
		return m.Datastore.dir, nil
	}
	switch {
	case m.Datastore.Dir == "" && !m.Datastore.InMemory:
		return "", errcode.TODO.Wrap(fmt.Errorf("--store.dir is empty"))
	case m.Datastore.Dir == InMemoryDir,
		m.Datastore.Dir == "",
		m.Datastore.InMemory:
		return InMemoryDir, nil
	}

	m.Datastore.dir = path.Join(m.Datastore.Dir, "account0") // account0 is a suffix that will be used with multi-account later

	_, err := os.Stat(m.Datastore.dir)
	switch {
	case os.IsNotExist(err):
		if err := os.MkdirAll(m.Datastore.dir, 0o700); err != nil {
			return "", errcode.TODO.Wrap(err)
		}
	case err != nil:
		return "", errcode.TODO.Wrap(err)
	}

	m.initLogger.Debug("datastore dir", zap.String("dir", m.Datastore.dir))
	return m.Datastore.dir, nil
}

func (m *Manager) GetRootDatastore() (datastore.Batching, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.getRootDatastore()
}

func (m *Manager) getRootDatastore() (datastore.Batching, error) {
	m.applyDefaults()

	if m.Datastore.rootDS != nil {
		return m.Datastore.rootDS, nil
	}

	dir, err := m.getDatastoreDir()
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	if dir == InMemoryDir {
		return sync_ds.MutexWrap(datastore.NewMapDatastore()), nil
	}

	var opts *badger.Options
	if m.Datastore.FileIO {
		opts = &badger.Options{
			Options: badger.DefaultOptions.WithValueLogLoadingMode(badger_opts.FileIO),
		}
	}

	ds, err := badger.NewDatastore(dir, opts)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}
	m.Datastore.rootDS = ds

	m.Datastore.rootDS = sync_ds.MutexWrap(m.Datastore.rootDS)
	m.initLogger.Debug("datastore", zap.Bool("in-memory", dir == InMemoryDir))
	return m.Datastore.rootDS, nil
}
