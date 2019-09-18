package main

import (
	"net"

	cnn "github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

func handleSrvPkt(conn net.Conn, ch chan<- *innerCmd, hdr, body []byte) {
	header := cnn.ParseHeader(hdr)
	ok := cnn.IsPrivateCmd(header.CmdID)
	if !ok {
		// this is a public cmd, just wrap as a server broadcast
		ch <- newServerCmd(innerCmdServerBroadcast, conn, hdr, body)
		return
	}

	// a private cmd
	switch header.CmdID {
	case cnn.CmdMasterSet:
		ch <- newServerCmd(innerCmdServerMasterSet, conn, hdr, body)
	case cnn.CmdBroadcastAll:
		ch <- newServerCmd(innerCmdServerBroadcastAll, conn, hdr, body)
	case cnn.CmdSessionKick:
		ch <- newServerCmd(innerCmdServerKick, conn, hdr, body)
	case cnn.CmdKickAll:
		ch <- newServerCmd(innerCmdServerKickAll, conn, nil, nil)
	case cnn.CmdSessionRoute:
		ch <- newServerCmd(innerCmdServerSetRoute, conn, hdr, body)
	}
}

func handleServerConn(conn net.Conn, ch chan<- *innerCmd) {
	defer dbg.Stacktrace()
	defer conn.Close()

	srvAddr := conn.RemoteAddr().String()
	log.Debug("server=%s incoming", srvAddr)

	err := cnn.HandleStream(conn, func(conn net.Conn, hdr, body []byte) {
		dupHdr := make([]byte, len(hdr))
		dupBody := make([]byte, len(body))
		copy(dupHdr, hdr)
		copy(dupBody, body)
		handleSrvPkt(conn, ch, dupHdr, dupBody)
	})

	// notify the server leave
	ch <- newNotifyCmd(innerCmdServerLeave, nil, srvAddr, 0, err)
	log.Debug("server=%s leave", srvAddr)
}
