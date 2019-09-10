package main

import (
	sf "github.com/xcsean/ApplicationEngine/core/shared/servicefmt"
)

type getcdConfig struct {
	Division string `xml:"division"`
	Log      sf.LogConfig `xml:"log"`
	RPC      sf.Endpoint `xml:"rpc"`
	Admin    sf.Endpoint `xml:"admin"`
	Mysql    sf.MysqlConnection `xml:"mysql"`
	Refresh  uint32 `xml:"refresh_interval"`
}

var config getcdConfig