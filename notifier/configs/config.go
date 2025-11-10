// Package config ...
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config ...
type Config struct {
	Kafka struct {
		Host            string `yaml:"host"`
		Port            string `yaml:"port"`
		TopicName       string `yaml:"order_topic"`
		ConsumerGroupID string `yaml:"consumer_group_id"`
		Brokers         string `yaml:"brokers"`
	} `yaml:"kafka"`
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
