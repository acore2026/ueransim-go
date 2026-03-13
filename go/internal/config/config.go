package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Slice struct {
	SST any `yaml:"sst"`
	SD  any `yaml:"sd"`
}

type Session struct {
	Type  string `yaml:"type"`
	APN   string `yaml:"apn"`
	Slice Slice  `yaml:"slice"`
}

type UEConfig struct {
	SUPI          string    `yaml:"supi"`
	MCC           string    `yaml:"mcc"`
	MNC           string    `yaml:"mnc"`
	Key           string    `yaml:"key"`
	OP            string    `yaml:"op"`
	OPType        string    `yaml:"opType"`
	AMF           string    `yaml:"amf"`
	IMEI          string    `yaml:"imei"`
	IMEISV        string    `yaml:"imeiSv"`
	TUNNetmask    string    `yaml:"tunNetmask"`
	GNBSearchList []string  `yaml:"gnbSearchList"`
	Sessions      []Session `yaml:"sessions"`
}

func (c UEConfig) NodeName() string {
	if c.SUPI != "" {
		return c.SUPI
	}
	if c.IMEI != "" {
		return c.IMEI
	}
	return "ue"
}

type AMFConfig struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type GNBConfig struct {
	Name            string      `yaml:"name"`
	MCC             string      `yaml:"mcc"`
	MNC             string      `yaml:"mnc"`
	NCI             string      `yaml:"nci"`
	IDLength        int         `yaml:"idLength"`
	TAC             int         `yaml:"tac"`
	LinkIP          string      `yaml:"linkIp"`
	NGAPIP          string      `yaml:"ngapIp"`
	GTPIP           string      `yaml:"gtpIp"`
	AMFConfigs      []AMFConfig `yaml:"amfConfigs"`
	Slices          []Slice     `yaml:"slices"`
	IgnoreStreamIDs bool        `yaml:"ignoreStreamIds"`
}

func (c GNBConfig) NodeName() string {
	if c.Name != "" {
		return c.Name
	}
	return fmt.Sprintf("gnb-%s%s", c.MCC, c.MNC)
}

func LoadUE(path string) (*UEConfig, error) {
	cfg := &UEConfig{}
	if err := loadYAML(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func LoadGNB(path string) (*GNBConfig, error) {
	cfg := &GNBConfig{}
	if err := loadYAML(path, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func loadYAML(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode yaml: %w", err)
	}

	return nil
}
