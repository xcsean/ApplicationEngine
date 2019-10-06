package main

import (
	"context"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"google.golang.org/grpc"
)

type vmEntityContext struct {
	division string
	version  string
	addr     string
	vmID     string
}
type vmEntityPushContext struct {
	conn   *grpc.ClientConn
	stream protocol.VMService_PushClient
}

func vmEntityPushInit(vmAddr string) (*vmEntityPushContext, error) {
	conn, err := grpc.Dial(vmAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := protocol.NewVMServiceClient(conn)
	stream, err := c.Push(context.Background())
	if err != nil {
		return nil, err
	}
	return &vmEntityPushContext{conn: conn, stream: stream}, nil
}

func vmEntityPushLoop(ent *vmEntityContext, pktChannel chan *protocol.SessionPacket, exitC chan struct{}, outChannel chan *innerCmd) {
	ctx, err := vmEntityPushInit(ent.addr)
	if err != nil {
		log.Error("vm push %s %s %s %s stream init failed: %s", ent.division, ent.version, ent.addr, ent.vmID, err.Error())
		outChannel <- newVMMCmd(innerCmdVMStreamInitFault, ent.division, ent.vmID, "")
		return
	}
	defer ctx.conn.Close()
	log.Info("vm push %s %s %s %s start", ent.division, ent.version, ent.addr, ent.vmID)

	for {
		exit := false
		select {
		case <-exitC:
			log.Info("vm push %s %s %s %s recv exit command, so exit", ent.division, ent.version, ent.addr, ent.vmID)
			exit = true
		case pkt := <-pktChannel:
			err = ctx.stream.Send(pkt)
			if err != nil {
				log.Error("vm push %s %s %s %s steam send failed: %s", ent.division, ent.version, ent.addr, ent.vmID, err.Error())
				outChannel <- newVMMCmd(innerCmdVMStreamSendFault, ent.division, ent.vmID, "")
				exit = true
			}
		}

		if exit {
			break
		}
	}

	log.Info("vm push %s %s %s %s exit", ent.division, ent.version, ent.addr, ent.vmID)
}

type vmEntityPullContext struct {
	conn   *grpc.ClientConn
	stream protocol.VMService_PullClient
}

func vmEntityPullInit(vmAddr string) (*vmEntityPullContext, error) {
	conn, err := grpc.Dial(vmAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := protocol.NewVMServiceClient(conn)
	stream, err := c.Pull(context.Background(), &protocol.StreamSessionPacketReq{Result: 1})
	if err != nil {
		return nil, err
	}
	return &vmEntityPullContext{conn: conn, stream: stream}, nil
}

func vmEntityPullLoop(ent *vmEntityContext, exitC chan struct{}, outChannel chan *innerCmd) {
	ctx, err := vmEntityPullInit(ent.addr)
	if err != nil {
		log.Error("vm pull %s %s %s %s stream init failed: %s", ent.division, ent.version, ent.addr, ent.vmID, err.Error())
		outChannel <- newVMMCmd(innerCmdVMStreamInitFault, ent.division, ent.vmID, "")
		return
	}
	defer ctx.conn.Close()
	log.Info("vm pull %s %s %s %s start", ent.division, ent.version, ent.addr, ent.vmID)

	for {
		pkt, err := ctx.stream.Recv()
		if err != nil {
			log.Error("vm pull %s %s %s %s stream recv failed: %s", ent.division, ent.version, ent.addr, ent.vmID, err.Error())
			outChannel <- newVMMCmd(innerCmdVMStreamRecvFault, ent.division, ent.vmID, "")
			break
		} else {
			cmd := newVMMCmd(innerCmdVMStreamRecvPkt, ent.division, ent.vmID, "")
			cmd.pkt = pkt
			outChannel <- cmd
		}
	}

	log.Info("vm pull %s %s %s %s exit", ent.division, ent.version, ent.addr, ent.vmID)
}
