package network

import (
	"context"
	"sync"

	"berty.tech/core/entity"
	"berty.tech/core/network/config"
	host "berty.tech/core/network/host"
	"berty.tech/core/pkg/tracing"
	"go.uber.org/zap"
)

type Network struct {
	config *config.Config

	host *host.BertyHost

	handler func(context.Context, *entity.Envelope) (*entity.Void, error)

	updating *sync.Mutex

	shutdown context.CancelFunc

	cache PeerCache
}

// Chainconfig.Options chains multiple options into a single option.
func ChainOptions(opts ...config.Option) config.Option {
	return func(cfg *config.Config) error {
		for _, opt := range opts {
			if err := opt(cfg); err != nil {
				return err
			}
		}
		return nil
	}
}

func New(ctx context.Context, opts ...config.Option) (*Network, error) {
	tracer := tracing.EnterFunc(ctx)
	defer tracer.Finish()

	ctx, cancel := context.WithCancel(ctx)

	var err error

	net := &Network{
		config:   &config.Config{},
		updating: &sync.Mutex{},
		shutdown: cancel,
		cache:    NewNoopCache(),
	}

	if err := net.config.Apply(ctx, opts...); err != nil {
		cancel()
		return nil, err
	}

	net.host, err = net.config.NewNode(ctx)
	if err != nil {
		cancel()
		return nil, err
	}

	if net.Config().PeerCache {
		net.cache = NewPeerCache(net.host)
	}

	net.init(ctx)

	return net, nil
}

func (net *Network) init(ctx context.Context) {
	net.host.SetStreamHandler(ProtocolID, net.handleEnvelope)
	net.logHostInfos()

	// bootstrap default peers
	// TOOD: infinite bootstrap + don't permit routing to provide when no peers are discovered
	if err := net.Bootstrap(ctx, false, net.config.Bootstrap...); err != nil {
		logger().Error(err.Error())
	}
}

// Update create new network and permit to override previous config
func (net *Network) Update(ctx context.Context, opts ...config.Option) error {
	net.updating.Lock()
	defer net.updating.Unlock()

	ctx, cancel := context.WithCancel(ctx)

	var err error

	update := &Network{
		config:   &config.Config{},
		updating: net.updating,
		handler:  net.handler,
		shutdown: cancel,
		cache:    NewNoopCache(),
	}

	if err := update.config.Apply(ctx, append([]config.Option{WithConfig(net.config)}, opts...)...); err != nil {
		cancel()
		return err
	}

	update.host, err = update.config.NewNode(ctx)
	if err != nil {
		cancel()
		return err
	}

	if update.Config().PeerCache {
		net.cache = NewPeerCache(update.host)
	}

	net.Close(ctx)

	*net = *update

	net.init(ctx)

	return nil
}

func (net *Network) Close(ctx context.Context) error {
	tracer := tracing.EnterFunc(ctx)
	defer tracer.Finish()

	net.shutdown()

	// FIXME: save cache to speedup next connections
	var err error

	// close host
	if net.host != nil {
		err = net.host.Close()
		if err != nil {
			logger().Error("p2p close error", zap.Error(err))
		}
	}

	return nil
}
