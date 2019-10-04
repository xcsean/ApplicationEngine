package main

import (
	"context"

	"github.com/xcsean/ApplicationEngine/core/protocol/ghost"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"google.golang.org/grpc"
)

type vmEntityContext struct {
	division string
	version  string
	addr     string
	vmID     string
}
type vmEntityStreamContext struct {
	conn   *grpc.ClientConn
	stream ghost.VMService_OnNotifyPacketClient
}

func vmEntityInitStream(vmAddr string) (*vmEntityStreamContext, error) {
	conn, err := grpc.Dial(vmAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := ghost.NewVMServiceClient(conn)
	stream, err := c.OnNotifyPacket(context.Background())
	if err != nil {
		return nil, err
	}
	return &vmEntityStreamContext{conn: conn, stream: stream}, nil
}

func vmEntityLoop(ent *vmEntityContext, pktChannel chan *ghost.GhostPacket, inChannel, outChannel chan *innerCmd) {
	var err error
	var ctx *vmEntityStreamContext
	for {
		exit := false
		select {
		case cmd := <-inChannel:
			cmdID := cmd.getID()
			if cmdID == innerCmdVMStart {
				log.Info("vm %s %s %s %s start", ent.division, ent.version, ent.addr, ent.vmID)
				ctx, err = vmEntityInitStream(ent.addr)
				if err != nil {
					log.Error("vm %s %s %s %s stream conn failed: %s", ent.division, ent.version, ent.addr, ent.vmID, err.Error())
					outChannel <- newVMMCmd(innerCmdVMStreamConnFault, ent.division, ent.vmID, "")
					exit = true
				} else {
					defer ctx.conn.Close()
					log.Info("vm %s %s %s %s stream ready", ent.division, ent.version, ent.addr, ent.vmID)
				}
			} else if cmdID == innerCmdVMShouldExit {
				log.Info("vm %s %s %s %s recv VMShouldExit command, so exit", ent.division, ent.version, ent.addr, ent.vmID)
				exit = true
			}
		case pkt := <-pktChannel:
			err = ctx.stream.Send(pkt)
			if err != nil {
				log.Info("vm %s %s %s %s steam send failed: %s", ent.division, ent.version, ent.addr, ent.vmID, err.Error())
				outChannel <- newVMMCmd(innerCmdVMStreamSendFault, ent.division, ent.vmID, "")
				exit = true
			}
		}

		if exit {
			break
		}
	}

	log.Info("vm %s %s %s %s exit", ent.division, ent.version, ent.addr, ent.vmID)
}
