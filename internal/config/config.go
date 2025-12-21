package config

import (
	"os"

	"github.com/aritradeveops/porichoy/internal/core/validation"
	"github.com/aritradeveops/porichoy/pkg/timex"
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

type JWT struct {
	Algorithm            string         `yaml:"algorithm" validate:"required"`
	SigningKeyResolver   string         `yaml:"signing_key_resolver" validate:"required,resolver"`
	VerifyingKeyResolver string         `yaml:"verifying_key_resolver" validate:"required,resolver"`
	Lifetime             timex.Duration `yaml:"lifetime" validate:"required,duration"`
}

type RefreshToken struct {
	Lifetime timex.Duration `yaml:"lifetime" validate:"required,duration"`
}

type Oauth struct {
	Lifetime timex.Duration `yaml:"lifetime" validate:"required,duration"`
}

type Authentication struct {
	JWT          JWT          `yaml:"jwt" validate:"required"`
	RefreshToken RefreshToken `yaml:"refresh_token" validate:"required"`
	Oauth        Oauth        `yaml:"oauth" validate:"required"`
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
