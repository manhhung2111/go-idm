package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
)

type ConfigFilePath string

type Config struct {
	Auth     Auth     `yaml:"auth"`
	Log      Log      `yaml:"log"`
	Database Database `yaml:"database"`
}

func NewConfig(filePath ConfigFilePath) (Config, error) {
	configBytes, err := os.ReadFile(string(filePath))
	if err != nil {
		return Config{}, fmt.Errorf("failed to read Yaml file: %w", err)
	}
	config := Config{}
	err = yaml.Unmarshal(configBytes, &config)
	if err != nil {
		return Config{}, fmt.Errorf("failed to unmarshal Yaml: %w", err)
	}

	return config, nil
}
