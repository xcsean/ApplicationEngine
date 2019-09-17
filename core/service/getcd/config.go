package main

import (
	svc "github.com/xcsean/ApplicationEngine/core/shared/service"
)

type getcdConfig struct {
	Division string `xml:"division"`
	Log      svc.LogConfig `xml:"log"`
	RPC      svc.Endpoint `xml:"rpc"`
	Admin    svc.Endpoint `xml:"admin"`
	Mysql    svc.MysqlConnection `xml:"mysql"`
	Refresh  uint32 `xml:"refresh_interval"`
}

var config getcdConfig