package main

import (
	"encoding/xml"
	"io/ioutil"

	svc "github.com/xcsean/ApplicationEngine/core/shared/service"
)

type gconndConfig struct {
	Division     string        `xml:"division"`
	Log          svc.LogConfig `xml:"log"`
	GetcdAddr    string        `xml:"getcd_addr"`
	RefreshTime	 uint32		   `xml:"getcd_refresh"`
	SrvQueueSize int32         `xml:"server_queue_size"`
	CliQueueSize int32         `xml:"client_queue_size"`
	CliMaxConns  int32         `xml:"client_max_connections"`
	ProfilerFlag int32         `xml:"profiler_enabled"`
	ProfilerTime int32         `xml:"profiler_refresh"`
	LogTraffic   int32         `xml:"log_traffic"`
	Categories   []string      `xml:"category"`
}

var config *gconndConfig

func newConfig(fileName string) (*gconndConfig, error) {
	fileData, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var cfg gconndConfig
	err = xml.Unmarshal(fileData, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *gconndConfig) GetID() (int64, error) {
	_, _, id, err := svc.ParseDivision(cfg.Division)
	return id, err
}

func (cfg *gconndConfig) GetGetcdAddr() string {
	return cfg.GetcdAddr
}

func (cfg *gconndConfig) IsTrafficEnabled() bool {
	return cfg.LogTraffic != 0
}

func (cfg *gconndConfig) IsProfilerEnabled() bool {
	return cfg.ProfilerFlag != 0
}
