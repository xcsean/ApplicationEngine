package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/xcsean/ApplicationEngine/core/shared/conn"
)

func lobbyLoop(srvAddr string, srvChannel chan<- string) {
	// try to connect gconnd as lobby
	c, err := net.Dial("tcp", srvAddr)
	if err != nil {
		fmt.Printf("[LOBBY] %s\n", err.Error())
		srvChannel <- "exit"
		return
	}
	defer c.Close()
	fmt.Printf("[LOBBY] connect to %s ok\n", srvAddr)

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
			fmt.Printf("[LOBBY] cmd=%d\n", h.CmdID)
			if h.CmdID == conn.CmdSessionEnter {
				_, sessionIDs, innerBody := conn.ParseSessionBody(b)
				sessionID := sessionIDs[0]

				// get the client address
				var pb conn.PrivateBody
				err = json.Unmarshal(innerBody, &pb)
				if err != nil {
					fmt.Printf("[LOBBY] client session=%d enter... %s\n", sessionID, err.Error())
				} else {
					fmt.Printf("[LOBBY] client session=%d addr=%s enter...\n", sessionID, pb.StrParam)
				}
			}
		} else {
			// wait the CmdMasterYou or CmdMasterNot
			switch h.CmdID {
			case conn.CmdMasterYou:
				fmt.Println("[LOBBY] I'm master, that's ok")
				isMaster = true
				srvChannel <- "master"
			case conn.CmdMasterNot:
				fmt.Println("[LOBBY] I can't be master, so exit")
				os.Exit(-1)
			}
		}
	})
	if err != nil {
		fmt.Printf("[LOBBY] %s\n", err.Error())
	}

	// notify lobby exit
	srvChannel <- "exit"
}
