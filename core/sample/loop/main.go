package main

import (
	"fmt"
	"os"
)

func printHelp() {
	fmt.Println("loop gconnd_ip cli_port srv_port")
}

func main() {
	if len(os.Args) < 4 {
		printHelp()
		return
	}

	cliAddr := fmt.Sprintf("%s:%s", os.Args[1], os.Args[2])
	srvAddr := fmt.Sprintf("%s:%s", "127.0.0.1", os.Args[3])

	fmt.Println(cliAddr)
	fmt.Println(srvAddr)

	srvChannel := make(chan string, 1000)
	cliChannel := make(chan string, 1000)

	go lobbyLoop(srvAddr, srvChannel)

	for {
		select {
		case cmd := <-srvChannel:
			if cmd == "exit" {
				break
			}
		case cmd := <-cliChannel:
			if cmd == "exit" {
				break
			}
		}
	}

	fmt.Println("exit...")
}