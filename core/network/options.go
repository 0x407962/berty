package network

import (
	"runtime"

	"berty.tech/core/network/config"
)

// Default options

// WithConfig
func WithConfig(override *config.Config) config.Option {
	return func(cfg *config.Config) error {
		cfg.MDNS = override.MDNS
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
}

func WithDefaultOptions() config.Option {
	return ChainOptions(
		EnableDefaultBootstrap(),
		EnablePing(),
		// EnableMDNS(),
		// EnableBLE(),
		EnableQUIC(),
		EnablePrivateNetwork(config.DefaultSwarmKey),
		DisableDHT(),
		EnableMetric(),
		DisableHOP(),
	)
}

func WithDefaultMobileOptions() config.Option {
	return ChainOptions(
		WithDefaultOptions(),
		DisableDHT(),
		DisableHOP(),
		// ...
	)
}

func WithDefaultRelayOptions() config.Option {
	return ChainOptions(
		WithDefaultOptions(),
		EnableHOP(),
		EnableDHT(),
		DisableMDNS(),
	)
}

// Custom options

func WithIdentity(identity string) config.Option {
	return func(cfg *config.Config) error {
		cfg.Identity = identity
		return nil
	}
}

func EnablePrivateNetwork(swarmKey string) config.Option {
	return func(cfg *config.Config) error {
		cfg.SwarmKey = swarmKey
		return nil
	}
}

func DisablePrivateNetwork(swarmKey string) config.Option {
	return func(cfg *config.Config) error {
		cfg.SwarmKey = ""
		return nil
	}
}

func EnableHOP() config.Option {
	return func(cfg *config.Config) error {
		cfg.HOP = true
		return nil
	}
}

func DisableHOP() config.Option {
	return func(cfg *config.Config) error {
		cfg.HOP = false
		return nil
	}
}

func EnableMDNS() config.Option {
	return func(cfg *config.Config) error {
		cfg.MDNS = true
		return nil
	}
}

func DisableMDNS() config.Option {
	return func(cfg *config.Config) error {
		cfg.MDNS = false
		return nil
	}
}

func EnableDefaultBootstrap() config.Option {
	return func(cfg *config.Config) error {
		cfg.DefaultBootstrap = true
		return nil
	}
}

func DisableDefaultBootstrap() config.Option {
	return func(cfg *config.Config) error {
		cfg.DefaultBootstrap = false
		return nil
	}
}

func Bootstrap(addr ...string) config.Option {
	return func(cfg *config.Config) error {
		cfg.Bootstrap = append(cfg.Bootstrap, addr...)
		return nil
	}
}

func EnableDHT() config.Option {
	return func(cfg *config.Config) error {
		cfg.DHT = true
		return nil
	}
}

func DisableDHT() config.Option {
	return func(cfg *config.Config) error {
		cfg.DHT = false
		return nil
	}
}

func EnablePing() config.Option {
	return func(cfg *config.Config) error {
		cfg.Ping = true
		return nil
	}
}

func DisablePing() config.Option {
	return func(cfg *config.Config) error {
		cfg.Ping = false
		return nil
	}
}

func EnableMetric() config.Option {
	return func(cfg *config.Config) error {
		cfg.Metric = true
		return nil
	}
}

func DisableMetric() config.Option {
	return func(cfg *config.Config) error {
		cfg.Metric = false
		return nil
	}
}

func EnableBLE() config.Option {
	return func(cfg *config.Config) error {
		if runtime.GOOS != "android" && runtime.GOOS != "darwin" {
			logger().Warn("ble not available on your platform: disabled")
			cfg.BLE = false
			return nil
		}
		cfg.BLE = true
		return nil
	}
}

func DisableBLE() config.Option {
	return func(cfg *config.Config) error {
		cfg.BLE = false
		return nil
	}
}

func EnableQUIC() config.Option {
	return func(cfg *config.Config) error {
		cfg.QUIC = true
		return nil
	}
}

func DisableQUIC() config.Option {
	return func(cfg *config.Config) error {
		cfg.QUIC = false
		return nil
	}
}
