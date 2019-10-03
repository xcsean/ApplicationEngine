package main

import (
	"context"
	"net"

	"github.com/xcsean/ApplicationEngine/core/protocol/ghost"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type myService struct{}

func (s *myService) OnNotifyStatus(ctx context.Context, req *ghost.NotifyStatusReq) (*ghost.NotifyStatusRsp, error) {
	sendRPCVMCmd(&hostCmd{Type: hostCmdNotifyStatus})
	rsp := &ghost.NotifyStatusRsp{Result: errno.OK, Desc: ""}
	return rsp, nil
}

func (s *myService) OnNotifyPacket(srv ghost.VMService_OnNotifyPacketServer) error {
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

func startRPC(ls net.Listener) {
	defer ls.Close()

	srv := grpc.NewServer()
	ghost.RegisterVMServiceServer(srv, &myService{})
	reflection.Register(srv)
	srv.Serve(ls)
}
