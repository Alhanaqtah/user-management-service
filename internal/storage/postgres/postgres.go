package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"user-management-service/internal/config"
	"user-management-service/internal/models"
	"user-management-service/internal/storage"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(cfg config.Storage) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := pgxpool.New(context.Background(), fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=%s",
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

	err = db.Ping(context.Background())
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	db.Exec(context.TODO(), `
	CREATE TYPE ROLE AS ENUM ('user', 'moderator', 'admin');
	
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) DEFAULT '',
		surname VARCHAR(255) DEFAULT '',
		username VARCHAR(255) NOT NULL,
		pass_hash BYTEA NOT NULL,
		phone_number VARCHAR(255) DEFAULT '',
		email VARCHAR(255) NOT NULL,
		role ROLE DEFAULT 'user' CHECK (role IN ('user', 'moderator', 'admin')),
		image_s3_path VARCHAR(255) DEFAULT '',
		is_blocked BOOLEAN DEFAULT false,
		created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		modified_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS groups (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users_groups (
		id BISERIAL PRIMARY KEY,
		user_id UUID NOT NULL,
		group_id UUID NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id),
		FOREIGN KEY (group_id) REFERENCES groups(id)
	);

	CREATE OR REPLACE FUNCTION update_modified_at()
	RETURNS TRIGGER AS $$
	BEGIN
		NEW.modified_at = CURRENT_TIMESTAMP;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	CREATE TRIGGER update_users_modified_at
	BEFORE UPDATE ON users
	FOR EACH ROW
	EXECUTE FUNCTION update_modified_at();
	`)

	return &Storage{db: db}, nil
}

func (s *Storage) UserByUUID(ctx context.Context, uuid string) (*models.User, error) {
	const op = "storage.postgres.UserByUUID"

	var user models.User

	// Fetch the user information
	userRow := s.db.QueryRow(ctx, `
        SELECT
            u.name,
            u.surname,
            u.username,
            u.phone_number,
            u.email,
            u.role,
            u.is_blocked,
            u.created_at,
            u.modified_at
        FROM
            users u
        WHERE
            u.id = $1`, uuid,
	)

	err := userRow.Scan(
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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Fetch the groups for the user
	groupRows, err := s.db.Query(ctx, `
        SELECT
            g.name
        FROM
            groups g
        JOIN
            users_groups ug ON g.id = ug.group_id
        WHERE
            ug.user_id = $1`, uuid,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer groupRows.Close()

	var groupNames []string
	for groupRows.Next() {
		var groupName string
		if err := groupRows.Scan(&groupName); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		groupNames = append(groupNames, groupName)
	}
	if err := groupRows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Assign the groups to the user
	user.Groups = groupNames

	// Output fields name, surname, phone_number, role, groups
	fmt.Printf("User: %v", user)

	return &user, nil
}

func (s *Storage) UserByName(ctx context.Context, username string) (*models.User, error) {
	const op = "storage.postgres.UserByName"

	row := s.db.QueryRow(ctx, `
		SELECT
			id,
			username,
			pass_hash,
			role
		FROM users WHERE username=$1`, username,
	)

	var user models.User
	err := row.Scan(
		&user.UUID,
		&user.Username,
		&user.PassHash,
		&user.Role,
	)
	if err != nil {
		if errors.As(err, &pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil
}

func (s *Storage) SearchEmail(ctx context.Context, email string) (bool, error) {
	const op = "storage.postgres.SearchEmail"

	var found bool
	err := s.db.QueryRow(ctx, `SELECT EXISTS(SELECT username FROM users WHERE email=$1)`, email).Scan(&found)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return found, nil
}

func (s *Storage) CreateNewUser(ctx context.Context, username string, email string, passHash []byte) error {
	const op = "storage.postgres.CreateNewUser"

	_, err := s.db.Exec(ctx, `INSERT INTO users (username, email, pass_hash) VALUES ($1, $2, $3)`, username, email, passHash)
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
	err := s.db.QueryRow(ctx, query, args...).Scan(&u.Name, &u.Surname, &u.Username, &u.PhoneNumber, &u.Email, &u.Role, &u.IsBlocked, &u.CreatedAt, &u.ModifiedAt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &u, nil
}

func (s *Storage) Delete(ctx context.Context, uuid string) error {
	const op = "storage.postgres.Delete"

	_, err := s.db.Exec(ctx, `DELETE FROM users WHERE id=$1`, uuid)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
