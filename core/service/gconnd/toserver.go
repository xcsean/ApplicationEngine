package main

import (
	"net"

	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

func handleSrvPkt(srvConn net.Conn, srvChannel chan<- *innerCmd, hdr, body []byte) {
	header := conn.ParseHeader(hdr)
	ok := conn.IsPrivateCmd(header.CmdID)
	if !ok {
		// this is a public cmd, just wrap as a server broadcast
		srvChannel <- newServerCmd(innerCmdServerBroadcast, srvConn, hdr, body)
		return
	}

	// a private cmd
	switch header.CmdID {
	case conn.CmdMasterSet:
		srvChannel <- newServerCmd(innerCmdServerMasterSet, srvConn, hdr, body)
	case conn.CmdBroadcastAll:
		srvChannel <- newServerCmd(innerCmdServerBroadcastAll, srvConn, hdr, body)
	case conn.CmdSessionKick:
		srvChannel <- newServerCmd(innerCmdServerKick, srvConn, hdr, body)
	case conn.CmdKickAll:
		srvChannel <- newServerCmd(innerCmdServerKickAll, srvConn, nil, nil)
	case conn.CmdSessionRoute:
		srvChannel <- newServerCmd(innerCmdServerSetRoute, srvConn, hdr, body)
	}
}

func handleServerConn(srvConn net.Conn, srvChannel chan<- *innerCmd) {
	defer dbg.Stacktrace()
	defer srvConn.Close()

	srvAddr := srvConn.RemoteAddr().String()
	log.Debug("server=%s incoming", srvAddr)

	err := conn.HandleStream(srvConn, func(_ net.Conn, hdr, body []byte) {
		dupHdr := make([]byte, len(hdr))
		dupBody := make([]byte, len(body))
		copy(dupHdr, hdr)
		copy(dupBody, body)
		handleSrvPkt(srvConn, srvChannel, dupHdr, dupBody)
	})

	// notify the server leave
	srvChannel <- newNotifyCmd(innerCmdServerLeave, nil, srvAddr, 0, err)
	log.Debug("server=%s leave", srvAddr)
}
