// Package config ...
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config ...
type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"service"`
	ProductService struct {
		Host  string `yaml:"host"`
		Port  string `yaml:"port"`
		Token string `yaml:"token"`
		Limit int    `yaml:"limit"`
		Burst int    `yaml:"burst"`
	} `yaml:"product_service"`
	LomsService struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"loms_service"`
	Jaeger struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"jaeger"`
}

// LoadConfig ...
func LoadConfig() (*Config, error) {
	f, err := os.Open(os.Getenv("CONFIG_FILE"))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			return
		}
	}()

	config := &Config{}
	if err := yaml.NewDecoder(f).Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}
