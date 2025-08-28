package config

import "time"

type APIConfig struct {
	Name   string       `json:"name" yaml:"name"`
	Server ServerConfig `json:"server" yaml:"server"`
	Auth   AuthConfig   `json:"auth" yaml:"auth"`
	DB     DBConfig     `json:"db" yaml:"db"`
	Redis  RedisConfig  `json:"redis" yaml:"redis"`
}

type AuthConfig struct {
	JWTExpiration time.Duration `json:"jwt_expiration" yaml:"jwt-expiration" validate:"required"`
	AllowSignup   bool          `json:"allow_signup" yaml:"allow-signup"`
	BcryptCost    uint8         `json:"bcrypt_cost" yaml:"bcrypt-cost" validate:"gte=8,lte=16"`

	PrivateKey string `json:"private_key" yaml:"private-key"`
	PublicKey  string `json:"public_key" yaml:"public-key"`
}

type DBConfig struct {
	URL             string `json:"url" yaml:"url" validate:"url"`
	MaxConns        int    `json:"max_conns" yaml:"max-conns"`
	SkipPreparation bool   `json:"skip_preparation" yaml:"skip-preparation"`
	Migrate         bool   `json:"migrate" yaml:"migrate"`
}

type RedisConfig struct {
	URL string `json:"url" yaml:"url" validate:"url"`
}

func WriteApiConfig(name string, cfg *APIConfig) error {
	return writeCfg(name, cfg)
}

func GetApiConfig(name string) (*APIConfig, error) {
	var cfg APIConfig
	err := getCfg(name, &cfg)
	return &cfg, err
}
