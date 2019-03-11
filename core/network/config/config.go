package config

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	peer "github.com/libp2p/go-libp2p-peer"
	pnet "github.com/libp2p/go-libp2p-pnet"
	swarm "github.com/libp2p/go-libp2p-swarm"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"

	"berty.tech/core/network/host"
	"berty.tech/core/network/metric"
	"berty.tech/core/network/protocol/ble"
	"berty.tech/core/network/protocol/mdns"
	"berty.tech/core/pkg/errorcodes"
	libp2p "github.com/libp2p/go-libp2p"
	circuit "github.com/libp2p/go-libp2p-circuit"
	libp2p_crypto "github.com/libp2p/go-libp2p-crypto"
	discovery "github.com/libp2p/go-libp2p-discovery"
	quic "github.com/libp2p/go-libp2p-quic-transport"
	libp2p_config "github.com/libp2p/go-libp2p/config"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	tcp "github.com/libp2p/go-tcp-transport"
	ws "github.com/libp2p/go-ws-transport"
)

const DefaultSwarmKey = `/key/swarm/psk/1.0.0/
/base16/
7beb018da4c79cb018e05305335d265046909f060c1b65e8eef94a107b9387cc`

var DefaultBootstrap = []string{
	"/ip4/51.158.71.240/udp/4004/quic/ipfs/QmeYFvq4VV5RU1k1wBw3J5ZZLYxTE6H3AAKMzCAavuBjTp",
	"/ip4/51.158.71.240/tcp/4004/ipfs/QmeYFvq4VV5RU1k1wBw3J5ZZLYxTE6H3AAKMzCAavuBjTp",
	"/ip4/51.158.71.240/tcp/443/ipfs/QmeYFvq4VV5RU1k1wBw3J5ZZLYxTE6H3AAKMzCAavuBjTp",
	"/ip4/51.158.71.240/tcp/80/ipfs/QmeYFvq4VV5RU1k1wBw3J5ZZLYxTE6H3AAKMzCAavuBjTp",
	"/ip4/51.158.67.118/udp/4004/quic/ipfs/QmS88MDaMZUQeEvVdRAFmMfMz96b19Y79VJ6wQJnf4dwoo",
	"/ip4/51.158.67.118/tcp/4004/ipfs/QmS88MDaMZUQeEvVdRAFmMfMz96b19Y79VJ6wQJnf4dwoo",
	"/ip4/51.158.67.118/tcp/443/ipfs/QmS88MDaMZUQeEvVdRAFmMfMz96b19Y79VJ6wQJnf4dwoo",
	"/ip4/51.158.67.118/tcp/80/ipfs/QmS88MDaMZUQeEvVdRAFmMfMz96b19Y79VJ6wQJnf4dwoo",
	"/ip4/51.15.221.60/udp/4004/quic/ipfs/QmZP7oAGikmrMLAmf7ooNtnarYdDWki4Wru2sJ5H5kgCw3",
	"/ip4/51.15.221.60/tcp/4004/ipfs/QmZP7oAGikmrMLAmf7ooNtnarYdDWki4Wru2sJ5H5kgCw3",
	"/ip4/51.15.221.60/tcp/443/ipfs/QmZP7oAGikmrMLAmf7ooNtnarYdDWki4Wru2sJ5H5kgCw3",
	"/ip4/51.15.221.60/tcp/80/ipfs/QmZP7oAGikmrMLAmf7ooNtnarYdDWki4Wru2sJ5H5kgCw3",
}

var BootstrapIpfs = []string{
	"/ip4/104.131.131.82/tcp/4001/ipfs/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
	"/ip4/104.236.179.241/tcp/4001/ipfs/QmSoLPppuBtQSGwKDZT2M73ULpjvfd3aZ6ha4oFGL1KrGM",
	"/ip4/104.236.76.40/tcp/4001/ipfs/QmSoLV4Bbm51jM9C4gDYZQ9Cy3U6aXMJDAbzgu2fzaDs64",
	"/ip4/128.199.219.111/tcp/4001/ipfs/QmSoLSafTMBsPKadTEgaXctDQVcqN88CNLHXMkTNwMKPnu",
	"/ip4/178.62.158.247/tcp/4001/ipfs/QmSoLer265NRgSp2LA3dPaeykiS1J6DifTC88f5uVQKNAd",
}

type Config struct {
	libp2p_config.Config `json:"-"`

	Bind []string

	MDNS bool
	DHT  bool

	// transport
	WS   bool
	TCP  bool
	BLE  bool
	QUIC bool

	DefaultBootstrap bool
	Bootstrap        []string

	Ping bool

	Metric bool

	HOP bool

	SwarmKey string

	Identity string

	Persist         bool `json:"-"`
	OverridePersist bool `json:"-"` // override persist config when apply
}

type Option func(cfg *Config) error

// Override override safely the current config
func (cfg *Config) Override(override *Config) error {
	cfg.MDNS = override.MDNS
	cfg.WS = override.WS
	cfg.TCP = override.TCP
	cfg.DHT = override.DHT
	cfg.BLE = override.BLE
	cfg.QUIC = override.QUIC
	cfg.DefaultBootstrap = override.DefaultBootstrap
	cfg.Bootstrap = override.Bootstrap
	cfg.Ping = override.Ping
	cfg.HOP = override.HOP
	cfg.Identity = override.Identity
	cfg.SwarmKey = override.SwarmKey
	return nil
}

// Apply applies the given options to the config, returning the first error
// encountered (if any).
func (cfg *Config) Apply(ctx context.Context, opts ...Option) error {
	libp2pOpts := []libp2p_config.Option{}

	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return err
		}
	}

	if cfg.OverridePersist {
		if err := cfg.OverridePersistConfig(); err != nil {
			return err
		}
	}

	if cfg.Persist {
		if err := cfg.ApplyPersistConfig(); err != nil {
			return err
		}
	}

	if cfg.DefaultBootstrap {
		cfg.Bootstrap = append(cfg.Bootstrap, DefaultBootstrap...)
	}

	libp2pOpts = append(libp2pOpts, libp2p.DefaultListenAddrs)
	if len(cfg.Bind) > 0 {
		libp2pOpts = append(libp2pOpts, libp2p.ListenAddrStrings(cfg.Bind...))
	}

	// add ws transport
	if cfg.WS {
		libp2pOpts = append(libp2pOpts, libp2p.Transport(ws.New))
	}

	// add tcp transport
	if cfg.TCP {
		libp2pOpts = append(libp2pOpts, libp2p.Transport(tcp.NewTCPTransport))
	}

	// add ble transport
	if cfg.BLE {
		libp2pOpts = append(libp2pOpts, libp2p.Transport(ble.NewTransport))
	}

	// add quic transport
	if cfg.QUIC {
		libp2pOpts = append(libp2pOpts, libp2p.Transport(quic.NewTransport))
	}

	// relay
	if cfg.HOP {
		libp2pOpts = append(libp2pOpts, libp2p.EnableRelay(circuit.OptActive, circuit.OptHop))
	} else {
		libp2pOpts = append(libp2pOpts, libp2p.EnableRelay(circuit.OptActive, circuit.OptDiscovery))
	}

	// private network
	if cfg.SwarmKey != "" {
		prot, err := pnet.NewProtector(strings.NewReader(cfg.SwarmKey))
		if err != nil {
			return err
		}
		libp2pOpts = append(libp2pOpts, libp2p.PrivateNetwork(prot))
	}

	// identity
	if cfg.Identity != "" {
		bytes, err := base64.StdEncoding.DecodeString(cfg.Identity)
		if err != nil {
			return errorcodes.ErrNetP2PIdentity.Wrap(err)
		}

		identity, err := libp2p_crypto.UnmarshalPrivateKey(bytes)
		if err != nil {
			return errorcodes.ErrNetP2PPublicKey.Wrap(err)
		}
		libp2pOpts = append(libp2pOpts, libp2p.Identity(identity))
	} else {
		libp2pOpts = append(libp2pOpts, libp2p.RandomIdentity)
	}

	libp2pOpts = append(libp2pOpts,
		libp2p.ConnectionManager(host.NewBertyConnMgr(ctx, 10, 20, time.Duration(60*time.Minute))))

	libp2pOpts = append(libp2pOpts, libp2p.NATPortMap())
	// override libp2p configuration
	err := cfg.Config.Apply(append(libp2pOpts, libp2p.FallbackDefaults)...)
	if err != nil {
		return err
	}

	// override conn manager

	// override ping service
	cfg.Config.DisablePing = true
	return nil
}

func (cfg *Config) NewNode(ctx context.Context) (*host.BertyHost, error) {
	var err error
	discoveries := []discovery.Discovery{}

	if cfg.Config.PeerKey == nil {
		return nil, fmt.Errorf("no peer key specified")
	}

	// Obtain Peer ID from public key
	pid, err := peer.IDFromPublicKey(cfg.Config.PeerKey.GetPublic())
	if err != nil {
		return nil, err
	}

	if cfg.Config.Peerstore == nil {
		return nil, fmt.Errorf("no peerstore specified")
	}

	if !cfg.Config.Insecure {
		cfg.Config.Peerstore.AddPrivKey(pid, cfg.Config.PeerKey)
		cfg.Config.Peerstore.AddPubKey(pid, cfg.Config.PeerKey.GetPublic())
	}

	// TODO: Make the swarm implementation configurable.
	swrm := swarm.NewSwarm(ctx, pid, cfg.Config.Peerstore, cfg.Config.Reporter)
	if cfg.Config.Filters != nil {
		swrm.Filters = cfg.Config.Filters
	}

	// use basic host
	h := &host.BertyHost{}
	h.Host, err = bhost.NewHost(ctx, swrm, &bhost.HostOpts{
		ConnManager:  cfg.Config.ConnManager,
		AddrsFactory: cfg.Config.AddrsFactory,
		NATManager:   cfg.Config.NATManager,
		EnablePing:   !cfg.Config.DisablePing,
	})
	if err != nil {
		swrm.Close()
		return nil, err
	}

	// upgrader
	upgrader := new(tptu.Upgrader)
	upgrader.Protector = cfg.Config.Protector
	upgrader.Filters = swrm.Filters
	if cfg.Config.Insecure {
		upgrader.Secure = makeInsecureTransport(pid)
	} else {
		upgrader.Secure, err = makeSecurityTransport(h, cfg.Config.SecurityTransports)
		if err != nil {
			h.Close()
			return nil, err
		}
	}

	upgrader.Muxer, err = makeMuxer(h, cfg.Config.Muxers)
	if err != nil {
		h.Close()
		return nil, err
	}

	tpts, err := makeTransports(h, upgrader, cfg.Config.Transports)
	if err != nil {
		h.Close()
		return nil, err
	}
	for _, t := range tpts {
		err = swrm.AddTransport(t)
		if err != nil {
			h.Close()
			return nil, err
		}
	}

	if cfg.Config.Relay {
		err := circuit.AddRelayTransport(swrm.Context(), h, upgrader, cfg.Config.RelayOpts...)
		if err != nil {
			h.Close()
			return nil, err
		}
	}

	// TODO: This method succeeds if listening on one address succeeds. We
	// should probably fail if listening on *any* addr fails.
	if err := h.Network().Listen(cfg.Config.ListenAddrs...); err != nil {
		h.Close()
		return nil, err
	}

	// Configure routing
	h.Routing, err = host.NewBertyRouting(ctx, h, cfg.DHT)
	if err != nil {
		h.Close()
		return nil, err
	}
	h.Network().Notify(h.Routing.(*host.BertyRouting))

	// crouter, ok := h.Routing.(routing.ContentRouting)
	// if !ok {
	// 	h.Close()
	// 	return nil, fmt.Errorf("cannot enable autorelay; no suitable routing for discovery")
	// }

	// routerDiscovery := discovery.NewRoutingDiscovery(h.Routing)

	// discoveries = append(discoveries, routerDiscovery)

	// configure mdns service
	if cfg.MDNS {
		if mdns, err := mdns.NewDiscovery(ctx, h); err != nil {
			return nil, err
		} else {
			discoveries = append(discoveries, mdns)
		}
	}

	// configure ping service
	if cfg.Ping {
		h.Ping = host.NewPingService(h)
	}

	// configure metric service
	if cfg.Metric {
		if !cfg.Ping {
			return nil, fmt.Errorf("cannot enable metric; ping is not enabled")
		}
		h.Metric = metric.NewBertyMetric(ctx, h, h.Ping)
		h.Network().Notify(h.Metric)
	}

	h.Discovery = host.NewBertyDiscovery(ctx, discoveries)

	return h, nil
}
