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

type vmMgr struct {
	sf  *id.Snowflake
	vms map[string]*vmEntity
	out chan *innerCmd
}

func newVMMgr(sf *id.Snowflake, outChannel chan *innerCmd) *vmMgr {
	return &vmMgr{
		sf:  sf,
		vms: make(map[string]*vmEntity),
		out: outChannel,
	}
}

func (vmm *vmMgr) addVM(division, version, addr string) (uint64, int32) {
	_, ok := vmm.vms[division]
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
	vmm.vms[vm.division] = vm
	in <- newVMMCmd(innerCmdVMStart, division, version, addr, uuid)
	go vmEntityLoop(pkt, in, vmm.out)
	return uuid, errno.OK
}

func (vmm *vmMgr) delVM(division string, uuid uint64) int32 {
	vm, ok := vmm.vms[division]
	if ok && vm.uuid == uuid {
		vm.in <- newVMMCmd(innerCmdVMShouldExit, division, "", "", uuid)
		close(vm.in)
		close(vm.pkt)
		delete(vmm.vms, division)
	}
	return errno.OK
}

func (vmm *vmMgr) debug(division, cmdOp, cmdParam string) int32 {
	vm, ok := vmm.vms[division]
	if !ok {
		return errno.HOSTVMNOTEXIST
	}

	if cmdOp == "status" {
		go debugNotifyStatusToVM(division, cmdParam)
	} else if cmdOp == "stream" {
		count, err := strconv.ParseInt(cmdParam, 10, 32)
		if err != nil {
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
	return errno.OK
}
