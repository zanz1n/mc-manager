package runner

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type mcOperator struct {
	UUID                uuid.UUID `json:"uuid"`
	Name                string    `json:"name"`
	Level               uint8     `json:"level"`
	BypassesPlayerLimit bool      `json:"bypassesPlayerLimit"`
}

func sanitizeEula(dataDir string) error {
	filePath := path.Join(dataDir, "eula.txt")
	return os.WriteFile(filePath, []byte("eula=true\n"), 0666)
}

func sanitizeMcProperties(dataDir string, instance *Instance) error {
	filePath := path.Join(dataDir, "server.properties")
	file, err := os.Open(filePath)

	config := make(map[string]string)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	} else {
		config, err = readMcProperties(file)
		if err != nil {
			return err
		}
	}

	if instance.Config.Difficulty != "" {
		config["difficulty"] = instance.Config.Difficulty
	} else {
		config["difficulty"] = "easy"
	}

	if instance.Limits.MaxPlayers != 0 {
		config["max-players"] = strconv.Itoa(int(instance.Limits.MaxPlayers))
	}

	if instance.Config.ViewDistance == 0 {
		instance.Config.ViewDistance = 8
	}
	if instance.Config.SimulationDistance == 0 {
		instance.Config.SimulationDistance = 7
	}

	config["motd"] = instance.Name

	config["view-distance"] = strconv.Itoa(int(instance.Config.ViewDistance))
	config["simulation-distance"] = strconv.Itoa(int(instance.Config.SimulationDistance))

	config["online-mode"] = strconv.FormatBool(!instance.Config.AllowPirate)
	config["query.port"] = strconv.Itoa(int(instance.Config.Port))
	config["spawn-protection"] = "0"

	file, err = os.Create(filePath)
	if err != nil {
		return err
	}

	return writeMcProperties(file, config)
}

func writeMcProperties(file io.WriteCloser, config map[string]string) error {
	defer file.Close()

	for k, v := range config {
		_, err := file.Write([]byte(k + "=" + v + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func readMcProperties(file io.ReadCloser) (map[string]string, error) {
	defer file.Close()

	config := make(map[string]string)
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')

		if equal := strings.Index(line, "="); equal >= 0 {
			if key := strings.TrimSpace(line[:equal]); len(key) > 0 {
				value := ""
				if len(line) > equal {
					value = strings.TrimSpace(line[equal+1:])
				}
				config[key] = value
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return config, nil
}
