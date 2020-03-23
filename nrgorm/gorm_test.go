package nrgorm

import (
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	newrelic "github.com/newrelic/go-agent"
	"github.com/smacker/newrelic-context/nrmock"
)

type Model struct {
	ID    int
	Value string
}

var db *gorm.DB
var testTxn newrelic.Transaction
var segmentsHistory []*nrmock.DatastoreSegment

func TestMain(m *testing.M) {
	var err error
	// prepare db
	os.Remove("./foo.db")
	db, err = gorm.Open("sqlite3", "./foo.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.CreateTable(&Model{}).Error; err != nil {
		panic(err)
	}
	if err := db.Create(&Model{Value: "to-select"}).Error; err != nil {
		panic(err)
	}
	AddGormCallbacks(db)

	// mock newrelic
	originalBuilder := segmentBuilder
	segmentBuilder = func(
		startTime newrelic.SegmentStartTime,
		product newrelic.DatastoreProduct,
		query string,
		operation string,
		collection string,
	) segment {
		segment := originalBuilder(startTime, product, query, operation, collection).(*newrelic.DatastoreSegment)
		mock := &nrmock.DatastoreSegment{DatastoreSegment: segment, StartTime: startTime}
		segmentsHistory = append(segmentsHistory, mock)
		return mock
	}

	app := &nrmock.NewrelicApp{}
	testTxn = app.StartTransaction("txn-name", nil, nil)

	os.Exit(m.Run())
}

func TestWrappedGorm(t *testing.T) {
	segmentsHistory = []*nrmock.DatastoreSegment{}
	txnDB := SetTxnToGorm(testTxn, db)
	dbInsert(t, txnDB)
	lastSegment := segmentsHistory[0]
	if lastSegment.Product != newrelic.DatastoreSQLite {
		t.Errorf("wrong product: %v", lastSegment.Product)
	}
	if lastSegment.ParameterizedQuery != `INSERT INTO "models" ("value") VALUES (?)` {
		t.Errorf("wrong query: %v", lastSegment.ParameterizedQuery)
	}
	if lastSegment.Operation != "INSERT" {
		t.Errorf("wrong operation: %v", lastSegment.Operation)
	}
	if lastSegment.Collection != "models" {
		t.Errorf("wrong collection: %v", lastSegment.Collection)
	}
	// gorm always tries to wrap insert into transaction
	lastSegment = segmentsHistory[1]
	if lastSegment.Operation != "COMMIT/ROLLBACK" {
		t.Errorf("wrong operation: %v", lastSegment.Operation)
	}
	if lastSegment.Collection != "models" {
		t.Errorf("wrong collection: %v", lastSegment.Collection)
	}
	// no transaction on select
	dbSelect(t, txnDB)
	// to update we have to create a row first +2 transactions
	dbUpdate(t, txnDB)
	lastSegment = segmentsHistory[5]
	if lastSegment.Operation != "UPDATE" {
		t.Errorf("wrong operation: %v", lastSegment.Operation)
	}
	// gorm always tries to wrap update into transaction
	lastSegment = segmentsHistory[6]
	if lastSegment.Operation != "COMMIT/ROLLBACK" {
		t.Errorf("wrong operation: %v", lastSegment.Operation)
	}
	if lastSegment.Collection != "models" {
		t.Errorf("wrong collection: %v", lastSegment.Collection)
	}
	// to delete we have to create a row first +2 transactions
	dbDelete(t, txnDB)
	lastSegment = segmentsHistory[9]
	if lastSegment.Operation != "DELETE" {
		t.Errorf("wrong operation: %v", lastSegment.Operation)
	}
	// gorm always tries to wrap delete into transaction
	lastSegment = segmentsHistory[10]
	if lastSegment.Operation != "COMMIT/ROLLBACK" {
		t.Errorf("wrong operation: %v", lastSegment.Operation)
	}
	if lastSegment.Collection != "models" {
		t.Errorf("wrong collection: %v", lastSegment.Collection)
	}

	dbSelectNoRecord(t, txnDB)
	lastSegment = segmentsHistory[11]
	if lastSegment.Operation != "SELECT" {
		t.Error("must report SELECT operation even no record result")
	}
	if lastSegment.ParameterizedQuery != `SELECT * FROM "models"  WHERE ("models"."value" = ?) ORDER BY "models"."id" ASC LIMIT 1` {
		t.Error("wrong query", lastSegment.ParameterizedQuery)
	}

	historyLen := len(segmentsHistory)
	dbInsert(t, db)
	dbSelect(t, db)
	dbUpdate(t, db)
	dbDelete(t, db)
	if len(segmentsHistory) > historyLen {
		t.Error("main db was affected")
	}
}

func TestDBManualTransaction(t *testing.T) {
	segmentsHistory = []*nrmock.DatastoreSegment{}
	txnDB := SetTxnToGorm(testTxn, db)

	// when transaction has been started manually gorm won't wrap insert/update/delete into another tx
	// commit time in such case must be measured manually by the user
	tx := txnDB.Begin()

	m := &Model{Value: "manual-tx-test"}
	if err := tx.Create(m).Error; err != nil {
		t.Error(err)
	}
	m.Value = "updated"
	if err := tx.Save(m).Error; err != nil {
		t.Error(err)
	}
	if err := tx.Delete(m).Error; err != nil {
		t.Error(err)
	}

	tx.Commit()

	if len(segmentsHistory) != 3 {
		t.Errorf("expected 3 segments, got: %v", len(segmentsHistory))
	}
	op := segmentsHistory[0].Operation
	if op != "INSERT" {
		t.Errorf("wrong operation: %v", op)
	}
	op = segmentsHistory[1].Operation
	if op != "UPDATE" {
		t.Errorf("wrong operation: %v", op)
	}
	op = segmentsHistory[2].Operation
	if op != "DELETE" {
		t.Errorf("wrong operation: %v", op)
	}
}

func dbInsert(t *testing.T, db *gorm.DB) {
	if err := db.Create(&Model{Value: "test"}).Error; err != nil {
		t.Error(err)
	}
}

func dbSelect(t *testing.T, db *gorm.DB) {
	if err := db.First(&Model{Value: "to-select"}).Error; err != nil {
		t.Error(err)
	}
}

func dbSelectNoRecord(t *testing.T, db *gorm.DB) {
	if err := db.Where(Model{Value: "not found"}).First(&Model{}).Error; err != gorm.ErrRecordNotFound {
		t.Error(err)
	}
}

func dbUpdate(t *testing.T, db *gorm.DB) {
	m := &Model{Value: "to-update"}
	if err := db.Create(m).Error; err != nil {
		t.Error(err)
	}
	m.Value = "updated"
	if err := db.Save(m).Error; err != nil {
		t.Error(err)
	}
}

func dbDelete(t *testing.T, db *gorm.DB) {
	m := &Model{Value: "to-update"}
	if err := db.Create(m).Error; err != nil {
		t.Error(err)
	}
	if err := db.Delete(m).Error; err != nil {
		t.Error(err)
	}
}
