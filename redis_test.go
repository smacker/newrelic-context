package nrcontext

import (
	"context"
	"github.com/alicebob/miniredis"
	"gopkg.in/redis.v4"
	"testing"
)

func TestRedis(t *testing.T) {
	// in-memory redis
	s, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer s.Close()

	// build ctx with newrelic
	app := &NewrelicAppMock{}
	txn := app.StartTransaction("txn-name", nil, nil)
	ctx := context.WithValue(context.Background(), txnKey, txn)

	// main client
	client := redis.NewClient(&redis.Options{
		Addr: s.Addr(),
	})

	// shouldn't see in log
	callCmd(client, "without wrapper")

	// should see in log
	ctxClient := WrapRedisClient(ctx, client)
	callCmd(ctxClient, "with wrapper")

	// check that we didn't affect main client
	callCmd(client, "without wrapper")
}

func callCmd(c *redis.Client, value string) {
	_, err := c.Set("foo", value, 0).Result()
	if err != nil {
		panic("Redis returned error")
	}
}
