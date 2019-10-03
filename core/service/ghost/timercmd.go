package main

type timerCmd struct {
	Type      uint8
	Userdata1 uint64
	Userdata2 uint64
}

const (
	timerCmdVMMOnTick           = 101
	timerCmdSessionWaitVerCheck = 201
	timerCmdSessionWaitBind     = 202
	timerCmdSessionWorking      = 203
	timerCmdSessionWaitUnbind   = 204
	timerCmdSessionWaitDelete   = 205
	timerCmdSessionDeleted      = 209
)

var (
	timerDesc map[uint8]string
)

func init() {
	timerDesc = make(map[uint8]string)
	timerDesc[0] = "startUp"
	timerDesc[timerCmdSessionWaitVerCheck] = "waitVerCheck"
	timerDesc[timerCmdSessionWaitBind] = "waitBind"
	timerDesc[timerCmdSessionWorking] = "working"
	timerDesc[timerCmdSessionWaitUnbind] = "waitUnbind"
	timerDesc[timerCmdSessionWaitDelete] = "waitDelete"
	timerDesc[timerCmdSessionDeleted] = "deleted"

}

func getTimerDesc(t uint8) string {
	s, ok := timerDesc[t]
	if ok {
		return s
	}
	return "unknown"
}
