package main

import (
	"context"
	"strings"

	"github.com/xcsean/ApplicationEngine/core/protocol/ghost"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/etc"
	"google.golang.org/grpc/peer"
)

var (
	reqChannel chan *innerCmd
)

type myService struct{}

func (s *myService) RegisterVM(ctx context.Context, req *ghost.RegisterVmReq) (*ghost.RegisterVmRsp, error) {
	defer dbg.Stacktrace()

	rsp := &ghost.RegisterVmRsp{Result: errno.OK}
	result := validateRemote(ctx, req.Division)
	if result != errno.OK {
		rsp.Result = result
		return rsp, nil
	}

	rspChannel := make(chan *rspRPC, 1)
	reqChannel <- newRPCReq(innerCmdRegisterVM, req.Division, req.Version, rspChannel)

	cmd := <-rspChannel
	result = cmd.getRPCRsp()

	rsp.Result = result
	return rsp, nil
}

func (s *myService) UnregisterVM(ctx context.Context, req *ghost.UnregisterVmReq) (*ghost.UnregisterVmRsp, error) {
	defer dbg.Stacktrace()

	rsp := &ghost.UnregisterVmRsp{Result: errno.OK}
	result := validateRemote(ctx, req.Division)
	if result != errno.OK {
		rsp.Result = result
		return rsp, nil
	}

	rspChannel := make(chan *rspRPC, 1)
	reqChannel <- newRPCReq(innerCmdUnregisterVM, req.Division, req.Version, rspChannel)

	cmd := <-rspChannel
	result = cmd.getRPCRsp()

	rsp.Result = result
	return rsp, nil
}

func (s *myService) LoadUserAsset(ctx context.Context, req *ghost.LoadUserassetReq) (*ghost.LoadUserassetRsp, error) {
	defer dbg.Stacktrace()

	rsp := &ghost.LoadUserassetRsp{Result: errno.OK}
	return rsp, nil
}

func (s *myService) SaveUserAsset(ctx context.Context, req *ghost.SaveUserassetReq) (*ghost.SaveUserassetRsp, error) {
	defer dbg.Stacktrace()

	rsp := &ghost.SaveUserassetRsp{}
	return rsp, nil
}

func (s *myService) SendPacket(ctx context.Context, req *ghost.SendPacketReq) (*ghost.SendPacketRsp, error) {
	defer dbg.Stacktrace()

	rsp := &ghost.SendPacketRsp{Result: errno.OK}
	return rsp, nil
}

func getRemoteIP(ctx context.Context) (string, int32) {
	pr, ok := peer.FromContext(ctx)
	if !ok {
		return "", errno.RPCDONOTHAVEPEERINFO
	}

	remoteAddr := pr.Addr.String()
	array := strings.Split(remoteAddr, ":")
	remoteIP := array[0]
	return remoteIP, errno.OK
}

func validateRemote(ctx context.Context, division string) int32 {
	ip, result := getRemoteIP(ctx)
	if result != errno.OK {
		return result
	}

	requiredIP, _, _, _, err := etc.SelectNode(division)
	if err != nil {
		return errno.NODENOTFOUNDINREGISTRY
	}
	if requiredIP != ip {
		return errno.NODEIPNOTEQUALREGISTRY
	}

	return errno.OK
}
