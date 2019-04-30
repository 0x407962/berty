package node

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"berty.tech/core/crypto/keypair"
	"berty.tech/core/crypto/sigchain"
	"berty.tech/core/entity"
	"berty.tech/core/network"
	network_metric "berty.tech/core/network/metric"
	"berty.tech/core/pkg/notification"
	"berty.tech/core/pkg/tracing"
	"berty.tech/core/pkg/zapring"
	"berty.tech/core/push"
	"github.com/gofrs/uuid"
	"github.com/gogo/protobuf/proto"
	"github.com/jinzhu/gorm"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

// Node is the top-level object of a Berty peer
type Node struct {
	clientCommitLogsSubscribers []clientCommitLogsSubscriber
	clientCommitLogsMutex       sync.Mutex
	clientEvents                chan *entity.Event
	clientEventsSubscribers     []clientEventSubscriber
	clientEventsMutex           sync.Mutex
	outgoingEvents              chan *entity.EventDispatch
	sqlDriver                   *gorm.DB
	config                      *entity.Config
	initDevice                  *entity.Device
	handleMutexInst             sync.Mutex
	notificationDriver          notification.Driver
	networkDriver               network.Driver
	networkMetric               network_metric.Metric
	asyncWaitGroupInst          sync.WaitGroup
	pubkey                      []byte // FIXME: use a crypto instance, i.e., enclave
	b64pubkey                   string // FIXME: same as above
	sigchain                    *sigchain.SigChain
	crypto                      keypair.Interface
	ring                        *zapring.Ring // log ring buffer
	rootSpan                    opentracing.Span
	rootContext                 context.Context // only used for tracing
	pushManager                 *push.Manager

	// devtools
	createdAt time.Time // used for uptime calculation
	devtools  struct {
		mapset map[string]string
	}

	shutdown chan struct{}
}

// New initializes a new Node object
func New(ctx context.Context, opts ...NewNodeOption) (*Node, error) {
	tracer := tracing.EnterFunc(ctx)
	// FIXME: defer tracer.Finish() // ???
	ctx = tracer.Context()

	n := &Node{
		// FIXME: fetch myself from db
		outgoingEvents: make(chan *entity.EventDispatch, 1),
		clientEvents:   make(chan *entity.Event, 1),
		createdAt:      time.Now().UTC(),
		rootSpan:       tracer.Span(),
		rootContext:    ctx,
		pushManager:    &push.Manager{},
		shutdown:       make(chan struct{}, 1),
	}

	// apply optioners
	for _, opt := range opts {
		opt(n)
	}

	// use NoopNotification by default
	if n.notificationDriver == nil {
		n.notificationDriver = notification.NewNoopNotification()
	}

	// check for misconfigurations based on optioners
	if err := n.Validate(); err != nil {
		return nil, errors.Wrap(err, "node is misconfigured")
	}

	// cache the signing pubkey
	var sc sigchain.SigChain
	if err := proto.Unmarshal(n.config.Myself.Sigchain, &sc); err != nil {
		return nil, errors.Wrap(err, "cannot get sigchain")
	}

	pubKey, err := n.crypto.GetPubKey()

	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve public key")
	}

	n.pubkey = pubKey
	n.b64pubkey = base64.StdEncoding.EncodeToString(n.pubkey)

	n.sigchain = &sc

	// configure network
	if n.networkDriver != nil {
		if err := n.UseNetworkDriver(n.rootContext, n.networkDriver); err != nil {
			return nil, errors.Wrap(err, "failed to setup network driver")
		}
	}

	return n, nil
}

// Shutdown closes object initialized by Node itself
//
// it should be called in a defer from the caller of New()
func (n *Node) Shutdown(ctx context.Context) {
	tracer := tracing.EnterFunc(ctx)
	defer tracer.Finish()
	close(n.shutdown)
	n.rootSpan.Finish()
}

// Validate returns an error if object is invalid
func (n *Node) Validate() error {
	if n == nil {
		return errors.New("missing required fields (node) to create a new Node")
	} else if n.sqlDriver == nil {
		return errors.New("missing required fields (gorm) to create a new Node")
	} else if n.initDevice == nil {
		return errors.New("missing required fields (initDevice) to create a new Node")
	} else if n.networkDriver == nil {
		return errors.New("missing required fields (networkDriver) to create a new Node")
	} else if n.notificationDriver == nil {
		return errors.New("missing required fields (notificationDriver) to create a new Node")
	} else if n.crypto == nil {
		return errors.New("missing required fields (crypto) to create a new Node")
	} else if n.config == nil {
		return errors.New("missing required fields (config) to create a new Node")
	} else if n.config.Myself == nil {
		return errors.New(fmt.Sprintf("missing required fields (config.Myself) to create a new Node %+v\n\n\n\n", n.config))
	}

	return nil
}

// NewNodeOption is a callback used to configure a Node during intiailization phase
type NewNodeOption func(n *Node)

// NewID returns a unique ID prefixed with our contact ID
func (n *Node) NewID() string {
	return fmt.Sprintf("%s:%s", n.config.Myself.ID, uuid.Must(uuid.NewV4()).String())
}

func (n *Node) DeviceID() string {
	return n.config.CurrentDevice.ID
}

func (n *Node) UserID() string {
	return n.config.Myself.ID
}

func (n *Node) PubKey() string {
	return n.b64pubkey
}

func (n *Node) handleMutex(ctx context.Context) func() {
	tracer := tracing.EnterFunc(ctx)
	defer tracer.Finish()
	// ctx = tracer.Context()

	n.handleMutexInst.Lock()
	return n.handleMutexInst.Unlock
}

func (n *Node) asyncWaitGroup(ctx context.Context) func() {
	tracer := tracing.EnterFunc(ctx)
	defer tracer.Finish()
	// ctx = tracer.Context()

	n.asyncWaitGroupInst.Add(1)
	return n.asyncWaitGroupInst.Done
}
