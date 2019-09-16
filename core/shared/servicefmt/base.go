package servicefmt

import (
	"fmt"
	"strconv"
	"strings"
)

// RegistryServerConfig the server config in registry
type RegistryServerConfig struct {
	App           string
	Server        string
	Division      string
	Node          string
	UseAgent      int32
	NodeStatus    int32
	ServiceStatus int32
}

// RegistryServiceConfig the service config in registry
type RegistryServiceConfig struct {
	App         string
	Server      string
	Division    string
	Node        string
	Service     string
	ServiceIP   string
	ServicePort int32
	RPCPort     int32
	AdminPort   int32
}

// RegistryGlobalConfig the global config in registry
type RegistryGlobalConfig struct {
	Category string
	Key      string
	Value    string
}

// RegistryProtocol the protocol limitation config in registry
type RegistryProtocol struct {
	ProtoID             int
	PlayerLimitEnable   int
	PlayerLimitCount    int
	PlayerLimitDuration int
	ServerLimitEnable   int
	ServerLimitCount    int
	ServerLimitDuration int
}

// MakeLookupKey make the lookup key by three factors
func MakeLookupKey(app, server, division string) string {
	return app + "!" + server + "!" + division
}

// MakeDivision make the division string by three factors
func MakeDivision(app, server string, id int64) string {
	return fmt.Sprintf("%s.%s.%d", app, server, id)
}

// ParseDivision parse the division to return three factors
func ParseDivision(division string) (string, string, int64, error) {
	s := strings.Split(division, ".")
	if len(s) != 3 {
		return "", "", 0, fmt.Errorf("division should be a.b.c format")
	}
	id, err := strconv.ParseInt(s[2], 10, 64)
	if err != nil {
		return "", "", 0, err
	}
	return s[0], s[1], id, nil
}
