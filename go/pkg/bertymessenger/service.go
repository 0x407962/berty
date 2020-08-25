package bertymessenger

import (
	"context"
	"errors"
	"time"

	"berty.tech/berty/v2/go/pkg/bertyprotocol"
	"berty.tech/berty/v2/go/pkg/bertytypes"
	"berty.tech/berty/v2/go/pkg/errcode"

	"github.com/golang/protobuf/proto" // nolint:staticcheck: not sure how to use the new protobuf api to unmarshal
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"moul.io/u"
	"moul.io/zapgorm2"
)

func New(client bertyprotocol.ProtocolServiceClient, opts *Opts) (Service, error) {
	optsCleanup, err := opts.applyDefaults()
	if err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	if err := initDB(opts.DB); err != nil {
		return nil, errcode.TODO.Wrap(err)
	}

	ctx, cancel := context.WithCancel(opts.Context)
	svc := service{
		protocolClient:  client,
		logger:          opts.Logger,
		startedAt:       time.Now(),
		protocolService: opts.ProtocolService,
		db:              opts.DB,
		dispatcher:      NewDispatcher(),
		cancelFn:        cancel,
		optsCleanup:     optsCleanup,
		ctx:             ctx,
	}

	icr, err := client.InstanceGetConfiguration(ctx, &bertytypes.InstanceGetConfiguration_Request{})
	if err != nil {
		return nil, err
	}
	pkStr := bytesToString(icr.GetAccountGroupPK())

	// get or create account in DB
	{
		var acc Account
		err := svc.db.First(&acc).Error
		switch {
		case err == gorm.ErrRecordNotFound: // account not found, create a new one
			svc.logger.Debug("account not found, creating a new one", zap.String("pk", pkStr))
			acc.State = Account_NotReady
			ret, err := svc.InstanceShareableBertyID(ctx, &InstanceShareableBertyID_Request{})
			if err != nil {
				return nil, err
			}
			acc.Link = ret.GetHTMLURL()
			acc.PublicKey = pkStr
			if err := svc.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&acc).Error; err != nil {
				return nil, err
			}
		case err != nil: // internal error
			return nil, err
		case err == nil && pkStr != acc.GetPublicKey(): // Check that we are connected to the correct node
			// FIXME: use errcode
			return nil, errors.New("messenger's account key does not match protocol's account key")
		default: // account exists, and public keys match
			// noop
		}
	}

	// Subscribe to account group metadata
	if err := svc.subscribeToMetadata(ctx, icr.GetAccountGroupPK()); err != nil {
		return nil, err
	}

	// subscribe to groups
	{
		var convs []Conversation
		err := svc.db.Find(&convs).Error
		if err != nil {
			return nil, err
		}
		for _, cv := range convs {
			gpk := cv.GetPublicKey()
			gpkb, err := stringToBytes(gpk)
			if err != nil {
				return nil, err
			}

			if err := svc.subscribeToMetadata(ctx, gpkb); err != nil {
				return nil, err
			}

			ms, err := svc.protocolClient.GroupMessageSubscribe(svc.ctx, &bertytypes.GroupMessageSubscribe_Request{GroupPK: gpkb})
			if err != nil {
				return nil, err
			}
			go func() {
				for {
					gme, err := ms.Recv()
					if err != nil {
						svc.logStreamingError("group message", err)
						return
					}

					var am AppMessage
					if err := proto.Unmarshal(gme.GetMessage(), &am); err != nil {
						svc.logger.Warn("failed to unmarshal AppMessage", zap.Error(err))
						return
					}
					err = handleAppMessage(&svc, gpk, gme, &am)
					if err != nil {
						svc.logger.Error("failed to handle app message", zap.Error(errcode.ErrInternal.Wrap(err)))
					}
				}
			}()
		}
	}

	// subscribe to group metadata for contacts in outgoing request sent state
	{
		var contacts []Contact
		err := svc.db.Find(&contacts).Error
		if err != nil {
			return nil, err
		}
		for _, c := range contacts {
			if c.State != Contact_OutgoingRequestSent {
				continue
			}
			gpk := c.GetConversationPublicKey()
			gpkb, err := stringToBytes(gpk)
			if err != nil {
				return nil, err
			}

			if err := svc.subscribeToMetadata(ctx, gpkb); err != nil {
				return nil, err
			}
		}
	}

	return &svc, nil
}

func (svc *service) subscribeToMetadata(ctx context.Context, gpkb []byte) error {
	s, err := svc.protocolClient.GroupMetadataSubscribe(ctx, &bertytypes.GroupMetadataSubscribe_Request{GroupPK: gpkb})
	if err != nil {
		return err
	}
	go func() {
		for {
			gme, err := s.Recv()
			if err != nil {
				svc.logStreamingError("group metadata", err)
				return
			}
			err = handleProtocolEvent(svc, gme)
			if err != nil {
				svc.logger.Error("failed to handle protocol event", zap.Error(errcode.ErrInternal.Wrap(err)))
			}
		}
	}()
	return nil
}

func (svc *service) Close() {
	svc.logger.Info("closing service")
	svc.cancelFn()
	svc.optsCleanup()
}

type Opts struct {
	Logger          *zap.Logger
	ProtocolService bertyprotocol.Service
	DB              *gorm.DB
	Context         context.Context
}

func (opts *Opts) applyDefaults() (func(), error) {
	cleanup := func() {}

	if opts.Context == nil {
		opts.Context = context.Background()
	}
	if opts.Logger == nil {
		opts.Logger = zap.NewNop()
	}
	if opts.DB == nil {
		opts.Logger.Warn("Messenger started without database, creating a volatile one in memory")
		zapLogger := zapgorm2.New(opts.Logger)
		zapLogger.SetAsDefault()
		db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{Logger: zapLogger})
		if err != nil {
			return nil, err
		}
		cleanup = func() {
			sqlDB, _ := db.DB()
			u.SilentClose(sqlDB)
		}
		opts.DB = db
	}

	return cleanup, nil
}

type service struct {
	logger          *zap.Logger
	protocolClient  bertyprotocol.ProtocolServiceClient
	startedAt       time.Time
	protocolService bertyprotocol.Service // optional, for debugging only
	db              *gorm.DB
	dispatcher      *Dispatcher
	cancelFn        func()
	optsCleanup     func()
	ctx             context.Context
}

type Service interface {
	MessengerServiceServer
	Close()
}

var _ Service = (*service)(nil)
