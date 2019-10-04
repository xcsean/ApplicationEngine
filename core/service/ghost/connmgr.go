package main

import (
	"net"

	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

type connCmd struct {
	cmdID uint8
	b1    []byte
	b2    []byte
}

func (cc *connCmd) getID() uint8 {
	return cc.cmdID
}

func (cc *connCmd) getConnCmd() ([]byte, []byte) {
	return cc.b1, cc.b2
}

func newConnCmd(cmdID uint8, b1, b2 []byte) *connCmd {
	return &connCmd{
		cmdID: cmdID,
		b1:    b1,
		b2:    b2,
	}
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

func connSend(pkt []byte) {
	if cnm == nil {
		return
	}

	defer dbg.Stacktrace()

	select {
	case cnm.out <- &connCmd{cmdID: innerCmdConnSessionDn, b1: pkt}:
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
			csk.Write(cmd.b1)
		case <-exitC:
			goto exit
		}
	}

exit:
	log.Info("conn send loop exit")
}

func connRecvLoop(csk net.Conn, connChannel chan *connCmd) {
	// try to request master
	conn.SendMasterSet(csk)

	isMaster := false
	err := conn.HandleStream(csk, func(_ net.Conn, hdr, body []byte) {
		h := conn.ParseHeader(hdr)
		if isMaster {
			// common packet deal, push to connChannel
			dupHdr := make([]byte, len(hdr))
			dupBody := make([]byte, len(body))
			copy(dupHdr, hdr)
			copy(dupBody, body)
			select {
			case connChannel <- newConnCmd(innerCmdConnSessionUp, dupHdr, dupBody):
			default:
				// just discard the packet
			}
		} else {
			// wait the CmdMasterYou or CmdMasterNot
			switch h.CmdID {
			case conn.CmdMasterYou:
				log.Info("I'm master, that's ok")
				// init the conn-manager
				isMaster = true
				initConnMgr(csk)
			case conn.CmdMasterNot:
				log.Fatal("I can't be master, so exit")
			}
		}
	})

	if err != nil {
		log.Error("client for gconnd exit, reason: %s", err.Error())
	} else {
		log.Info("client for gconnd exit")
	}
	connChannel <- newConnCmd(innerCmdConnExit, nil, nil)
}
