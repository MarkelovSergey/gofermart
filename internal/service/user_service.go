package service

import (
	"context"

	"github.com/MarkelovSergey/gofermart/internal/auth"
	"github.com/MarkelovSergey/gofermart/internal/storage"
)

type UserService interface {
	RegisterUser(ctx context.Context, login, password string) (string, error)
	LoginUser(ctx context.Context, login, password string) (string, error)
}

type userService struct {
	storage    storage.Storage
	jwtManager *auth.JWTManager
}

func NewUserService(storage storage.Storage, jwtManager *auth.JWTManager) *userService {
	return &userService{
		storage,
		jwtManager,
	}
}

func (s *userService) RegisterUser(ctx context.Context, login, password string) (string, error) {
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return "", err
	}

	user, err := s.storage.CreateUser(ctx, login, passwordHash)
	if err != nil {
		return "", err
	}

	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *userService) LoginUser(ctx context.Context, login, password string) (string, error) {
	user, err := s.storage.GetUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}

	if !auth.CheckPassword(password, user.PasswordHash) {
		return "", storage.ErrUserNotFound
	}

	token, err := s.jwtManager.GenerateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}
