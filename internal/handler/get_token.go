package handler

import (
	"geoservice/internal/entity"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-chi/jwtauth/v5"
)

// CreationTokenHandler godoc
// @Summary Создание токена для авторизации
// @Tags get token
// @Accept json
// @Produce json
// @Success 200 {object} entity.CreationTokenResponse "OK"
// @Failure 500 {object} entity.ErrorResponse "Internal Server Error"
// @Router /creation-token [get]
func CreationTokenHandler(tokenAuth *jwtauth.JWTAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := generateToken(tokenAuth)
		if err != nil {
			c.JSON(http.StatusInternalServerError, entity.ErrorResponse{
				Error:   "Не удалось сгенерировать токен",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, entity.CreationTokenResponse{
			Token:     token,
			CreatedAt: time.Now(),
		})
	}
}

// generateToken генерирует JWT токен для авторизации
func generateToken(tokenAuth *jwtauth.JWTAuth) (string, error) {
	claims := map[string]interface{}{
		"name": "admin",
		"exp":  time.Now().Add(240 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	_, tokenString, err := tokenAuth.Encode(claims)
	return tokenString, err
}
