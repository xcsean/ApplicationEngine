package main

import (
	"context"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	svc "github.com/xcsean/ApplicationEngine/core/shared/service"
)

type myService struct{}

func (s *myService) QueryRegistry(ctx context.Context, req *protocol.QueryRegistryReq) (*protocol.QueryRegistryRsp, error) {
	defer dbg.Stacktrace()

	rsp := &protocol.QueryRegistryRsp{Result: errno.OK}

	server := getServerMap()
	serverLen := getServerCount(server)

	service := getServiceMap()
	serviceLen := getServiceCount(service)

	serverArray := make([]*protocol.RegistryServer, serverLen)
	i := 0
	for _, v := range server {
		c := v
		s := &protocol.RegistryServer{
			App:           c.App,
			Server:        c.Server,
			Division:      c.Division,
			Node:          c.Node,
			UseAgent:      c.UseAgent,
			NodeStatus:    c.NodeStatus,
			ServiceStatus: c.ServiceStatus,
		}
		serverArray[i] = s
		i++
	}
	rsp.Servers = serverArray

	serviceArray := make([]*protocol.RegistryService, serviceLen)
	i = 0
	for _, v := range service {
		for e := v.Front(); e != nil; e = e.Next() {
			c := e.Value.(svc.RegistryServiceConfig)
			s := &protocol.RegistryService{
				App:         c.App,
				Server:      c.Server,
				Division:    c.Division,
				Node:        c.Node,
				Service:     c.Service,
				ServiceIp:   c.ServiceIP,
				ServicePort: c.ServicePort,
				RpcPort:     c.RPCPort,
				AdminPort:   c.AdminPort,
			}
			serviceArray[i] = s
			i++
		}
	}
	rsp.Services = serviceArray

	return rsp, nil
}

func (s *myService) QueryGlobalConfig(ctx context.Context, req *protocol.QueryGlobalConfigReq) (*protocol.QueryGlobalConfigRsp, error) {
	defer dbg.Stacktrace()

	rsp := &protocol.QueryGlobalConfigRsp{Result: errno.OK}

	allConfig := getGlobalConfigMap()
	var entries []*protocol.CategoryEntry
	if len(req.Categories) == 0 {
		entries = make([]*protocol.CategoryEntry, 0, len(allConfig))
		for category, categoryArray := range allConfig {
			kv := make(map[string]string)
			for _, v := range categoryArray {
				kv[v.Key] = v.Value
			}
			entries = append(entries, &protocol.CategoryEntry{Category: category, Kv: kv})
		}
	} else {
		entries = make([]*protocol.CategoryEntry, 0, len(req.Categories))
		for _, category := range req.Categories {
			categoryArray, ok := allConfig[category]
			if !ok {
				log.Error("category:%s not exist", category)
				continue
			}
			kv := make(map[string]string)
			for _, v := range categoryArray {
				kv[v.Key] = v.Value
			}
			entries = append(entries, &protocol.CategoryEntry{Category: category, Kv: kv})
		}
	}

	rsp.Entries = entries
	return rsp, nil
}

func (s *myService) QueryProtoLimit(ctx context.Context, req *protocol.QueryProtoLimitReq) (*protocol.QueryProtoLimitRsp, error) {
	defer dbg.Stacktrace()

	rsp := &protocol.QueryProtoLimitRsp{Result: errno.OK}

	entries := make([]*protocol.ProtoLimitEntry, 0, len(protoLimitArr))
	for _, v := range protoLimitArr {
		entries = append(entries, &protocol.ProtoLimitEntry{
			ProtoId:             int32(v.ProtoID),
			PlayerLimitEnable:   int32(v.PlayerLimitEnable),
			PlayerLimitCount:    int32(v.PlayerLimitCount),
			PlayerLimitDuration: int32(v.PlayerLimitDuration),
			ServerLimitEnable:   int32(v.ServerLimitEnable),
			ServerLimitCount:    int32(v.ServerLimitCount),
			ServerLimitDuration: int32(v.ServerLimitDuration),
		})
	}

	rsp.Entries = entries
	return rsp, nil
}
