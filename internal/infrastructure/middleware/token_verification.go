package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-chi/jwtauth/v5"
	"net/http"
	"strings"
)

func JWTVerifier(tokenAuth *jwtauth.JWTAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		// проверяем заголовок
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		// проверяем на шаблон: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid Authorization header"})
			return
		}
		rawToken := parts[1]

		// проверяем токен
		token, err := jwtauth.VerifyToken(tokenAuth, rawToken)
		if err != nil || token == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("jwt_claims", token)
		c.Next()
	}
}
