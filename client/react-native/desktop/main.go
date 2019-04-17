package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"berty.tech/client/react-native/desktop/coreinterface"
	"berty.tech/core/daemon"
	network_config "berty.tech/core/network/config"
	"berty.tech/core/pkg/deviceinfo"
	"berty.tech/core/pkg/logmanager"
	"go.uber.org/zap"

	astilectron "github.com/asticode/go-astilectron"
	bootstrap "github.com/asticode/go-astilectron-bootstrap"
	astilog "github.com/asticode/go-astilog"
	"github.com/shibukawa/configdir"
)

// Vars
var (
	AppName  string
	BuiltAt  string
	debug    = astilectron.PtrBool(true) // flag.Bool("d", false, "enables the debug mode")
	homepage = flag.String("h", "index.html", "overrides default resource url (useful when having a local dev web build)")
)

func getStorageDir() (string, error) {
	storagePath := configdir.New("Berty Technologies", "Berty")
	storageDirs := storagePath.QueryFolders(configdir.Global)
	if len(storageDirs) == 0 {
		return "", errors.New("no storage path found")
	}

	if err := storageDirs[0].CreateParentDir(""); err != nil {
		return "", err
	}

	return storageDirs[0].Path, nil
}

func main() {
	storagePath, err := getStorageDir()
	if err != nil {
		panic(err)
	}

	if err = deviceinfo.SetStoragePath(storagePath); err != nil {
		panic(err)
	}

	fmt.Println(storagePath)
	sqlConfig := &daemon.SQLConfig{
		Path: fmt.Sprintf("%s/%s", storagePath, "berty.state.db"),
		Key:  "s3cur3",
	}

	config := &daemon.Config{
		SqlOpts:          sqlConfig,
		GrpcBind:         ":1337",
		GqlBind:          ":1338",
		HideBanner:       true,
		DropDatabase:     false,
		InitOnly:         false,
		WithBot:          false,
		Notification:     true,
		ApnsCerts:        []string{},
		ApnsDevVoipCerts: []string{},
		FcmAPIKeys:       []string{},
		PrivateKeyFile:   "",
		PeerCache:        true,
		Identity:         "",
		Bootstrap:        network_config.DefaultBootstrap,
		NoP2P:            false,
		BindP2P:          []string{},
		TransportP2P:     []string{},
		Hop:              true,
		Ble:              true,
		Mdns:             true,
		DhtServer:        true,
		PrivateNetwork:   true,
		SwarmKeyPath:     "",
	}

	// Init
	flag.Parse()

	t := true
	logman, err := logmanager.New(logmanager.Opts{
		RingSize:      10 * 1024 * 1024,
		LogLevel:      "debug",
		LogNamespaces: "core.*,vendor.gorm*,client.react-native.*",
		LogDirectory:  os.Getenv("HOME") + "/Library/Logs", // FIXME: win, linux
	})
	if err != nil {
		panic(err)
	}

	logman.SetGlobal()

	zap.L().Debug("Berty desktop client started")
	astilog.SetDefaultLogger()

	homepageUrl := "index.html"
	if homepage != nil {
		homepageUrl = *homepage
	}

	startRequest := &daemon.StartRequest{
		Nickname: "daemon-desktop",
	}

	d, err := NewDaemonDesktop()
	if err != nil {
		panic(err)
	}

	if err := d.Initialize(context.Background(), config); err != nil {
		panic(err)
	}

	if _, err := d.bridge.Start(context.Background(), startRequest); err != nil {
		panic(err)
	}

	zap.L().Debug("Berty desktop client started")

	// Run bootstrap
	logger().Debug(fmt.Sprintf("Running app built at %s", BuiltAt))
	if err := bootstrap.Run(bootstrap.Options{
		Adapter:       coreinterface.SetAstilectron,
		Asset:         Asset,
		AssetDir:      AssetDir,
		RestoreAssets: RestoreAssets,
		AstilectronOptions: astilectron.Options{
			AppName:            AppName,
			AppIconDarwinPath:  "resources/icon.icns",
			AppIconDefaultPath: "resources/icon.png",
		},
		Debug: *debug,
		MenuOptions: []*astilectron.MenuItemOptions{{
			Label: astilectron.PtrStr("File"),
			SubMenu: []*astilectron.MenuItemOptions{
				{Role: astilectron.MenuItemRoleClose},
				{Role: astilectron.MenuItemRoleQuit, Label: astilectron.PtrStr("Quit Berty")},
			},
		},
			{
				Role: astilectron.MenuItemRoleEditMenu,
				SubMenu: []*astilectron.MenuItemOptions{
					{Role: astilectron.MenuItemRoleUndo},
					{Role: astilectron.MenuItemRoleRedo},
					{Role: astilectron.MenuItemRoleCut},
					{Role: astilectron.MenuItemRoleCopy},
					{Role: astilectron.MenuItemRolePaste},
					{Role: astilectron.MenuItemRoleDelete},
					{Role: astilectron.MenuItemRoleSelectAll},
				},
			}},
		OnWait: coreinterface.SetNotificationDriver,
		Windows: []*bootstrap.Window{{
			Homepage:       homepageUrl,
			MessageHandler: d.handleMessages,
			Options: &astilectron.WindowOptions{
				BackgroundColor: astilectron.PtrStr("#333"),
				Width:           astilectron.PtrInt(1060),
				Height:          astilectron.PtrInt(640),
				WebPreferences: &astilectron.WebPreferences{
					DevTools: &t,
				},
			},
		}},
	}); err != nil {
		logger().Error("running bootstrap failed", zap.Error(err))
	}
}
