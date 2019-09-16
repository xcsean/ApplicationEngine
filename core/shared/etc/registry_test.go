package etc

import (
	"testing"

	"github.com/xcsean/ApplicationEngine/core/protocol/getcd"
	"github.com/xcsean/ApplicationEngine/core/shared/log"
)

func TestQueryRegistry(t *testing.T) {
	log.SetupMainLogger("./", "etc.log", "debug")
	log.Debug("query registry...")

	// fill the response
	rsp := &getcd.QueryRegistryRsp{
		Result: 0,
		Servers: make([]*getcd.RegistryServer, 0),
		Services: make([]*getcd.RegistryService, 0),
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
	saveRegistry(rsp)

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
	use1 := IsUseAgent("app.globby.1")
	if use1 != false {
		t.Errorf("IsUseAgent failed")
		return
	}
	use2 := IsUseAgent("app.globby.2")
	if use2 != true {
		t.Errorf("IsUseAgent failed")
		return
	}
	t.Logf("use agent: %v %v", use1, use2)
}