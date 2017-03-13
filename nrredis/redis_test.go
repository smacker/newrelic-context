package nrredis

import (
	"os"
	"testing"

	"github.com/alicebob/miniredis"
	newrelic "github.com/newrelic/go-agent"
	"github.com/smacker/newrelic-context/nrmock"
	"gopkg.in/redis.v5"
)

var client *redis.Client
var testTxn newrelic.Transaction
var lastSegment *nrmock.DatastoreSegment

func TestMain(m *testing.M) {
	// in-memory redis
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()
	client = redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	// mock newrelic
	originalBuilder := segmentBuilder
	segmentBuilder = func(txn newrelic.Transaction, product newrelic.DatastoreProduct, operation string) segment {
		segment := originalBuilder(txn, product, operation).(newrelic.DatastoreSegment)
		mock := &nrmock.DatastoreSegment{DatastoreSegment: segment, Txn: txn}
		lastSegment = mock
		return mock
	}

	app := &nrmock.NewrelicApp{}
	testTxn = app.StartTransaction("txn-name", nil, nil)

	os.Exit(m.Run())
}

func TestWrapRedis(t *testing.T) {
	lastSegment = nil
	callSet(t, client, "without wrapper")
	if lastSegment != nil {
		t.Fatal("newrelic segment was created for bare redis client")
	}

	ctxClient := WrapRedisClient(testTxn, client)
	callSet(t, ctxClient, "with wrapper")
	if lastSegment == nil {
		t.Fatal("newrelic segment was created for wrapped client call")
	}

	lastSegment = nil
	callSet(t, client, "without wrapper")
	if lastSegment != nil {
		t.Fatal("main client was affected")
	}
}

func TestSegmentParams(t *testing.T) {
	txnClient := WrapRedisClient(testTxn, client)
	callSet(t, txnClient, "with wrapper")
	if !lastSegment.Finished {
		t.Error("segment should be finished")
	}
	if lastSegment.Txn != testTxn {
		t.Error("incorrect transaction was passed")
	}
	if lastSegment.Product != newrelic.DatastoreRedis {
		t.Error("wrong product")
	}
	if lastSegment.Operation != "set" {
		t.Error("wrong operation")
	}
	callGet(t, txnClient)
	if lastSegment.Operation != "get" {
		t.Error("wrong operation")
	}
}

func callSet(t *testing.T, c *redis.Client, value string) {
	_, err := c.Set("foo", value, 0).Result()
	if err != nil {
		t.Fatalf("Redis returned error: %v", err)
	}
}

func callGet(t *testing.T, c *redis.Client) {
	_, err := c.Get("foo").Result()
	if err != nil {
		t.Fatalf("Redis returned error: %v", err)
	}
}
