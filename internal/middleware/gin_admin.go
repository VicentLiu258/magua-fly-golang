package middleware

import (
	"net/http"

	"magua-fly-golang/internal/domain"

	"github.com/gin-gonic/gin"
)

// GinRequireAdmin 需在 GinAuth 之后注册，仅允许 role=admin。
func GinRequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetGinClaims(c)
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		if domain.UserRole(claims.Role) != domain.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}
