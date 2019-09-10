package main

import (
	"container/list"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/proto"
	"github.com/xcsean/ApplicationEngine/core/protocol/getcd"
	"github.com/xcsean/ApplicationEngine/core/shared/dbg"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
	sf "github.com/xcsean/ApplicationEngine/core/shared/servicefmt"
)

var (
	// registry server & service
	serverLock   sync.RWMutex
	serverMap    map[string]*sf.RegistryServerConfig
	serviceLock  sync.RWMutex
	serviceMap   map[string]*list.List

	// registry global config & proto limitation
	globalConfigMap map[string][]*sf.RegistryGlobalConfig
	protoLimitArr   []sf.RegistryProtocol
)

func getServerMap() map[string]*sf.RegistryServerConfig {
	serverLock.RLock()
	defer serverLock.RUnlock()
	return serverMap
}

func getServiceMap() map[string]*list.List {
	serviceLock.RLock()
	defer serviceLock.RUnlock()
	return serviceMap
}

func getGlobalConfigMap() map[string][]*sf.RegistryGlobalConfig {
	return globalConfigMap
}

func loadServerFromRegistry(db *sql.DB) (map[string]*sf.RegistryServerConfig, error) {
	// load t_server table
	rows, err := db.Query("SELECT app, server, division, node, use_agent, node_status, service_status FROM t_server")
	if err != nil {
		log.Error("select t_server failed: %s", err.Error())
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]*sf.RegistryServerConfig, 100)
	for rows.Next() {
		var app string
		var server string
		var division string
		var node string
		var useAgent int32
		var nodeStatus int32
		var serviceStatus int32
		err = rows.Scan(&app, &server, &division, &node, &useAgent, &nodeStatus, &serviceStatus)
		if err != nil {
			log.Error("scan failed: %s", err.Error())
			return nil, err
		}

		log.Debug("server: app=%s, server=%s, division=%s, node=%s, use_agent=%v, node_status=%v, service_status=%v",
			app, server, division, node, useAgent, nodeStatus, serviceStatus)
		c := &sf.RegistryServerConfig{
			App:           app,
			Server:        server,
			Division:      division,
			Node:          node,
			UseAgent:      useAgent,
			NodeStatus:    nodeStatus,
			ServiceStatus: serviceStatus,
		}

		key := sf.MakeLookupKey(app, server, division)
		m[key] = c
	}
	return m, nil
}

func loadServiceFromRegistry(db *sql.DB) (map[string]*list.List, error) {
	// load t_service table
	rows, err := db.Query("SELECT app, server, division, node, service, service_ip, service_port, admin_port, rpc_port FROM t_service")
	if err != nil {
		log.Error("select t_service failed: %s", err.Error())
		return nil, err
	}
	defer rows.Close()

	m := make(map[string]*list.List, 100)
	for rows.Next() {
		var app string
		var server string
		var division string
		var node string
		var service string
		var serviceIP string
		var servicePort int32
		var adminPort int32
		var rpcPort int32
		err = rows.Scan(&app, &server, &division, &node, &service, &serviceIP, &servicePort, &adminPort, &rpcPort)
		if err != nil {
			log.Error("scan failed: %s", err.Error())
			return nil, err
		}

		log.Debug("service: app=%s, server=%s, division=%s, node=%s, service=%s, serviceIp=%s, servicePort=%d, adminPort=%d, rpcPort=%d",
			app, server, division, node, service, serviceIP, servicePort, adminPort, rpcPort)
		c := sf.RegistryServiceConfig{
			App:         app,
			Server:      server,
			Division:    division,
			Node:        node,
			Service:     service,
			ServiceIP:   serviceIP,
			ServicePort: servicePort,
			AdminPort:   adminPort,
			RPCPort:     rpcPort,
		}

		key := sf.MakeLookupKey(app, server, division)
		l, ok := m[key]
		if ok {
			l.PushBack(c)
		} else {
			l = list.New()
			m[key] = l
			l.PushBack(c)
		}
	}
	return m, nil
}

func loadGlobalConfigFromRegistry(db *sql.DB) (map[string][]*sf.RegistryGlobalConfig, error) {
	// load t_global_config table
	rows, err := db.Query("SELECT t_category, t_key, t_value FROM t_global_config")
	if err != nil {
		log.Error("SELECT t_category, t_key, t_value FROM t_global_config failed:%s", err.Error())
		return nil, err
	}
	defer rows.Close()

	s := make(map[string][]*sf.RegistryGlobalConfig)
	for rows.Next() {
		var category, key, value string
		err = rows.Scan(&category, &key, &value)
		if err != nil {
			log.Error("scan failed, %s", err.Error())
			return nil, err
		}
		s[category] = append(s[category], &sf.RegistryGlobalConfig{
			Category: category,
			Key:      key,
			Value:    value,
		})
	}
	return s, nil
}

func loadProtocolLimitFromRegistry(db *sql.DB) ([]sf.RegistryProtocol, error) {
	// load t_protocol table
	rows, err := db.Query("SELECT proto_id, player_limit_enable, player_limit_count, player_limit_duration, server_limit_enable, server_limit_count, server_limit_duration FROM t_protocol")
	if err != nil {
		log.Error("SELECT * from protocol failed:%v", err)
		return nil, err
	}
	defer rows.Close()

	entries := make([]sf.RegistryProtocol, 0)
	for rows.Next() {
		var protoID int
		var playerLimitEnable, playerLimitCount, playerLimitDuration int
		var serverLimitEnable, serverLimitCount, serverLimitDuration int
		err = rows.Scan(&protoID, &playerLimitEnable, &playerLimitCount, &playerLimitDuration,
			&serverLimitEnable, &serverLimitCount, &serverLimitDuration)
		if err != nil {
			log.Error("scan protocol failed:%v", err)
			return nil, err
		}
		entries = append(entries, sf.RegistryProtocol{
			ProtoID:             protoID,
			PlayerLimitEnable:   playerLimitEnable,
			PlayerLimitCount:    playerLimitCount,
			PlayerLimitDuration: playerLimitDuration,
			ServerLimitEnable:   serverLimitEnable,
			ServerLimitCount:    serverLimitCount,
			ServerLimitDuration: serverLimitDuration,
		})
	}
	return entries, nil
}

func loadRegistryFromMysql(s string) error {
	defer dbg.Stacktrace()

	db, err := sql.Open("mysql", s)
	if err = db.Ping(); err != nil {
		return err
	}
	defer db.Close()

	server, err := loadServerFromRegistry(db)
	if err != nil {
		return err
	}

	service, err := loadServiceFromRegistry(db)
	if err != nil {
		return err
	}

	globalConfig, err := loadGlobalConfigFromRegistry(db)
	if err != nil {
		return err
	}

	protoLimits, err := loadProtocolLimitFromRegistry(db)
	if err != nil {
		return err
	}

	// save!
	serverLock.Lock()
	serverMap = server
	serverLock.Unlock()

	serviceLock.Lock()
	serviceMap = service
	serviceLock.Unlock()

	globalConfigMap = globalConfig
	protoLimitArr = protoLimits

	// dump!
	dumpRegistryServerConfig()
	dumpRegistryServiceConfig()
	dumpRegistryGlobalConfig()
	dumpRegistryProtoLimit()

	log.Info("load registry from mysql ok")
	return nil
}

func loadRegistryPeriodically(s string, t uint32) {
	tick := time.NewTicker(time.Duration(t) * time.Second)
	for {
		select {
		case <-tick.C:
			log.Debug("it's time to load registry from mysql")
			loadRegistryFromMysql(s)
		}
	}
}

func dumpRegistryServerConfig() {
	log.Info("registry server ===>")
	for k, v := range serverMap {
		log.Info("key=%s", k)
		c := v
		log.Info("app=%s, server=%s, division=%s, node=%s, use_agent=%v, node_status=%v, service_status=%v",
			c.App, c.Server, c.Division, c.Node, c.UseAgent, c.NodeStatus, c.ServiceStatus)
	}
	log.Info("<===")
}

func dumpRegistryServiceConfig() {
	log.Info("registry service ===>")
	for k, v := range serviceMap {
		log.Info("key=%s", k)
		for e := v.Front(); e != nil; e = e.Next() {
			c := e.Value.(sf.RegistryServiceConfig)
			log.Info("service: app=%s, server=%s, division=%s, node=%s, service=%s, serviceIp=%s, servicePort=%d, adminPort=%d, rpcPort=%d",
				c.App, c.Server, c.Division, c.Node, c.Service, c.ServiceIP, c.ServicePort, c.AdminPort, c.RPCPort)
		}
	}
	log.Info("<===")
}

func dumpRegistryGlobalConfig() {
	log.Info("registry global config ===>")
	for category, array := range globalConfigMap {
		str := ""
		for k, v := range array {
			str += fmt.Sprintf("%v:%s", k, v)
		}
		log.Info("category:%s %s", category, str)
	}
	log.Info("<===")
}

func dumpRegistryProtoLimit() {
	log.Info("registry proto limit ===>")
	for _, v := range protoLimitArr {
		log.Info("protoID:%d player(%d %d %d) server(%d %d %d)",
			v.ProtoID, v.PlayerLimitEnable, v.PlayerLimitCount, v.PlayerLimitDuration,
			v.ServerLimitEnable, v.ServerLimitCount, v.ServerLimitDuration)
	}
	log.Info("<===")
}

func getServerCount(m map[string]*sf.RegistryServerConfig) uint32 {
	return uint32(len(m))
}

func getServiceCount(m map[string]*list.List) uint32 {
	var count uint32
	for _, v := range m {
		count += uint32(v.Len())
	}
	return count
}

func getRegistryPbFormat() ([]byte, error) {
	rsp := &getcd.QueryRegistryRsp{Result: 0}

	server := getServerMap()
	serverLen := getServerCount(server)

	service := getServiceMap()
	serviceLen := getServiceCount(service)

	serverArray := make([]*getcd.RegistryServer, serverLen)
	i := 0
	for _, v := range server {
		c := v
		s := &getcd.RegistryServer{
			App:           c.App,
			Server:        c.Server,
			Division:      c.Division,
			Node:          c.Node,
			UseAgent:      c.UseAgent,
			NodeStatus:    c.NodeStatus,
			ServiceStatus: c.ServiceStatus,
		}
		serverArray[i] = s
		i ++
	}
	rsp.Servers = serverArray

	serviceArray := make([]*getcd.RegistryService, serviceLen)
	i = 0
	for _, v := range service {
		for e := v.Front(); e != nil; e = e.Next() {
			c := e.Value.(sf.RegistryServiceConfig)
			s := &getcd.RegistryService{
				App:         c.App,
				Server:      c.Server,
				Division:    c.Division,
				Node:        c.Node,
				Service:     c.Service,
				ServiceIp:   c.ServiceIP,
				ServicePort: c.ServicePort,
				RpcPort:     c.RPCPort,
				AdminPort:   c.AdminPort,
			}
			serviceArray[i] = s
			i ++
		}
	}
	rsp.Services = serviceArray

	return proto.Marshal(rsp)
}

func getRegistryServerString() string {
	m := getServerMap()
	s := ""
	for k, v := range m {
		s += k
		s += "\n"
		c := v
		s1 := fmt.Sprintf("app=%s, server=%s, division=%s, node=%s, use_agent=%v, node_status=%v, service_status=%v",
			c.App, c.Server, c.Division, c.Node, c.UseAgent, c.NodeStatus, c.ServiceStatus)
		s += s1
		s += "\n"
	}
	return s
}

func getRegistryServiceString() string {
	m := getServiceMap()
	s := ""
	for k, v := range m {
		s += k
		s += "\n"
		for e := v.Front(); e != nil; e = e.Next() {
			c := e.Value.(sf.RegistryServiceConfig)
			s1 := fmt.Sprintf("app=%s, server=%s, division=%s, node=%s, service=%s, serviceip=%s, serviceport=%d, adminport=%d, rpcport=%d",
				c.App, c.Server, c.Division, c.Node, c.Service, c.ServiceIP, c.ServicePort, c.AdminPort, c.RPCPort)
			s += s1
			s += "\n"
		}
	}
	return s
}
