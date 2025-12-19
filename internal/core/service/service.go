package service

import (
	"github.com/aritradeveops/porichoy/internal/config"
	"github.com/aritradeveops/porichoy/internal/persistence/repository"
)

type Service struct {
	config     *config.Config
	repository repository.Querier
}

func New(config *config.Config, repository repository.Querier) *Service {
	return &Service{
		config:     config,
		repository: repository,
	}
}
