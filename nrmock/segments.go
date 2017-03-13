package nrmock

import newrelic "github.com/newrelic/go-agent"

type DatastoreSegment struct {
	newrelic.DatastoreSegment
	Txn       newrelic.Transaction
	StartTime newrelic.SegmentStartTime
	Finished  bool
}

func (m *DatastoreSegment) End() {
	m.Finished = true
}
