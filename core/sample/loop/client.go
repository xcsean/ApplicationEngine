package main

import (
	"fmt"
	"net"
	"strings"

	ui "github.com/jroimartin/gocui"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
)

const (
	clientView      = "clientView"
	clientViewTitle = "client message"
	clientEdit      = "clientEdit"
	clientEditTitle = "client input"
)

var (
	kbdChannel chan string
	netChannel chan *netCmd
	cliAddr    string
	cliConn    net.Conn
)

func getClientView() string {
	return clientView
}

func getClientViewTitle() string {
	return clientViewTitle
}

func getClientEdit() string {
	return clientEdit
}

func getClientEditTitle() string {
	return clientEditTitle
}

func sendClientKeyboard(text string) {
	kbdChannel <- text
}

type netCmd struct {
	cmdID uint8
	hdr   []byte
	body  []byte
}

func clientLoop(addr string, g *ui.Gui) {
	cliAddr = addr
	kbdChannel = make(chan string, 100)
	netChannel = make(chan *netCmd, 100)
	cliLog := func(s string) {
		g.Update(func(g *ui.Gui) error {
			v, _ := g.View(clientView)
			fmt.Fprintf(v, "%s\n", s)
			return nil
		})
	}

	// wait for message from netChannel & kbdChannel
	for {
		select {
		case cmd := <-kbdChannel:
			dealKeyboard(cmd, cliLog)
		case cmd := <-netChannel:
			cliLog(fmt.Sprintf("[CLIENT] net cmd=%d\n", cmd.cmdID))
		}
	}
}

func dealKeyboard(cmd string, cliLog func(s string)) {
	cmd = strings.Replace(cmd, "\n", "", -1)
	if cmd == "conn" {
		if cliConn == nil {
			c, err := net.Dial("tcp", cliAddr)
			if err != nil {
				cliLog(fmt.Sprintf("[C] connect %s failed: %s", cliAddr, err.Error()))
			} else {
				cliConn = c
				cliLog(fmt.Sprintf("[C] connect %s ok", cliAddr))
			}
		} else {
			cliLog("[C] alreay connected!")
		}
	} else if cmd == "disc" {
		if cliConn == nil {
			cliLog("[C] please type 'conn' first!")
		} else {
			cliConn.Close()
			cliConn = nil
			cliLog(fmt.Sprintf("[C] disconnect from %s ok", cliAddr))
		}
	}
}

func netLoop(c net.Conn, netChannel chan<- *netCmd) {
	conn.HandleStream(c, func(_ net.Conn, hdr, body []byte) {
		netChannel <- &netCmd{cmdID: 1, hdr: hdr, body: body}
	})
	netChannel <- &netCmd{cmdID: 0}
}
