package main

// TODO: replace this with the time-wheel style

import (
	"time"
)

type timerMgr struct {
	C chan *timerCmd
	global map[string]bool
}

var (
	tmm *timerMgr
)

func init() {
	tmm = &timerMgr{
		C: make(chan *timerCmd, 1000),
		global: make(map[string]bool),
	}
}

func tmmGetChannel() chan *timerCmd {
	return tmm.C
}

func tmmAddDelayTask(d time.Duration, task func(c chan *timerCmd)) {
	go func(d time.Duration, task func(c chan *timerCmd)) {
		time.Sleep(d)
		task(tmm.C)
	}(d, task)
}

func tmmAddGlobalPeriodicTask(name string, d time.Duration, task func(c chan *timerCmd)) {
	if name == "" {
		return
	}

	_, ok := tmm.global[name]
	if !ok {
		tmm.global[name] = true
		go func(d time.Duration, task func(c chan *timerCmd)) {
			for {
				time.Sleep(d)
				task(tmm.C)
			}
		}(d, task)
	}
}
