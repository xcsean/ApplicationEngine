package main

import (
	"encoding/xml"

	"github.com/xcsean/ApplicationEngine/core/shared/etc"
	svc "github.com/xcsean/ApplicationEngine/core/shared/service"
)

type vmcliConfig struct {
	Division    string   `xml:"division"`
	GetcdAddr   string   `xml:"getcd_addr"`
	RefreshTime uint32   `xml:"getcd_refresh"`
	Gconnd      string   `xml:"gconnd"`
	Ghost       string   `xml:"ghost"`
	Categories  []string `xml:"category"`
}

var config *vmcliConfig

func newConfig(fileName string) (*vmcliConfig, error) {
	fileData, err := etc.ReadFromXMLFile(fileName)
	if err != nil {
		return nil, err
	}

	var cfg vmcliConfig
	err = xml.Unmarshal(fileData, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *vmcliConfig) getID() (int64, error) {
	_, _, id, err := svc.ParseDivision(cfg.Division)
	return id, err
}

func (cfg *vmcliConfig) getGetcdAddr() string {
	return cfg.GetcdAddr
}
