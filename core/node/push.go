package node

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"

	"berty.tech/core/api/p2p"
	"berty.tech/core/entity"
	"berty.tech/core/pkg/deviceinfo"
	"berty.tech/core/pkg/errorcodes"
	"berty.tech/core/push"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func WithPushManager(pushManager *push.Manager) NewNodeOption {
	return func(n *Node) {
		n.pushManager = pushManager
	}
}

func WithPushTokenSubscriber() NewNodeOption {
	return func(n *Node) {
		var err error
		ctx := context.Background()

		packageID := deviceinfo.PackageID()

		go func() {
			tokenSubscription := n.notificationDriver.SubscribeToken()

			for {
				select {
				case token := <-tokenSubscription:
					{
						currentToken := &entity.DevicePushConfig{}

						if err = n.sql(ctx).First(&currentToken, &entity.DevicePushConfig{PushType: token.Type}).Error; err != nil {
							logger().Info("unable to get push token", zap.Error(err))
						}

						pushID := &push.PushNativeIdentifier{
							PackageID:   packageID,
							DeviceToken: token.Hash(),
						}

						pushIDBytes, err := pushID.Marshal()
						if err != nil {
							logger().Error("unable to serialize push id", zap.Error(err))
							continue
						}

						if len(token.Value) > 0 && bytes.Compare(currentToken.PushID, pushIDBytes) == 0 {
							continue
						}

						if len(currentToken.PushID) > 0 {
							_, err = n.DevicePushConfigRemove(ctx, currentToken)

							logger().Error("unable to delete existing push token", zap.Error(err))
						}

						if len(token.Value) > 0 {
							_, err = n.DevicePushConfigCreate(ctx, &entity.DevicePushConfig{
								RelayID:  []byte{},
								PushID:   pushIDBytes,
								PushType: token.Type,
							})

							logger().Error("unable to create push token", zap.Error(err))
						}
					}
				case <-n.shutdown:
					logger().Debug("node push token subscriber shutdown")
					n.notificationDriver.UnsubscribeToken(tokenSubscription)
				}
			}
		}()
	}
}

func WithPushNotificationSubscriber() NewNodeOption {
	return func(n *Node) {
		ctx := context.Background()
		go func() {
			notificationSubscription := n.notificationDriver.Subscribe()

			for {
				select {
				case notification := <-notificationSubscription:
					{
						logger().Debug("node receive push notification")

						payload := push.Payload{}
						if err := json.Unmarshal([]byte(notification.Body), &payload); err != nil {
							logger().Error(errorcodes.ErrNodePushNotifSub.Wrap(err).Error())
							continue
						}

						b64Envelope := payload.BertyEnvelope
						if b64Envelope == "" {
							logger().Error(errorcodes.ErrNodePushNotifSub.Wrap(errors.New("berty-envelope is missing")).Error())
							continue
						}

						bytesEnvelope, err := base64.StdEncoding.DecodeString(string(b64Envelope))
						if err != nil {
							logger().Error(errorcodes.ErrNodePushNotifSub.Wrap(err).Error())
							continue
						}

						envelope := &p2p.Envelope{}
						if err := envelope.Unmarshal(bytesEnvelope); err != nil {
							logger().Error(errorcodes.ErrNodePushNotifSub.Wrap(err).Error())
							continue
						}

						if err := n.handleEnvelope(ctx, envelope); err != nil {
							logger().Error(errorcodes.ErrNodePushNotifSub.Wrap(err).Error())
							continue
						}
					}
				case <-n.shutdown:
					logger().Debug("node push notification subscriber shutdown")
					n.notificationDriver.Unsubscribe(notificationSubscription)
				}
			}
		}()
	}
}
