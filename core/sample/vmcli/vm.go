package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	ui "github.com/jroimartin/gocui"
	"github.com/xcsean/ApplicationEngine/core/protocol"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"google.golang.org/grpc"
)

const (
	vmView      = "VMView"
	vmTitle     = "vm message"
	vmEdit      = "VMEdit"
	vmEditTitle = "vm input"
	version     = "1.1.1.1"
)

var (
	kbdVMChannel chan string
	rpcVMChannel chan *hostCmd
	sndVMChannel chan *protocol.SessionPacket
	hostAddr     string
	division     string
	vmID         uint64
)

func getVMView() string {
	return vmView
}

func getVMTitle() string {
	return vmTitle
}

func getVMEdit() string {
	return vmEdit
}

func getVMEditTitle() string {
	return vmEditTitle
}

func sendVMKeyboard(text string) {
	kbdVMChannel <- text
}

func sendRPCVMCmd(cmd *hostCmd) {
	rpcVMChannel <- cmd
}

func vmLoop(addr, vmAddr string, g *ui.Gui) {
	hostAddr = addr
	division = config.Division
	kbdVMChannel = make(chan string, 100)
	rpcVMChannel = make(chan *hostCmd, 100)
	sndVMChannel = make(chan *protocol.SessionPacket, 100)
	vmLog := func(s string) {
		g.Update(func(g *ui.Gui) error {
			v, _ := g.View(vmView)
			fmt.Fprintln(v, s)
			return nil
		})
	}

	// bind the service to vmAddr
	ls, err := net.Listen("tcp", vmAddr)
	if err != nil {
		vmLog(fmt.Sprintf("[VM] can't bind %s", vmAddr))
		time.Sleep(1 * time.Second)
		os.Exit(-1)
	}
	go startRPC(ls)

	// init the send goroutine which send pkts to host
	exitC := make(chan struct{}, 1)
	go hostSendLoop(addr, exitC, sndVMChannel, vmLog)

	// wait for message from kbdVMChannel & netVMChannel
	for {
		select {
		case cmd := <-kbdVMChannel:
			dealVMKeyboard(cmd, vmLog)
		case cmd := <-rpcVMChannel:
			dealHostRPC(cmd, vmLog)
		}
	}
}

func dealVMKeyboard(text string, vmLog func(s string)) {
	text = strings.Replace(text, "\n", "", -1)
	text = strings.Replace(text, "\t", " ", -1)
	array := strings.Fields(text)
	if array == nil {
		return
	}

	cmd := array[0]
	if cmd == "register" {
		callRegisterVM(division, version, vmLog)
	} else if cmd == "unregister" {
		callUnregisterVM(division, version, vmLog)
	} else if cmd == "debug" {
		if len(array) >= 2 {
			cmdOp := array[1]
			cmdParam := ""
			if len(array) >= 3 {
				cmdParam = array[2]
			}
			if cmdOp == "dump" {
				cmdParam = fmt.Sprintf("%d", vmID)
			}
			callDebug(division, cmdOp, cmdParam, vmLog)
		} else {
			vmLog("help debug: debug op param")
		}
	}

}

func dealHostRPC(cmd *hostCmd, vmLog func(s string)) {
	switch cmd.Type {
	case hostCmdRemoteClosed:
		vmLog("[CB] stream closed")
	case hostCmdNotifyStatus:
		vmLog("[CB] notify status")
	case hostCmdNotifyPacket:
		dealHostPkt(cmd, vmLog)
	}
}

func dealHostPkt(cmd *hostCmd, vmLog func(s string)) {
	switch cmd.Pkt.Common.CmdId {
	case cmdLogin:
		var rb cmdBody
		err := json.Unmarshal([]byte(cmd.Pkt.Common.Body), &rb)
		if err != nil {
			return
		}
		sUUID, ok := rb.Kv["uuid"]
		if !ok {
			return
		}
		uuid, err := strconv.ParseInt(sUUID, 10, 64)
		if err != nil {
			return
		}
		go callBindSession(cmd.Pkt.Sessions[0], uint64(uuid), vmLog)
	case uint32(protocol.Packet_PRIVATE_NOTIFY_VM_UNBIND):
		var rb protocol.PacketReservedBody
		err := json.Unmarshal([]byte(cmd.Pkt.Common.Body), &rb)
		if err != nil {
			return
		}
		sUUID, ok := rb.Kv["uuid"]
		if !ok {
			return
		}
		uuid, err := strconv.ParseInt(sUUID, 10, 64)
		if err != nil {
			return
		}
		go callUnbindSession(cmd.Pkt.Sessions[0], uint64(uuid), vmLog)
	default:
		vmLog(fmt.Sprintf("[CB] notify packet cmd=%d body=%s", cmd.Pkt.Common.CmdId, cmd.Pkt.Common.Body))
	}
}

func callRegisterVM(division, version string, vmLog func(s string)) {
	callGhost(func(c protocol.GhostServiceClient, ctx context.Context) error {
		req := &protocol.RegisterVmReq{
			Division: division,
			Version:  version,
		}
		rsp, err := c.RegisterVM(ctx, req)
		if err != nil {
			return err
		}

		vmLog(fmt.Sprintf("[VM] register result=%d, uuid=%d", rsp.Result, rsp.Vmid))
		if rsp.Result == errno.OK {
			vmID = rsp.Vmid
		}
		return nil
	}, 3)
}

func callUnregisterVM(division, version string, vmLog func(s string)) {
	callGhost(func(c protocol.GhostServiceClient, ctx context.Context) error {
		req := &protocol.UnregisterVmReq{
			Division: division,
			Version:  version,
			Vmid:     vmID,
		}
		rsp, err := c.UnregisterVM(ctx, req)
		if err != nil {
			return err
		}

		vmLog(fmt.Sprintf("[VM] unregister result=%d", rsp.Result))
		if rsp.Result == errno.OK {
			vmID = 0
		}
		return nil
	}, 3)
}

func callBindSession(sessionID, uuid uint64, vmLog func(s string)) {
	vmLog(fmt.Sprintf("[VM] bind sesssion=%d uuid=%d", sessionID, uuid))
	result := int32(0)
	callGhost(func(c protocol.GhostServiceClient, ctx context.Context) error {
		req := &protocol.BindSessionReq{
			Division:  division,
			Sessionid: sessionID,
			Uuid:      uuid,
		}
		rsp, err := c.BindSession(ctx, req)
		if err != nil {
			return err
		}

		vmLog(fmt.Sprintf("[VM] bind session result=%d", rsp.Result))
		result = rsp.Result
		return nil
	}, 3)

	if result == int32(errno.OK) {
		innerBody := &cmdBody{
			StrParam: "",
			Kv:       make(map[string]string),
		}
		innerBody.Kv["result"] = fmt.Sprintf("%d", result)
		innerBody.Kv["uuid"] = fmt.Sprintf("%d", uuid)
		body, _ := json.Marshal(innerBody)
		pkt := &protocol.SessionPacket{
			Common: &protocol.Packet{
				CmdId:     cmdLogin,
				UserData:  0,
				Timestamp: 0,
				Body:      string(body[:]),
			},
			Sessions: []uint64{sessionID},
		}
		hostSend(pkt)
	} else if result == int32(errno.HOSTVMBINDNEEDRETRY) {
		time.Sleep(1 * time.Second)
		go callBindSession(sessionID, uuid, vmLog)
	}
}

func callUnbindSession(sessionID, uuid uint64, vmLog func(s string)) {
	vmLog(fmt.Sprintf("[VM] unbind session=%d uuid=%d", sessionID, uuid))
	callGhost(func(c protocol.GhostServiceClient, ctx context.Context) error {
		req := &protocol.UnbindSessionReq{
			Division:  division,
			Sessionid: sessionID,
			Uuid:      uuid,
		}
		rsp, err := c.UnbindSession(ctx, req)
		if err != nil {
			return err
		}

		vmLog(fmt.Sprintf("[VM] unbind session result=%d", rsp.Result))
		return nil
	}, 3)
}

func callDebug(division, cmdOp, cmdParam string, vmLog func(s string)) {
	callGhost(func(c protocol.GhostServiceClient, ctx context.Context) error {
		req := &protocol.DebugReq{
			Division: division,
			Cmdop:    cmdOp,
			Cmdparam: cmdParam,
		}
		rsp, err := c.Debug(ctx, req)
		if err != nil {
			return err
		}

		vmLog(fmt.Sprintf("[VM] debug result=%d desc='%s'", rsp.Result, rsp.Desc))
		return nil
	}, 3)
}

func callGhost(handler func(c protocol.GhostServiceClient, ctx context.Context) error, timeout time.Duration) error {
	conn, err := grpc.Dial(hostAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	c := protocol.NewGhostServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	return handler(c, ctx)
}

func hostSend(pkt *protocol.SessionPacket) {
	sndVMChannel <- pkt
}

type hostContext struct {
	conn   *grpc.ClientConn
	stream protocol.GhostService_SendPacketClient
}

func hostInitStream(addr string) (*hostContext, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	c := protocol.NewGhostServiceClient(conn)
	stream, err := c.SendPacket(context.Background())
	if err != nil {
		return nil, err
	}
	return &hostContext{conn: conn, stream: stream}, nil
}

func hostSendLoop(addr string, in chan struct{}, out chan *protocol.SessionPacket, vmLog func(s string)) {
	ctx, err := hostInitStream(addr)
	if err != nil {
		vmLog(fmt.Sprintf("[VM] stream init failed: %s", err.Error()))
		return
	}

	vmLog("[VM] host send loop start")
	for {
		exit := false
		select {
		case <-in:
			exit = true
		case pkt := <-out:
			err = ctx.stream.Send(pkt)
			if err != nil {
				vmLog(fmt.Sprintf("[VM] send to host: %s failed: %s", addr, err.Error()))
			}
		}
		if exit {
			break
		}
	}
	vmLog("[VM] host send loop exit")
}
