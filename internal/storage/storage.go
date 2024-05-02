package storage

import "errors"

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserExists       = errors.New("user already exists")
	ErrEmailNotFound    = errors.New("email not found")
	ErrNoFieldsToUpdate = errors.New("no fields to update")
)
