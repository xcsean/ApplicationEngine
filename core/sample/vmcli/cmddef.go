package main

const (
	cmdLoginReq = 1
	cmdLoginRsp = 2
	cmdSAY      = 3
)

type cmdBody struct {
	StrParam string
	Kv       map[string]string
}
