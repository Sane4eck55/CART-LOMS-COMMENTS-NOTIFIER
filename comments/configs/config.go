// Package configs ...
package configs

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
	DBShards []DBShard `yaml:"db_shards"`
}

// DBShard ...
type DBShard struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
	ShardID  string `yaml:"shard_id"`
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
