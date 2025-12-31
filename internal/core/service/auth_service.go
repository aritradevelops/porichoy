package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/aritradeveops/porichoy/internal/core/cryptoutil"
	"github.com/aritradeveops/porichoy/internal/core/jwtutil"
	"github.com/aritradeveops/porichoy/internal/core/validation"
	"github.com/aritradeveops/porichoy/internal/persistence/repository"
	"github.com/aritradeveops/porichoy/internal/pkg/logger"
	"github.com/aritradeveops/porichoy/pkg/timex"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	ResponseTypeCode  = "code"
	ResponseTypeToken = "token"
)

type RegisterUserPayload struct {
	Name     string `json:"name,omitempty" validate:"required,alphaspace,min=5"`
	Email    string `json:"email,omitempty" validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required"`
}

type RegisterUserResponse struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

type LoginUserPayload struct {
	Email     string `json:"email,omitempty" validate:"required,email"`
	Password  string `json:"password,omitempty" validate:"required"`
	UserAgent string `json:"user_agent,omitempty" validate:"required"`
	UserIP    string `json:"user_ip,omitempty" validate:"required"`
	Host      string `json:"host,omitempty" validate:"required"`
}
type LoginUserResponse struct {
	AccessToken        string    `json:"access_token,omitempty"`
	RefreshToken       string    `json:"refresh_token,omitempty"`
	AccessTokenExpiry  time.Time `json:"access_token_expiry,omitempty"`
	RefreshTokenExpiry time.Time `json:"refresh_token_expiry,omitempty"`
}

type Oauth2Payload struct {
	ClientID            string `json:"client_id" validate:"required"`
	ResponseType        string `json:"response_type" validate:"required,oneof=code token"`
	RedirectURI         string `json:"redirect_uri" validate:"required"`
	CodeChallenge       string `json:"code_challenge"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	State               string `json:"state"`
	LoginHint           string `json:"login_hint"`
	Nonce               string `json:"nonce"`
}
type Oauth2TokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenLifetime  time.Time `json:"access_token_lifetime"`
	RefreshToken         string    `json:"refresh_token"`
	RefreshTokenLifetime time.Time `json:"refresh_token_lifetime"`
}
type Oauth2CodeResponse struct {
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri"`
}

type Oauth2Response struct {
	App                 repository.App         `json:"app"`
	OauthConfig         repository.OauthConfig `json:"oauth_config"`
	Oauth2TokenResponse *Oauth2TokenResponse   `json:"oauth2_token_response"`
	Oauth2CodeResponse  *Oauth2CodeResponse    `json:"oauth2_code_response"`
}

type Oauth2TokenPayload struct {
	ClientID     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
	GrantType    string `json:"grant_type" validate:"required,oneof=authorization_code client_credentials"`
	Code         string `json:"code" validate:"required"`
	RedirectURI  string `json:"redirect_uri" validate:"required"`
	UserAgent    string `json:"user_agent" validate:"required"`
	UserIP       string `json:"user_ip" validate:"required"`
}

type Oauth2ConsentPayload struct {
	ClientID string `json:"client_id" validate:"required"`
}

type Oauth2ConsentResponse struct {
	AppName string `json:"app_name"`
}

const (
	OauthCodeLifetime = 10 * time.Minute
)

var (
	ErrInvalidLoginCredentials = errors.New("auth_service: invalid email or password")
	ErrInvalidLoginMethod      = errors.New("auth_service: invalid login method")
	ErrUserExists              = errors.New("auth_service: user already exists")
	ErrDeactivatedUser         = errors.New("auth_service: user account deactivated")
	ErrInvalidOauthCall        = errors.New("auth_service: invalid oauth call")
	ErrInvalidRedirectUri      = errors.New("auth_service: invalid redirect uri")
	ErrInternalError           = errors.New("auth_service: internal error")
)

func (s *Service) RegisterUser(ctx context.Context, payload RegisterUserPayload, isRootUser bool) (RegisterUserResponse, error) {
	var response RegisterUserResponse
	var user repository.User
	// validate payload
	errs := validation.Validate(payload)
	if errs != nil {
		return response, errs
	}
	// validate password
	errs = validation.ValidatePassword(payload.Password)
	if errs != nil {
		return response, errs
	}
	// check if user already exists
	user, err := s.repository.FindUserByEmail(ctx, payload.Email)
	if err == nil {
		return response, ErrUserExists
	}
	var userId uuid.UUID
	// create new user
	if !isRootUser {
		userId = uuid.New()
	} else {
		userId = uuid.Nil
	}
	user, err = s.repository.RegisterUser(ctx, repository.RegisterUserParams{
		ID:        userId,
		Email:     payload.Email,
		Name:      payload.Name,
		CreatedBy: userId,
	})
	if err != nil {
		return response, err
	}
	// hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return response, err
	}
	// store the password
	err = s.repository.CreatePasswordForUser(ctx, repository.CreatePasswordForUserParams{
		HashedPassword: string(hashedPassword),
		CreatedBy:      user.ID,
	})
	if err != nil {
		return response, err
	}

	response.ID = user.ID.String()
	response.Name = user.Name
	response.Email = user.Email
	return response, nil
}

func (s *Service) LoginUser(ctx context.Context, payload LoginUserPayload) (LoginUserResponse, error) {
	// validate payload
	var response LoginUserResponse
	errs := validation.Validate(payload)
	if errs != nil {
		return response, errs
	}
	rootApp, err := s.repository.FindRootApp(ctx)
	if err != nil {
		logger.Error().Err(err).Msg("zero")
		return response, err
	}

	user, err := s.repository.FindUserByEmail(ctx, payload.Email)
	if err != nil {
		logger.Error().Err(err).Msg("one")
		return response, ErrInvalidLoginCredentials
	}

	if user.DeactivatedAt != nil {
		return response, ErrDeactivatedUser
	}

	passwd, err := s.repository.FindUserPassword(ctx, user.ID)
	if err != nil {
		// if err is does not exist then throw invalid login method
		logger.Error().Err(err).Msg("two")
		return response, ErrInvalidLoginMethod
	}

	// compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(passwd.HashedPassword), []byte(payload.Password))
	if err != nil {
		logger.Error().Err(err).Msg("three")
		return response, ErrInvalidLoginCredentials
	}

	// sign tokens
	dp := ""
	if user.Dp.Valid {
		dp = user.Dp.String
	}
	accessToken, err := jwtutil.Sign(rootApp.OauthConfig.JwtAlgo, jwtutil.JwtPayload{
		UserID: user.ID.String(),
		Name:   user.Name,
		Email:  user.Email,
		Dp:     dp,
	}, rootApp.OauthConfig.JwtSecretResolver.String, rootApp.App.Domain, rootApp.App.Domain,
		timex.Duration(rootApp.OauthConfig.JwtLifetime).Duration())

	if err != nil {
		logger.Error().Err(err).Msg("four")
		return response, err
	}

	refreshToken, err := cryptoutil.GenerateHash(32)
	if err != nil {
		return response, err
	}

	// create session
	err = s.repository.CreateSession(ctx, repository.CreateSessionParams{
		UserID:       user.ID,
		AppID:        rootApp.App.ID,
		RefreshToken: refreshToken,
		UserIp:       payload.UserIP,
		UserAgent:    payload.UserAgent,
		ExpiresAt:    time.Now().Add(timex.Duration(rootApp.OauthConfig.RefreshTokenLifetime).Duration()),
		CreatedBy:    user.ID,
	})
	if err != nil {
		return response, err
	}

	response.AccessToken = accessToken
	response.RefreshToken = refreshToken
	response.AccessTokenExpiry = time.Now().Add(timex.Duration(rootApp.OauthConfig.JwtLifetime).Duration())
	response.RefreshTokenExpiry = time.Now().Add(timex.Duration(rootApp.OauthConfig.RefreshTokenLifetime).Duration())
	return response, nil
}

func (s *Service) LogoutUser(ctx context.Context, initiator string) error {
	return s.repository.DeleteSession(ctx, uuid.MustParse(initiator))
}

func (s *Service) Oauth2(ctx context.Context, initiator string, payload Oauth2Payload) (Oauth2Response, error) {
	errs := validation.Validate(payload)
	var response Oauth2Response
	if errs != nil {
		return response, errs
	}
	app, err := s.repository.FindAppByClientID(ctx, payload.ClientID)
	if err != nil {
		logger.Error().Err(err).Msg("one")
		return response, ErrInvalidOauthCall
	}
	response.App = app.App
	response.OauthConfig = app.OauthConfig

	redirectUri, err := url.Parse(payload.RedirectURI)
	if err != nil {
		logger.Error().Err(err).Msg("two")
		return response, ErrInvalidOauthCall
	}
	fmt.Println(app.OauthConfig.RedirectUris, strings.Split(redirectUri.String(), "?")[0])
	if !slices.Contains(app.OauthConfig.RedirectUris, strings.Split(redirectUri.String(), "?")[0]) {
		logger.Error().Err(err).Msg("three")
		return response, ErrInvalidRedirectUri
	}

	if payload.ResponseType == ResponseTypeCode {
		code, err := cryptoutil.GenerateHash(32)
		if err != nil {
			logger.Error().Err(err).Msg("four")
			return response, ErrInternalError
		}
		err = s.repository.CreateOauthCall(ctx, repository.CreateOauthCallParams{
			AppID:     app.App.ID,
			Code:      code,
			UserID:    uuid.MustParse(initiator),
			ExpiresAt: time.Now().Add(OauthCodeLifetime),
		})
		if err != nil {
			logger.Error().Err(err).Msg("five")
			return response, ErrInternalError
		}
		q := redirectUri.Query()
		q.Add("code", code)
		redirectUri.RawQuery = q.Encode()

		response.Oauth2CodeResponse = &Oauth2CodeResponse{
			Code:        code,
			RedirectURI: redirectUri.String(),
		}
		return response, nil
	} else if payload.ResponseType == ResponseTypeToken {
		// TODO: and as it is inherently insecure so it's not important as of now
	}

	return response, nil
}

func (s *Service) Oauth2ConsentResponse(ctx context.Context, payload Oauth2ConsentPayload) (Oauth2ConsentResponse, error) {
	errs := validation.Validate(payload)
	var resp Oauth2ConsentResponse
	if errs != nil {
		return resp, errs
	}
	app, err := s.repository.FindAppByClientID(ctx, payload.ClientID)
	if err != nil {
		return resp, err
	}
	resp.AppName = app.App.Name
	return resp, nil
}

func (s *Service) Token(ctx context.Context, payload Oauth2TokenPayload) (Oauth2TokenResponse, error) {
	var resp Oauth2TokenResponse

	errs := validation.Validate(payload)
	if errs != nil {
		return resp, errs
	}

	app, err := s.repository.FindAppByClientID(ctx, payload.ClientID)
	if err != nil {
		return resp, err
	}

	if app.OauthConfig.ClientSecret != payload.ClientSecret {
		return resp, ErrInvalidOauthCall
	}

	// if app.OauthConfig.GrantType != payload.GrantType {
	// 	return resp, ErrInvalidOauthCall
	// }

	if payload.GrantType == "authorization_code" {
		oauthCall, err := s.repository.FindOauthCallByCode(ctx, payload.Code)
		if err != nil {
			return resp, err
		}

		user, err := s.repository.FindUserByID(ctx, oauthCall.UserID)
		if err != nil {
			return resp, err
		}
		logger.Info().Any("user", app).Msg("app")
		accessToken, err := jwtutil.Sign(app.OauthConfig.JwtAlgo, jwtutil.JwtPayload{
			UserID: user.ID.String(),
			Name:   user.Name,
			Email:  user.Email,
			Dp:     user.Dp.String,
		}, app.OauthConfig.JwtSecretResolver.String, app.App.Domain, s.config.Http.Host, timex.Duration(app.OauthConfig.JwtLifetime).Duration())

		if err != nil {
			return resp, err
		}

		resp.AccessToken = accessToken
		resp.AccessTokenLifetime = time.Now().Add(timex.Duration(app.OauthConfig.JwtLifetime).Duration())

		refreshToken, err := cryptoutil.GenerateHash(64)
		if err != nil {
			return resp, err
		}

		err = s.repository.CreateSession(ctx, repository.CreateSessionParams{
			UserID:       user.ID,
			AppID:        app.App.ID,
			RefreshToken: refreshToken,
			UserIp:       payload.UserIP,
			UserAgent:    payload.UserAgent,
			ExpiresAt:    time.Now().Add(timex.Duration(app.OauthConfig.RefreshTokenLifetime).Duration()),
			CreatedBy:    user.ID,
		})
		if err != nil {
			return resp, err
		}

		resp.RefreshToken = refreshToken
		resp.RefreshTokenLifetime = time.Now().Add(timex.Duration(app.OauthConfig.RefreshTokenLifetime).Duration())
	}

	return resp, nil
}
