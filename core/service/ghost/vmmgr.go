package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/xcsean/ApplicationEngine/core/protocol/ghost"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/etc"
	"github.com/xcsean/ApplicationEngine/core/shared/id"
)

type vmEntity struct {
	vmID      uint64
	division  string
	version   string
	addr      string
	pkt       chan *ghost.GhostPacket
	in        chan *innerCmd
	checkTime int64
}

type vmEntityStatus struct {
	curLoad int64
	maxLoad int64
}

type vmStatus map[string]vmEntityStatus

type vmMgr struct {
	sf    *id.Snowflake
	out   chan *innerCmd
	vms   map[string]*vmEntity
	vers  map[string]vmStatus // fast-index
	addrs map[string]bool     // fast-index
}

var (
	vmmgr *vmMgr
)

func newVMMgr(sf *id.Snowflake, outChannel chan *innerCmd) *vmMgr {
	vmmgr = &vmMgr{
		sf:    sf,
		out:   outChannel,
		vms:   make(map[string]*vmEntity),
		vers:  make(map[string]vmStatus),
		addrs: make(map[string]bool),
	}
	tmmAddGlobalPeriodicTask("vmm", 1*time.Second, func(c chan *timerCmd) {
		c <- &timerCmd{Type: timerCmdVMMOnTick}
	})
	return vmmgr
}

func getVMMgr() *vmMgr {
	return vmmgr
}

func (vmm *vmMgr) addVM(division, version, addr string) (uint64, int32) {
	_, ok := vmm.addrs[addr]
	if ok {
		return 0, errno.HOSTVMADDRALREADYEXIST
	}
	_, ok = vmm.vms[division]
	if ok {
		return 0, errno.HOSTVMALREADYEXIST
	}

	vmID, _ := vmm.sf.NextID()
	pkt := make(chan *ghost.GhostPacket, 1000)
	in := make(chan *innerCmd, 10)
	checkTime := time.Now().Unix() + etc.GetInt64WithDefault("global", "keepAlive", 10)
	vm := &vmEntity{
		vmID:      vmID,
		division:  division,
		version:   version,
		addr:      addr,
		pkt:       pkt,
		in:        in,
		checkTime: checkTime,
	}

	// division ---> vm
	vmm.vms[vm.division] = vm
	// ip:rpc_port ---> exist
	vmm.addrs[vm.addr] = true
	// version ---> [division ---> status]
	status, ok := vmm.vers[version]
	if !ok {
		status = make(map[string]vmEntityStatus)
		vmm.vers[version] = status
	}
	status[division] = vmEntityStatus{curLoad: 0, maxLoad: 5000}

	in <- newVMMCmd(innerCmdVMStart, division, version, addr, vmID)
	go vmEntityLoop(pkt, in, vmm.out)
	return vmID, errno.OK
}

func (vmm *vmMgr) delVM(division string, vmID uint64) int32 {
	vm, ok := vmm.vms[division]
	if ok && vm.vmID == vmID {
		vm.in <- newVMMCmd(innerCmdVMShouldExit, division, vm.version, vm.addr, vmID)
		close(vm.in)
		close(vm.pkt)
		status, ok := vmm.vers[vm.version]
		if ok {
			delete(status, division)
			if len(status) == 0 {
				delete(vmm.vers, vm.version)
			}
		}
		delete(vmm.addrs, vm.addr)
		delete(vmm.vms, division)
	}
	return errno.OK
}

func (vmm *vmMgr) onTick() {
	now := time.Now().Unix()

	// check keep-alive with all vm(s)
	interval := etc.GetInt64WithDefault("global", "keepAlive", 10)
	for _, vm := range vmm.vms {
		if now >= vm.checkTime {
			vm.pkt <- &ghost.GhostPacket{
				CmdId:     conn.CmdPing,
				UserData:  0,
				Timestamp: uint32(now),
				Sessions:  []uint64{0},
				Body:      "KEEP-ALIVE",
			}
			vm.checkTime = now + interval
		}
	}
}

func (vmm *vmMgr) pick(ver string) (string, int32) {
	status, ok := vmm.vers[ver]
	if !ok {
		return "", errno.HOSTVMUNAVAILABLEBYVER
	}

	minLoad := int64(99999999)
	division := ""
	for dv, es := range status {
		if es.curLoad < minLoad {
			minLoad = es.curLoad
			division = dv
		}
	}

	if division == "" {
		return "", errno.HOSTVMUNAVAILABLEBYVER
	}
	return division, errno.OK
}

func (vmm *vmMgr) dump() string {
	s := ""
	for _, vm := range vmm.vms {
		s = s + fmt.Sprintf("%s ", vm.division)
	}
	return s
}

func (vmm *vmMgr) debug(division, cmdOp, cmdParam string) (string, int32) {
	s := ""
	if cmdOp == "status" {
		go debugNotifyStatusToVM(division, cmdParam)
	} else if cmdOp == "dump" {
		s = vmm.dump()
	} else if cmdOp == "ping" {
		vm, ok := vmm.vms[division]
		if !ok {
			return s, errno.HOSTVMNOTEXIST
		}
		count, err := strconv.ParseInt(cmdParam, 10, 64)
		if err != nil {
			count = 1
		}
		if count <= 0 {
			count = 1
		}
		if count > 10 {
			count = 10
		}
		for i := int64(0); i < count; i++ {
			vm.pkt <- &ghost.GhostPacket{
				CmdId:     conn.CmdPing,
				UserData:  uint32(i),
				Timestamp: uint32(time.Now().Unix()),
				Sessions:  []uint64{0},
				Body:      "DEBUG-PING",
			}
		}
	}
	return s, errno.OK
}

func (vmm *vmMgr) forward(division string, sessionID uint64, header *conn.Header, body []byte) int32 {
	vm, ok := vmm.vms[division]
	if !ok {
		return errno.HOSTVMNOTEXIST
	}

	pkt := &ghost.GhostPacket{
		CmdId:     uint32(header.CmdID),
		UserData:  header.UserData,
		Timestamp: header.Timestamp,
		Sessions:  []uint64{sessionID},
		Body:      string(body)}
	select {
	case vm.pkt <- pkt:
		return errno.OK
	default:
		return errno.HOSTVMSENDCHANNELFULL
	}
}
