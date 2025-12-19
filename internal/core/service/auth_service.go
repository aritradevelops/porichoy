package service

import (
	"context"
	"errors"
	"time"

	"github.com/aritradeveops/porichoy/internal/core/cryptoutil"
	"github.com/aritradeveops/porichoy/internal/core/jwtutil"
	"github.com/aritradeveops/porichoy/internal/core/validation"
	"github.com/aritradeveops/porichoy/internal/persistence/repository"
	"github.com/aritradeveops/porichoy/internal/pkg/logger"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUserPayload struct {
	Name     string `json:"name,omitempty" validate:"required,alphaspace,min=5"`
	Email    string `json:"email,omitempty" validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required"`
}

type LoginUserPayload struct {
	Email    string `json:"email,omitempty" validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required"`
}

type AuthTokens struct {
	AccessToken        string    `json:"access_token,omitempty"`
	RefreshToken       string    `json:"refresh_token,omitempty"`
	AccessTokenExpiry  time.Time `json:"access_token_expiry,omitempty"`
	RefreshTokenExpiry time.Time `json:"refresh_token_expiry,omitempty"`
}

var (
	ErrInvalidLoginCredentials = errors.New("auth_service: invalid email or password")
	ErrInvalidLoginMethod      = errors.New("auth_service: invalid login method")
	ErrUserExists              = errors.New("auth_service: user already exists")
	ErrDeactivatedUser         = errors.New("auth_service: user account deactivated")
)

// Errors:
//   - ErrUserExists
func (s *Service) RegisterUser(ctx context.Context, payload RegisterUserPayload) (repository.User, error) {
	var user repository.User
	// validate payload
	errs := validation.Validate(payload)
	if errs != nil {
		return user, errs
	}
	// validate password
	errs = validation.ValidatePassword(payload.Password)
	if errs != nil {
		return user, errs
	}
	// check if user already exists
	user, err := s.repository.FindUserByEmail(ctx, payload.Email)
	if err == nil {
		return user, ErrUserExists
	}
	// create new user
	userId := uuid.New()
	user, err = s.repository.RegisterUser(ctx, repository.RegisterUserParams{
		ID:        userId,
		Email:     payload.Email,
		Name:      payload.Name,
		CreatedBy: userId,
	})
	if err != nil {
		return user, err
	}
	// hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)
	if err != nil {
		return user, err
	}
	// store the password
	err = s.repository.CreatePasswordForUser(ctx, repository.CreatePasswordForUserParams{
		HashedPassword: string(hashedPassword),
		CreatedBy:      user.ID,
	})
	if err != nil {
		return user, err
	}

	return user, nil
}

func (s *Service) LoginUser(ctx context.Context, payload LoginUserPayload) (AuthTokens, error) {
	// validate payload
	var tokens AuthTokens
	errs := validation.Validate(payload)
	if errs != nil {
		return tokens, errs
	}
	user, err := s.repository.FindUserByEmail(ctx, payload.Email)
	if err != nil {
		logger.Error().Err(err).Msg("")
		return tokens, ErrInvalidLoginCredentials
	}

	if user.DeactivatedAt != nil {
		return tokens, ErrDeactivatedUser
	}

	passwd, err := s.repository.GetUserPassword(ctx, user.ID)
	if err != nil {
		// if err is does not exist then throw invalid login method
		logger.Error().Err(err).Msg("")
		return tokens, ErrInvalidLoginMethod
	}

	// compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(passwd.HashedPassword), []byte(payload.Password))
	if err != nil {
		logger.Error().Err(err).Msg("")
		return tokens, ErrInvalidLoginCredentials
	}

	// sign tokens
	dp := ""
	if user.Dp.Valid {
		dp = user.Dp.String
	}
	jwtConfig := s.config.Authentication.JWT
	accessToken, err := jwtutil.Sign(jwtConfig.Algorithm, jwtutil.JwtPayload{
		UserID: user.ID.String(),
		Name:   user.Name,
		Email:  user.Email,
		Dp:     dp,
	}, jwtConfig.SigningKeyResolver, "localhost", "localhost", jwtConfig.ParsedLifetime())

	if err != nil {
		return tokens, err
	}

	refreshToken, err := cryptoutil.GenerateHash(32)
	if err != nil {
		return tokens, err
	}

	tokens.AccessToken = accessToken
	tokens.RefreshToken = refreshToken
	tokens.AccessTokenExpiry = time.Now().Add(5 * time.Hour)
	tokens.RefreshTokenExpiry = time.Now().Add(5 * time.Minute)
	return tokens, nil
}
