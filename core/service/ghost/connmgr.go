package main

import (
	"net"

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

	close(cnm.out)
	close(cnm.exitC)
	cnm.csk = nil
	cnm = nil
}

func connSendLoop(csk net.Conn, out chan *connCmd, exitC chan struct{}) {
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
