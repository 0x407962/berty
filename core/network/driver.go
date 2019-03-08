package network

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"berty.tech/core/entity"
	"berty.tech/core/network/helper"
	"berty.tech/core/network/metric"
	"berty.tech/core/pkg/tracing"
	ggio "github.com/gogo/protobuf/io"
	cid "github.com/ipfs/go-cid"
	ipfsaddr "github.com/ipfs/go-ipfs-addr"
	libp2p_discovery "github.com/libp2p/go-libp2p-discovery"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	protocol "github.com/libp2p/go-libp2p-protocol"
	ma "github.com/multiformats/go-multiaddr"
	mh "github.com/multiformats/go-multihash"
	"go.uber.org/zap"
)

var _ Driver = (*Network)(nil)

var ProtocolID = protocol.ID("berty/p2p/envelope")

func (net *Network) logHostInfos() {
	var addrs []string

	for _, addr := range net.host.Addrs() {
		addrs = append(addrs, addr.String())
	}

	logger().Debug("Host", zap.String("ID", net.host.ID().Pretty()), zap.Strings("Addrs", addrs))
}

func (net *Network) getPeerInfo(ctx context.Context, addr string) (*pstore.PeerInfo, error) {
	tracer := tracing.EnterFunc(ctx, addr)
	defer tracer.Finish()
	// ctx = tracer.Context()

	iaddr, err := ipfsaddr.ParseString(addr)
	if err != nil {
		return nil, err
	}

	return pstore.InfoFromP2pAddr(iaddr.Multiaddr())
}

func (net *Network) Protocols(ctx context.Context, p *metric.Peer) ([]string, error) {
	tracer := tracing.EnterFunc(ctx, p)
	defer tracer.Finish()
	// ctx = tracer.Context()

	peerid, err := peer.IDB58Decode(p.ID)
	if err != nil {
		return nil, fmt.Errorf("get protocols error: `%s`", err)
	}

	return net.host.Peerstore().GetProtocols(peerid)
}

func (net *Network) Addrs() []ma.Multiaddr {
	return net.host.Addrs()
}

func (net *Network) ID(ctx context.Context) *metric.Peer {
	tracer := tracing.EnterFunc(ctx)
	defer tracer.Finish()
	// ctx = tracer.Context()
	addrs := []string{}
	for _, addr := range net.host.Addrs() {
		addrs = append(addrs, addr.String())
	}

	return &metric.Peer{
		ID:         net.host.ID().Pretty(),
		Addrs:      addrs,
		Connection: metric.ConnectionType_CONNECTED,
	}
}

func (net *Network) handleNewPovider(providerID string, pi pstore.PeerInfo) {
	logger().Debug("new providers", zap.String("providers ID", providerID), zap.String("peer ID", pi.ID.Pretty()))
}

func (net *Network) Peerstore(ctx context.Context) pstore.Peerstore {
	tracer := tracing.EnterFunc(ctx)
	defer tracer.Finish()
	// ctx = tracer.Context()

	return net.host.Peerstore()
}

func (net *Network) Bootstrap(ctx context.Context, sync bool, addrs ...string) error {
	tracer := tracing.EnterFunc(ctx, sync, addrs)
	defer tracer.Finish()
	ctx = tracer.Context()

	bf := net.BootstrapPeerAsync
	if sync {
		bf = net.BootstrapPeer
	}

	for _, addr := range addrs {
		if err := bf(ctx, addr); err != nil {
			return err
		}
	}

	return nil
}

func (net *Network) BootstrapPeerAsync(ctx context.Context, addr string) error {
	tracer := tracing.EnterFunc(ctx, addr)
	defer tracer.Finish()
	ctx = tracer.Context()

	go func() {
		if err := net.BootstrapPeer(ctx, addr); err != nil {
			logger().Warn("Bootstrap error", zap.String("addr", addr), zap.Error(err))
		}
	}()

	return nil
}

func (net *Network) BootstrapPeer(ctx context.Context, bootstrapAddr string) error {
	tracer := tracing.EnterFunc(ctx, bootstrapAddr)
	defer tracer.Finish()
	ctx = tracer.Context()

	if bootstrapAddr == "" {
		return nil
	}

	logger().Debug("Bootstraping peer", zap.String("addr", bootstrapAddr))
	pinfo, err := net.getPeerInfo(ctx, bootstrapAddr)
	if err != nil {
		return err
	}

	// Even if we can't connect, bootstrap peers are trusted peers, add it to
	// the peerstore so we can connect later in case of failure
	net.host.Peerstore().AddAddrs(pinfo.ID, pinfo.Addrs, pstore.PermanentAddrTTL)
	net.host.ConnManager().TagPeer(pinfo.ID, "bootstrap", 2)
	return net.host.Connect(ctx, *pinfo)
}

func (net *Network) Discover(ctx context.Context) {
	libp2p_discovery.Advertise(ctx, net.host.Discovery, "berty")
	go func() {
		for {
			peers, err := libp2p_discovery.FindPeers(ctx, net.host.Discovery, "berty", 0)
			if err != nil {
				logger().Error("network discover error", zap.String("err", err.Error()))
				continue
			}
			for _, pi := range peers {
				net.Connect(ctx, pi)
			}
		}
	}()
}

// Connect ensures there is a connection between this host and the peer with
// given peer.ID.
func (net *Network) Connect(ctx context.Context, pi pstore.PeerInfo) error {
	// first, check if we're already connectenet.
	if net.host.Network().Connectedness(pi.ID) == inet.Connected {
		return nil
	}
	// if we were given some addresses, keep + use them.
	if len(pi.Addrs) > 0 {
		net.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, pstore.TempAddrTTL)
	}

	// Check if we have some addresses in our recent memory.
	addrs := net.host.Peerstore().Addrs(pi.ID)
	if len(addrs) < 1 {
		// no addrs? find some with the routing system.
		var err error
		pi, err = net.host.Routing.FindPeer(ctx, pi.ID)
		if err != nil {
			return err
		}
		return net.host.Connect(ctx, pi)
	}

	// if we're here, we got some addrs. let's use our wrapped host to connect.
	pi.Addrs = addrs
	return net.host.Connect(ctx, pi)
}

func (net *Network) Dial(ctx context.Context, peerID string, pid protocol.ID) (net.Conn, error) {
	tracer := tracing.EnterFunc(ctx, peerID, pid)
	defer tracer.Finish()
	ctx = tracer.Context()

	return helper.NewDialer(net.host, pid)(ctx, peerID)
}

func (net *Network) createCid(id string) (cid.Cid, error) {
	h, err := mh.Sum([]byte(id), mh.SHA2_256, -1)
	if err != nil {
		return cid.Cid{}, err
	}

	return cid.NewCidV0(h), nil
}

func (net *Network) Find(ctx context.Context, pid peer.ID) (pstore.PeerInfo, error) {
	return net.host.Routing.FindPeer(ctx, pid)
}

func (net *Network) Emit(ctx context.Context, e *entity.Envelope) error {
	tracer := tracing.EnterFunc(ctx, e)
	defer tracer.Finish()
	ctx = tracer.Context()

	return net.EmitTo(ctx, e.GetChannelID(), e)
}

func (net *Network) EmitTo(ctx context.Context, channel string, e *entity.Envelope) error {
	tracer := tracing.EnterFunc(ctx, channel, e)
	defer tracer.Finish()
	ctx = tracer.Context()

	logger().Debug("looking for peers", zap.String("channel", channel))
	ss, err := net.FindProvidersAndWait(ctx, channel, true)
	if err != nil {
		return err
	}

	// @TODO: we need to split this, and let the node do the logic to try
	// back if the send fail with the given peer

	logger().Debug("found peers", zap.String("channel", channel), zap.Int("number", len(ss)))

	mu := sync.Mutex{}
	wg := sync.WaitGroup{}
	wg.Add(len(ss))

	ok := false
	for i, s := range ss {
		go func(pi pstore.PeerInfo, index int) {
			gotracer := tracing.EnterFunc(ctx, index)
			goctx := tracer.Context()

			defer gotracer.Finish()
			defer wg.Done()

			if err := net.SendTo(goctx, pi, e); err != nil {
				logger().Warn("sendTo", zap.Error(err))
				return
			}

			mu.Lock()
			ok = true
			mu.Unlock()
			return
		}(s, i)
	}

	// wait until all goroutines are done
	wg.Wait()
	if !ok {
		return fmt.Errorf("unable to send evenlope to at last one peer")
	}

	return nil
}

func (net *Network) SendTo(ctx context.Context, pi pstore.PeerInfo, e *entity.Envelope) error {
	peerID := pi.ID.Pretty()
	if pi.ID == net.host.ID() {
		return fmt.Errorf("cannot dial to self")
	}

	logger().Debug("connecting", zap.String("peerID", peerID))
	if err := net.Connect(ctx, pi); err != nil {
		return fmt.Errorf("failed to connect: `%s`", err.Error())
	}

	s, err := net.host.NewStream(ctx, pi.ID, ProtocolID)
	if err != nil {
		return fmt.Errorf("new stream failed: `%s`", err.Error())
	}

	logger().Debug("sending envelope", zap.String("peerID", peerID))
	pbw := ggio.NewDelimitedWriter(s)
	if err := pbw.WriteMsg(e); err != nil {
		return fmt.Errorf("write stream: `%s`", err.Error())
	}

	go inet.FullClose(s)
	return nil
}

func (net *Network) handleEnvelope(s inet.Stream) {
	logger().Debug("receiving envelope")
	if net.handler == nil {
		logger().Error("handler is not set")
		return
	}

	pbr := ggio.NewDelimitedReader(s, inet.MessageSizeMax)
	for {
		e := &entity.Envelope{}
		switch err := pbr.ReadMsg(e); err {
		case io.EOF:
			s.Close()
			return
		case nil: // do noting, everything fine
		default:
			s.Reset()
			logger().Error("invalid envelope", zap.Error(err))
			return
		}

		// @TODO: get opentracing context
		net.handler(context.Background(), e)
	}

}

func (net *Network) FindProvidersAndWait(ctx context.Context, id string, cache bool) ([]pstore.PeerInfo, error) {
	c, err := net.createCid(id)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	piChan := net.host.Routing.FindProvidersAsync(ctx, c, 10)

	piSlice := []pstore.PeerInfo{}
	for {
		select {
		case pi := <-piChan:
			if pi.ID != "" {
				piSlice = append(piSlice, pi)
			}
		case <-ctx.Done():
			if len(piSlice) == 0 {
				return nil, errors.New("no providers found")
			}
			return piSlice, nil
		}
	}

}

func (net *Network) Join(ctx context.Context, id string) error {
	c, err := net.createCid(id)
	if err != nil {
		return err
	}

	return net.host.Routing.Provide(ctx, c, true)
}

func (net *Network) OnEnvelopeHandler(f func(context.Context, *entity.Envelope) (*entity.Void, error)) {
	net.handler = f
}

func (net *Network) PingOtherNode(ctx context.Context, destination string) (err error) {
	peerid, err := peer.IDB58Decode(destination)
	if err != nil {
		return err
	}

	ch, err := net.host.Ping.Ping(ctx, peerid)
	if err != nil {
		return err
	}

	<-ch
	return nil
}

func (net *Network) Metric() metric.Metric {
	return net.host.Metric
}
