package timer

import (
	"testing"
	"time"
)

func TestTimerMgr(t *testing.T) {
	tmm := New()
	tmm.Start()
	tmm.AddDelayTask(time.Second, func() {
		t.Log("delay task executed")
	})
	tmm.AddPeriodicTask(time.Second, func() {
		t.Log("periodic task execute")
	})

	time.Sleep(10 * time.Second)

	t.Log("timer mgr stop")
	tmm.Stop()
}
