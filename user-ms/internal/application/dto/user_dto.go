package dto

import "time"

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
}

type UpdateUserRequest struct {
	FirstName string      `json:"first_name,omitempty"`
	LastName  string      `json:"last_name,omitempty"`
	Address   *AddressDTO `json:"address,omitempty"`
}

type UserResponse struct {
	ID        string      `json:"id"`
	Email     string      `json:"email"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	Address   *AddressDTO `json:"address,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type AddressDTO struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	ZipCode string `json:"zip_code"`
	Country string `json:"country"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponse struct {
	User  *UserResponse `json:"user"`
	Token string        `json:"token"`
}

type UserListResponse struct {
	Users []UserResponse `json:"users"`
	Total int            `json:"total"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}
