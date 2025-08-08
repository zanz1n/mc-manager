package config

type Config struct {
	Docker DockerConfig `json:"docker" yaml:"docker"`
	Data   DataConfig   `json:"data" yaml:"data"`
}

type DockerConfig struct {
	URL         string `json:"url" yaml:"url" validate:"url"`
	Prefix      string `json:"prefix" yaml:"prefix"`
	NetworkName string `json:"network_name" yaml:"network-name"`
}

type DataConfig struct {
	DataDir string `json:"data_dir" yaml:"data-dir"`
}
