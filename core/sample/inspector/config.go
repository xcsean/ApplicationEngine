package main

import (
	"encoding/xml"

	"github.com/xcsean/ApplicationEngine/core/shared/etc"
)

type inspectorConfig struct {
	Cmd      string `xml:"cmd"`
	Prefix   string `xml:"prefix"`
	Postfix  string `xml:"postfix"`
	Interval int32  `xml:"interval"`
	Dir      string `xml:"dir"`
}

var config *inspectorConfig

func newConfig(fileName string) (*inspectorConfig, error) {
	fileData, err := etc.ReadFromXMLFile(fileName)
	if err != nil {
		return nil, err
	}

	var cfg inspectorConfig
	err = xml.Unmarshal(fileData, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (cfg *inspectorConfig) getCmd() string {
	return cfg.Cmd
}

func (cfg *inspectorConfig) getPrefix() string {
	return cfg.Prefix
}

func (cfg *inspectorConfig) getPostfix() string {
	return cfg.Postfix
}

func (cfg *inspectorConfig) getInterval() int32 {
	if cfg.Interval <= 0 {
		return 10
	}
	return cfg.Interval
}

func (cfg *inspectorConfig) getDir() string {
	return cfg.Dir
}
