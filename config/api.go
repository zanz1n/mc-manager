package config

type APIConfig struct {
	Name   string       `json:"name" yaml:"name"`
	Server ServerConfig `json:"server" yaml:"server"`
	Auth   AuthConfig   `json:"auth" yaml:"auth"`
	DB     DBConfig     `json:"db" yaml:"db"`
	Redis  RedisConfig  `json:"redis" yaml:"redis"`
}

func WriteApiConfig(name string, cfg *APIConfig) error {
	return writeCfg(name, cfg)
}

func GetApiConfig(name string) (*APIConfig, error) {
	var cfg APIConfig
	err := getCfg(name, &cfg)
	return &cfg, err
}
