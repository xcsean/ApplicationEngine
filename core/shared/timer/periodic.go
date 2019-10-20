package timer

import "time"

type periodicTask struct {
	interval time.Duration
}

func (pt *periodicTask) Next(prev time.Time) time.Time {
	return prev.Add(pt.interval)
}
