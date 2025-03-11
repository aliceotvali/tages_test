package config

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Port        int           `yaml:"port"`
	Timeout     time.Duration `yaml:"timeout"`
	StoragePath string        `yaml:"storage_path"`
	CreateLimit int           `yaml:"create_limit"`
	GetLimit    int           `yaml:"get_limit"`
	ListLimit   int           `yaml:"list_limit"`
}

// ParseConfig получает и обрабатывает конфиг.
func ParseConfig(path string) (Config, error) {
	var config Config

	filename, err := filepath.Abs(path)
	if err != nil {
		return config, err
	}

	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return config, err
	}

	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return config, err
	}

	return config, nil
}

// Validate валидирует конфиг.
func (c *Config) Validate() error {
	if c.Port <= 0 {
		return errors.New("port must be greater than zero")
	}

	if c.CreateLimit <= 0 {
		return errors.New("create_limit must be greater than zero")
	}

	if c.GetLimit <= 0 {
		return errors.New("get_limit must be greater than zero")
	}

	if c.ListLimit <= 0 {
		return errors.New("list_limit must be greater than zero")
	}

	if c.StoragePath == "" {
		c.StoragePath = "binary_files"
	}

	if c.Timeout <= 0 {
		return errors.New("timeout can not be negative or 0")
	}

	return nil
}
