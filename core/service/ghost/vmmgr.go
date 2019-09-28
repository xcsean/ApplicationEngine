package main

import (
	"github.com/xcsean/ApplicationEngine/core/shared/errno"
)

type vmEntity struct {
	division string
	version  string
}

type vmMgr struct {
	vms map[string]vmEntity
}

func newVMMgr() *vmMgr {
	return &vmMgr{
		vms: make(map[string]vmEntity),
	}
}

func (vmm *vmMgr) addVM(vm *vmEntity) int32 {
	_, ok := vmm.vms[vm.division]
	if ok {
		return errno.HOSTVMALREADYEXIST
	}
	vmm.vms[vm.division] = *vm
	return errno.OK
}

func (vmm *vmMgr) delVM(division string) int32 {
	delete(vmm.vms, division)
	return errno.OK
}

func (vmm *vmMgr) debug(division, cmdLine string) int32 {
	_, ok := vmm.vms[division]
	if !ok {
		return errno.HOSTVMNOTEXIST
	}

	if cmdLine == "status" {
		go debugNotifyStatusToVM(division)
	}
	return errno.OK
}
