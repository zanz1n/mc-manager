package config

import (
	"net"
	"strings"
	"time"
)

type ServerConfig struct {
	IP               net.IP `json:"ip" yaml:"ip" validate:"required"`
	Port             uint16 `json:"port" yaml:"port" validate:"required"`
	Password         string `json:"password" yaml:"password"`
	EnableReflection bool   `json:"enable_reflection" yaml:"enable-reflection"`

	TLS *ServerTLSConfig `json:"tls" yaml:"tls"`
}

type ServerTLSConfig struct {
	Enable       bool   `json:"enable" yaml:"enable"`
	AutoGenerate bool   `json:"auto_generate" yaml:"auto-generate"`
	Certificate  string `json:"cert" yaml:"cert"`
	Key          string `json:"key" yaml:"key"`
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
	Type            string `json:"type" yaml:"type"`
	URL             string `json:"url" yaml:"url"`
	MaxConns        int    `json:"max_conns" yaml:"max-conns"`
	SkipPreparation bool   `json:"skip_preparation" yaml:"skip-preparation"`
	Migrate         bool   `json:"migrate" yaml:"migrate"`

	kind   string
	driver string
}

func (c *DBConfig) DBKind() string {
	if c.kind != "" {
		return c.kind
	}

	switch c.Type {
	case "postgres", "postgresql", "pg", "pgx", "pgx/v5":
		c.kind = "postgres"

	case "", "sqlite3", "local", "sqlite", "file", "fs":
		c.kind = "sqlite"
	}

	return c.kind
}

func (c *DBConfig) DriverName() string {
	if c.driver != "" {
		return c.driver
	}

	switch c.DBKind() {
	case "postgres":
		c.driver = "pgx/v5"

	case "sqlite":
		c.driver = "sqlite"
	}

	return c.driver
}

func (c *DBConfig) ConnString() string {
	if c.DriverName() == "sqlite" {
		if !strings.HasPrefix(c.URL, "file:") {
			return "file:" + c.URL + "?_time_integer_format=unix_milli&_inttotime=true"
		}
	}
	return c.URL
}

type CacheConfig struct {
	Type         string        `json:"type" yaml:"type"`
	URL          string        `json:"url" yaml:"url"`
	SaveInterval time.Duration `json:"save_interval" yaml:"save-interval"`
}
