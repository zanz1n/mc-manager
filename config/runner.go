package config

type RunnerConfig struct {
	Server ServerConfig `json:"server" yaml:"server"`
	Docker DockerConfig `json:"docker" yaml:"docker"`
	Data   DataConfig   `json:"data" yaml:"data"`
}

func WriteRunnerConfig(name string, cfg *RunnerConfig) (err error) {
	return writeCfg(name, cfg)
}

func GetRunnerConfig(name string) (*RunnerConfig, error) {
	var cfg RunnerConfig
	err := getCfg(name, &cfg)
	return &cfg, err
}
