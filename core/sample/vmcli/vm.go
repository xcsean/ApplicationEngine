package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	ui "github.com/jroimartin/gocui"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
)

const (
	vmView  = "VMView"
	vmTitle = "vm message"
)

func getVMView() string {
	return vmView
}

func getVMTitle() string {
	return vmTitle
}

func vmLoop(hostAddr string, g *ui.Gui) {
	// delay 1 second
	time.Sleep(1 * time.Second)

	vmLog := func(s string) {
		g.Update(func(g *ui.Gui) error {
			v, _ := g.View(vmView)
			fmt.Fprintln(v, s)
			return nil
		})
	}

	// try to connect gconnd as lobby
	c, err := net.Dial("tcp", hostAddr)
	if err != nil {
		vmLog(fmt.Sprintf("[S] %s", err.Error()))
		return
	}
	defer c.Close()

	vmLog(fmt.Sprintf("[S] connect to %s ok", hostAddr))

	// try to request master
	conn.SendMasterSet(c)

	// handle the stream between gconnd and self
	isMaster := false
	err = conn.HandleStream(c, func(_ net.Conn, hdr, body []byte) {
		h := conn.ParseHeader(hdr)
		if isMaster {
			// common packet deal
			b := make([]byte, len(body))
			copy(b, body)
			if h.CmdID == conn.CmdSessionEnter {
				_, sessionIDs, innerBody := conn.ParseSessionBody(b)
				sessionID := sessionIDs[0]
				// get the client address
				var pb conn.PrivateBody
				err = json.Unmarshal(innerBody, &pb)
				if err != nil {
					vmLog(fmt.Sprintf("[P] client session=%d enter... %s", sessionID, err.Error()))
				} else {
					vmLog(fmt.Sprintf("[P] client session=%d addr=%s enter...", sessionID, pb.StrParam))
				}
			} else if h.CmdID == conn.CmdSessionLeave {
				_, sessionIDs, innerBody := conn.ParseSessionBody(b)
				sessionID := sessionIDs[0]
				// get the client address
				var pb conn.PrivateBody
				err = json.Unmarshal(innerBody, &pb)
				if err != nil {
					vmLog(fmt.Sprintf("[P] client session=%d leave... %s", sessionID, err.Error()))
				} else {
					vmLog(fmt.Sprintf("[P] client session=%d addr=%s leave...", sessionID, pb.StrParam))
				}
			} else if h.CmdID == cmdSAY {
				_, sessionIDs, innerBody := conn.ParseSessionBody(b)
				sessionID := sessionIDs[0]
				var say sayBody
				err := json.Unmarshal(innerBody, &say)
				if err == nil {
					vmLog(fmt.Sprintf("[P] client session=%d say '%s'", sessionID, say.StrParam))
				}
			}
		} else {
			// wait the CmdMasterYou or CmdMasterNot
			switch h.CmdID {
			case conn.CmdMasterYou:
				vmLog(fmt.Sprintf("[S] I'm master, that's ok"))
				isMaster = true
				// active the client input
				g.Update(func(g *ui.Gui) error {
					g.SetCurrentView(getClientEdit())
					return nil
				})
			case conn.CmdMasterNot:
				vmLog(fmt.Sprintf("[S] I can't be master, so exit"))
			}
		}
	})

	if err != nil {
		vmLog(fmt.Sprintf("[S] %s", err.Error()))
	}
}
