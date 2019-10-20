package main

import (
	"context"
	"net"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	reqChannel chan<- *innerCmd
)

type myService struct{}

func (s *myService) RegisterMaster(ctx context.Context, req *protocol.RegisterMasterReq) (*protocol.RegisterMasterRsp, error) {
	rsp := &protocol.RegisterMasterRsp{Result: errno.OK}
	return rsp, nil
}

func (s *myService) AllocSessionID(ctx context.Context, req *protocol.SessionAllocReq) (*protocol.SessionAllocRsp, error) {
	rsp := &protocol.SessionAllocRsp{Result: errno.OK}
	return rsp, nil
}

func (s *myService) IsSessionAlive(ctx context.Context, req *protocol.SessionAliveReq) (*protocol.SessionAliveRsp, error) {
	rsp := &protocol.SessionAliveRsp{Result: errno.OK}
	return rsp, nil
}

func startRPCLoop(ls net.Listener, rpcChannel chan<- *innerCmd) {
	defer ls.Close()

	reqChannel = rpcChannel

	srv := grpc.NewServer()
	protocol.RegisterGconndServiceServer(srv, &myService{})
	reflection.Register(srv)
	srv.Serve(ls)

	log.Info("RPC service exit")
}
