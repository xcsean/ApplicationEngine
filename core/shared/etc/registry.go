package etc

import (
	"container/list"
	"fmt"
	"reflect"
	"time"
	"sync"

	"github.com/xcsean/ApplicationEngine/core/protocol/getcd"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	rc "github.com/xcsean/ApplicationEngine/core/shared/errno"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	sf "github.com/xcsean/ApplicationEngine/core/shared/servicefmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	// server data
	serverMap  map[string]*sf.RegistryServerConfig
	serverLock sync.RWMutex

	// service data
	serviceMap  map[string]*list.List
	serviceLock sync.RWMutex

	// the service-all object for RPC-call
	ss *serviceAll

	// global config
	gc *globalConfig

	// others...
	getcdAddr string
	registryLastPrintTime time.Time
)

func init() {
	serverMap = make(map[string]*sf.RegistryServerConfig)
	serviceMap = make(map[string]*list.List)
	ss = newServiceAll()
	gc = newGlobalConfig()
}

func getServerMap() map[string]*sf.RegistryServerConfig {
	serverLock.RLock()
	defer serverLock.RUnlock()
	return serverMap
}

func setServerMap(newer map[string]*sf.RegistryServerConfig) {
	serverLock.Lock()
	defer serverLock.Unlock()
	serverMap = newer
}

func getServiceMap() map[string]*list.List {
	serviceLock.RLock()
	defer serviceLock.RUnlock()
	return serviceMap
}

func setServiceMap(newer map[string]*list.List) {
	serviceLock.Lock()
	defer serviceLock.Unlock()
	serviceMap = newer
}

func saveService(rsp *getcd.QueryRegistryRsp) {
	if rsp.Result != rc.OK {
		return
	}

	// make server data
	server := make(map[string]*sf.RegistryServerConfig)
	for _, e := range rsp.Servers {
		s := &sf.RegistryServerConfig{
			App:           e.App,
			Server:        e.Server,
			Division:      e.Division,
			Node:          e.Node,
			UseAgent:      e.UseAgent,
			NodeStatus:    e.NodeStatus,
			ServiceStatus: e.ServiceStatus,
		}
		key := sf.MakeLookupKey(s.App, s.Server, s.Division)
		server[key] = s
	}
	oldServer := getServerMap()
	setServerMap(server)

	// make service data
	service := make(map[string]*list.List)
	for _, e := range rsp.Services {
		s := sf.RegistryServiceConfig{
			App:         e.App,
			Server:      e.Server,
			Division:    e.Division,
			Node:        e.Node,
			Service:     e.Service,
			ServiceIP:   e.ServiceIp,
			ServicePort: e.ServicePort,
			AdminPort:   e.AdminPort,
			RPCPort:     e.RpcPort,
		}
		key := sf.MakeLookupKey(s.App, s.Server, s.Division)
		l, ok := service[key]
		if ok {
			l.PushBack(s)
		} else {
			l = list.New()
			service[key] = l
			l.PushBack(s)
		}
	}
	oldService := getServiceMap()
	setServiceMap(service)

	// make service-all object
	ss.replace(server, service)

	// dump the server & service
	dumpService(oldServer, server, oldService, service)
}

func dumpService(oldServer, newServer map[string]*sf.RegistryServerConfig, oldService, newService map[string]*list.List) {
	if reflect.DeepEqual(oldServer, newServer) &&
		reflect.DeepEqual(oldService, newService) &&
		registryLastPrintTime.Hour() == time.Now().Hour() {
		return
	}

	registryLastPrintTime = time.Now()
	for _, v := range newServer {
		log.Info("server: %+v", v)
	}
	for _, v := range newService {
		for e := v.Front(); e != nil; e = e.Next() {
			log.Info("service: %+v", e.Value)
		}
	}
	ss.dump()
}

func queryServicePeriodically(t uint32) {
	tick := time.NewTicker(time.Duration(t) * time.Second)
	for {
		select {
		case <-tick.C:
			if err := QueryService(); err != nil {
				log.Error("query service failed: %s", err.Error())
			}
		}
	}
}

func saveGlobalConfig(rsp *getcd.QueryGlobalConfigRsp) {
	if rsp.Result != rc.OK {
		return
	}

	newer := make(map[string]*categoryEntry)
	for _, e := range rsp.Entries {
		cat := e.Category
		newerEntry, ok := newer[cat]
		if !ok {
			newerEntry = &categoryEntry{
				category: cat,
				kv: make(map[string]string),
			}
			newer[cat] = newerEntry
		}
		for k, v := range e.Kv {
			newerEntry.kv[k] = v
		}
	}

	gc.replace(newer)
	gc.dump()
}

// SetGetcdAddr set the address of getcd service
func SetGetcdAddr(addr string) {
	getcdAddr = addr
	log.Info("getcd service set to %s", getcdAddr)
}

// StartQueryServiceLoop start a timer for query service from getcd
func StartQueryServiceLoop(t uint32) error {
	log.Debug("begin a query service loop from registry, duration=%d", t)
	go queryServicePeriodically(t)
	return nil
}

// QueryService query registry service from getcd
func QueryService() error {
	defer dbg.Stacktrace()

	log.Debug("query registry service from %s begin...", getcdAddr)
	conn, err := grpc.Dial(getcdAddr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	c := getcd.NewGetcdServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rsp, err := c.QueryRegistry(ctx, &getcd.QueryRegistryReq{})
	if err != nil {
		return err
	}
	
	log.Debug("query registry service result: %d", rsp.Result)
	saveService(rsp)
	return nil
}

// QueryEndpoint query an endpoint by app & server & division format
func QueryEndpoint(app string, server string, division string) (string, int32, int32, int32, error) {
	key := sf.MakeLookupKey(app, server, division)
	s1 := getServerMap()
	s2 := getServiceMap()

	l1, ok := s1[key]
	if !ok {
		return "", 0, 0, 0, fmt.Errorf("server %s not found", key)
	}
	l2, ok := s2[key]
	if !ok {
		return "", 0, 0, 0, fmt.Errorf("service %s not found", key)
	}

	for e := l2.Front(); e != nil; e = e.Next() {
		c := e.Value.(sf.RegistryServiceConfig)
		if c.Node == l1.Node {
			return c.ServiceIP, c.ServicePort, c.AdminPort, c.RPCPort, nil
		}
	}
	return "", 0, 0, 0, fmt.Errorf("node=%s not found in registry", l1.Node)
}

// PickEndpoint pick an endpoint by division format
func PickEndpoint(division string) (string, int32, int32, int32, error) {
	app, server, _, err := sf.ParseDivision(division)
	if err != nil {
		return "", 0, 0, 0, err
	}
	return QueryEndpoint(app, server, division)
}

// SelectEndpoint select an endpoint by service format
//  "app.server.service" format, will use round-robin algorithm to find a node
//  "app.server.1" format, will find the node directly, round-robin skipped
func SelectEndpoint(service string) (string, int32, int32, error) {
	// try to find the node by "app.server.service"
	ip, servicePort, rpcPort, err := ss.round(service)
	if err == nil {
		return ip, servicePort, rpcPort, err
	}

	// and try to find the node by "app.server.1"
	ip, servicePort, _, rpcPort, err = PickEndpoint(service)
	if err == nil {
		return ip, servicePort, rpcPort, err
	}

	// not found
	return "", 0, 0, fmt.Errorf("service=%s not found", service)
}

// IsServiceUseAgent tell whether a node use agent or not
func IsServiceUseAgent(division string) bool {
	app, server, _, err := sf.ParseDivision(division)
	if err != nil {
		return false
	}
	key := sf.MakeLookupKey(app, server, division)
	s := getServerMap()
	status, ok := s[key]
	if ok {
		return (status.UseAgent != 0)
	}
	return false
}

// QueryGlobalConfig query a config from global config
func QueryGlobalConfig(category, key string) (string, bool) {
	return gc.getValue(category, key)
}

// InGlobalConfig tell whether the key in category and has a sub-string like 'pattern'
func InGlobalConfig(category, key, pattern string) bool {
	return gc.contains(category, key, pattern)
}
