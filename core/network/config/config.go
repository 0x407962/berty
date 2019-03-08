package config

import (
	"context"
	"time"

	"berty.tech/core/network/host"
	"berty.tech/core/network/metric"
	"berty.tech/core/network/protocol/mdns"
	discovery "github.com/libp2p/go-libp2p-discovery"
	libp2p_host "github.com/libp2p/go-libp2p-host"
	routing "github.com/libp2p/go-libp2p-routing"
	libp2p_config "github.com/libp2p/go-libp2p/config"
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
	libp2p_config.Config

	MDNS bool
	DHT  bool

	DefaultBootstrap bool
	Bootstrap        []string

	Ping bool

	Metric bool
}

type Option func(cfg *Config) error

// Apply applies the given options to the config, returning the first error
// encountered (if any).
func (cfg *Config) Apply(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return err
		}
	}
	return nil
}

func (cfg *Config) NewNode(ctx context.Context) (*host.BertyHost, error) {

	var err error

	var routingOpt *host.BertyRouting
	discoveries := []discovery.Discovery{}

	if cfg.DefaultBootstrap {
		cfg.Bootstrap = append(cfg.Bootstrap, DefaultBootstrap...)
	}

	// setup dht for libp2p routing host
	cfg.Config.Routing = func(h libp2p_host.Host) (routing.PeerRouting, error) {
		// configure DHT
		routingOpt, err = host.NewBertyRouting(ctx, h, cfg.DHT)
		if err != nil {
			return nil, err
		}
		return routingOpt, nil
	}

	// override conn manager
	cfg.Config.ConnManager = host.NewBertyConnMgr(ctx, 10, 20, time.Duration(60*1000))

	// override ping service
	cfg.Config.DisablePing = true

	h, err := cfg.Config.NewNode(ctx)
	if err != nil {
		return nil, err
	}

	// configure mdns service
	if cfg.MDNS {
		if mdns, err := mdns.NewDiscovery(ctx, h); err != nil {
			return nil, err
		} else {
			discoveries = append(discoveries, mdns)
		}
	}

	// configure ping service
	var pingOpt *host.PingService
	if cfg.Ping {
		pingOpt = host.NewPingService(h)
	}

	// configure metric service
	var metricOpt metric.Metric
	if cfg.Metric {
		metricOpt = metric.NewBertyMetric(ctx, h, pingOpt)
		h.Network().Notify(metricOpt)
	}

	return host.NewBertyHost(ctx, h, &host.BertyHostOptions{
		Discovery: host.NewBertyDiscovery(ctx, discoveries),
		Routing:   routingOpt,
		Metric:    metricOpt,
		Ping:      pingOpt,
	})
}
