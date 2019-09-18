package main

import (
	"net"

	cnn "github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

func handleClientConn(conn net.Conn, sessionID uint64, srvMst string, ch chan<- *innerCmd) {
	defer dbg.Stacktrace()
	defer conn.Close()

	cliAddr := conn.RemoteAddr().String()
	log.Debug("client=%s incoming, session=%d, master=%s", cliAddr, sessionID, srvMst)

	handleCliPkt := func(sessionID uint64, ch chan<- *innerCmd, hdr, body []byte) {
		ch <- newClientCmd(innerCmdClientUp, sessionID, hdr, body)
	}
	err := cnn.HandleStream(conn, func(conn net.Conn, hdr, body []byte) {
		header := cnn.ParseHeader(hdr)
		if cnn.IsPrivateCmd(header.CmdID) {
			return
		}

		dupHdr := make([]byte, len(hdr))
		dupBody := make([]byte, len(body))
		copy(dupHdr, hdr)
		copy(dupBody, body)
		handleCliPkt(sessionID, ch, dupHdr, dupBody)
	})

	// notify the client leave
	ch <- newNotifyCmd(innerCmdClientLeave, nil, cliAddr, sessionID, err)
}
