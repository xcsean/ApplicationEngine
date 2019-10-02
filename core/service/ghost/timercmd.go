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
