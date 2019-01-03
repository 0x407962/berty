package node

import (
	"context"
	"fmt"
	"reflect"

	"berty.tech/core/api/node"
	"berty.tech/core/api/p2p"
	"berty.tech/core/entity"
	"berty.tech/core/sql"
	"github.com/jinzhu/gorm"
	opentracing "github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

// WithSQL registers a gorm connection as the node database
func WithSQL(sql *gorm.DB) NewNodeOption {
	return func(n *Node) {
		n.sqlDriver = sql.Set("gorm:auto_preload", true).Unscoped()
		sql.Callback().Create().Register("berty:after_create", func(scope *gorm.Scope) { n.handleCommitLog("create", scope) })
		sql.Callback().Update().Register("berty:after_update", func(scope *gorm.Scope) { n.handleCommitLog("update", scope) })
		sql.Callback().Delete().Register("berty:after_delete", func(scope *gorm.Scope) { n.handleCommitLog("delete", scope) })
	}
}

// sql returns a gorm.DB object with opentracing context
func (n *Node) sql(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return n.sqlDriver.Set("rootSpan", n.rootSpan)
	}
	if span := opentracing.SpanFromContext(ctx); span != nil {
		return n.sqlDriver.Set("rootSpan", span)
	}
	return n.sqlDriver
}

func (n *Node) handleCommitLog(operation string, scope *gorm.Scope) {
	// same usage as https://github.com/jinzhu/gorm/blob/master/scope.go#L241

	if scope.HasError() {
		return
	}

	if scope.Value == nil {
		return
	}

	if indirectScopeValue := scope.IndirectValue(); indirectScopeValue.Kind() == reflect.Slice {
		for i := 0; i < indirectScopeValue.Len(); i++ {
			n.sendCommitLog(n.createCommitLog(operation, indirectScopeValue.Index(i)))
		}
	} else {
		n.sendCommitLog(n.createCommitLog(operation, indirectScopeValue))
	}
}

func (n *Node) sendCommitLog(commitLog *node.CommitLog) {
	if commitLog == nil {
		return
	}

	logger().Debug("commit log", zap.Stringer("commit log", commitLog))

	n.clientCommitLogsMutex.Lock()
	defer n.clientCommitLogsMutex.Unlock()
	for _, sub := range n.clientCommitLogsSubscribers {
		sub.queue <- commitLog
	}
}

func (n *Node) createCommitLog(operation string, reflectValue reflect.Value) *node.CommitLog {
	var err error

	// Only get address from non-pointer
	if reflectValue.CanAddr() && reflectValue.Kind() != reflect.Ptr {
		reflectValue = reflectValue.Addr()
	}

	log := &node.CommitLog{}

	switch operation {
	case "create":
		log.Operation = node.CommitLog_Create
	case "update":
		log.Operation = node.CommitLog_Update
	case "delete":
		log.Operation = node.CommitLog_Delete
	default:
		logger().Warn(fmt.Sprintf("undefined operation %+v", operation))
		return nil
	}

	switch data := reflectValue.Interface().(type) {
	case *entity.Contact:
		if operation != "delete" {
			data, err = sql.ContactByID(n.sqlDriver, data.ID)
			if err != nil {
				return nil
			}
		}
		log.Entity = &node.CommitLog_Entity{Contact: data}
	case *entity.Device:
		log.Entity = &node.CommitLog_Entity{Device: data}
	case *entity.Conversation:
		if operation != "delete" {
			data, err = sql.ConversationByID(n.sqlDriver, data.ID)
			if err != nil {
				return nil
			}
		}
		log.Entity = &node.CommitLog_Entity{Conversation: data}
	case *entity.ConversationMember:
		log.Entity = &node.CommitLog_Entity{
			ConversationMember: data,
		}
	case *p2p.Event:
		log.Entity = &node.CommitLog_Entity{Event: data}
	default:
		logger().Warn(fmt.Sprintf("unhandled entity %+v", data))
		return nil
	}
	return log
}
