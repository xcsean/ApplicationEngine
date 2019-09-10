package main

import (
	"testing"
	"time"

	"github.com/xcsean/ApplicationEngine/core/protocol/getcd"
	rc "github.com/xcsean/ApplicationEngine/core/shared/errno"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func TestCallGetcd(t * testing.T) {
	getcdAddr := "127.0.0.1:17000"
	conn, err := grpc.Dial(getcdAddr, grpc.WithInsecure())
	if err != nil {
		t.Errorf("connect to getcd=%s failed: %s", getcdAddr, err.Error())
		return
	}
	defer conn.Close()

	c := getcd.NewGetcdServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	categories := []string{"global"}
	r, err := c.QueryGlobalConfig(ctx, &getcd.QueryGlobalConfigReq{Categories: categories})
	if err != nil {
		t.Errorf("call c.QueryGlobalConfig failed: %s", err.Error())
		return
	}

	if r.Result != rc.OK {
		t.Errorf("c.QueryGlobalConfig result=%d", r.Result)
		return
	}

	t.Logf("entries=%v", r.Entries)
	t.Logf("test ok")
}