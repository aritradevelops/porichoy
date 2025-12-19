package jwtutil

import (
	"fmt"
	"time"

	"github.com/aritradeveops/porichoy/internal/pkg/resolver"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtPayload struct {
	UserID string `json:"user_id,omitempty"`
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
	Dp     string `json:"dp,omitempty"`
}

type Claims struct {
	JwtPayload
	jwt.RegisteredClaims
}

func Sign(alg string, payload JwtPayload, secretResolver string, aud string, iss string, lifetime time.Duration) (string, error) {
	method := jwt.GetSigningMethod(alg)
	factory := resolver.NewResolverFactory()
	r, err := factory.Auto(secretResolver)
	if err != nil {
		return "", fmt.Errorf("jwtutil: could not resolver secret: %v", err)
	}
	secret, err := r.Resolve(secretResolver)
	if err != nil {
		return "", fmt.Errorf("jwtutil: could not resolver secret: %v", err)
	}
	secretStr := secret.(string)

	now := time.Now()
	token := jwt.NewWithClaims(method, Claims{
		JwtPayload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    iss,
			Subject:   payload.Email,
			Audience:  []string{aud},
			ExpiresAt: jwt.NewNumericDate(now.Add(lifetime)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	})
	switch method.(type) {
	case *jwt.SigningMethodHMAC:
		signed, err := token.SignedString([]byte(secretStr))
		return signed, err
	case *jwt.SigningMethodRSA:
		key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(secretStr))
		if err != nil {
			return "", err
		}
		signed, err := token.SignedString(key)
		if err != nil {
			return "", err
		}
		return signed, err

	default:
		return "", fmt.Errorf("jwtutil: %s is not implemented", method.Alg())
	}
}

func Verify(alg string, token string, secretResolver string) (*JwtPayload, error) {
	method := jwt.GetSigningMethod(alg)
	factory := resolver.NewResolverFactory()
	r, err := factory.Auto(secretResolver)
	if err != nil {
		return nil, fmt.Errorf("jwtutil: could not resolve secret: %v", err)
	}
	secret, err := r.Resolve(secretResolver)
	if err != nil {
		return nil, fmt.Errorf("jwtutil: could not resolve secret: %v", err)
	}
	secretStr := secret.(string)

	claims := &Claims{}
	switch method.(type) {
	case *jwt.SigningMethodHMAC:
		parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
			return []byte(secretStr), nil
		})
		if err != nil {
			return nil, err
		}
		if !parsed.Valid {
			return nil, fmt.Errorf("jwtutil: invalid token")
		}
		return &claims.JwtPayload, nil
	case *jwt.SigningMethodRSA:
		key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(secretStr))
		if err != nil {
			return nil, err
		}
		parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
			return key, nil
		})
		if err != nil {
			return nil, err
		}
		if !parsed.Valid {
			return nil, fmt.Errorf("jwtutil: invalid token")
		}
		return &claims.JwtPayload, err

	default:
		return &claims.JwtPayload, fmt.Errorf("jwtutil: %s is not implemented", method.Alg())
	}
}
