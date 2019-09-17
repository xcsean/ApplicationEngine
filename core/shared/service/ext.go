package service

import (
	"fmt"
)

// Endpoint end-point define
type Endpoint struct {
	IP   string `xml:"ip"`
	Port string `xml:"port"`
}

// Addr get the address in xxxx:yyy format
func (e *Endpoint) Addr() string {
	return fmt.Sprintf("%s:%s", e.IP, e.Port)
}

// MysqlConnection mysql config
type MysqlConnection struct {
	Endpoint `xml:"endpoint"`
	Database string `xml:"database"`
	Username string `xml:"username"`
	Password string `xml:"password"`
}

// RedisConnection redis config
type RedisConnection struct {
	Endpoint `xml:"endpoint"`
	AuthPass string `xml:"auth_pass"`
}

// LogConfig the config for logger
type LogConfig struct {
	Dir      string `xml:"dir"`
	FileName string `xml:"filename"`
	LogLevel string `xml:"log_level"`
}
