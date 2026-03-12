package middleware

import (
	"context"

	"magua-fly-golang/internal/pkg/jwt"
)

type contextKey string

const claimsKey contextKey = "claims"

// WithClaims stores JWT claims in context.
func WithClaims(ctx context.Context, c *jwt.Claims) context.Context {
	return context.WithValue(ctx, claimsKey, c)
}

// GetClaims returns claims from context or nil.
func GetClaims(ctx context.Context) *jwt.Claims {
	c, _ := ctx.Value(claimsKey).(*jwt.Claims)
	return c
}
