package middleware

import (
	"github.com/gin-gonic/gin"
)

// GinCORS adds permissive CORS headers for development.
func GinCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// When the browser sends credentials (cookies/authorization), using "*" for
		// Access-Control-Allow-Origin will be rejected by the browser.
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		// Allow browser to include credentials (e.g. cookies) in cross-origin requests.
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// Mirror preflight requested headers to avoid failing on custom headers.
		// Example: Access-Control-Request-Headers: content-type, authorization, x-requested-with
		if acrh := c.Request.Header.Get("Access-Control-Request-Headers"); acrh != "" {
			c.Header("Access-Control-Allow-Headers", acrh)
		} else {
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
