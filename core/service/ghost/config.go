package main

import (
	"encoding/xml"

	"github.com/xcsean/ApplicationEngine/core/shared/etc"
	svc "github.com/xcsean/ApplicationEngine/core/shared/service"
)

type ghostConfig struct {
	Division    string              `xml:"division"`
	Log         svc.LogConfig       `xml:"log"`
	Mysql       svc.MysqlConnection `xml:"mysql"`
	GetcdAddr   string              `xml:"getcd_addr"`
	RefreshTime uint32              `xml:"getcd_refresh"`
	Gconnd      string              `xml:"gconnd"`
	Categories  []string            `xml:"category"`
}

var config *ghostConfig

func newConfig(fileName string) (*ghostConfig, error) {
	fileData, err := etc.ReadFromXMLFile(fileName)
	if err != nil {
		return nil, err
	}

	var cfg ghostConfig
	err = xml.Unmarshal(fileData, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *ghostConfig) getID() (int64, error) {
	_, _, id, err := svc.ParseDivision(cfg.Division)
	return id, err
}
