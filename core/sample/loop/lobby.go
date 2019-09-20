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
	lobbyView  = "lobbyView"
	lobbyTitle = "lobby message"
)

func getLobbyView() string {
	return lobbyView
}

func getLobbyTitle() string {
	return lobbyTitle
}

func lobbyLoop(srvAddr string, g *ui.Gui) {
	// delay 1 second
	time.Sleep(1 * time.Second)

	lobbyLog := func(s string) {
		g.Update(func(g *ui.Gui) error {
			v, _ := g.View(lobbyView)
			fmt.Fprintln(v, s)
			return nil
		})
	}

	// try to connect gconnd as lobby
	c, err := net.Dial("tcp", srvAddr)
	if err != nil {
		lobbyLog(fmt.Sprintf("[S] %s", err.Error()))
		return
	}
	defer c.Close()

	lobbyLog(fmt.Sprintf("[S] connect to %s ok", srvAddr))

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
					lobbyLog(fmt.Sprintf("[P] client session=%d enter... %s", sessionID, err.Error()))
				} else {
					lobbyLog(fmt.Sprintf("[P] client session=%d addr=%s enter...", sessionID, pb.StrParam))
				}
			} else if h.CmdID == conn.CmdSessionLeave {
				_, sessionIDs, innerBody := conn.ParseSessionBody(b)
				sessionID := sessionIDs[0]
				// get the client address
				var pb conn.PrivateBody
				err = json.Unmarshal(innerBody, &pb)
				if err != nil {
					lobbyLog(fmt.Sprintf("[P] client session=%d leave... %s", sessionID, err.Error()))
				} else {
					lobbyLog(fmt.Sprintf("[P] client session=%d addr=%s leave...", sessionID, pb.StrParam))
				}
			}
		} else {
			// wait the CmdMasterYou or CmdMasterNot
			switch h.CmdID {
			case conn.CmdMasterYou:
				lobbyLog(fmt.Sprintf("[S] I'm master, that's ok"))
				isMaster = true
				// active the client input
				g.Update(func(g *ui.Gui) error {
					g.SetCurrentView(getClientEdit())
					return nil
				})
			case conn.CmdMasterNot:
				lobbyLog(fmt.Sprintf("[S] I can't be master, so exit"))
			}
		}
	})

	if err != nil {
		lobbyLog(fmt.Sprintf("[S] %s", err.Error()))
	}
}
