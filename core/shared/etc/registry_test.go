package etc

import (
	"testing"

	"github.com/xcsean/ApplicationEngine/core/protocol/getcd"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

func testService(t *testing.T) {
	// fill the response
	rsp := &getcd.QueryRegistryRsp{
		Result: 0,
		Servers: nil,
		Services: nil,
	}
	srv := &getcd.RegistryServer{
		App: "app",
		Server: "globby",
		Division: "app.globby.1",
		Node: "192.168.1.22",
		UseAgent: 0,
		NodeStatus: 0,
		ServiceStatus: 0,
	}
	rsp.Servers = append(rsp.Servers, srv)
	svc := &getcd.RegistryService{
		App: "app",
		Server: "globby",
		Division: "app.globby.1",
		Node: "192.168.1.22",
		Service: "app.globby.playersave",
		ServiceIp: "192.168.1.22",
		ServicePort: 17001,
		RpcPort: 17002,
		AdminPort: 17003,
	}
	rsp.Services = append(rsp.Services, svc)

	srv = &getcd.RegistryServer{
		App: "app",
		Server: "globby",
		Division: "app.globby.2",
		Node: "192.168.1.23",
		UseAgent: 1,
		NodeStatus: 0,
		ServiceStatus: 0,
	}
	rsp.Servers = append(rsp.Servers, srv)
	svc = &getcd.RegistryService{
		App: "app",
		Server: "globby",
		Division: "app.globby.2",
		Node: "192.168.1.23",
		Service: "app.globby.playersave",
		ServiceIp: "192.168.1.23",
		ServicePort: 17001,
		RpcPort: 17002,
		AdminPort: 17003,
	}
	rsp.Services = append(rsp.Services, svc)

	// save to etc
	saveService(rsp)

	// test pick
	division := "app.globby.1"
	ip, _, _, _, err := PickEndpoint(division)
	if err != nil {
		t.Errorf("PickEndpoint failed: %s", err.Error())
		return
	}
	if ip != "192.168.1.22" {
		t.Errorf("PickEndpoint failed")
		return
	}
	t.Logf("%s ip=%s", division, ip)

	// test round
	first, _, _, err := SelectEndpoint("app.globby.playersave")
	second, _, _, err := SelectEndpoint("app.globby.playersave")
	third, _, _, err := SelectEndpoint("app.globby.playersave")
	if first == second {
		t.Errorf("SelectEndpoint failed")
		return
	}
	if first != third {
		t.Errorf("SelectEndpoint failed")
	}
	t.Logf("first=%s, second=%s, third=%s", first, second, third)

	// test use agent
	use1 := IsServiceUseAgent("app.globby.1")
	if use1 != false {
		t.Errorf("IsServiceUseAgent failed")
		return
	}
	use2 := IsServiceUseAgent("app.globby.2")
	if use2 != true {
		t.Errorf("IsServiceUseAgent failed")
		return
	}
	t.Logf("use agent: %v %v", use1, use2)
}

func testGlobalConfig(t *testing.T) {
	rsp := &getcd.QueryGlobalConfigRsp{
		Result: 0,
		Entries: nil,
	}

	// add permission
	cat := "permission"
	entry := &getcd.CategoryEntry{
		Category: cat,
		Kv: make(map[string]string),
	}
	entry.Kv["ipAdminList"] = "127.0.0.1,192.168.1.10"
	entry.Kv["ipWhiteList"] = "127.0.0.1"
	rsp.Entries = append(rsp.Entries, entry)

	// add global
	cat = "global"
	entry = &getcd.CategoryEntry{
		Category: cat,
		Kv: make(map[string]string),
	}
	entry.Kv["wechatLogin"] = "1"
	entry.Kv["qqLogin"] = "0"
	entry.Kv["devLogin"] = "0"
	rsp.Entries = append(rsp.Entries, entry)

	// add permission again
	cat = "permission"
	entry = &getcd.CategoryEntry{
		Category: cat,
		Kv: make(map[string]string),
	}
	entry.Kv["ipBlackList"] = "192.168.1.11"
	rsp.Entries = append(rsp.Entries, entry)
	
	// save to etc
	saveGlobalConfig(rsp)

	// test in global config
	ip := "192.168.1.10"
	exist := InGlobalConfig("permission", "ipAdminList", ip)
	if !exist {
		t.Errorf("InGlobalConfig failed")
		return
	}
	t.Logf("ip %s is in global config", ip)

	ip = "192.168.1.11"
	exist = InGlobalConfig("permission", "ipAdminList", ip)
	if exist {
		t.Errorf("InGlobalConfig failed")
		return
	}
	t.Logf("ip %s isn't in global config", ip)

	ip2, exist := QueryGlobalConfig("permission", "ipBlackList")
	if !exist {
		t.Errorf("QueryGlobalConfig failed")
		return
	}
	t.Logf("ip black list is %s", ip2)

	if ip != ip2 {
		t.Errorf("QueryGlobalConfig failed")
		return
	}
	t.Logf("ip %s equal to ip2 %s", ip, ip2)
}

func testNetwork(t *testing.T) {
	division := "app.globby.1"
	ok, err := CanProvideService(division)
	if err != nil {
		t.Logf("CanProvideService failed: %s", err.Error())
	} else {
		if ok {
			t.Logf("self can provide service %s", division)
		} else {
			t.Logf("self can't provide service %s", division)
		}
	}

	loopback := "127.0.0.1"
	ok = HaveAddress(loopback)
	if ok {
		t.Logf("self have loopback addr: %s", loopback)
	}
}

func TestRegistry(t *testing.T) {
	log.SetupMainLogger("./", "etc.log", "debug")

	// test service
	testService(t)

	// test global config
	testGlobalConfig(t)

	// test network
	testNetwork(t)
}