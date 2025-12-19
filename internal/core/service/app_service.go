package service

import (
	"context"

	"github.com/aritradeveops/porichoy/internal/core/cryptoutil"
	"github.com/aritradeveops/porichoy/internal/core/validation"
	"github.com/aritradeveops/porichoy/internal/persistence/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateAppPayload struct {
	Name               string   `json:"name" validate:"required,min=3"`
	Domain             string   `json:"domain" validate:"required,fqdn"`
	LandingUrl         string   `json:"landing_url" validate:"required,url"`
	Logo               string   `json:"logo,omitempty" validate:"omitempty,url"`
	RedirectUris       []string `json:"redirect_uris" validate:"required,min=1,dive,url"`
	SuccessCallbackUrl string   `json:"success_callback_url" validate:"required,url"`
	ErrorCallbackUrl   string   `json:"error_callback_url" validate:"required,url"`
	JwtAlgo            string   `json:"jwt_algo" validate:"required,jwt_algo"`
	JwtSecretResolver  string   `json:"jwt_secret_resolver" validate:"required,resolver"`
}

func (s *Service) CreateApp(ctx context.Context, initiator string, payload CreateAppPayload) (repository.App, error) {
	var app repository.App
	errs := validation.Validate(payload)
	if errs != nil {
		return app, errs
	}

	// TODO: think about this field
	clientId := payload.Domain
	app, err := s.repository.CreateApp(ctx, repository.CreateAppParams{
		Name:       payload.Name,
		Domain:     payload.Domain,
		LandingUrl: payload.LandingUrl,
		Logo:       pgtype.Text{String: payload.Logo, Valid: payload.Logo != ""},
		CreatedBy:  uuid.MustParse(initiator),
		ClientID:   clientId,
	})

	if err != nil {
		return app, err
	}

	clientSecret, err := cryptoutil.GenerateHash(64)
	if err != nil {
		return app, err
	}

	err = s.repository.CreateOauthInfo(ctx, repository.CreateOauthInfoParams{
		ClientSecret:       clientSecret,
		RedirectUris:       payload.RedirectUris,
		SuccessCallbackUrl: payload.SuccessCallbackUrl,
		ErrorCallbackUrl:   payload.ErrorCallbackUrl,
		JwtAlgo:            payload.JwtAlgo,
		JwtSecretResolver:  pgtype.Text{String: payload.JwtSecretResolver, Valid: true},
		AppID:              app.ID,
		CreatedBy:          uuid.MustParse(initiator),
	})

	return app, err
}
