package storage

import (
	"context"

	"github.com/CssHammer/go-template/models"
)

type Storage interface {
	HealthCheck() error
	GetUser(ctx context.Context, id int) (*models.User, error)
	CreateUser(ctx context.Context, user models.User) error
}
