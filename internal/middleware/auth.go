package middleware

import (
	"net/http"
	"strings"

	"magua-fly-golang/internal/pkg/jwt"
)

const BearerPrefix = "Bearer "

// Auth extracts JWT from Authorization header and sets claims in context (or 401).
func Auth(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, BearerPrefix) {
				http.Error(w, `{"error":"missing or invalid authorization"}`, http.StatusUnauthorized)
				return
			}
			tokenString := strings.TrimPrefix(auth, BearerPrefix)
			claims, err := jwtManager.Parse(tokenString)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			ctx := WithClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth parses JWT if present but does not require it.
func OptionalAuth(jwtManager *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth != "" && strings.HasPrefix(auth, BearerPrefix) {
				tokenString := strings.TrimPrefix(auth, BearerPrefix)
				if claims, err := jwtManager.Parse(tokenString); err == nil {
					r = r.WithContext(WithClaims(r.Context(), claims))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
