package rpc_handle

import (
	"geoservice/internal/entity"
	"geoservice/internal/service"
)

// UserRPCHandler обрабатывает RPC-запросы для пользователей
type UserRPCHandler struct {
	userService service.UserService
}

func NewUserRPCHandler(userService service.UserService) *UserRPCHandler {
	return &UserRPCHandler{userService: userService}
}

// RPC-запросы/ответы

type CreateUserRequest = entity.CreateUserRequest
type CreateUserResponse = entity.UserResponse

type GetUserRequest struct {
	ID string
}
type GetUserResponse = entity.UserResponse

type UpdateUserRequest struct {
	ID   string
	Data entity.UpdateUserRequest
}
type UpdateUserResponse = entity.UserResponse

type DeleteUserRequest struct {
	ID string
}
type DeleteUserResponse struct{}

type ListUsersRequest struct {
	Search string `form:"search"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}
type ListUsersResponse = entity.UserListResponse

// RPC методы

func (h *UserRPCHandler) CreateUser(req CreateUserRequest, resp *CreateUserResponse) error {
	user, err := h.userService.CreateUser(nil, req)
	if err != nil {
		return err
	}
	*resp = *user
	return nil
}

func (h *UserRPCHandler) GetUser(req GetUserRequest, resp *CreateUserResponse) error {
	user, err := h.userService.GetUser(nil, req.ID)
	if err != nil {
		return err
	}
	*resp = *user
	return nil
}

func (h *UserRPCHandler) UpdateUser(req UpdateUserRequest, resp *UpdateUserResponse) error {
	user, err := h.userService.UpdateUser(nil, req.ID, req.Data)
	if err != nil {
		return err
	}
	*resp = *user
	return nil
}

func (h *UserRPCHandler) DeleteUser(req DeleteUserRequest, resp *DeleteUserResponse) error {
	if err := h.userService.DeleteUser(nil, req.ID); err != nil {
		return err
	}
	*resp = DeleteUserResponse{}
	return nil
}

func (h *UserRPCHandler) ListUsers(req ListUsersRequest, resp *ListUsersResponse) error {
	cond := entity.Conditions{
		Limit:  req.Limit,
		Offset: req.Offset,
		Search: req.Search,
	}
	users, err := h.userService.ListUsers(nil, cond)
	if err != nil {
		return err
	}
	*resp = *users
	return nil
}
