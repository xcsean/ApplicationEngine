package main

import (
	"net"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"github.com/xcsean/ApplicationEngine/core/shared/packet"
)

func handleSrvPkt(srvConn net.Conn, srvChannel chan<- *innerCmd, hdr, body []byte) {
	header := conn.ParseHeader(hdr)
	cmdID := protocol.PacketType(header.CmdID)
	if packet.IsPublicCmdID(cmdID) {
		// this is a public cmd, just wrap as a server broadcast
		srvChannel <- newServerCmd(innerCmdServerBroadcast, srvConn, hdr, body)
		return
	}

	// a private cmd
	switch cmdID {
	case protocol.Packet_PRIVATE_MASTER_SET:
		srvChannel <- newServerCmd(innerCmdServerMasterSet, srvConn, hdr, body)
	case protocol.Packet_PRIVATE_BROADCAST_ALL:
		srvChannel <- newServerCmd(innerCmdServerBroadcastAll, srvConn, hdr, body)
	case protocol.Packet_PRIVATE_SESSION_KICK:
		srvChannel <- newServerCmd(innerCmdServerKick, srvConn, hdr, body)
	case protocol.Packet_PRIVATE_KICK_ALL:
		srvChannel <- newServerCmd(innerCmdServerKickAll, srvConn, nil, nil)
	case protocol.Packet_PRIVATE_SESSION_ROUTE:
		srvChannel <- newServerCmd(innerCmdServerSetRoute, srvConn, hdr, body)
	}
}

func srvRecvLoop(srvConn net.Conn, srvChannel chan<- *innerCmd) {
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

func startSrvLoop(ls net.Listener, srvChannel chan<- *innerCmd) {
	defer dbg.Stacktrace()
	defer ls.Close()

	// notify listen start
	srvChannel <- newNotifyCmd(innerCmdServerListenStart, nil, "", 0, nil)

	for {
		c, err := ls.Accept()
		if err != nil {
			log.Error("accept server failed: %s", err.Error())
			continue
		}

		// notify a new server is incoming
		srvChannel <- newNotifyCmd(innerCmdServerIncoming, c, "", 0, nil)
	}
}
