package config

import (
	"gopkg.in/yaml.v3"
	"os"
)

type DatabaseConfig struct {
	URL            string `yaml:"url"`
	MaxConnections int    `yaml:"max_connections"`
}

type KafkaConfig struct {
	BrokerURL string `yaml:"broker_url"`
	Topic     string `yaml:"topic"`
}

type Config struct {
	Port     string         `yaml:"port"`
	Database DatabaseConfig `yaml:"database"`
	Kafka    KafkaConfig    `yaml:"kafka"`
}

func LoadConfig(filename string) (*Config, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
