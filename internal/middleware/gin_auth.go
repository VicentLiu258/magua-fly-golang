package middleware

import (
	"net/http"
	"strings"

	"magua-fly-golang/internal/domain"
	"magua-fly-golang/internal/pkg/jwt"
	"magua-fly-golang/internal/repository"

	"github.com/gin-gonic/gin"
)

const GinClaimsKey = "claims"

// GinAuth 校验 JWT；非 admin 须为在职（employment_status=active）且未软删。
func GinAuth(jwtManager *jwt.Manager, users repository.UserRepository) gin.HandlerFunc {
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
		if domain.UserRole(claims.Role) != domain.RoleAdmin {
			u, err := users.GetByID(c.Request.Context(), claims.UserID)
			if err != nil || u.EmploymentStatus != domain.EmploymentActive {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "employment inactive"})
				return
			}
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
