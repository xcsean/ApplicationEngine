package main

import (
	"fmt"
	"net"
	"time"

	"github.com/xcsean/ApplicationEngine/core/shared/conn"
)

type netCmd struct {
	cmdID uint8
	hdr []byte
	body []byte
}

func clientLoop(cliAddr string, cliChannel chan<- string) {
	// delay 1 second
	time.Sleep(1*time.Second)
	fmt.Println("[CLIENT] start...")

	// try to connect as client
	c, err := net.Dial("tcp", cliAddr)
	if err != nil {
		fmt.Printf("[CLIENT] %s\n", err.Error())
		cliChannel <- "exit"
		return
	}
	defer c.Close()

	// create a network routine & channel
	netChannel := make(chan *netCmd, 100)
	go netLoop(c, netChannel)

	// create a keyboard routine & channel
	kbdChannel := make(chan string, 100)
	go kbdLoop(kbdChannel)

	// wait for message from netChannel & kbdChannel
	for {
		exit := false
		select {
		case cmd := <-kbdChannel:
			if cmd == "exit" {
				exit = true
			}
		case cmd := <-netChannel:
			fmt.Printf("[CLIENT] net cmd=%d\n", cmd.cmdID)
			if cmd.cmdID == 0 {
				exit = true
			}
		}

		if exit {
			break
		}
	}

	// notify lobby exit
	cliChannel <- "exit"
}

func kbdLoop(kbdChannel chan<- string) {
	line := ""
	for {
		fmt.Printf("[CLIENT]$> ")
		fmt.Scanln(&line)
		kbdChannel <- line
		if line == "exit" {
			break
		}
	}
}

func netLoop(c net.Conn, netChannel chan<- *netCmd) {
	conn.HandleStream(c, func(_ net.Conn, hdr, body []byte) {
		netChannel <- &netCmd{cmdID: 1, hdr: hdr, body: body, }
	})
	netChannel <- &netCmd{cmdID: 0, }
}