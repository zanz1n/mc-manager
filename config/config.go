package config

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `json:"server" yaml:"server"`
	Docker DockerConfig `json:"docker" yaml:"docker"`
	Data   DataConfig   `json:"data" yaml:"data"`
}

type ServerConfig struct {
	IP               net.IP `json:"ip" yaml:"ip" validate:"required"`
	Port             uint16 `json:"port" yaml:"port" validate:"required"`
	Password         string `json:"password" yaml:"password"`
	EnableReflection bool   `json:"enable_reflection" yaml:"enable-reflection"`
}

type DockerConfig struct {
	Prefix      string `json:"prefix" yaml:"prefix" validate:"required"`
	NetworkName string `json:"network_name" yaml:"network-name" validate:"required"`
}

type DataConfig struct {
	DataDir string `json:"data_dir" yaml:"data-dir" validate:"required"`
}

func WriteConfig(name string, cfg *Config) (err error) {
	ext := filepath.Ext(name)

	var buf []byte
	switch ext {
	case ".yaml", ".yml":
		buf, err = yaml.Marshal(cfg)
	case ".json", ".jsonc":
		buf, err = json.MarshalIndent(cfg, "", "  ")
	default:
		err = fmt.Errorf(
			"failed to locate config file at '%s': unknown extension %s",
			name,
			ext,
		)
	}

	if err != nil {
		return
	}
	err = os.WriteFile(name, buf, 0666)
	return
}

func GetConfig(name string) (*Config, error) {
	file, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(name)
	var cfg Config

	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(file, &cfg)
	case ".json", ".jsonc":
		err = json.Unmarshal(file, &cfg)
	default:
		return nil, fmt.Errorf(
			"failed to open config file at '%s': unknown extension %s",
			name,
			ext,
		)
	}

	return &cfg, err
}
