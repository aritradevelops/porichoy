package jwtutil

import (
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const testRSAPrivateKey = `
-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQDvvZroFSsPykfINHHhO8eqVPYZP8ox4HznyJ5LfGJS72mIQMWo
SjvrZyHJPvrFEinpL6DTw3eLbrtlVmgrCCIFhJ5hBl7x/hwCH2nthgAAzMvY+YG4
I5S/j7K9ZUx39INaWz0xK/8SQJeBVDwU+7Oc8NTvge9N8LyfXLuEbevQxQIDAQAB
AoGBAKOwVvM9eNBoMcjNni/GDFcNeZbVyi1x9HahsQsjW1L7KpggeZSlUvIw0Y3B
1aD2/Oy9W1cbcCUgrwzLCYbQH8FIczxvBYEznkNsxlhE3q0LCyThgpxRzhrdMn0f
Qm+8KDZncWMiiFH9tX7hKkVKlUBy28v6wtNqZPZPbhi9vA2FAkEA+WWmwTryowFB
18SHcsnWQFQ+Gyz5VUAzQpGyApzI/ChnUUI/1/5LU8we5+50URAErUIoyFfM3cLm
RWKJFU7hnwJBAPYWgUOXLxSIWtQE6VZyGy79ZeBXj6Itr9PKCuaQlu6+g/AmoBby
Ag1JirIMOM6wpPJoR/0udAR1gWOG3RsJ2xsCQCHJ2tjNErhw4CnKb4tmuwdGIo/t
/O3G3+sB8DsYYMaA9tZ0gk/SHQSCYCGOFeYxpGCQ2ROjiZb149q8qdPgNwMCQAZi
JxA9x7bcop6FUhgv9YyOfioHm241iS4RO58neQLQZlPAbL6roGn/0l0z+/VAl8bB
9bwXjGLhOW3/fZTJ+KkCQF4+Hlzwg1l3ZAsdCu+PimDaZ/BksGjrjdriWiNL61Tv
Q55N5n3tj40eWsKl6XRFgE2AYZ8nGg5P/Fr5GBdDK6A=
-----END RSA PRIVATE KEY-----
`
const testRSAPublicKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDvvZroFSsPykfINHHhO8eqVPYZ
P8ox4HznyJ5LfGJS72mIQMWoSjvrZyHJPvrFEinpL6DTw3eLbrtlVmgrCCIFhJ5h
Bl7x/hwCH2nthgAAzMvY+YG4I5S/j7K9ZUx39INaWz0xK/8SQJeBVDwU+7Oc8NTv
ge9N8LyfXLuEbevQxQIDAQAB
-----END PUBLIC KEY-----`

const testHmacKey = "sssshhhhhhhhhh!"

func TestJwtSign_AllAlgorithms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		alg      string
		resolver string
		keyFunc  jwt.Keyfunc
	}{
		// HMAC
		{
			name:     "HS256",
			alg:      "HS256",
			resolver: fmt.Sprintf("literal://%s", testHmacKey),
			keyFunc: func(token *jwt.Token) (any, error) {
				return []byte(testHmacKey), nil
			},
		},
		{
			name:     "HS384",
			alg:      "HS384",
			resolver: fmt.Sprintf("literal://%s", testHmacKey),
			keyFunc: func(token *jwt.Token) (any, error) {
				return []byte(testHmacKey), nil
			},
		},
		{
			name:     "HS512",
			alg:      "HS512",
			resolver: fmt.Sprintf("literal://%s", testHmacKey),
			keyFunc: func(token *jwt.Token) (any, error) {
				return []byte(testHmacKey), nil
			},
		},

		// RSA
		{
			name:     "RS256",
			alg:      "RS256",
			resolver: "literal://" + testRSAPrivateKey,
			keyFunc: func(token *jwt.Token) (any, error) {
				return jwt.ParseRSAPublicKeyFromPEM([]byte(testRSAPublicKey))
			},
		},
		{
			name:     "RS384",
			alg:      "RS384",
			resolver: "literal://" + testRSAPrivateKey,
			keyFunc: func(token *jwt.Token) (any, error) {
				return jwt.ParseRSAPublicKeyFromPEM([]byte(testRSAPublicKey))
			},
		},
		{
			name:     "RS512",
			alg:      "RS512",
			resolver: "literal://" + testRSAPrivateKey,
			keyFunc: func(token *jwt.Token) (any, error) {
				return jwt.ParseRSAPublicKeyFromPEM([]byte(testRSAPublicKey))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			token, err := Sign(
				tt.alg,
				JwtPayload{
					UserID: "user_" + tt.name,
					Email:  "user_" + tt.name,
				},
				tt.resolver,
				"localhost",
				"localhost",
				10*time.Second,
			)

			assert.NoError(t, err)
			assert.NotEmpty(t, token)
			parsed, err := jwt.Parse(token, tt.keyFunc)
			assert.NoError(t, err)
			assert.True(t, parsed.Valid)
		})
	}
}
