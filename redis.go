package nrcontext

import (
	"context"
	"strings"

	"github.com/newrelic/go-agent"
	"gopkg.in/redis.v5"
)

type redisWrapper func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error

// Function for client.WrapProcess that mesures time of commands in newrelic
func RedisWrapper(txn newrelic.Transaction) redisWrapper {
	return func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
		return func(cmd redis.Cmder) error {
			defer newrelic.DatastoreSegment{
				StartTime: newrelic.StartSegmentNow(txn),
				Product:   newrelic.DatastoreRedis,
				Operation: strings.Split(cmd.String(), " ")[0],
			}.End()

			// It's too difficult to mock
			// fmt.Println(cmd)

			return oldProcess(cmd)
		}
	}
}

// Gets transaction from Context and applies RedisWrapper, returns cloned client
func WrapRedisClient(ctx context.Context, c *redis.Client) *redis.Client {
	txn := GetTnxFromContext(ctx)
	if txn == nil {
		return c
	}

	cCopyP := c.WithContext(ctx)
	cCopyP.WrapProcess(RedisWrapper(txn))

	return cCopyP
}
