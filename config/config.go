package config

import (
	"net"
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
	return writeCfg(name, cfg)
}

func GetConfig(name string) (*Config, error) {
	var cfg Config
	err := getCfg(name, &cfg)
	return &cfg, err
}
