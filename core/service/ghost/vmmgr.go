package main

import (
	"strconv"
	"time"

	"github.com/xcsean/ApplicationEngine/core/protocol/ghost"
	"github.com/xcsean/ApplicationEngine/core/shared/conn"
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/id"
)

type vmEntity struct {
	uuid     uint64
	division string
	version  string
	addr     string
	pkt      chan *ghost.GhostPacket
	in       chan *innerCmd
}

type vmEntityStatus struct {
	curLoad uint64
	maxLoad uint64
}

type vmStatus map[string]vmEntityStatus

type vmMgr struct {
	sf    *id.Snowflake
	out   chan *innerCmd
	vms   map[string]*vmEntity
	vers  map[string]vmStatus // fast-index
	addrs map[string]bool     // fast-index
}

func newVMMgr(sf *id.Snowflake, outChannel chan *innerCmd) *vmMgr {
	return &vmMgr{
		sf:    sf,
		out:   outChannel,
		vms:   make(map[string]*vmEntity),
		vers:  make(map[string]vmStatus),
		addrs: make(map[string]bool),
	}
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

	uuid, _ := vmm.sf.NextID()
	pkt := make(chan *ghost.GhostPacket, 1000)
	in := make(chan *innerCmd, 10)
	vm := &vmEntity{
		uuid:     uuid,
		division: division,
		version:  version,
		addr:     addr,
		pkt:      pkt,
		in:       in,
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

	in <- newVMMCmd(innerCmdVMStart, division, version, addr, uuid)
	go vmEntityLoop(pkt, in, vmm.out)
	return uuid, errno.OK
}

func (vmm *vmMgr) delVM(division string, uuid uint64) int32 {
	vm, ok := vmm.vms[division]
	if ok && vm.uuid == uuid {
		vm.in <- newVMMCmd(innerCmdVMShouldExit, division, vm.version, vm.addr, uuid)
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

func (vmm *vmMgr) dump(division string, uuid uint64) string {
	vm, ok := vmm.vms[division]
	if !ok {
		return ""
	}
	if vm.uuid != uuid {
		return ""
	}

	return "dump"
}

func (vmm *vmMgr) debug(division, cmdOp, cmdParam string) (string, int32) {
	s := ""
	vm, ok := vmm.vms[division]
	if !ok {
		return s, errno.HOSTVMNOTEXIST
	}

	if cmdOp == "status" {
		go debugNotifyStatusToVM(division, cmdParam)
	} else if cmdOp == "dump" {
		uuid, _ := strconv.ParseInt(cmdParam, 10, 64)
		s = vmm.dump(division, uint64(uuid))
	} else if cmdOp == "ping" {
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
		sessions := make([]uint64, 1)
		sessions[0] = 12345678
		for i := int64(0); i < count; i++ {
			vm.pkt <- &ghost.GhostPacket{
				CmdId:     conn.CmdPing,
				UserData:  uint32(i),
				Timestamp: uint32(time.Now().Unix()),
				Sessions:  sessions,
				Body:      "ABCDEFG",
			}
		}
	}
	return s, errno.OK
}
