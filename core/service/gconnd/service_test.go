package main

import (
	"context"
	"testing"
	"time"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"google.golang.org/grpc"
)

func TestAllocSessionIDs(t *testing.T) {
	getcdAddr := "127.0.0.1:18003"
	conn, err := grpc.Dial(getcdAddr, grpc.WithInsecure())
	if err != nil {
		t.Errorf("connect to getcd=%s failed: %s", getcdAddr, err.Error())
		return
	}
	defer conn.Close()

	c := protocol.NewGconndServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &protocol.SessionAllocReq{Count: 100}
	rsp, err := c.AllocSessionID(ctx, req)
	if err != nil {
		t.Errorf("call c.AllocSessionID failed: %s", err.Error())
		return
	}

	if rsp.Result != errno.OK {
		t.Errorf("call c.AllocSessionID result=%d", rsp.Result)
		return
	}

	t.Logf("result=%d", rsp.Result)
	t.Logf("sessionIDs=%v", rsp.SessionIds)
}

func TestIsSessionAlive(t *testing.T) {
	getcdAddr := "127.0.0.1:18003"
	conn, err := grpc.Dial(getcdAddr, grpc.WithInsecure())
	if err != nil {
		t.Errorf("connect to getcd=%s failed: %s", getcdAddr, err.Error())
		return
	}
	defer conn.Close()

	c := protocol.NewGconndServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &protocol.SessionAliveReq{SessionId: 123456}
	rsp, err := c.IsSessionAlive(ctx, req)
	if err != nil {
		t.Errorf("call c.IsSessionAlive failed: %s", err.Error())
		return
	}

	if rsp.Result != errno.OK {
		t.Errorf("call c.IsSessionAlive result=%d", rsp.Result)
		return
	}

	t.Logf("result=%d", rsp.Result)
	t.Logf("status=%v", rsp.Status)
}
