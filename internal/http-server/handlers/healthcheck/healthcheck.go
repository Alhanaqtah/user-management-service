package healthcheck

import (
	"net/http"

	resp "user-managment-service/internal/lib/response"

	"github.com/go-chi/render"
)

func Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render.JSON(w, r, resp.Ok())
	}
}
