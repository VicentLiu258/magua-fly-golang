package middleware

import (
	"net/http"
	"strings"

	"magua-fly-golang/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

const GinClaimsKey = "claims"

// GinAuth extracts JWT and sets claims in gin.Context; returns 401 if invalid.
func GinAuth(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, BearerPrefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid authorization"})
			return
		}
		tokenString := strings.TrimPrefix(auth, BearerPrefix)
		claims, err := jwtManager.Parse(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Set(GinClaimsKey, claims)
		c.Next()
	}
}

// GetGinClaims returns claims from gin context (from GinAuth).
func GetGinClaims(c *gin.Context) *jwt.Claims {
	v, ok := c.Get(GinClaimsKey)
	if !ok {
		return nil
	}
	claims, _ := v.(*jwt.Claims)
	return claims
}
