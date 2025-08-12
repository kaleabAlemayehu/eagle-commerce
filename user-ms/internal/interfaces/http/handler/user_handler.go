package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/application/dto"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

type UserHandler struct {
	userService domain.UserService
}

func NewUserHandler(userService domain.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// @Summary Create a new user
// @Description Create a new user with email and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User data"
// @Success 201 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Router /signup [post]
func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user := &domain.User{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	newUser, token, err := h.userService.RegisterUser(r.Context(), user)
	if err != nil {
		if validationErrors := utils.GetValidationErrors(err); len(validationErrors) > 0 {
			utils.SendValidationErrorResponse(w, validationErrors)
			return
		}
		utils.SendErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	registerRes := dto.AuthResponse{
		User:  h.toUserResponse(newUser),
		Token: token,
	}
	utils.SendSuccessResponse(w, http.StatusCreated, registerRes)
}

// @Summary Get user by ID
// @Description Get user details by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	user, err := h.userService.GetUser(r.Context(), id)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	userRes := h.toUserResponse(user)
	utils.SendSuccessResponse(w, http.StatusOK, userRes)
}

// @Summary List users
// @Description Get paginated list of users
// @Tags users
// @Produce json
// @Param limit query int false "Limit" default(10)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} dto.Response
// @Router /users [get]
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 10
	}

	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	users, err := h.userService.ListUsers(r.Context(), limit, offset)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SendSuccessResponse(w, http.StatusOK, users)
}

// @Summary Put user by ID
// @Description Update user details by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 201 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	user := &domain.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Address: domain.Address{
			Street:  req.Address.Street,
			City:    req.Address.City,
			State:   req.Address.State,
			ZipCode: req.Address.ZipCode,
			Country: req.Address.Country,
		},
		UpdatedAt: time.Now(),
	}

	updatedUser, err := h.userService.UpdateUser(r.Context(), id, user)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}

	userRes := h.toUserResponse(updatedUser)
	utils.SendSuccessResponse(w, http.StatusOK, userRes)
}

// @Summary Delete user by ID
// @Description Delete user details by ID
// @Tags users
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} dto.Response
// @Failure 404 {object} dto.Response
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.userService.DeleteUser(r.Context(), id); err != nil {
		utils.SendErrorResponse(w, http.StatusNotFound, "User not found")
		return
	}
	utils.SendSuccessResponse(w, http.StatusOK, "User Successfully Deleted")
}

// @Summary Login user
// @Description Login user
// @Tags users
// @Produce json
// @Success 200 {object} dto.Response
// @Failure 400 {object} dto.Response
// @Failure 401 {object} dto.Response
// @Router /users/login [post]
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.SendErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	user, token, err := h.userService.AuthenticateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		utils.SendErrorResponse(w, http.StatusUnauthorized, "Unauthorized user")
		return
	}

	loginRes := dto.AuthResponse{
		User: h.toUserResponse(user), Token: token,
	}
	utils.SendSuccessResponse(w, http.StatusOK, loginRes)
}

func (h *UserHandler) toUserResponse(u *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        u.ID.String(),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Address: &dto.AddressDTO{
			Street:  u.Address.Street,
			City:    u.Address.City,
			State:   u.Address.State,
			ZipCode: u.Address.ZipCode,
			Country: u.Address.Country,
		},
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
