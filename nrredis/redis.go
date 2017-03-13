package nrredis

import (
	"strings"

	newrelic "github.com/newrelic/go-agent"
	redis "gopkg.in/redis.v5"
)

// WrapRedisClient adds newrelic measurements for commands and returns cloned client
func WrapRedisClient(txn newrelic.Transaction, c *redis.Client) *redis.Client {
	if txn == nil {
		return c
	}

	// clone using context
	ctx := c.Context()
	copy := c.WithContext(ctx)

	copy.WrapProcess(func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			defer segmentBuilder(txn, newrelic.DatastoreRedis, strings.Split(cmd.String(), " ")[0]).End()

			return oldProcess(cmd)
		}
	})
	return copy
}

type segment interface {
	End()
}

// create segment through function to be able to test it
var segmentBuilder = func(txn newrelic.Transaction, product newrelic.DatastoreProduct, operation string) segment {
	return newrelic.DatastoreSegment{
		StartTime: newrelic.StartSegmentNow(txn),
		Product:   product,
		Operation: operation,
	}
}
