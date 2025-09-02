package config

import (
	"net"
	"time"
)

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
