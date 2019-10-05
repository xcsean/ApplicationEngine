package main

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/etc"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
)

var (
	reqChannel chan *innerCmd
)

type myService struct{}

func (s *myService) RegisterVM(ctx context.Context, req *protocol.RegisterVmReq) (*protocol.RegisterVmRsp, error) {
	defer dbg.Stacktrace()

	rsp := &protocol.RegisterVmRsp{Result: errno.OK}
	addr, result := validateRemote(ctx, req.Division)
	if result != errno.OK {
		rsp.Result = result
		return rsp, nil
	}

	rspChannel := make(chan *rspRPC, 1)
	reqChannel <- newRPCReq(innerCmdRegisterVM, req.Division, req.Version, addr, rspChannel)

	cmd := <-rspChannel
	result, sVMID := cmd.getRPCRsp()
	vmID, _ := parseUint64(sVMID)
	log.Debug("register vm %s %s, result=%d", req.Division, req.Version, result)

	rsp.Result = result
	rsp.Vmid = vmID
	return rsp, nil
}

func (s *myService) UnregisterVM(ctx context.Context, req *protocol.UnregisterVmReq) (*protocol.UnregisterVmRsp, error) {
	defer dbg.Stacktrace()

	rsp := &protocol.UnregisterVmRsp{Result: errno.OK}
	_, result := validateRemote(ctx, req.Division)
	if result != errno.OK {
		rsp.Result = result
		return rsp, nil
	}

	rspChannel := make(chan *rspRPC, 1)
	reqChannel <- newRPCReq(innerCmdUnregisterVM, req.Division, req.Version, fmt.Sprintf("%d", req.Vmid), rspChannel)

	cmd := <-rspChannel
	result, _ = cmd.getRPCRsp()
	log.Debug("unregister vm %s %s, result=%d", req.Division, req.Version, result)

	rsp.Result = result
	return rsp, nil
}

func (s *myService) SendPacket(srv protocol.GhostService_SendPacketServer) error {
	for {
		if pkt, err := srv.Recv(); err == nil {
			data, err := conn.MakeSessionPkt(pkt.Sessions, uint16(pkt.CmdId), pkt.UserData, pkt.Timestamp, []byte(pkt.Body))
			if err == nil {
				reqChannel <- newRPCReq(innerCmdSendPacket, string(data[:]), "", "", nil)
			} else {
				log.Error("MakeSessionPkt failed: %s", err.Error())
			}
		} else {
			log.Error("%s", err.Error())
			break
		}
	}
	return nil
}

func (s *myService) BindSession(ctx context.Context, req *protocol.BindSessionReq) (*protocol.BindSessionRsp, error) {
	defer dbg.Stacktrace()

	rsp := &protocol.BindSessionRsp{Result: errno.OK, Division: req.Division, Sessionid: req.Sessionid, Uuid: req.Uuid}
	_, result := validateRemote(ctx, req.Division)
	if result != errno.OK {
		rsp.Result = result
		return rsp, nil
	}

	rspChannel := make(chan *rspRPC, 1)
	reqChannel <- newRPCReq(innerCmdBindSession, req.Division, fmt.Sprintf("%d", req.Sessionid), fmt.Sprintf("%d", req.Uuid), rspChannel)

	cmd := <-rspChannel
	result, _ = cmd.getRPCRsp()
	log.Debug("bind session=%d uuid=%d, result=%d", req.Sessionid, req.Uuid, result)

	rsp.Result = result
	return rsp, nil
}

func (s *myService) UnbindSession(ctx context.Context, req *protocol.UnbindSessionReq) (*protocol.UnbindSessionRsp, error) {
	defer dbg.Stacktrace()

	rsp := &protocol.UnbindSessionRsp{Result: errno.OK, Division: req.Division, Sessionid: req.Sessionid, Uuid: req.Uuid}
	_, result := validateRemote(ctx, req.Division)
	if result != errno.OK {
		rsp.Result = result
		return rsp, nil
	}

	rspChannel := make(chan *rspRPC, 1)
	reqChannel <- newRPCReq(innerCmdUnbindSession, req.Division, fmt.Sprintf("%d", req.Sessionid), fmt.Sprintf("%d", req.Uuid), rspChannel)

	cmd := <-rspChannel
	result, _ = cmd.getRPCRsp()
	log.Debug("unbind session=%d uuid=%d, result=%d", req.Sessionid, req.Uuid, result)

	rsp.Result = result
	return rsp, nil
}

func (s *myService) LockUserAsset(ctx context.Context, req *protocol.LockUserassetReq) (*protocol.LockUserassetRsp, error) {
	defer dbg.Stacktrace()

	rsp := &protocol.LockUserassetRsp{Result: errno.OK}
	return rsp, nil
}

func (s *myService) UnlockUserAsset(ctx context.Context, req *protocol.UnlockUserassetReq) (*protocol.UnlockUserassetRsp, error) {
	defer dbg.Stacktrace()

	rsp := &protocol.UnlockUserassetRsp{Result: errno.OK}
	return rsp, nil
}

func (s *myService) Debug(ctx context.Context, req *protocol.DebugReq) (*protocol.DebugRsp, error) {
	defer dbg.Stacktrace()

	rsp := &protocol.DebugRsp{Result: errno.OK, Desc: ""}
	_, result := validateRemote(ctx, req.Division)
	if result != errno.OK {
		rsp.Result = result
		return rsp, nil
	}

	rspChannel := make(chan *rspRPC, 1)
	reqChannel <- newRPCReq(innerCmdDebug, req.Division, req.Cmdop, req.Cmdparam, rspChannel)

	cmd := <-rspChannel
	result, desc := cmd.getRPCRsp()
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

func startRPC(ls net.Listener, rpcChannel chan *innerCmd) {
	defer ls.Close()

	reqChannel = rpcChannel

	srv := grpc.NewServer()
	protocol.RegisterGhostServiceServer(srv, &myService{})
	reflection.Register(srv)
	srv.Serve(ls)

	log.Info("RPC service exit")
}
