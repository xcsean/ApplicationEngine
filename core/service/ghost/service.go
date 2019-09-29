package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/xcsean/ApplicationEngine/core/protocol/ghost"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/etc"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"google.golang.org/grpc/peer"
)

var (
	reqChannel chan *innerCmd
)

type myService struct{}

func (s *myService) RegisterVM(ctx context.Context, req *ghost.RegisterVmReq) (*ghost.RegisterVmRsp, error) {
	defer dbg.Stacktrace()

	rsp := &ghost.RegisterVmRsp{Result: errno.OK}
	addr, result := validateRemote(ctx, req.Division)
	if result != errno.OK {
		rsp.Result = result
		return rsp, nil
	}

	rspChannel := make(chan *rspRPC, 1)
	reqChannel <- newRPCReq(innerCmdRegisterVM, req.Division, req.Version, addr, 0, rspChannel)

	cmd := <-rspChannel
	result, uuid, _ := cmd.getRPCRsp()
	log.Debug("register vm %s %s, result=%d", req.Division, req.Version, result)

	rsp.Result = result
	rsp.Uuid = uuid
	return rsp, nil
}

func (s *myService) UnregisterVM(ctx context.Context, req *ghost.UnregisterVmReq) (*ghost.UnregisterVmRsp, error) {
	defer dbg.Stacktrace()

	rsp := &ghost.UnregisterVmRsp{Result: errno.OK}
	_, result := validateRemote(ctx, req.Division)
	if result != errno.OK {
		rsp.Result = result
		return rsp, nil
	}

	rspChannel := make(chan *rspRPC, 1)
	reqChannel <- newRPCReq(innerCmdUnregisterVM, req.Division, req.Version, "", req.Uuid, rspChannel)

	cmd := <-rspChannel
	result, _, _ = cmd.getRPCRsp()
	log.Debug("unregister vm %s %s, result=%d", req.Division, req.Version, result)

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

func (s *myService) Debug(ctx context.Context, req *ghost.DebugReq) (*ghost.DebugRsp, error) {
	defer dbg.Stacktrace()

	rsp := &ghost.DebugRsp{Result: errno.OK, Desc: ""}
	_, result := validateRemote(ctx, req.Division)
	if result != errno.OK {
		rsp.Result = result
		return rsp, nil
	}

	rspChannel := make(chan *rspRPC, 1)
	reqChannel <- newRPCReq(innerCmdDebug, req.Division, req.Cmdop, req.Cmdparam, 0, rspChannel)

	cmd := <-rspChannel
	result, _, desc := cmd.getRPCRsp()
	log.Debug("debug vm %s op='%s' param='%s', result=%d", req.Division, req.Cmdop, req.Cmdparam, result)

	rsp.Result = result
	rsp.Desc = desc
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

func validateRemote(ctx context.Context, division string) (string, int32) {
	ip, result := getRemoteIP(ctx)
	if result != errno.OK {
		return "", result
	}

	requiredIP, _, _, rpcPort, err := etc.PickEndpoint(division)
	if err != nil {
		return "", errno.NODENOTFOUNDINREGISTRY
	}
	if requiredIP != ip {
		return "", errno.NODEIPNOTEQUALREGISTRY
	}

	return fmt.Sprintf("%s:%d", requiredIP, rpcPort), errno.OK
}
