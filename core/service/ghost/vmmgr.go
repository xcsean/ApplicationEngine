package main

import (
	"sync"

	"github.com/xcsean/ApplicationEngine/core/shared/errno"
)

type vmEntity struct {
	division string
	version  string
}

type vmMgr struct {
	vms  map[string]vmEntity
	lock sync.RWMutex
}

func newVMMgr() *vmMgr {
	return &vmMgr{
		vms: make(map[string]vmEntity),
	}
}

func (vmm *vmMgr) addVM(vm *vmEntity) int32 {
	vmm.lock.Lock()
	defer vmm.lock.Unlock()

	_, ok := vmm.vms[vm.division]
	if ok {
		return errno.HOSTVMALREADYEXIST
	}
	vmm.vms[vm.division] = *vm
	return errno.OK
}

func (vmm *vmMgr) delVM(division string) int32 {
	vmm.lock.Lock()
	defer vmm.lock.Unlock()

	delete(vmm.vms, division)
	return errno.OK
}
