package main

import (
	"net"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"github.com/xcsean/ApplicationEngine/core/shared/packet"
)

func recvCliLoop(cliConn net.Conn, sessionID uint64, srvMst string, cliChannel chan<- *innerCmd) {
	defer dbg.Stacktrace()
	defer cliConn.Close()

	cliAddr := cliConn.RemoteAddr().String()
	log.Debug("client=%s incoming, session=%d, master=%s", cliAddr, sessionID, srvMst)

	err := conn.HandleStream(cliConn, func(_ net.Conn, hdr, body []byte) {
		header := conn.ParseHeader(hdr)
		cmdID := protocol.PacketType(header.CmdID)
		if packet.IsPrivateCmdID(cmdID) {
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

func startCliLoop(ls net.Listener, cliChannel chan<- *innerCmd) {
	defer dbg.Stacktrace()
	defer ls.Close()

	// notify listen start
	cliChannel <- newNotifyCmd(innerCmdClientListenStart, nil, "", 0, nil)

	for {
		c, err := ls.Accept()
		if err != nil {
			log.Error("accept client failed: %s", err.Error())
			continue
		}

		// notify a new client is incoming
		cliChannel <- newNotifyCmd(innerCmdClientIncoming, c, "", 0, nil)
	}

	// notify listen stop
	cliChannel <- newNotifyCmd(innerCmdClientListenStop, nil, "", 0, nil)
	log.Info("acceptor for client exit")
}
