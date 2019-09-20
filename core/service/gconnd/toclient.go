package main

import (
	"net"

	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

func handleClientConn(cliConn net.Conn, sessionID uint64, srvMst string, cliChannel chan<- *innerCmd) {
	defer dbg.Stacktrace()
	defer cliConn.Close()

	cliAddr := cliConn.RemoteAddr().String()
	log.Debug("client=%s incoming, session=%d, master=%s", cliAddr, sessionID, srvMst)

	err := conn.HandleStream(cliConn, func(_ net.Conn, hdr, body []byte) {
		header := conn.ParseHeader(hdr)
		if conn.IsPrivateCmd(header.CmdID) {
			return
		}

		dupHdr := make([]byte, len(hdr))
		dupBody := make([]byte, len(body))
		copy(dupHdr, hdr)
		copy(dupBody, body)
		cliChannel <- newClientCmd(innerCmdClientUp, sessionID, hdr, body)
	})

	// notify the client leave
	cliChannel <- newNotifyCmd(innerCmdClientLeave, nil, cliAddr, sessionID, err)
}
