package users

import "log/slog"

type Service interface {
}

type handler struct {
	log     slog.Logger
	service Service
}
