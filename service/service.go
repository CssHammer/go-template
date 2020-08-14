package service

import (
	"context"

	"github.com/CssHammer/go-template/cache"
	"github.com/CssHammer/go-template/models"
	"github.com/CssHammer/go-template/storage"
)

type Service interface {
	GetUser(ctx context.Context, id int) (*models.User, error)
	CreateUser(ctx context.Context, user models.User) error
}

type DefaultService struct {
	storage storage.Storage
	cache   cache.Cache
}

func New(storage storage.Storage, cache cache.Cache) *DefaultService {
	return &DefaultService{
		storage: storage,
		cache:   cache,
	}
}
