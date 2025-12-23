package config

import (
	"os"

	"github.com/aritradeveops/porichoy/internal/core/validation"
	_ "github.com/joho/godotenv/autoload"
	"go.yaml.in/yaml/v3"
)

type Http struct {
	Host string `yaml:"host" validate:"required"`
	Port int    `yaml:"port" validate:"required"`
}

type Database struct {
	URIResolver string `yaml:"uri_resolver" validate:"required"`
}

type UI struct {
	Template string `yaml:"template" validate:"oneof=vanilla"`
}

type Config struct {
	Http     Http     `yaml:"http" validate:"required"`
	Database Database `yaml:"database" validate:"required"`
	UI       UI       `yaml:"ui" validate:"required"`
}

func LoadConfig() (*Config, error) {
	config := &Config{}
	yamlFile, err := os.ReadFile("porichoy.yml")
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(yamlFile, config)
	if err != nil {
		return nil, err
	}

	// if err := resolveSecrets(config); err != nil {
	// 	return nil, err
	// }

	if err := validation.Validate(config); err != nil {
		return nil, err
	}

	return config, nil
}
