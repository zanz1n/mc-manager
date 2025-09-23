package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/zanz1n/mc-manager/config"
)

var configFile = flag.String(
	"config",
	"/etc/mc/config.yaml",
	"the path of the config file",
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

	cfg, err := config.GetRunnerConfig(*configFile)
	if err != nil {
		log.Fatalln("Failed to get config:", err)
	}

	err = config.WriteRunnerConfig(*configFile, cfg)
	if err != nil {
		log.Fatalln("Failed to format config:", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sig := <-endCh
		log.Printf("Received signal %s: closing server ...\n", sig.String())
		cancel()
	}()

	Run(ctx, cfg)
}
