package bertydemo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"berty.tech/berty/go/internal/group"
	"berty.tech/berty/go/internal/ipfsutil"
	"berty.tech/berty/go/internal/orbitutil/identityberty"
	"berty.tech/berty/go/internal/orbitutil/storegroup"
	"berty.tech/berty/go/pkg/errcode"
	orbitdb "berty.tech/go-orbit-db"
	"berty.tech/go-orbit-db/stores/operation"
	cid "github.com/ipfs/go-cid"
	ipfs_core "github.com/ipfs/go-ipfs/core"
	ipfs_interface "github.com/ipfs/interface-go-ipfs-core"
	"github.com/libp2p/go-libp2p-core/crypto"
)

type Client struct {
	api       ipfs_interface.CoreAPI
	node      *ipfs_core.IpfsNode
	odb       orbitdb.OrbitDB
	logs      map[string]orbitdb.EventLogStore
	logsMutex *sync.Mutex
}

type Opts struct {
	OrbitDBDirectory string
}

func New(opts *Opts) (*Client, error) {
	ctx := context.Background()

	api, node, err := ipfsutil.NewInMemoryCoreAPI(ctx)
	if err != nil {
		return nil, err
	}

	if opts.OrbitDBDirectory == "" {
		opts.OrbitDBDirectory = ":memory:"
	}

	odb, err := orbitdb.NewOrbitDB(ctx, api, &orbitdb.NewOrbitDBOptions{Directory: &opts.OrbitDBDirectory})
	if err != nil {
		return nil, err
	}

	logs := make(map[string]orbitdb.EventLogStore)

	logsMutex := &sync.Mutex{}

	return &Client{api, node, odb, logs, logsMutex}, nil
}

func intPtr(i int) *int {
	return &i
}

func (d *Client) logFromToken(ctx context.Context, token string) (orbitdb.EventLogStore, error) {
	d.logsMutex.Lock()
	defer d.logsMutex.Unlock()

	if log, ok := d.logs[token]; ok {
		// I tried to avoid this map but orbitdb recreates log instances and looses operations even when fed with the same args
		return log, nil
	}
	sigkb, err := hex.DecodeString(token)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}
	sigk, err := crypto.UnmarshalEd25519PrivateKey(sigkb)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}
	ks := identityberty.NewBertySignedKeyStore()
	err = ks.SetKey(sigk)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}
	g := &group.Group{PubKey: sigk.GetPublic(), SigningKey: sigk}
	opts, err := storegroup.DefaultOptions(g, &orbitdb.CreateDBOptions{}, ks)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}
	log, err := d.odb.Log(ctx, "DemoLog", opts)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}
	d.logs[token] = log
	return log, nil
}

func (d *Client) LogToken(ctx context.Context, _ *LogToken_Request) (*LogToken_Reply, error) {
	sigk, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}
	sigkb, err := sigk.Raw()
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}
	return &LogToken_Reply{LogToken: hex.EncodeToString(sigkb)}, nil
}

func operationCidString(op operation.Operation) string {
	return op.GetEntry().GetHash().String()
}

func (d *Client) LogAdd(ctx context.Context, req *LogAdd_Request) (*LogAdd_Reply, error) {
	log, err := d.logFromToken(ctx, req.GetLogToken())
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	op, err := log.Add(ctx, req.GetData())
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}
	return &LogAdd_Reply{Cid: operationCidString(op)}, nil
}

func convertLogOperationToProtobufLogOperation(op operation.Operation) *LogOperation {
	if op == nil {
		return nil
	}
	return &LogOperation{
		Name:  op.GetOperation(),
		Value: op.GetValue(),
		Cid:   operationCidString(op),
	}
}

func (d *Client) LogGet(ctx context.Context, req *LogGet_Request) (*LogGet_Reply, error) {
	log, err := d.logFromToken(ctx, req.GetLogToken())
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	c, err := cid.Decode(req.GetCid())
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	op, err := log.Get(ctx, c)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	pop := convertLogOperationToProtobufLogOperation(op)
	return &LogGet_Reply{Operation: pop}, nil
}

func maybeDecodeCid(str string) *cid.Cid {
	c, err := cid.Decode(str)
	if err != nil {
		return nil
	}
	return &c
}

func decodeStreamOptions(opts *LogStreamOptions) *orbitdb.StreamOptions {
	if opts == nil {
		return nil
	}
	return &orbitdb.StreamOptions{
		GT:     maybeDecodeCid(opts.GetGT()),
		GTE:    maybeDecodeCid(opts.GetGTE()),
		LT:     maybeDecodeCid(opts.GetLT()),
		LTE:    maybeDecodeCid(opts.GetLTE()),
		Amount: intPtr(int(opts.GetAmount())),
	}
}

func (d *Client) LogList(ctx context.Context, req *LogList_Request) (*LogList_Reply, error) {
	log, err := d.logFromToken(ctx, req.GetLogToken())
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	opts := decodeStreamOptions(req.GetOptions())

	ops, err := log.List(ctx, opts)
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	protoOps := make([]*LogOperation, len(ops))
	for i, op := range ops {
		protoOps[i] = convertLogOperationToProtobufLogOperation(op)
	}

	return &LogList_Reply{Operations: protoOps}, nil
}

func (d *Client) LogStream(req *LogStream_Request, srv DemoService_LogStreamServer) error {
	// Hack using List until go-orbit-db Stream is fixed
	ctx := srv.Context()
	log, err := d.logFromToken(ctx, req.GetLogToken())
	if err != nil {
		return errcode.TODO.Wrap(err)
	}

	opts := decodeStreamOptions(req.GetOptions())
	if opts == nil {
		opts = &orbitdb.StreamOptions{}
	}

	dc := ctx.Done()
	for {
		if dc != nil {
			select {
			case _, done := <-dc:
				if done {
					return nil
				}
			default:
			}
		} // else we are in an infinite loop
		ops, err := log.List(ctx, opts)
		if err != nil {
			return err
		}
		if len(ops) == 0 {
			continue
		}
		for _, op := range ops {
			pop := convertLogOperationToProtobufLogOperation(op)
			if err = srv.Send(pop); err != nil {
				return err
			}
		}
		lastOpCid := ops[len(ops)-1].GetEntry().GetHash()
		opts.GT = &lastOpCid
		time.Sleep(1 * time.Second)
	}
}

func (d *Client) Close() error {
	err := d.odb.Close()
	if err != nil {
		return err
	}
	return d.node.Close()
}
