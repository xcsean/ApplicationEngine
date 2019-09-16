package etc

import (
	"container/list"
	"fmt"
	"sync"

	"github.com/xcsean/ApplicationEngine/core/shared/log"
	sf "github.com/xcsean/ApplicationEngine/core/shared/servicefmt"
)

// serviceGroup contains services which belong to the same type
//  as the field 'service' of table 't_service' in 'registry.sql'
type serviceGroup struct {
	array  []*sf.RegistryServiceConfig
	cursor uint32
}

// serviceAll contains service-groups which belong to various type
//  type is the map key generated by sf.MakeLookupKey
type serviceAll struct {
	curr map[string]*serviceGroup
	lock sync.RWMutex
}

func newServiceAll() *serviceAll {
	return &serviceAll{
		curr: make(map[string]*serviceGroup),
	}
}

func (sa *serviceAll) fillNewer(newer map[string]*serviceGroup, key string,
	serviceList *list.List, serverMap map[string]*sf.RegistryServerConfig) {
	sa.lock.RLock()
	defer sa.lock.RUnlock()

	for e := serviceList.Front(); e != nil; e = e.Next() {
		c := e.Value.(sf.RegistryServiceConfig)
		if c.Service == "" {
			// if service name is empty, just skip
			continue
		}
		s, ok := serverMap[key]
		if !ok {
			// if server is not exist, just skip
			continue
		}
		if s.NodeStatus != 0 || s.ServiceStatus != 0 {
			// if server status & service status are not 0, just skip
			continue
		}
		grp, ok := newer[c.Service]
		if ok {
			grp.array = append(grp.array, &c)
		} else {
			cursor := uint32(0)
			currgrp, ok := sa.curr[c.Service]
			if ok {
				// inherit the cursor
				cursor = currgrp.cursor
			}
			grp = &serviceGroup{
				array:  nil,
				cursor: cursor,
			}
			grp.array = append(grp.array, &c)
			newer[c.Service] = grp
		}
	}
}

func (sa *serviceAll) replace(serverMap map[string]*sf.RegistryServerConfig,
	serviceMap map[string]*list.List) {
	newer := make(map[string]*serviceGroup)
	for key, serviceList := range serviceMap {
		sa.fillNewer(newer, key, serviceList, serverMap)
	}

	// replace the current
	sa.lock.Lock()
	defer sa.lock.Unlock()
	sa.curr = newer
}

func (sa *serviceAll) round(service string) (string, int32, int32, error) {
	sa.lock.Lock()
	defer sa.lock.Unlock()

	grp, ok := sa.curr[service]
	if !ok {
		return "", 0, 0, fmt.Errorf("service=%s not exist", service)
	}

	index := grp.cursor
	if index >= uint32(len(grp.array)) {
		index = 0
	}

	// do a round-robin select
	s := grp.array[index]
	index++
	grp.cursor = index
	return s.ServiceIP, s.ServicePort, s.RPCPort, nil
}

func (sa *serviceAll) dump() {
	sa.lock.RLock()
	defer sa.lock.RUnlock()

	for k, v := range sa.curr {
		log.Info("key=%s", k)
		log.Info("cursor=%v", v.cursor)
		for i := 0; i < len(v.array); i++ {
			log.Info("[%d]=%v", i, v.array[i])
		}
	}
}
