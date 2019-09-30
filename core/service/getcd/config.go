package main

import (
	"encoding/xml"

	"github.com/xcsean/ApplicationEngine/core/shared/etc"
	svc "github.com/xcsean/ApplicationEngine/core/shared/service"
)

type getcdConfig struct {
	Division string              `xml:"division"`
	Log      svc.LogConfig       `xml:"log"`
	RPC      svc.Endpoint        `xml:"rpc"`
	Admin    svc.Endpoint        `xml:"admin"`
	Mysql    svc.MysqlConnection `xml:"mysql"`
	Refresh  uint32              `xml:"refresh_interval"`
}

var config *getcdConfig

func newConfig(fileName string) (*getcdConfig, error) {
	fileData, err := etc.ReadFromXMLFile(fileName)
	if err != nil {
		return nil, err
	}

	var cfg getcdConfig
	err = xml.Unmarshal(fileData, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *getcdConfig) getID() (int64, error) {
	_, _, id, err := svc.ParseDivision(cfg.Division)
	return id, err
}