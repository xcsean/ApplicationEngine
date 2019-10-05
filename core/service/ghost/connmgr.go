package main

import (
	"net"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"github.com/xcsean/ApplicationEngine/core/shared/packet"
)

type connCmd struct {
	cmdID uint8
	data  []byte
	pkt   *protocol.SessionPacket
}

func (cc *connCmd) getID() uint8 {
	return cc.cmdID
}

type connMgr struct {
	csk   net.Conn
	out   chan *connCmd
	exitC chan struct{}
}

var (
	cnm *connMgr
)

func initConnMgr(csk net.Conn) {
	if cnm == nil {
		cnm = &connMgr{
			csk:   csk,
			out:   make(chan *connCmd, 3000),
			exitC: make(chan struct{}, 1),
		}
		go connSendLoop(cnm.csk, cnm.out, cnm.exitC)
	}
}

func finiConnMgr() {
	if cnm == nil {
		return
	}

	exitC := cnm.exitC
	cnm = nil
	close(exitC)
}

func connSend(pkt *protocol.SessionPacket) {
	if cnm == nil {
		return
	}

	defer dbg.Stacktrace()

	select {
	case cnm.out <- &connCmd{cmdID: innerCmdConnSessionDn, pkt: pkt}:
	default:
		// just discard the packet
	}
}

// the below connXXXXLoop run themself goroutine, not main goroutine!!!

func connSendLoop(csk net.Conn, out chan *connCmd, exitC chan struct{}) {
	defer csk.Close()
	defer close(out)

	// TODO down sampler & clipped
	for {
		select {
		case cmd := <-out:
			if cmd.pkt != nil {
				data, err := packet.TransformRPCToSocket(cmd.pkt)
				if err == nil {
					csk.Write(data)
				}
			} else if cmd.data != nil {
				csk.Write(cmd.data)
			}
		case <-exitC:
			goto exit
		}
	}

exit:
	log.Info("conn send loop exit")
}

func connRecvLoop(csk net.Conn, connChannel chan *connCmd) {
	// try to request master
	pkt, _ := packet.MakeMasterSet()
	csk.Write(pkt)

	isMaster := false
	err := conn.HandleStream(csk, func(_ net.Conn, hdr, body []byte) {
		if isMaster {
			// common packet deal, push to connChannel
			dupHdr := make([]byte, len(hdr))
			dupBody := make([]byte, len(body))
			copy(dupHdr, hdr)
			copy(dupBody, body)
			pkt := packet.TransformSocketToRPC(dupHdr, dupBody)
			cmd := &connCmd{cmdID: innerCmdConnSessionUp, pkt: pkt}
			select {
			case connChannel <- cmd:
			default:
				// just discard the packet
			}
		} else {
			// wait the CmdMasterYou or CmdMasterNot
			h := conn.ParseHeader(hdr)
			cmdID := protocol.PacketType(h.CmdID)
			switch cmdID {
			case protocol.Packet_PRIVATE_MASTER_YOU:
				log.Info("I'm master, that's ok")
				// init the conn-manager
				isMaster = true
				initConnMgr(csk)
			case protocol.Packet_PRIVATE_MASTER_NOT:
				log.Fatal("I can't be master, so exit")
			}
		}
	})

	if err != nil {
		log.Error("client for gconnd exit, reason: %s", err.Error())
	} else {
		log.Info("client for gconnd exit")
	}
	connChannel <- &connCmd{cmdID: innerCmdConnExit}
}
