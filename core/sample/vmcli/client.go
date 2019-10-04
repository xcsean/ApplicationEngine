package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
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
			fmt.Fprintln(v, s)
			return nil
		})
	}

	// wait for message from netChannel & kbdChannel
	for {
		select {
		case cmd := <-kbdChannel:
			dealCliKeyboard(cmd, cliLog)
		case cmd := <-netChannel:
			dealNetCmd(cmd, cliLog)
		}
	}
}

func dealCliKeyboard(text string, cliLog func(s string)) {
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
				go netLoop(c, netChannel)
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
	} else if cmd == "ver" {
		if cliConn == nil {
			cliLog("[C] please type 'conn' first!")
		} else {
			if len(array) >= 2 {
				var rb conn.ReservedBody
				rb.StrParam = array[1]
				body, _ := json.Marshal(rb)
				pkt := conn.MakeCommonPkt(conn.CmdVerCheck, 0, 0, body)
				cliConn.Write(pkt)
				cliLog(fmt.Sprintf("[C] version=%s", array[1]))
			} else {
				cliLog("[C] help: version 1.1.1.1")
			}
		}
	} else if cmd == "login" {
		if cliConn == nil {
			cliLog("[C] please type 'conn' first!")
		} else {
			if len(array) >= 2 {
				_, err := strconv.ParseInt(array[1], 10, 64)
				if err == nil {
					innerBody := &cmdBody{
						StrParam: "",
						Kv:       make(map[string]string),
					}
					innerBody.Kv["uuid"] = array[1]
					innerBody.Kv["token"] = "1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ"
					body, _ := json.Marshal(innerBody)
					pkt := conn.MakeCommonPkt(cmdLogin, 0, 0, body)
					cliConn.Write(pkt)
					cliLog(fmt.Sprintf("[C] login with uuid='%s'", array[1]))
				} else {
					cliLog(fmt.Sprintf("[C] uuid error: %s", err.Error()))
				}
			} else {
				cliLog("[C] help: login 10001")
			}
		}
	} else if cmd == "say" {
		if cliConn == nil {
			cliLog("[C] please type 'conn' first!")
		} else {
			if len(array) >= 2 {
				var innerBody cmdBody
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

func dealNetCmd(cmd *netCmd, cliLog func(s string)) {
	if cmd.cmdID == 1 {
		header := conn.ParseHeader(cmd.hdr)
		switch header.CmdID {
		case conn.CmdVerCheck:
			var rb conn.ReservedBody
			err := json.Unmarshal(cmd.body, &rb)
			if err == nil {
				cliLog(fmt.Sprintf("[S] ver-check: %s", rb.StrParam))
			} else {
				cliLog(fmt.Sprintf("[S] ver-check parse body failed: %s", err.Error()))
			}
		case cmdLogin:
			var innerBody cmdBody
			err := json.Unmarshal(cmd.body, &innerBody)
			if err == nil {
				cliLog(fmt.Sprintf("[S] result=%s uuid=%s", innerBody.Kv["result"], innerBody.Kv["uuid"]))
			}
		default:
			cliLog(fmt.Sprintf("[S] unknown cmd=%d", cmd.cmdID))
		}
	} else if cmd.cmdID == 0 {
		cliLog(fmt.Sprintf("[S] connection closed by remote"))
		if cliConn != nil {
			cliConn.Close()
			cliConn = nil
		}
	}
}
