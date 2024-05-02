package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"user-managment-service/internal/config"
	"user-managment-service/internal/models"
	"user-managment-service/internal/storage"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	pool *pgx.Conn
}

func New(cfg config.Storage) (*Storage, error) {
	const op = "storage.postgres.New"

	pool, err := pgx.Connect(context.Background(), fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
		cfg.SSLMode,
	))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) UserByUUID(ctx context.Context, username string) (*models.User, error) {
	const op = "storage.postgres.UserByUUID"

	row := s.pool.QueryRow(ctx, `
		SELECT
			name,
			surname,
			username,
			phone_number,
			email,
			role,
			is_blocked,
			created_at,
			modified_at
		FROM users WHERE id=$1`, username,
	)

	var groupID sql.NullInt64
	var user models.User
	err := row.Scan(
		&user.Name,
		&user.Surname,
		&user.Username,
		&user.PhoneNumber,
		&user.Email,
		&user.Role,
		&user.IsBlocked,
		&user.CreatedAt,
		&user.ModifiedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if groupID.Valid {
		user.GroupID = groupID.Int64
	}

	return &user, nil
}

func (s *Storage) UserByName(ctx context.Context, username string) (*models.User, error) {
	const op = "storage.postgres.UserByName"

	row := s.pool.QueryRow(ctx, `
		SELECT
			id,
			name,
			surname,
			username,
			pass_hash,
			phone_number,
			email,
			role,
			group_id,
			image_s3_path,
			is_blocked,
			created_at,
			modified_at
		FROM users WHERE username=$1`, username,
	)

	var groupID sql.NullInt64
	var user models.User
	err := row.Scan(
		&user.UUID,
		&user.Name,
		&user.Surname,
		&user.Username,
		&user.PassHash,
		&user.PhoneNumber,
		&user.Email,
		&user.Role,
		&groupID,
		&user.ImageS3Path,
		&user.IsBlocked,
		&user.CreatedAt,
		&user.ModifiedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if groupID.Valid {
		user.GroupID = groupID.Int64
	}

	return &user, nil
}

func (s *Storage) SearchEmail(ctx context.Context, email string) (bool, error) {
	const op = "storage.postgres.SearchEmail"

	var found bool
	err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT username FROM users WHERE email=$1)`, email).Scan(&found)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return found, nil
}

func (s *Storage) CreateNewUser(ctx context.Context, username string, email string, passHash []byte) error {
	const op = "storage.postgres.CreateNewUser"

	_, err := s.pool.Exec(ctx, `INSERT INTO users (username, email, pass_hash) VALUES ($1, $2, $3)`, username, email, passHash)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) PatchUser(ctx context.Context, uuid string, user *models.User) (*models.User, error) {
	const op = "storage.postgres.PatchUser"

	var args []interface{}
	var attrs []string
	if user.Name != "" {
		args = append(args, user.Name)
		attrs = append(attrs, "name=$")
	}
	if user.Surname != "" {
		args = append(args, user.Surname)
		attrs = append(attrs, "surname=$")
	}
	if user.Username != "" {
		args = append(args, user.Username)
		attrs = append(attrs, "username=$")
	}
	if user.PhoneNumber != "" {
		args = append(args, user.PhoneNumber)
		attrs = append(attrs, "phone_number=$")
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrNoFieldsToUpdate)
	}

	for i := range attrs {
		attrs[i] += strconv.Itoa(i + 1)
	}

	args = append(args, time.Now().UTC().Format(time.RFC3339), uuid)

	queryAttrs := strings.Join(attrs, ", ")

	query := "UPDATE users SET " + queryAttrs + ", modified_at=$" + strconv.Itoa(len(attrs)+1) + " WHERE id=$" + strconv.Itoa(len(attrs)+2) + " RETURNING name, surname, username, phone_number, email, role, is_blocked, created_at, modified_at"

	var u models.User
	err := s.pool.QueryRow(ctx, query, args...).Scan(&u.Name, &u.Surname, &u.Username, &u.PhoneNumber, &u.Email, &u.Role, &u.IsBlocked, &u.CreatedAt, &u.ModifiedAt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &u, nil
}

func (s *Storage) Delete(ctx context.Context, uuid string) error {
	const op = "storage.postgres.CreateNewUser"

	_, err := s.pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, uuid)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetByModerator(ctx context.Context, moderatorID, userID int64) (*models.User, error) {
	const op = "storage.postgres.GetByModerator"

	var user models.User
	err := s.pool.QueryRow(ctx, ``, moderatorID, userID).Scan(&user)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}
