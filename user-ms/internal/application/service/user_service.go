package service

import (
	"errors"

	argon "github.com/alexedwards/argon2id"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/models"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

type UserServiceImpl struct {
	repo       domain.UserRepository
	natsClient *messaging.NATSClient
}

func NewUserService(repo domain.UserRepository, natsClient *messaging.NATSClient) domain.UserService {
	return &UserServiceImpl{
		repo:       repo,
		natsClient: natsClient,
	}
}

func (s *UserServiceImpl) CreateUser(user *domain.User) error {
	// Validate user data
	if err := utils.ValidateStruct(user); err != nil {
		return err
	}

	// Check if user already exists
	existingUser, _ := s.repo.GetByEmail(user.Email)
	if existingUser != nil {
		return errors.New("user already exists")
	}

	// Hash password
	hashedPassword, err := argon.CreateHash(user.Password, argon.DefaultParams)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// Create user
	if err := s.repo.Create(user); err != nil {
		return err
	}

	// Publish event
	event := models.Event{
		Type:   models.UserCreatedEvent,
		Source: "user-service",
		Data: map[string]interface{}{
			"user_id": user.ID.Hex(),
			"email":   user.Email,
		},
	}
	s.natsClient.Publish("user.created", event)

	return nil
}

func (s *UserServiceImpl) GetUser(id string) (*domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserServiceImpl) UpdateUser(id string, user *domain.User) error {
	if err := utils.ValidateStruct(user); err != nil {
		return err
	}

	return s.repo.Update(id, user)
}

func (s *UserServiceImpl) DeleteUser(id string) error {
	return s.repo.Delete(id)
}

func (s *UserServiceImpl) ListUsers(limit, offset int) ([]*domain.User, error) {
	return s.repo.List(limit, offset)
}

func (s *UserServiceImpl) AuthenticateUser(email, password string) (*domain.User, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	match, err := argon.ComparePasswordAndHash(user.Password, password)
	if err != nil || !match {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}
