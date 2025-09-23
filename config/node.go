package config

type NodeConfig struct {
	Server ServerConfig `json:"server" yaml:"server"`
	Docker DockerConfig `json:"docker" yaml:"docker"`
	Data   DataConfig   `json:"data" yaml:"data"`
}

func WriteRunnerConfig(name string, cfg *NodeConfig) (err error) {
	return writeCfg(name, cfg)
}

func GetRunnerConfig(name string) (*NodeConfig, error) {
	var cfg NodeConfig
	err := getCfg(name, &cfg)
	return &cfg, err
}
