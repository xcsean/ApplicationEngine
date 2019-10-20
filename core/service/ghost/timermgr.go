package main

import (
	"time"

	"github.com/xcsean/ApplicationEngine/core/shared/timer"
)

var (
	tmgr *timer.Mgr
	tmmc chan *timerCmd
)

func init() {
	tmmc = make(chan *timerCmd, 1000)
	tmgr = timer.New()
	tmgr.Start()
}

func tmmGetChannel() chan *timerCmd {
	return tmmc
}

func tmmAddDelayTask(d time.Duration, task func(c chan *timerCmd)) {
	tmgr.AddDelayTask(d, func() { task(tmmc) })
}

func tmmAddGlobalPeriodicTask(name string, d time.Duration, task func(c chan *timerCmd)) {
	tmgr.AddPeriodicTask(d, func() { task(tmmc) })
}
