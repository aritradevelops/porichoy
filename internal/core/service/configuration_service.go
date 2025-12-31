package service

import (
	"context"

	"github.com/aritradeveops/porichoy/internal/core/validation"
	"github.com/aritradeveops/porichoy/internal/pkg/logger"
)

type ConfigurationPayload struct {
	RootUser RegisterUserPayload `json:"root_user"`
	RootApp  CreateAppPayload    `json:"root_app"`
}

func (s *Service) Configure(ctx context.Context, config ConfigurationPayload) error {
	// validate
	errs := validation.Validate(config)
	if errs != nil {
		return errs
	}

	user, err := s.RegisterUser(ctx, config.RootUser, true)
	if err != nil {
		return err
	}
	logger.Info().Any("root user", user).Msg("root user registered successfully!")
	app, err := s.CreateApp(ctx, user.ID, config.RootApp)
	if err != nil {
		return err
	}
	logger.Info().Any("root app", app).Msg("root app created successfully!")
	return nil
}
