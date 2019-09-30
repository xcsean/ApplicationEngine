package main

type timerCmd struct {
	Type      uint8
	Userdata1 uint64
	Userdata2 uint64
}

const (
	timerCmdVMMOnTick = 101
	timerCmdSessionWaitVerCheck = 201
	timerCmdSessionWaitBindUser = 202
)
