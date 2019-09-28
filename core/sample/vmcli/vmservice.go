package main

import (
	"context"
	"fmt"
	"net"

	"github.com/xcsean/ApplicationEngine/core/protocol/ghost"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type myService struct{}

func (s *myService) OnNotifyStatus(ctx context.Context, req *ghost.NotifyStatusReq) (*ghost.NotifyStatusRsp, error) {
	sendRPCVMText(fmt.Sprintf("[VM] notify status = %d", req.Status))
	rsp := &ghost.NotifyStatusRsp{Result: errno.OK, Desc: ""}
	return rsp, nil
}

func (s *myService) OnNotifyPacket(srv ghost.VMService_OnNotifyPacketServer) error {
	return nil
}

func startRPC(ls net.Listener) {
	defer ls.Close()

	srv := grpc.NewServer()
	ghost.RegisterVMServiceServer(srv, &myService{})
	reflection.Register(srv)
	srv.Serve(ls)
}
