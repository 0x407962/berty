package main

import (
	"context"
	"flag"

	ff "github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
	"go.uber.org/zap"

	"berty.tech/berty/v2/go/pkg/bertytypes"
	"berty.tech/berty/v2/go/pkg/errcode"
)

func daemonCommand() *ffcli.Command {
	fsBuilder := func() (*flag.FlagSet, error) {
		fs := flag.NewFlagSet("berty daemon", flag.ExitOnError)
		manager.SetupLoggingFlags(fs)              // also available at root level
		manager.SetupLocalMessengerServerFlags(fs) // we want to configure a local messenger server
		manager.SetupDefaultGRPCListenersFlags(fs)
		manager.SetupMetricsFlags(fs)
		return fs, nil
	}

	return &ffcli.Command{
		Name:           "daemon",
		ShortUsage:     "berty [global flags] daemon [flags]",
		ShortHelp:      "start a full Berty instance (Berty Protocol + Berty Messenger)",
		Options:        []ff.Option{ff.WithEnvVarPrefix("BERTY")},
		FlagSetBuilder: fsBuilder,
		UsageFunc:      usageFunc,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) > 0 {
				return flag.ErrHelp
			}

			logger, err := manager.GetLogger()
			if err != nil {
				return err
			}

			// since this command is daemon, we want to be sure to run a local daemon with protocol and messenger
			{
				_, err := manager.GetLocalProtocolServer()
				if err != nil {
					return err
				}
				_, err = manager.GetLocalMessengerServer()
				if err != nil {
					return err
				}
			}

			// connect to the local client
			{
				protocolClient, err := manager.GetProtocolClient()
				if err != nil {
					return err
				}
				info, err := protocolClient.InstanceGetConfiguration(ctx, &bertytypes.InstanceGetConfiguration_Request{})
				if err != nil {
					return errcode.TODO.Wrap(err)
				}
				logger.Named("main").Info("daemon initialized", zap.String("peer-id", info.PeerID), zap.Strings("listeners", info.Listeners))
			}

			return manager.RunWorkers()
		},
	}
}
