package config

import (
	"fmt"
	"os"

	"github.com/manhhung2111/go-idm/config"
	"gopkg.in/yaml.v2"
)

type ConfigFilePath string

type Config struct {
	GRPC     GRPC     `yaml:"grpc"`
	HTTP     HTTP     `yaml:"http"`
	Auth     Auth     `yaml:"auth"`
	Log      Log      `yaml:"log"`
	Database Database `yaml:"database"`
	Cache    Cache    `yaml:"cache"`
}

func NewConfig(filePath ConfigFilePath) (Config, error) {
	var (
		configBytes = config.DefaultConfigBytes
		config      = Config{}
		err         error
	)

	if filePath != "" {
		configBytes, err = os.ReadFile(string(filePath))
		if err != nil {
			return Config{}, fmt.Errorf("failed to read YAML file: %w", err)
		}
	}

	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal Yaml: %w", err)
	}

	return config, nil
}
