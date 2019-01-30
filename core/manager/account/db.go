package account

import (
	"encoding/json"
	"fmt"

	"berty.tech/core/pkg/deviceinfo"
	"berty.tech/core/pkg/tracing"
	"github.com/jinzhu/gorm"
	opentracing "github.com/opentracing/opentracing-go"
)

//
// state db
//

type StateDB struct {
	Gorm *gorm.DB `gorm:"-"`
	gorm.Model

	StartCounter int

	JSONNetConf string
	BotMode     bool
	LocalGRPC   bool
}

func OpenStateDB(path string, initialState StateDB) (*StateDB, error) {
	// open db
	db, err := gorm.Open("sqlite3", deviceinfo.GetStoragePath()+"/"+path)
	if err != nil {
		return nil, err
	}

	// create tables and columns
	if err := db.AutoMigrate(&StateDB{}).Error; err != nil {
		return nil, err
	}

	// preload last state
	var state StateDB
	if err := db.Last(&state).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err == gorm.ErrRecordNotFound {
		state = initialState
		if err := db.Save(&state).Error; err != nil {
			return nil, err
		}
	}

	state.Gorm = db

	return &state, nil
}

func (state StateDB) String() string {
	out, _ := json.Marshal(state)
	return string(out)
}

func (state *StateDB) Save() error {
	return state.Gorm.Save(state).Error
}

func (state *StateDB) Create() error {
	return state.Gorm.Create(state).Error
}

func (state *StateDB) Close() {
	state.Gorm.Close()
}

//
// app db
//

func gormCreateSubSpan(scope *gorm.Scope, operationName string) {
	if span, ok := scope.Get("rootSpan"); ok {
		rootSpan := span.(opentracing.Span)
		operationName = fmt.Sprintf("gorm::%s", operationName)
		subSpan := rootSpan.Tracer().StartSpan(
			operationName,
			opentracing.ChildOf(rootSpan.Context()),
		)
		subSpan.SetTag("component", "gorm")
		scope.Set("subSpan", subSpan)
	}
}

func gormFinishSubSpan(scope *gorm.Scope) {
	if span, ok := scope.Get("subSpan"); ok {
		subSpan := span.(opentracing.Span)
		defer subSpan.Finish()
		subSpan.LogKV("sql", scope.SQL)
	}
}

func WithDatabase(opts *DatabaseOptions) NewOption {
	return func(a *Account) error {
		tracer := tracing.EnterFunc(a.rootContext, opts)
		defer tracer.Finish()
		ctx := tracer.Context()

		if opts == nil {
			opts = &DatabaseOptions{}
		}

		a.dbDir = deviceinfo.GetStoragePath() + "/" + opts.Path

		a.dbDrop = opts.Drop
		if err := a.openDatabase(ctx); err != nil {
			return err
		}

		if a.tracer != nil {
			// create
			a.db.Callback().Create().Before("gorm:before_create").Register("jaeger:before_create", func(scope *gorm.Scope) { gormCreateSubSpan(scope, fmt.Sprintln("insert", scope.TableName())) })
			a.db.Callback().Create().Before("gorm:after_create").Register("jaeger:after_create", func(scope *gorm.Scope) { gormFinishSubSpan(scope) })

			// update
			a.db.Callback().Update().Before("gorm:before_update").Register("jaeger:before_update", func(scope *gorm.Scope) { gormCreateSubSpan(scope, fmt.Sprintln("update", scope.TableName())) })
			a.db.Callback().Update().Before("gorm:after_update").Register("jaeger:after_update", func(scope *gorm.Scope) { gormFinishSubSpan(scope) })

			// delete
			a.db.Callback().Delete().Before("gorm:before_delete").Register("jaeger:before_delete", func(scope *gorm.Scope) { gormCreateSubSpan(scope, fmt.Sprintln("delete", scope.TableName())) })
			a.db.Callback().Delete().Before("gorm:after_delete").Register("jaeger:after_delete", func(scope *gorm.Scope) { gormFinishSubSpan(scope) })

			// query
			a.db.Callback().Query().Before("gorm:query").Register("jaeger:before_query", func(scope *gorm.Scope) { gormCreateSubSpan(scope, fmt.Sprintln("select", scope.TableName())) })
			a.db.Callback().Query().Before("gorm:after_query").Register("jaeger:after_query", func(scope *gorm.Scope) { gormFinishSubSpan(scope) })
		}
		return nil
	}
}
