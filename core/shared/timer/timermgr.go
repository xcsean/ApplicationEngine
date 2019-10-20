package timer

import (
	"time"

	tw "github.com/RussellLuo/timingwheel"
)

// Mgr timer manager for delay/scheduler features
type Mgr struct {
	t *tw.TimingWheel
}

// New new a timer manager
func New() *Mgr {
	return &Mgr{
		t: tw.NewTimingWheel(time.Millisecond, 20),
	}
}

// Start start the timer manager
func (mm *Mgr) Start() {
	mm.t.Start()
}

// Stop stop the timer manager
func (mm *Mgr) Stop() {
	mm.t.Stop()
}

// AddDelayTask add a timer for task execute
func (mm *Mgr) AddDelayTask(d time.Duration, f func()) {
	mm.t.AfterFunc(d, f)
}

// AddPeriodicTask add a task for execute periodic
func (mm *Mgr) AddPeriodicTask(d time.Duration, f func()) {
	s := &periodicTask{interval: d}
	mm.t.ScheduleFunc(s, f)
}
