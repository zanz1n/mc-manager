package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/dto"
)

var (
	configFile = flag.String(
		"config",
		"/etc/mc/config.yaml",
		"the path of the config file",
	)
	migrateOpt = flag.Bool(
		"migrate",
		false,
		"executes migrations automatically",
	)
)

func init() {
	flag.Parse()

	if cenv := os.Getenv("CONFIG_FILE"); cenv != "" {
		*configFile = cenv
	}
}

func main() {
	endCh := make(chan os.Signal, 1)
	signal.Notify(endCh, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.GetApiConfig(*configFile)
	if err != nil {
		log.Fatalln("Failed to get config:", err)
	}

	if cfg.LocalNode != nil {
		if cfg.LocalNode.ID == 0 {
			cfg.LocalNode.ID = dto.NewSnowflake()
		}
	}

	err = config.WriteApiConfig(*configFile, cfg)
	if err != nil {
		log.Fatalln("Failed to format config:", err)
	}

	if *migrateOpt {
		cfg.DB.Migrate = true
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sig := <-endCh
		log.Printf("Received signal %s: closing server ...\n", sig.String())
		cancel()
	}()

	Run(ctx, cfg)
}
