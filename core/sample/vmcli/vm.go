package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	ui "github.com/jroimartin/gocui"
	"github.com/xcsean/ApplicationEngine/core/protocol/ghost"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"google.golang.org/grpc"
)

const (
	vmView      = "VMView"
	vmTitle     = "vm message"
	vmEdit      = "VMEdit"
	vmEditTitle = "vm input"
	division    = "app.td.2"
	version     = "1.1.1.1"
)

var (
	kbdVMChannel chan string
	rpcVMChannel chan string
	hostAddr     string
	vmsvcAddr    string
	vmUuid       uint64
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

func sendRPCVMText(text string) {
	rpcVMChannel <- text
}

func vmLoop(addr, vmAddr string, g *ui.Gui) {
	hostAddr = addr
	vmsvcAddr = vmAddr
	kbdVMChannel = make(chan string, 100)
	rpcVMChannel = make(chan string, 100)
	vmLog := func(s string) {
		g.Update(func(g *ui.Gui) error {
			v, _ := g.View(vmView)
			fmt.Fprintln(v, s)
			return nil
		})
	}

	g.Update(func(g *ui.Gui) error {
		g.SetCurrentView(getVMEdit())
		return nil
	})

	// bind the service to vmsvcAddr
	ls, err := net.Listen("tcp", vmsvcAddr)
	if err != nil {
		vmLog(fmt.Sprintf("[VM] can't bind %s", vmsvcAddr))
		time.Sleep(1 * time.Second)
		os.Exit(-1)
	}
	go startRPC(ls)

	// wait for message from kbdVMChannel & netVMChannel
	for {
		select {
		case cmd := <-kbdVMChannel:
			dealVMKeyboard(cmd, vmLog)
		case cmd := <-rpcVMChannel:
			vmLog(cmd)
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
			callDebug(division, cmdOp, cmdParam, vmLog)
		} else {
			vmLog("help debug: debug cmdline")
		}
	}

}

func callRegisterVM(division, version string, vmLog func(s string)) {
	callGhost(func(c ghost.GhostServiceClient, ctx context.Context) error {
		req := &ghost.RegisterVmReq{
			Division: division,
			Version:  version,
		}
		rsp, err := c.RegisterVM(ctx, req)
		if err != nil {
			return err
		}

		vmLog(fmt.Sprintf("[VM] register result=%d, uuid=%d", rsp.Result, rsp.Uuid))
		if rsp.Result == errno.OK {
			vmUuid = rsp.Uuid
		}
		return nil
	}, 3)
}

func callUnregisterVM(division, version string, vmLog func(s string)) {
	callGhost(func(c ghost.GhostServiceClient, ctx context.Context) error {
		req := &ghost.UnregisterVmReq{
			Division: division,
			Version:  version,
			Uuid:     vmUuid,
		}
		rsp, err := c.UnregisterVM(ctx, req)
		if err != nil {
			return err
		}

		vmLog(fmt.Sprintf("[VM] unregister result=%d", rsp.Result))
		if rsp.Result == errno.OK {
			vmUuid = 0
		}
		return nil
	}, 3)
}

func callDebug(division, cmdOp, cmdParam string, vmLog func(s string)) {
	callGhost(func(c ghost.GhostServiceClient, ctx context.Context) error {
		req := &ghost.DebugReq{
			Division: division,
			Cmdop:    cmdOp,
			Cmdparam: cmdParam,
		}
		rsp, err := c.Debug(ctx, req)
		if err != nil {
			return err
		}

		vmLog(fmt.Sprintf("[VM] debug result=%d", rsp.Result))
		return nil
	}, 3)
}

func callGhost(handler func(c ghost.GhostServiceClient, ctx context.Context) error, timeout time.Duration) error {
	conn, err := grpc.Dial(hostAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	c := ghost.NewGhostServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()

	return handler(c, ctx)
}
