package nrmock

import (
	"net/http"

	newrelic "github.com/newrelic/go-agent"
)

type Transaction struct {
	name     string
	WasEnded bool
}

// interface

func (t *Transaction) End() error {
	t.WasEnded = true
	return nil
}
func (t *Transaction) Ignore() error                                    { return nil }
func (t *Transaction) SetName(name string) error                        { return nil }
func (t *Transaction) NoticeError(err error) error                      { return nil }
func (t *Transaction) AddAttribute(key string, value interface{}) error { return nil }
func (t *Transaction) StartSegmentNow() newrelic.SegmentStartTime {
	return newrelic.SegmentStartTime{}
}
func (t *Transaction) Header() http.Header       { return http.Header{} }
func (t *Transaction) Write([]byte) (int, error) { return 0, nil }
func (t *Transaction) WriteHeader(int)           {}

func (t *Transaction) CreateDistributedTracePayload() newrelic.DistributedTracePayload {
	return nil
}
func (t *Transaction) AcceptDistributedTracePayload(newrelic.TransportType, interface{}) error {
	return nil
}

// test helpers

func (t *Transaction) GetName() string {
	return t.name
}
