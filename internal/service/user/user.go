package user

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"user-management-service/internal/models"
	"user-management-service/internal/storage"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrNoFieldsToUpdate = errors.New("no fields to update")
)

type Storage interface {
	UserByUUID(ctx context.Context, uuid string) (*models.User, error)
	PatchUser(ctx context.Context, uuid string, user *models.User) (*models.User, error)
	Delete(ctx context.Context, uuid string) error
}
type Service struct {
	log     *slog.Logger
	storage Storage
}

func New(log *slog.Logger, storage Storage) *Service {
	return &Service{
		log:     log,
		storage: storage,
	}
}

func (s *Service) UserByUUID(uuid string) (*models.User, error) {
	const op = "service.user.UserByUUID"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user, err := s.storage.UserByUUID(ctx, uuid)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Service) PatchUser(uuid string, user *models.User) (*models.User, error) {
	const op = "service.user.PatchUser"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	info, err := s.storage.PatchUser(ctx, uuid, user)
	if err != nil {
		if errors.Is(err, storage.ErrNoFieldsToUpdate) {
			return nil, fmt.Errorf("%s: %w", op, ErrNoFieldsToUpdate)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return info, nil
}

func (s *Service) Delete(uuid string) error {
	const op = "service.user.Delete"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := s.storage.Delete(ctx, uuid)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
