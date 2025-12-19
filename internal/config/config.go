package config

import (
	"os"
	"strconv"
	"time"

	"github.com/aritradeveops/porichoy/internal/core/validation"
	"go.yaml.in/yaml/v3"
)

type Http struct {
	Host string `yaml:"host" validate:"required"`
	Port int    `yaml:"port" validate:"required"`
}

type Database struct {
	URIResolver string `yaml:"uri_resolver" validate:"required"`
}

type JWT struct {
	Algorithm            string `yaml:"algorithm" validate:"required"`
	SigningKeyResolver   string `yaml:"signing_key_resolver" validate:"required,resolver"`
	VerifyingKeyResolver string `yaml:"verifying_key_resolver" validate:"required,resolver"`
	Lifetime             string `yaml:"lifetime" validate:"required,duration"`
}

func (j JWT) ParsedLifetime() time.Duration {
	duration, err := time.ParseDuration(j.Lifetime)
	if err != nil {
		isD := j.Lifetime[len(j.Lifetime)-1] == 'd'
		if isD {
			rest := j.Lifetime[:len(j.Lifetime)-1]
			days, _ := strconv.Atoi(rest)
			return time.Duration(days) * 24 * time.Hour
		}
	}
	return duration
}

type Authentication struct {
	JWT                  JWT    `yaml:"jwt" validate:"required"`
	RefreshTokenLifetime string `yaml:"refresh_token_lifetime" validate:"required,duration"`
}

type Config struct {
	Http           Http           `yaml:"http" validate:"required"`
	Database       Database       `yaml:"database" validate:"required"`
	Authentication Authentication `yaml:"authentication" validate:"required"`
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
