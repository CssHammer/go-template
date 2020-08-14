package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/CssHammer/go-template/models"
)

func (s *DefaultService) GetUser(ctx context.Context, id int) (*models.User, error) {
	if id == 0 {
		return nil, ErrNotValidRequest{Reason: "id can not be 0"}
	}

	// try read from cache
	userJson, err := s.cache.Get(ctx, fmt.Sprint(id))
	if err != nil {
		return nil, fmt.Errorf("cache get (id: %d): %w", id, err)
	}

	if len(userJson) > 0 {
		var user models.User
		err = json.Unmarshal([]byte(userJson), &user)
		if err != nil {
			return nil, fmt.Errorf("unmarshal cached user (json: %s): %w", userJson, err)
		}

		return &user, nil
	}

	// read from storage
	user, err := s.storage.GetUser(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("storage: get user (id: %d): %w", id, err)
	}

	// cache data
	userBytes, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("marshal user (user: %v): %w", user, err)
	}

	ttl := 300 * time.Second
	err = s.cache.Set(ctx, fmt.Sprint(id), string(userBytes), ttl)
	if err != nil {
		return nil, fmt.Errorf("cache set (id: %d, userJson: %s, ttl: %s): %w", id, userJson, ttl.String(), err)
	}

	return user, nil
}

func (s *DefaultService) CreateUser(ctx context.Context, user models.User) error {
	if user.ID != 0 {
		return ErrNotValidRequest{Reason: "id must be 0"}
	}

	if user.Name == "" {
		return ErrNotValidRequest{Reason: "name is empty"}
	}

	// write to storage
	err := s.storage.CreateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("storage: create user (user: %v): %w", user, err)
	}

	return nil
}
