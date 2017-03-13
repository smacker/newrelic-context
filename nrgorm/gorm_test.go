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
var lastSegment *nrmock.DatastoreSegment

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
		segment := originalBuilder(startTime, product, query, operation, collection).(newrelic.DatastoreSegment)
		mock := &nrmock.DatastoreSegment{DatastoreSegment: segment, StartTime: startTime}
		lastSegment = mock
		return mock
	}

	app := &nrmock.NewrelicApp{}
	testTxn = app.StartTransaction("txn-name", nil, nil)

	os.Exit(m.Run())
}

func TestWrappedGorm(t *testing.T) {
	txnDB := SetTxnToGorm(testTxn, db)
	dbInsert(t, txnDB)
	if lastSegment.Product != newrelic.DatastoreSQLite {
		t.Error("wrong product")
	}
	if lastSegment.ParameterizedQuery != `INSERT INTO "models" ("value") VALUES (?)` {
		t.Error("wrong query")
	}
	if lastSegment.Operation != "INSERT" {
		t.Error("wrong operation")
	}
	if lastSegment.Collection != "models" {
		t.Error("wrong collection")
	}
	dbSelect(t, txnDB)
	if lastSegment.Operation != "SELECT" {
		t.Error("wrong operation")
	}
	dbUpdate(t, txnDB)
	if lastSegment.Operation != "UPDATE" {
		t.Error("wrong operation")
	}
	dbDelete(t, txnDB)
	if lastSegment.Operation != "DELETE" {
		t.Error("wrong operation")
	}

	lastSegment = nil
	dbInsert(t, db)
	dbSelect(t, db)
	dbUpdate(t, db)
	dbDelete(t, db)
	if lastSegment != nil {
		t.Error("main db was affected")
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
