package entity

import (
	"berty.tech/core/pkg/errorcodes"
	"go.uber.org/zap"
)

func (c *Config) Validate() error {
	if c == nil {
		err := errorcodes.ErrCfgMissing.New()
		logger().Warn("Config.Validate", zap.Error(err))
		return err
	}

	if c.Myself == nil {
		err := errorcodes.ErrCfgMyself.New()
		logger().Warn("Config.Validate", zap.Error(err))
		return err
	}

	if len(c.Myself.Devices) < 1 {
		err := errorcodes.ErrCfgDevices.New()
		logger().Warn("Config.Validate", zap.Error(err))
		return err
	}

	return nil
}

func (c *Config) Filtered() *Config {
	return &Config{
		ID:              c.ID,
		PushRelayIDAPNS: c.PushRelayIDAPNS,
		PushRelayIDFCM:  c.PushRelayIDFCM,
		UpdatedAt:       c.UpdatedAt,
		CreatedAt:       c.CreatedAt,
	}
}
