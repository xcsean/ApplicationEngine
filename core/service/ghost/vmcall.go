package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/xcsean/ApplicationEngine/core/protocol/ghost"
	"github.com/xcsean/ApplicationEngine/core/shared/etc"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"google.golang.org/grpc"
)

func callVM(vmAddr string, handler func(c ghost.VMServiceClient, ctx context.Context) error, timeout time.Duration) error {
	conn, err := grpc.Dial(vmAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	c := ghost.NewVMServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	return handler(c, ctx)
}

func debugNotifyStatusToVM(division, cmdParam string) {
	ip, _, _, rpcPort, err := etc.PickEndpoint(division)
	if err != nil {
		log.Error("debug notify to vm %s failed: %s", division, err.Error())
		return
	}
	vmAddr := fmt.Sprintf("%s:%d", ip, rpcPort)
	status, err := strconv.ParseInt(cmdParam, 10, 32)
	if err != nil {
		status = 0
	}
	err = callVM(vmAddr, func(c ghost.VMServiceClient, ctx context.Context) error {
		req := &ghost.NotifyStatusReq{
			Status: uint32(status),
		}
		rsp, err := c.OnNotifyStatus(ctx, req)
		if err != nil {
			return err
		}
		log.Debug("call vm %s OnNotifyStatus rsp=%v", division, rsp)
		return nil
	}, 3)
	if err != nil {
		log.Error("call vm %s OnNotifyStatus failed: %s", division, err.Error())
	}
}
