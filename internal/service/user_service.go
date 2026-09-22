package service

import (
	"context"
	"errors"
	"go-postgres-api/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(ctx context.Context, id int64) (*repository.User, error) {
	if id <= 0 {
		return nil, errors.New("id must be greater than 0")
	}
	return s.repo.FindByID(ctx, id)
}

func (s *UserService) CreateUser(ctx context.Context, name, email string) (*repository.User, error) {
	if name == "" || email == "" {
		return nil, errors.New("name and email cannot be empty")
	}
	return s.repo.Create(ctx, name, email)
}