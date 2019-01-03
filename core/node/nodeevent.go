package node

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"go.uber.org/zap"

	"berty.tech/core/api/node"
	"berty.tech/core/pkg/tracing"
)

func (n *Node) EnqueueNodeEvent(ctx context.Context, kind node.Kind, attributes proto.Message) {
	tracer := tracing.EnterFunc(ctx, kind, attributes)
	defer tracer.Finish()
	// ctx = tracer.Context()

	event, err := node.NewEvent(kind, attributes)
	if err != nil {
		logger().Error("failed to create node.NodeStarted event")
	} else {
		n.clientEvents <- event
	}
}

func (n *Node) LogBackgroundError(ctx context.Context, err error) {
	logger().Error("background error", zap.Error(err))
	n.EnqueueNodeEvent(ctx, node.Kind_BackgroundError, &node.BackgroundErrorAttrs{
		ErrMsg: err.Error(),
	})
}

func (n *Node) LogBackgroundWarn(ctx context.Context, err error) {
	logger().Warn("background warn", zap.Error(err))
	n.EnqueueNodeEvent(ctx, node.Kind_BackgroundWarn, &node.BackgroundWarnAttrs{
		ErrMsg: err.Error(),
	})
}

func (n *Node) LogBackgroundDebug(ctx context.Context, msg string) {
	logger().Debug("background debug", zap.String("msg", msg))
	n.EnqueueNodeEvent(ctx, node.Kind_Debug, &node.DebugAttrs{
		Msg: msg,
	})
}
