package jwtutil

import (
	"fmt"
	"time"

	"github.com/aritradeveops/porichoy/pkg/resolver"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtPayload struct {
	UserID   string `json:"user_id,omitempty"`
	Name     string `json:"name,omitempty"`
	Email    string `json:"email,omitempty"`
	Dp       string `json:"dp,omitempty"`
	Resolver string `json:"resolver,omitempty"`
}

type Claims struct {
	JwtPayload
	jwt.RegisteredClaims
}

func Sign(alg string, payload JwtPayload, secretResolver string, aud string, iss string, lifetime time.Duration) (string, error) {
	method := jwt.GetSigningMethod(alg)
	factory := resolver.NewResolverFactory()
	payload.Resolver = secretResolver
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

func Verify(token string) (*JwtPayload, error) {
	factory := resolver.NewResolverFactory()

	claims := &Claims{}

	parsed, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (any, error) {
		r, err := factory.Auto(claims.Resolver)
		if err != nil {
			return nil, fmt.Errorf("jwtutil: could not resolve secret: %v", err)
		}
		secret, err := r.Resolve(claims.Resolver)
		if err != nil {
			return nil, fmt.Errorf("jwtutil: could not resolve secret: %v", err)
		}
		switch t.Method.(type) {
		case *jwt.SigningMethodHMAC:
			return []byte(secret.(string)), nil
		case *jwt.SigningMethodRSA:
			key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(secret.(string)))
			if err != nil {
				return nil, err
			}
			return key, nil
		default:
			return nil, fmt.Errorf("jwtutil: %s is not implemented", t.Method.Alg())
		}
	})
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, fmt.Errorf("jwtutil: invalid token")
	}
	return &claims.JwtPayload, nil
}
