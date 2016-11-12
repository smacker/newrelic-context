package nrcontext

import (
	"github.com/newrelic/go-agent"
	"net/http"
	"time"
)

type NewrelicAppMock struct {
	Tnx *TransactionMock
}

func (a *NewrelicAppMock) StartTransaction(name string, w http.ResponseWriter, r *http.Request) newrelic.Transaction {
	a.Tnx = &TransactionMock{name, false}
	return a.Tnx
}

// just implement interface
func (a *NewrelicAppMock) RecordCustomEvent(eventType string, params map[string]interface{}) error {
	return nil
}
func (a *NewrelicAppMock) WaitForConnection(timeout time.Duration) error {
	return nil
}
func (a *NewrelicAppMock) Shutdown(timeout time.Duration) {}

type TransactionMock struct {
	name     string
	WasEnded bool
}

// interface
func (t *TransactionMock) End() error {
	t.WasEnded = true
	return nil
}
func (t *TransactionMock) Ignore() error                                    { return nil }
func (t *TransactionMock) SetName(name string) error                        { return nil }
func (t *TransactionMock) NoticeError(err error) error                      { return nil }
func (t *TransactionMock) AddAttribute(key string, value interface{}) error { return nil }
func (t *TransactionMock) StartSegmentNow() newrelic.SegmentStartTime {
	return newrelic.SegmentStartTime{}
}
func (t *TransactionMock) Header() http.Header       { return http.Header{} }
func (t *TransactionMock) Write([]byte) (int, error) { return 0, nil }
func (t *TransactionMock) WriteHeader(int)           {}

// test helpers
func (t *TransactionMock) GetName() string {
	return t.name
}
