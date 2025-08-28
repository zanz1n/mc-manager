package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func writeCfg(name string, v any) (err error) {
	ext := filepath.Ext(name)

	var buf []byte
	switch ext {
	case ".yaml", ".yml":
		buf, err = yaml.Marshal(v)
	case ".json", ".jsonc":
		buf, err = json.MarshalIndent(v, "", "  ")
	default:
		err = fmt.Errorf(
			"failed to locate config file at '%s': unknown extension %s",
			name,
			ext,
		)
	}

	if err != nil {
		return
	}
	err = os.WriteFile(name, buf, 0666)
	return
}

func getCfg(name string, v any) error {
	file, err := os.ReadFile(name)
	if err != nil {
		return err
	}

	ext := filepath.Ext(name)

	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(file, v)
	case ".json", ".jsonc":
		err = json.Unmarshal(file, v)
	default:
		err = fmt.Errorf(
			"failed to open config file at '%s': unknown extension %s",
			name,
			ext,
		)
	}

	return err
}
