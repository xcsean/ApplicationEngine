package main

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	reqChannel chan<- *reqRPC
)

type myService struct{}

func (s *myService) RegisterMaster(ctx context.Context, req *protocol.RegisterMasterReq) (*protocol.RegisterMasterRsp, error) {
	rsp := &protocol.RegisterMasterRsp{Result: errno.OK}
	return rsp, nil
}

func (s *myService) AllocSessionID(ctx context.Context, req *protocol.SessionAllocReq) (*protocol.SessionAllocRsp, error) {
	rsp := &protocol.SessionAllocRsp{Result: errno.OK}
	rspChannel := make(chan *rspRPC, 1)
	reqChannel <- &reqRPC{Type: innerCmdRPCAllocSessionID, StrParam: fmt.Sprintf("%d", req.Count), Rsp: rspChannel}
	cmd := <-rspChannel
	rsp.Result = cmd.Result
	ss := strings.Split(cmd.StrParam, ",")
	for i := 0; i < len(ss); i++ {
		sessionID, _ := strconv.ParseUint(ss[i], 10, 64)
		if sessionID != 0 {
			rsp.SessionIds = append(rsp.SessionIds, sessionID)
		}
	}
	return rsp, nil
}

func (s *myService) IsSessionAlive(ctx context.Context, req *protocol.SessionAliveReq) (*protocol.SessionAliveRsp, error) {
	rsp := &protocol.SessionAliveRsp{Result: errno.OK}
	rspChannel := make(chan *rspRPC, 1)
	reqChannel <- &reqRPC{Type: innerCmdRPCIsSessionAlive, StrParam: fmt.Sprintf("%d", req.SessionId), Rsp: rspChannel}
	cmd := <-rspChannel
	status, _ := strconv.ParseInt(cmd.StrParam, 10, 64)
	rsp.Result = cmd.Result
	rsp.Status = int32(status)
	return rsp, nil
}

func startRPCLoop(ls net.Listener, rpcChannel chan<- *reqRPC) {
	defer ls.Close()

	reqChannel = rpcChannel

	srv := grpc.NewServer()
	protocol.RegisterGconndServiceServer(srv, &myService{})
	reflection.Register(srv)
	srv.Serve(ls)

	log.Info("RPC service exit")
}
