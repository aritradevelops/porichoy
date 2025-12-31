package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/zalando/go-keyring"
)

type CreateAppPayload struct {
	Name string `json:"name" validate:"required,min=3"`
	// TODO: think about this field
	Domain               string   `json:"domain" validate:"required"`
	LandingUrl           string   `json:"landing_url" validate:"required,url"`
	Logo                 string   `json:"logo,omitempty" validate:"omitempty,url"`
	RedirectUris         []string `json:"redirect_uris" validate:"required,min=1,dive,url"`
	RedirectUri          string   `json:"redirect_uri" validate:"required,url"`
	SuccessCallbackUrl   string   `json:"success_callback_url" validate:"required,url"`
	ErrorCallbackUrl     string   `json:"error_callback_url" validate:"required,url"`
	JwtAlgo              string   `json:"jwt_algo" validate:"required,jwt_algo"`
	JwtSecretResolveFrom string   `json:"jwt_secret_resolve_from" validate:"oneof=env db literal"`
	JwtSecretResolver    string   `json:"jwt_secret_resolver" validate:"required,resolver"`
	JwtLifetime          string   `json:"jwt_lifetime" validate:"required,duration"`
	RefreshTokenLifetime string   `json:"refresh_token_lifetime" validate:"required,duration"`
}

var appAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new app",
	Long:  `Add a new app`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := addApp(cmd.Context()); err != nil {
			cobra.CheckErr(err)
		}
	},
}

func addApp(ctx context.Context) error {

	accessToken, err := keyring.Get("porichoy", "access_token")
	if err != nil {
		return err
	}

	questions := []*survey.Question{
		{
			Name: "Name",
			Prompt: &survey.Input{
				Message: "App name:",
			},
			Validate: survey.Required,
		},
		{
			Name: "Domain",
			Prompt: &survey.Input{
				Message: "App domain:",
			},
			Validate: survey.Required,
		},
		{
			Name: "RedirectUri",
			Prompt: &survey.Multiline{
				Message: "Redirect URIs (one per line):",
			},
			Validate: survey.Required,
		},
		{
			Name: "JwtAlgo",
			Prompt: &survey.Select{
				Options: []string{"HS256", "RS256", "JWKS"},
				Message: "Please select an option:",
			},
			Validate: survey.Required,
		},
		{
			Name: "JwtSecretResolveFrom",
			Prompt: &survey.Select{
				Options: []string{"env", "file", "literal"},
				Message: "Please select an option:",
			},
		},
		{
			Name: "JwtSecretResolver",
			Prompt: &survey.Input{
				Message: "Jwt resolver:",
			},
			Validate: survey.Required,
		},
		{
			Name: "JwtLifetime",
			Prompt: &survey.Input{
				Message: "JWT lifetime:",
			},
			Validate: survey.Required,
		},
		{
			Name: "RefreshTokenLifetime",
			Prompt: &survey.Input{
				Message: "Refresh token lifetime:",
			},
			Validate: survey.Required,
		},
		{
			Name: "LandingUrl",
			Prompt: &survey.Input{
				Message: "Landing URL:",
			},
			Validate: survey.Required,
		},
		{
			Name: "ErrorCallbackUrl",
			Prompt: &survey.Input{
				Message: "Error callback URL:",
			},
			Validate: survey.Required,
		},
		{
			Name: "SuccessCallbackUrl",
			Prompt: &survey.Input{
				Message: "Success callback URL:",
			},
			Validate: survey.Required,
		},
	}

	var payload CreateAppPayload
	if err := survey.Ask(questions, &payload); err != nil {
		return err
	}

	payload.RedirectUris = []string{payload.RedirectUri}
	payload.JwtSecretResolver = payload.JwtSecretResolveFrom + "://" + payload.JwtSecretResolver
	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "http://localhost:8080/api/v1/apps/create", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to create app: %s", string(body))
	}

	fmt.Println("App created successfully")
	fmt.Println("App:", string(body))
	return nil
}
