package nrcontext

import (
	"context"
	"github.com/newrelic/go-agent"
	"gopkg.in/redis.v4"
	"strings"
)

type redisWrapper func(oldProcess func(cmd redis.Cmder) error) func(cmd redis.Cmder) error

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

func WrapRedisClient(ctx context.Context, c *redis.Client) *redis.Client {
	txn := GetTnxFromContext(ctx)
	if txn == nil {
		return c
	}

	cCopy := *c
	cCopyP := &cCopy

	cCopyP.WrapProcess(RedisWrapper(txn))

	return cCopyP
}
