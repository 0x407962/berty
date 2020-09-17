package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"berty.tech/berty/v2/go/pkg/bertymessenger"
	"berty.tech/berty/v2/go/pkg/errcode"
	"github.com/peterbourgon/ff/v3/ffcli"
	"moul.io/godev"
)

func systemInfoCommand() *ffcli.Command {
	var (
		fs               = flag.NewFlagSet("info", flag.ExitOnError)
		refreshEveryFlag time.Duration
	)
	manager.SetupLocalMessengerServerFlags(fs) // by default, start a new local messenger server,
	manager.SetupRemoteNodeFlags(fs)           // but allow to set a remote server instead
	fs.DurationVar(&refreshEveryFlag, "info.refresh", refreshEveryFlag, "refresh every DURATION (0: no refresh)")

	return &ffcli.Command{
		Name:       "info",
		ShortUsage: "berty [global flags] info [flags]",
		ShortHelp:  "display system info",
		FlagSet:    fs,
		Exec: func(ctx context.Context, args []string) error {
			// messenger client
			messenger, err := manager.GetMessengerClient()
			if err != nil {
				return err
			}

			for {
				ret, err := messenger.SystemInfo(ctx, &bertymessenger.SystemInfo_Request{})
				if err != nil {
					return errcode.TODO.Wrap(err)
				}

				if ret.Messenger.ProtocolInSameProcess {
					ret.Messenger.Process = nil
				}

				if refreshEveryFlag == 0 {
					fmt.Println(godev.PrettyJSONPB(ret))
					break
				}
				/// clear screen
				print("\033[H\033[2J")
				fmt.Println(godev.PrettyJSONPB(ret))
				time.Sleep(refreshEveryFlag)
			}

			return nil
		},
	}
}
