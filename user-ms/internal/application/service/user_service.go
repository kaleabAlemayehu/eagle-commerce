package service

import (
	"context"
	"errors"

	argon "github.com/alexedwards/argon2id"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/domain"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/infrastructure/messaging"
	"github.com/kaleabAlemayehu/eagle-commerce/services/user-ms/internal/infrastructure/repository"
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

func (s *UserServiceImpl) CreateUser(ctx context.Context, user *domain.User) error {
	// Validate user data
	if err := utils.ValidateStruct(user); err != nil {
		return err
	}

	// Check if user already exists
	existingUser, err := s.repo.GetByEmail(ctx, user.Email)
	if err != nil {
		if !errors.Is(err, repository.ErrorUserNotFound) {
			return err
		}
	}
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
	if err := s.repo.Create(ctx, user); err != nil {
		return err
	}

	// Publish event
	return s.nats.PublishUserCreated(user)
}

func (s *UserServiceImpl) GetUser(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *UserServiceImpl) UpdateUser(ctx context.Context, id string, user *domain.User) (*domain.User, error) {
	if err := utils.ValidateStruct(user); err != nil {
		return nil, err
	}
	updatedUser, err := s.repo.Update(ctx, id, user)
	if err != nil {
		return nil, err
	}
	return updatedUser, s.nats.PublishUserUpdated(updatedUser)
}

func (s *UserServiceImpl) DeleteUser(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	return s.nats.PublishUserDeleted(id)
}

func (s *UserServiceImpl) ListUsers(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *UserServiceImpl) AuthenticateUser(ctx context.Context, email, password string) (*domain.User, string, error) {
	user, err := s.repo.GetByEmail(ctx, email)
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
