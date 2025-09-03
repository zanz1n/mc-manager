package config

import "github.com/zanz1n/mc-manager/internal/dto"

type APIConfig struct {
	Name   string       `json:"name" yaml:"name"`
	Server ServerConfig `json:"server" yaml:"server"`
	Auth   AuthConfig   `json:"auth" yaml:"auth"`
	DB     DBConfig     `json:"db" yaml:"db"`
	Redis  RedisConfig  `json:"redis" yaml:"redis"`

	LocalNode *APILocalNodeConfig `json:"runner" yaml:"runner"`
}

type APILocalNodeConfig struct {
	Enable bool          `json:"enable" yaml:"enable"`
	ID     dto.Snowflake `json:"id" yaml:"id"`
	Docker DockerConfig  `json:"docker" yaml:"docker"`
	Data   DataConfig    `json:"data" yaml:"data"`
}

func WriteApiConfig(name string, cfg *APIConfig) error {
	return writeCfg(name, cfg)
}

func GetApiConfig(name string) (*APIConfig, error) {
	var cfg APIConfig
	err := getCfg(name, &cfg)
	return &cfg, err
}
