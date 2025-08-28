package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"database/sql"
	"encoding/pem"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/valkey-io/valkey-go"
	"github.com/zanz1n/mc-manager/config"
	"github.com/zanz1n/mc-manager/internal/db"
	"github.com/zanz1n/mc-manager/internal/kv"
	sqlembed "github.com/zanz1n/mc-manager/sql"
)

func openKV(ctx context.Context, cfg *config.APIConfig) (kv.KVStorer, error) {
	opt, err := valkey.ParseURL(cfg.Redis.URL)
	if err != nil {
		return nil, err
	}

	cli, err := valkey.NewClient(opt)
	if err != nil {
		return nil, err
	}

	return kv.NewRedisKV(cli), nil
}

func openDB(ctx context.Context, cfg *config.APIConfig) (db.Querier, *sql.DB, error) {
	sqldb, err := sql.Open("pgx/v5", cfg.DB.URL)
	if err != nil {
		return nil, nil, err
	}
	sqldb.SetMaxOpenConns(cfg.DB.MaxConns)

	if !cfg.DB.SkipPreparation {
		if err = sqldb.PingContext(ctx); err != nil {
			return nil, nil, err
		}
	}

	if cfg.DB.Migrate {
		if err = migrate(ctx, sqldb); err != nil {
			return nil, nil, fmt.Errorf("migrate: %w", err)
		}
	}

	var q db.Querier
	if !cfg.DB.SkipPreparation {
		if q, err = db.Prepare(ctx, sqldb); err != nil {
			return nil, nil, err
		}
	} else {
		q = db.New(sqldb)
	}

	return q, sqldb, nil
}

func loadEd25519(cfg *config.APIConfig) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pub, err := parseKeyFile[ed25519.PublicKey](cfg.Auth.PublicKey, false)
	if err != nil {
		return generateEd25519(cfg)
	}

	priv, err := parseKeyFile[ed25519.PrivateKey](cfg.Auth.PrivateKey, true)
	if err != nil {
		return nil, nil, err
	}

	return pub, priv, nil
}

func generateEd25519(cfg *config.APIConfig) (ed25519.PublicKey, ed25519.PrivateKey, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("generate Ed25519 key: %w", err)
	}

	err = marshalKeyFile(cfg.Auth.PublicKey, pub, false)
	if err != nil {
		return nil, nil, err
	}

	err = marshalKeyFile(cfg.Auth.PrivateKey, priv, true)
	if err != nil {
		return nil, nil, err
	}

	return pub, priv, err
}

func parseKeyFile[T any](name string, private bool) (res T, err error) {
	file, err := os.ReadFile(name)
	if err != nil {
		return
	}

	block, _ := pem.Decode(file)

	var key any
	if private {
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	} else {
		key, err = x509.ParsePKIXPublicKey(block.Bytes)
	}

	if err != nil {
		return
	}
	res = key.(T)

	return
}

func marshalKeyFile(name string, key any, private bool) (err error) {
	var data []byte
	if private {
		data, err = x509.MarshalPKCS8PrivateKey(key)
	} else {
		data, err = x509.MarshalPKIXPublicKey(key)
	}

	if err != nil {
		return
	}

	file, err := os.Create(name)
	if err != nil {
		return
	}

	mode := "PUBLIC"
	if private {
		mode = "PRIVATE"
	}

	err = pem.Encode(file, &pem.Block{
		Type:  fmt.Sprintf("BEGIN %s KEY", mode),
		Bytes: data,
	})
	return
}

func migrate(ctx context.Context, db *sql.DB) error {
	logger := slog.NewLogLogger(slog.Default().Handler(), slog.LevelInfo)
	logger.SetPrefix("Database: ")

	goose.SetLogger(logger)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	goose.SetBaseFS(sqlembed.Migrations)
	goose.SetSequential(true)

	return goose.UpContext(ctx, db, "migrations")
}
