package nrcontext

import (
	"context"
	"github.com/newrelic/go-agent"
	"net/http"
)

// Gets NewRelic transaction from context and wraps client transport with newrelic RoundTripper
func WrapHTTPClient(ctx context.Context, c *http.Client) {
	txn := GetTnxFromContext(ctx)
	if txn == nil {
		return
	}
	c.Transport = newrelic.NewRoundTripper(txn, c.Transport)
}
