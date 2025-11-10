// Package config ...
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config ...
type Config struct {
	Server struct {
		Host     string `yaml:"host"`
		HTTPPort string `yaml:"http_port"`
		GRPCPort string `yaml:"grpc_port"`
	} `yaml:"service"`
	DataBaseMaster struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"db_name"`
	} `yaml:"db_master"`
	DataBaseReplica struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"db_name"`
	} `yaml:"db_replica"`
	Kafka struct {
		Host      string `yaml:"host"`
		Port      string `yaml:"port"`
		TopicName string `yaml:"order_topic"`
		Brokers   string `yaml:"brokers"`
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
