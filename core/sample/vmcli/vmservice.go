package main

import (
	"context"
	"net"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type myService struct{}

func (s *myService) NotifyStatus(ctx context.Context, req *protocol.NotifyStatusReq) (*protocol.NotifyStatusRsp, error) {
	sendRPCVMCmd(&hostCmd{Type: hostCmdNotifyStatus})
	rsp := &protocol.NotifyStatusRsp{Result: errno.OK, Desc: ""}
	return rsp, nil
}

func (s *myService) Push(srv protocol.VMService_PushServer) error {
	for {
		if pkt, err := srv.Recv(); err == nil {
			cmd := &hostCmd{Type: hostCmdNotifyPacket, Pkt: pkt}
			sendRPCVMCmd(cmd)
		} else {
			sendRPCVMCmd(&hostCmd{Type: hostCmdRemoteClosed})
			break
		}
	}
	return nil
}

func (s *myService) Pull(_ *protocol.StreamSessionPacketReq, srv protocol.VMService_PullServer) error {
	for {
		pkt := <-sndVMChannel
		err := srv.Send(pkt)
		if err != nil {
			break
		}
	}
	return nil
}

func startRPC(ls net.Listener) {
	defer ls.Close()

	srv := grpc.NewServer()
	protocol.RegisterVMServiceServer(srv, &myService{})
	reflection.Register(srv)
	srv.Serve(ls)
}
