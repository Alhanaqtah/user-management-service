package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"user-management-service/internal/config"
	"user-management-service/internal/lib/jwt"
	"user-management-service/internal/models"
	"user-management-service/internal/storage"

	gojwt "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound  = errors.New("user not found")
	ErrUserExists    = errors.New("user already exists")
	ErrTokenRevoked  = errors.New("token revoked")
	ErrEmailNotFound = errors.New("email not found")
)

type Storage interface {
	UserByName(ctx context.Context, username string) (*models.User, error)
	UserByUUID(ctx context.Context, uuid string) (*models.User, error)
	SearchEmail(ctx context.Context, email string) (bool, error)
	CreateNewUser(ctx context.Context, username, email string, passHash []byte) error
}

type Cash interface {
	AddToBlaclist(ctx context.Context, token string) error
	SearchInBlacklist(ctx context.Context, token string) (bool, error)
}

type Broker interface {
	ResetPassword(ctx context.Context, email string) error
}

type Service struct {
	log      *slog.Logger
	storage  Storage
	cash     Cash
	broker   Broker
	tokenCfg config.Token
}

func New(log *slog.Logger, storage Storage, cash Cash, broker Broker, token config.Token) *Service {
	return &Service{
		log:      log,
		storage:  storage,
		cash:     cash,
		broker:   broker,
		tokenCfg: token,
	}
}

func (s *Service) SignUp(username, email, password string) error {
	const op = "service.auth.SignUp"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	u, err := s.storage.UserByName(ctx, username)
	if err != nil && !errors.Is(err, storage.ErrUserNotFound) {
		return fmt.Errorf("%s: %w", op, err)
	}
	if u != nil {
		return fmt.Errorf("%s: %w", op, storage.ErrUserExists)
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.storage.CreateNewUser(ctx, username, email, passHash)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Service) Login(username, password string) (string, string, error) {
	const op = "service.auth.Login"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	user, err := s.storage.UserByName(ctx, username)
	if err != nil {
		if errors.As(err, &storage.ErrUserNotFound) {
			return "", "", fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	accessToken, err := jwt.NewAccessToken(user, s.tokenCfg)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	refreshToken, err := jwt.NewRefreshToken(user, s.tokenCfg)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, refreshToken, nil
}

func (s *Service) RefreshToken(token string) (string, string, error) {
	const op = "service.auth.RefreshToken"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Search token in blacklist
	found, err := s.cash.SearchInBlacklist(ctx, token)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	if found {
		return "", "", fmt.Errorf("%s: %w", op, ErrTokenRevoked)
	}

	// Add old token to blacklist
	err = s.cash.AddToBlaclist(ctx, token)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	// Parse refresh token to get it's claims
	jwtToken, err := gojwt.Parse(token, func(t *gojwt.Token) (interface{}, error) {
		return []byte(s.tokenCfg.JWT.Secret), nil
	})
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	claims := jwtToken.Claims.(gojwt.MapClaims)

	// Get UUID from token's claims
	uuid, err := jwt.GetClaim(claims, "sub")
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	// Get user info to form tokens
	user, err := s.storage.UserByUUID(ctx, uuid)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	// Form new access token
	accessToken, err := jwt.NewAccessToken(user, s.tokenCfg)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	// Form new refresh token
	refreshToken, err := jwt.NewRefreshToken(user, s.tokenCfg)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	return accessToken, refreshToken, nil
}

func (s *Service) ResetPassword(email string) error {
	const op = "service.auth.ResetPassword"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	found, err := s.storage.SearchEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	if !found {
		return fmt.Errorf("%s: %w", op, storage.ErrEmailNotFound)
	}

	err = s.broker.ResetPassword(ctx, email)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
