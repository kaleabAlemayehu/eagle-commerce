package service

import (
	"errors"

	argon "github.com/alexedwards/argon2id"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/infrastructure/messaging"
	sharedMiddlware "github.com/kaleabAlemayehu/eagle-commerce/shared/middleware"
	"github.com/kaleabAlemayehu/eagle-commerce/shared/utils"
)

type UserServiceImpl struct {
	repo domain.UserRepository
	nats *messaging.UserEventPublisher
	auth *sharedMiddlware.Auth
}

func NewUserService(repo domain.UserRepository, nats *messaging.UserEventPublisher, auth *sharedMiddlware.Auth) domain.UserService {
	return &UserServiceImpl{
		repo: repo,
		nats: nats,
		auth: auth,
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
	return s.nats.PublishUserCreated(user)
}

func (s *UserServiceImpl) GetUser(id string) (*domain.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserServiceImpl) UpdateUser(id string, user *domain.User) error {
	if err := utils.ValidateStruct(user); err != nil {
		return err
	}

	if err := s.repo.Update(id, user); err != nil {
		return err
	}
	return s.nats.PublishUserUpdated(user)
}

func (s *UserServiceImpl) DeleteUser(id string) error {
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	return s.nats.PublishUserDeleted(id)
}

func (s *UserServiceImpl) ListUsers(limit, offset int) ([]*domain.User, error) {
	return s.repo.List(limit, offset)
}

func (s *UserServiceImpl) AuthenticateUser(email, password string) (*domain.User, string, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}
	match, err := argon.ComparePasswordAndHash(user.Password, password)
	if err != nil || !match {
		return nil, "", errors.New("invalid credentials")
	}
	token, err := s.auth.GenerateJWT(user.ID.String(), user.Email)
	if err != nil {
		return nil, "", errors.New("unable to generate JWT")
	}

	return user, token, nil
}
