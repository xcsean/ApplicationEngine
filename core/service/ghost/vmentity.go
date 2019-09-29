package main

import (
	"context"

	"github.com/xcsean/ApplicationEngine/core/protocol/ghost"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"google.golang.org/grpc"
)

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

func vmEntityLoop(pktChannel chan *ghost.GhostPacket, inChannel, outChannel chan *innerCmd) {
	var err error
	var ctx *vmEntityStreamContext
	division, version, addr, uuid := "", "", "", uint64(0)
	for {
		exit := false
		select {
		case cmd := <-inChannel:
			cmdID := cmd.getID()
			if cmdID == innerCmdVMStart {
				division, version, addr, uuid = cmd.getVMMCmd()
				log.Info("vm %s %s %s %d start", division, version, addr, uuid)
				ctx, err = vmEntityInitStream(addr)
				if err != nil {
					log.Error("vm %s %s %s %d stream conn failed: %s", division, version, addr, uuid, err.Error())
					outChannel <- newVMMCmd(innerCmdVMStreamConnFault, division, version, addr, uuid)
					exit = true
				} else {
					defer ctx.conn.Close()
					log.Info("vm %s %s %s %d stream ready", division, version, addr, uuid)
				}
			} else if cmdID == innerCmdVMShouldExit {
				exit = true
			}
		case pkt := <-pktChannel:
			err = ctx.stream.Send(pkt)
			if err != nil {
				log.Info("vm %s %s %s %d steam send failed: %s", division, version, addr, uuid, err.Error())
				outChannel <- newVMMCmd(innerCmdVMStreamSendFault, division, version, addr, uuid)
				exit = true
			}
		}

		if exit {
			break
		}
	}

	log.Info("vm %s %s %s %d exit", division, version, addr, uuid)
}
