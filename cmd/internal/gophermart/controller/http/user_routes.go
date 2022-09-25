package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/service"
	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/storage"
	"github.com/go-chi/chi/v5"
)

type userRoutes struct {
	gophermartService service.GophermartService
}

func newUserRoutes(router *chi.Mux, s service.GophermartService) {
	r := &userRoutes{gophermartService: s}

	router.Route("/api", func(router chi.Router) {
		router.Route("/user", func(router chi.Router) {
			router.Post("/register", r.userRegisterHandler)
		})
	})
}

func (u *userRoutes) userRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if !isValidContentType(r, "application/json", "application/x-gzip") {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	reader, err := getBodyReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	req := &AuthRequest{}

	if err = json.NewDecoder(reader).Decode(&req); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	err = u.gophermartService.AddUser(ctx, req.Login, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrorEmptyValue) {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		if errors.Is(err, storage.ErrorLoginIsAlreadyUsed) {
			http.Error(w, "", http.StatusConflict)
			return
		}
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
