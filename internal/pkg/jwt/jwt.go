package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

// Claims holds JWT claims (user id and role).
type Claims struct {
	UserID int64  `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// Manager issues and validates JWT tokens.
type Manager struct {
	secret []byte
	expiry time.Duration
}

// NewManager creates a JWT manager.
func NewManager(secret string, expiry time.Duration) *Manager {
	return &Manager{
		secret: []byte(secret),
		expiry: expiry,
	}
}

// Issue creates a new token for the user.
func (m *Manager) Issue(userID int64, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Parse validates the token and returns claims.
func (m *Manager) Parse(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
