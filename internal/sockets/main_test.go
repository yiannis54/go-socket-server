package sockets

import (
	"testing"

	"go.uber.org/goleak"
)

func TestMain(m *testing.M) {
	ignoreOpenCensus := goleak.IgnoreTopFunction("go.opencensus.io/stats/view.(*worker).start")
	goleak.VerifyTestMain(m, ignoreOpenCensus)
}
