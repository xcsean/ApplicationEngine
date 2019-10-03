package main

const (
	cmdLogin = 1
	cmdSAY   = 3
)

type cmdBody struct {
	StrParam string
	Kv       map[string]string
}
