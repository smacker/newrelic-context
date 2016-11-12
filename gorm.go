package nrcontext

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/newrelic/go-agent"
)

const (
	txnGormKey   = "newrelicTransaction"
	startTimeKey = "newrelicStartTime"
)

// Sets transaction from Context to gorm settings, returns cloned DB
func SetTxnToGorm(ctx context.Context, db *gorm.DB) *gorm.DB {
	txn := GetTnxFromContext(ctx)
	if txn == nil {
		return db
	}
	return db.Set(txnGormKey, txn)
}

// Adds callbacks to NewRelic, you should call SetTxnToGorm to make them work
func AddGormCallbacks(db *gorm.DB) {
	dialect := db.Dialect().GetName()
	var product newrelic.DatastoreProduct
	switch dialect {
	case "postgres":
		product = newrelic.DatastorePostgres
	case "mysql":
		product = newrelic.DatastoreMySQL
	case "sqlite3":
		product = newrelic.DatastoreSQLite
	case "mssql":
		product = newrelic.DatastoreMSSQL
	default:
		return
	}
	callbacks := newCallbacks(product)
	registerCallbacks(db, "create", callbacks)
	registerCallbacks(db, "query", callbacks)
	registerCallbacks(db, "update", callbacks)
	registerCallbacks(db, "delete", callbacks)
}

type callbacks struct {
	product newrelic.DatastoreProduct
}

func newCallbacks(product newrelic.DatastoreProduct) *callbacks {
	return &callbacks{product}
}

func (c *callbacks) beforeCreate(scope *gorm.Scope) { c.before(scope) }
func (c *callbacks) afterCreate(scope *gorm.Scope)  { c.after(scope, "INSERT") }
func (c *callbacks) beforeQuery(scope *gorm.Scope)  { c.before(scope) }
func (c *callbacks) afterQuery(scope *gorm.Scope)   { c.after(scope, "SELECT") }
func (c *callbacks) beforeUpdate(scope *gorm.Scope) { c.before(scope) }
func (c *callbacks) afterUpdate(scope *gorm.Scope)  { c.after(scope, "UPDATE") }
func (c *callbacks) beforeDelete(scope *gorm.Scope) { c.before(scope) }
func (c *callbacks) afterDelete(scope *gorm.Scope)  { c.after(scope, "DELETE") }

func (c *callbacks) before(scope *gorm.Scope) {
	txn, ok := scope.Get(txnGormKey)
	if !ok {
		return
	}
	scope.Set(startTimeKey, newrelic.StartSegmentNow(txn.(newrelic.Transaction)))
}

func (c *callbacks) after(scope *gorm.Scope, operation string) {
	if scope.HasError() {
		return
	}
	startTime, ok := scope.Get(startTimeKey)
	if !ok {
		return
	}
	newrelic.DatastoreSegment{
		StartTime:          startTime.(newrelic.SegmentStartTime),
		Product:            c.product,
		ParameterizedQuery: scope.SQL,
		Operation:          operation,
		Collection:         scope.TableName(),
	}.End()
	// It's too difficult to mock
	// uncomment it for tests
	// fmt.Println("Product:", c.product)
	// fmt.Println("ParameterizedQuery:", scope.SQL)
	// fmt.Println("Operation:", operation)
	// fmt.Println("Collection:", scope.TableName())
}

func registerCallbacks(db *gorm.DB, name string, c *callbacks) {
	beforeName := fmt.Sprintf("newrelic:%v_before", name)
	afterName := fmt.Sprintf("newrelic:%v_after", name)
	gormCallbackName := fmt.Sprintf("gorm:%v", name)
	// gorm does some magic, if you pass CallbackProcessor here - nothing works
	switch name {
	case "create":
		db.Callback().Create().Before(gormCallbackName).Register(beforeName, c.beforeCreate)
		db.Callback().Create().After(gormCallbackName).Register(afterName, c.afterCreate)
	case "query":
		db.Callback().Query().Before(gormCallbackName).Register(beforeName, c.beforeQuery)
		db.Callback().Query().After(gormCallbackName).Register(afterName, c.afterQuery)
	case "update":
		db.Callback().Update().Before(gormCallbackName).Register(beforeName, c.beforeUpdate)
		db.Callback().Update().After(gormCallbackName).Register(afterName, c.afterUpdate)
	case "delete":
		db.Callback().Delete().Before(gormCallbackName).Register(beforeName, c.beforeDelete)
		db.Callback().Delete().After(gormCallbackName).Register(afterName, c.afterDelete)
	}
}
