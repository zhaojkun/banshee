package notifier

import (
	"testing"

	"github.com/eleme/banshee/util"
)

func TestWrapNumber(t *testing.T) {
	util.Must(t, wrapNumber(0) == "0")
	util.Must(t, wrapNumber(1) == "1")
	util.Must(t, wrapNumber(100) == "100")
	util.Must(t, wrapNumber(10100) == "10.1K")
}

func TestGraphiteName(t *testing.T) {
	util.Must(t, graphiteName("timer.count_ps.name") == "stats.timers.name.count_ps")
}
