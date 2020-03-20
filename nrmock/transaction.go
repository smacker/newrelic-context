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

func (t *Transaction) SetWebRequest(newrelic.WebRequest) error {
	return nil
}

func (t *Transaction) SetWebResponse(http.ResponseWriter) newrelic.Transaction {
	return t
}

func (t *Transaction) Application() newrelic.Application {
	return &NewrelicApp{}
}

func (t *Transaction) BrowserTimingHeader() (*newrelic.BrowserTimingHeader, error) {
	return nil, nil
}

func (t *Transaction) NewGoroutine() newrelic.Transaction {
	return t
}

func (t *Transaction) GetTraceMetadata() newrelic.TraceMetadata {
	return newrelic.TraceMetadata{}
}

func (t *Transaction) GetLinkingMetadata() newrelic.LinkingMetadata {
	return newrelic.LinkingMetadata{}
}

func (t *Transaction) IsSampled() bool {
	return false
}

// test helpers

func (t *Transaction) GetName() string {
	return t.name
}
