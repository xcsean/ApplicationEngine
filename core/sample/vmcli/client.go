package main

import (
	"encoding/json"
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
	connAddr   string
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
	connAddr = addr
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

func dealKeyboard(text string, cliLog func(s string)) {
	text = strings.Replace(text, "\n", "", -1)
	text = strings.Replace(text, "\t", " ", -1)
	array := strings.Fields(text)
	if array == nil {
		return
	}

	cmd := array[0]
	if cmd == "conn" {
		if cliConn == nil {
			c, err := net.Dial("tcp", connAddr)
			if err != nil {
				cliLog(fmt.Sprintf("[C] connect %s failed: %s", connAddr, err.Error()))
			} else {
				cliConn = c
				cliLog(fmt.Sprintf("[C] connect %s ok", connAddr))
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
			cliLog(fmt.Sprintf("[C] disconnect from %s ok", connAddr))
		}
	} else if cmd == "say" {
		if cliConn == nil {
			cliLog("[C] please type 'conn' first!")
		} else {
			if len(array) >= 2 {
				var innerBody sayBody
				innerBody.StrParam = array[1]
				body, _ := json.Marshal(innerBody)
				pkt := conn.MakeCommonPkt(cmdSAY, 0, 0, body)
				cliConn.Write(pkt)
				cliLog(fmt.Sprintf("[C] say '%s'", array[1]))
			}
		}
	}
}

func netLoop(c net.Conn, netChannel chan<- *netCmd) {
	conn.HandleStream(c, func(_ net.Conn, hdr, body []byte) {
		netChannel <- &netCmd{cmdID: 1, hdr: hdr, body: body}
	})
	netChannel <- &netCmd{cmdID: 0}
}
