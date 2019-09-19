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

	fmt.Println("[MAIN] loop start...")
	cliAddr := fmt.Sprintf("%s:%s", os.Args[1], os.Args[2])
	srvAddr := fmt.Sprintf("%s:%s", "127.0.0.1", os.Args[3])

	srvChannel := make(chan string, 100)
	cliChannel := make(chan string, 100)

	go lobbyLoop(srvAddr, srvChannel)

	for {
		exit := false
		select {
		case cmd := <-srvChannel:
			if cmd == "exit" {
				fmt.Println("[MAIN] lobby exit...")
				exit = true
			} else if cmd == "master" {
				fmt.Println("[MAIN] lobby is master now")
				go clientLoop(cliAddr, cliChannel)
			}
		case cmd := <-cliChannel:
			if cmd == "exit" {
				fmt.Println("[MAIN] client exit...")
				exit = true
			}
		}

		if exit {
			break
		}
	}

	fmt.Println("[MAIN] loop exit...")
}