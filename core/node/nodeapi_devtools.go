package node

import (
	"bufio"
	"compress/gzip"
	"context"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"berty.tech/core/pkg/errorcodes"

	"github.com/brianvoe/gofakeit"
	"github.com/gogo/protobuf/proto"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"

	"berty.tech/core"
	"berty.tech/core/api/node"
	"berty.tech/core/api/p2p"
	"berty.tech/core/crypto/keypair"
	"berty.tech/core/crypto/sigchain"
	"berty.tech/core/entity"
	"berty.tech/core/pkg/deviceinfo"
	"berty.tech/core/pkg/logmanager"
	"berty.tech/core/pkg/tracing"
	"berty.tech/core/sql"
	"berty.tech/core/testrunner"
)

func (n *Node) generateFakeContact(ctx context.Context) (*entity.Contact, error) {
	var span opentracing.Span
	span, ctx = tracing.EnterFunc(ctx)
	defer span.Finish()

	var (
		username   = gofakeit.Username()
		devicename = fmt.Sprintf("%s's phone", username)
	)
	if rand.Intn(3) > 0 {
		username = fmt.Sprintf("%s %s", gofakeit.FirstName(), gofakeit.LastName())
	}

	priv, err := rsa.GenerateKey(crand.Reader, 512)
	if err != nil {
		return nil, errorcodes.ErrCryptoKeyGen.Wrap(err)
	}
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, errorcodes.ErrCryptoKeyDecode.Wrap(err)
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(priv.Public())
	if err != nil {
		return nil, errorcodes.ErrCryptoKeyDecode.Wrap(err)
	}
	kp := keypair.InsecureCrypto{}
	if err := kp.SetPrivateKeyData(privBytes); err != nil {
		return nil, errorcodes.ErrCrypto.Wrap(err)
	}
	sc := sigchain.SigChain{}
	if err := sc.Init(&kp, pubBytes); err != nil {
		return nil, errorcodes.ErrCryptoSigchain.Wrap(err)
	}
	scBytes, err := proto.Marshal(&sc)
	if err != nil {
		return nil, errorcodes.ErrCryptoSigchain.Wrap(err)
	}
	contact := &entity.Contact{
		ID:          base64.StdEncoding.EncodeToString(pubBytes),
		DisplayName: username,
		Status:      entity.Contact_Status(rand.Intn(5) + 1),
		Sigchain:    scBytes,
		Devices: []*entity.Device{
			{
				ID:         base64.StdEncoding.EncodeToString(pubBytes),
				Name:       devicename,
				Status:     entity.Device_Status(rand.Intn(3) + 1),
				ApiVersion: p2p.Version,
			},
		},
	}
	sql := n.sql(ctx)
	if err := sql.Set("gorm:association_autoupdate", true).Save(&contact).Error; err != nil {
		return nil, errorcodes.ErrDbCreate.Wrap(err)
	}
	return contact, nil
}

func (n *Node) GenerateFakeData(ctx context.Context, input *node.Void) (*node.Void, error) {
	var span opentracing.Span
	span, ctx = tracing.EnterFunc(ctx, input)
	defer span.Finish()

	// FIXME: enable mutex, but allow calling submethod, i.e., node.CreateConversation
	//n.handleMutex.Lock()
	//defer n.handleMutex.Unlock()

	contacts := []*entity.Contact{}
	for i := 0; i < 10; i++ {
		contact, err := n.generateFakeContact(ctx)
		if err != nil {
			return nil, err
		}
		contacts = append(contacts, contact)
	}

	for i := 0; i < 10; i++ {
		if err := func() error {
			var span opentracing.Span
			span, ctx = opentracing.StartSpanFromContext(ctx, "new conversation")
			defer span.Finish()

			contactsMembers := []*entity.Contact{}
			for j := 0; j < rand.Intn(2)+1; j++ {
				contactsMembers = append(contactsMembers, &entity.Contact{
					ID: contacts[rand.Intn(len(contacts))].ID,
				})
			}
			if _, err := n.conversationCreate(ctx, &node.ConversationCreateInput{
				Contacts: contactsMembers,
				Title:    strings.Title(fmt.Sprintf("%s %s", gofakeit.HipsterWord(), gofakeit.HackerNoun())),
				Topic:    gofakeit.HackerPhrase(),
			}); err != nil {
				return errorcodes.ErrDbCreate.Wrap(err)
			}
			return nil
		}(); err != nil {
			return nil, err
		}
	}

	/*
		// enqueue fake incoming event
		in := n.NewContactEvent(&entity.Contact{ID: "abcde"}, p2p.Kind_DevtoolsMapset)
		if err := n.EnqueueClientEvent(in); err != nil {
			return nil, err
		}

		// enqueue fake outgoing event
		out := n.NewContactEvent(&entity.Contact{ID: "abcde"}, p2p.Kind_DevtoolsMapset)
		if err := n.EnqueueOutgoingEvent(out); err != nil {
			return nil, err
		}
	*/

	return &node.Void{}, nil
}

func (n *Node) NodeInfos(ctx context.Context) (*deviceinfo.DeviceInfos, error) {
	span, _ := tracing.EnterFunc(ctx)
	defer span.Finish()

	infos := deviceinfo.DeviceInfos{}

	// versions
	infos.Add(deviceinfo.NewInfo("node", "versions").SetJSON(deviceinfo.VersionInfo{
		Core:    core.Everything(),
		P2PApi:  p2p.Version,
		NodeAPI: node.Version,
	}))

	// dates
	type dateInfo struct {
		At  time.Time
		Ago string
	}
	dates := map[string]dateInfo{}
	dates["node-creation"] = dateInfo{
		n.createdAt,
		fmt.Sprintf("%s", time.Since(n.createdAt)),
	}
	dates["db-creation"] = dateInfo{
		n.config.CreatedAt,
		fmt.Sprintf("%s", time.Since(n.config.CreatedAt)),
	}
	infos.Add(deviceinfo.NewInfo("node", "dates").SetJSON(dates))

	// sql
	sqlMetrics := map[string]int{}
	// FIXME: add some info about last addition date, last modification date
	db := n.sql(ctx)
	for _, table := range sql.AllTables() {
		var count int
		if err := db.Table(table).Count(&count).Error; err != nil {
			sqlMetrics[table] = -1
			// FIXME: append error
		} else {
			sqlMetrics[table] = count
		}
	}
	infos.Add(deviceinfo.NewInfo("node", "sql-entries").SetJSON(sqlMetrics))

	// queues
	infos.Add(deviceinfo.NewInfo("node", "queues").SetJSON(map[string]int{
		"client-events":      len(n.clientEvents),
		"client-subscribers": len(n.clientEventsSubscribers),
		"outgoing-events":    len(n.outgoingEvents),
	}))

	// crypto
	infos.Add(deviceinfo.NewInfo("node", "pubkey").SetString(n.b64pubkey))
	infos.Add(deviceinfo.NewInfo("node", "sigchain").SetJSON(n.sigchain))
	peer, err := n.ID(ctx, nil)
	infos.Add(deviceinfo.NewInfo("node", "peer").SetJSON(peer).SetError(err))

	// links
	infos.Add(deviceinfo.NewInfo("links", "jaeger-url").SetURL("http://jaeger.berty.io:16686/search?service=" + url.PathEscape(n.initDevice.Name+":mobile")))

	// config
	infos.Add(deviceinfo.NewInfo("node", "init-device").SetJSON(n.initDevice))

	return &infos, nil
}

func (n *Node) DeviceInfos(ctx context.Context, input *node.Void) (*deviceinfo.DeviceInfos, error) {
	var span opentracing.Span
	span, ctx = tracing.EnterFunc(ctx, input)
	defer span.Finish()

	nodeInfos, err := n.NodeInfos(ctx)
	if err != nil {
		return nil, errorcodes.ErrUndefined.Wrap(err)
	}

	runtimeInfos, err := deviceinfo.Runtime()
	if err != nil {
		return nil, err
	}

	return deviceinfo.Merge(nodeInfos, runtimeInfos), nil
}

func (n *Node) RunIntegrationTests(ctx context.Context, input *node.IntegrationTestInput) (*node.IntegrationTestOutput, error) {
	span, _ := tracing.EnterFunc(ctx, input)
	defer span.Finish()

	tests := listIntegrationTests()

	output := &node.IntegrationTestOutput{
		StartedAt: time.Now().UTC(),
		Name:      input.Name,
	}

	testFunc, ok := tests[input.Name]

	if ok == false {
		output.Success = false
		output.Verbose = "Test not found"
	} else {
		output.Success, output.Verbose = testrunner.TestRunner(input.Name, testFunc)
	}

	output.FinishedAt = time.Now().UTC()

	return output, nil
}

func (n *Node) AppVersion(ctx context.Context, input *node.Void) (*node.AppVersionOutput, error) {
	span, _ := tracing.EnterFunc(ctx, input)
	defer span.Finish()

	return &node.AppVersionOutput{Version: core.Version}, nil
}

func (n *Node) TestPanic(ctx context.Context, input *node.Void) (*node.Void, error) {
	span, _ := tracing.EnterFunc(ctx, input)
	defer span.Finish()

	panic("panic from client")
}

func (n *Node) TestLogBackgroundError(ctx context.Context, input *node.Void) (*node.Void, error) {
	span, _ := tracing.EnterFunc(ctx, input)
	defer span.Finish()

	n.LogBackgroundError(ctx, errorcodes.ErrUndefined.Wrap(errors.New("just an error test")))
	return &node.Void{}, nil
}

func (n *Node) TestLogBackgroundWarn(ctx context.Context, input *node.Void) (*node.Void, error) {
	span, _ := tracing.EnterFunc(ctx, input)
	defer span.Finish()

	n.LogBackgroundWarn(ctx, errorcodes.ErrUndefined.Wrap(errors.New("just a warn test")))
	return &node.Void{}, nil
}

func (n *Node) TestLogBackgroundDebug(ctx context.Context, input *node.Void) (*node.Void, error) {
	span, _ := tracing.EnterFunc(ctx, input)
	defer span.Finish()

	n.LogBackgroundDebug(ctx, "just a debug test")
	return &node.Void{}, nil
}

func (n *Node) DebugRequeueEvent(ctx context.Context, input *node.EventIDInput) (*p2p.Event, error) {
	var span opentracing.Span
	span, ctx = tracing.EnterFunc(ctx, input)
	defer span.Finish()

	sql := n.sql(ctx)
	var event p2p.Event
	if err := sql.First(&event, "ID = ?", input.EventID).Error; err != nil {
		return nil, errorcodes.ErrDbNothingFound.Wrap(err)
	}

	if err := n.EventRequeue(ctx, &event); err != nil {
		return nil, errorcodes.ErrNetQueue.Wrap(err)
	}

	return &event, nil
}

func (n *Node) DebugRequeueAll(ctx context.Context, _ *node.Void) (*node.Void, error) {
	var span opentracing.Span
	span, ctx = tracing.EnterFunc(ctx)
	defer span.Finish()

	if _, err := n.EventsRetry(ctx, time.Now()); err != nil {
		return nil, errorcodes.ErrNetQueue.Wrap(err)
	}

	return &node.Void{}, nil
}

func (n *Node) LogStream(input *node.LogStreamInput, stream node.Service_LogStreamServer) error {
	span, _ := tracing.EnterFunc(stream.Context(), input)
	defer span.Finish()

	if n.ring == nil {
		return errorcodes.ErrUndefined.Wrap(fmt.Errorf("ring not configured"))
	}

	// FIXME: support Continue
	// FIXME: support LogLevel
	// FIXME: support Namespaces
	// FIXME: support Last

	r, w := io.Pipe()

	go func() {
		_, _ = n.ring.WriteTo(w)
		w.Close()
	}()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		if err := stream.Send(&node.LogEntry{
			Line: scanner.Text(),
		}); err != nil {
			return errorcodes.ErrUndefined.Wrap(err)
		}

	}
	return nil
}

func (n *Node) LogfileList(_ *node.Void, stream node.Service_LogfileListServer) error {
	span, _ := tracing.EnterFunc(stream.Context())
	defer span.Finish()

	dir := logmanager.G().LogDirectory()
	if dir == "" {
		return fmt.Errorf("no log directory configured")
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() == true {
			continue
		}
		modTime := f.ModTime()

		// creation date is dependent of the architecture, faking it for now
		createTime := modTime
		//stat := f.Sys().(*syscall.Stat_t)
		//createTime := time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))

		if err := stream.Send(&node.LogfileEntry{
			Path:      f.Name(),
			Filesize:  int32(f.Size()),
			CreatedAt: &createTime,
			UpdatedAt: &modTime,
		}); err != nil {
			return errorcodes.ErrUndefined.Wrap(err)
		}
	}
	return nil
}

func (n *Node) LogfileRead(input *node.LogfileReadInput, stream node.Service_LogfileReadServer) error {
	span, _ := tracing.EnterFunc(stream.Context(), input)
	defer span.Finish()

	dir := logmanager.G().LogDirectory()
	if dir == "" {
		return fmt.Errorf("no log directory configured")
	}

	var reader io.Reader

	filepath := path.Join(dir, input.Path)
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()
	reader = file

	if strings.HasSuffix(input.Path, ".gz") {
		zfile, err := gzip.NewReader(file)
		if err != nil {
			return err
		}
		defer zfile.Close()
		reader = zfile
	}

	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		if err := stream.Send(&node.LogEntry{
			Line: scanner.Text(),
		}); err != nil {
			return err
		}
	}

	return scanner.Err()
}
