package http_handle

import (
	"net/http"
	"strconv"

	"geoservice/internal/entity"
	"geoservice/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// CreateUser godoc
// @Summary Добавить нового пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param user body entity.CreateUserRequest true "Введите значения в поля: Email, Name, Age, Location"
// @Success 201 {object} entity.UserResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 409 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	var req entity.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user with this email already exists" {
			statusCode = http.StatusConflict
		} else if err.Error()[:10] == "validation" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, entity.ErrorResponse{
			Error:   "Failed to create user",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUser godoc
// @Summary Получить пользователя по ID
// @Tags users
// @Produce json
// @Param id path string true "Введите ID пользователя"
// @Success 200 {object} entity.UserResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: "User ID is required",
		})
		return
	}

	// Get user
	user, err := h.userService.GetUser(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, entity.ErrorResponse{
			Error:   "Failed to get user",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser godoc
// @Summary Обновить информацию о пользователе
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "Введите ID пользователя"
// @Param user body entity.UpdateUserRequest true "Редактируйте значения полей: Email, Name, Age, Location"
// @Success 200 {object} entity.UserResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 409 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: "User ID is required",
		})
		return
	}

	var req entity.UpdateUserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error:   "Invalid request body",
			Message: err.Error(),
		})
		return
	}

	// Update user
	user, err := h.userService.UpdateUser(c.Request.Context(), id, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		} else if err.Error() == "user with this email already exists" {
			statusCode = http.StatusConflict
		} else if err.Error()[:10] == "validation" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, entity.ErrorResponse{
			Error:   "Failed to update user",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary Удалить пользователя по ID
// @Tags users
// @Produce json
// @Param id path string true "Введите ID пользователя"
// @Success 204 "No Content"
// @Failure 400 {object} entity.ErrorResponse
// @Failure 404 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /users/{id}/delete [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, entity.ErrorResponse{
			Error: "User ID is required",
		})
		return
	}

	// Delete user
	err := h.userService.DeleteUser(c.Request.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, entity.ErrorResponse{
			Error:   "Failed to delete user",
			Message: err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// ListUsers godoc
// @Summary Получить список пользователей с выборкой по количеству и параметру
// @Tags users
// @Produce json
// @Param limit query int false "Введите количество пользователей (default: 10, max: 100)" minimum(1) maximum(100)
// @Param offset query int false "Введите порядковый номер пользователя с которого начать выборку (default: 0)" minimum(0)
// @Param search query string false "Введите параметр для поиска"
// @Success 200 {object} entity.UserListResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var conditions entity.Conditions

	// Parse query parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			conditions.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			conditions.Offset = offset
		}
	}

	conditions.Search = c.Query("search")

	// List users
	response, err := h.userService.ListUsers(c.Request.Context(), conditions)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error()[:10] == "validation" {
			statusCode = http.StatusBadRequest
		}

		c.JSON(statusCode, entity.ErrorResponse{
			Error:   "Failed to list users",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
