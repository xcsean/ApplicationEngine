package main

import (
	"context"

	"github.com/xcsean/ApplicationEngine/core/protocol/ghost"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
)

type myService struct{}

func (s *myService) RegisterVM(ctx context.Context, req *ghost.RegisterVmReq) (*ghost.RegisterVmRsp, error) {
	defer dbg.Stacktrace()

	rsp := &ghost.RegisterVmRsp{Result: errno.OK}
	return rsp, nil
}

func (s *myService) UnregisterVM(ctx context.Context, req *ghost.UnregisterVmReq) (*ghost.UnregisterVmRsp, error) {
	defer dbg.Stacktrace()

	rsp := &ghost.UnregisterVmRsp{Result: errno.OK}
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
