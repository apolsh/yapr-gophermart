package http

import (
	"encoding/json"
	"errors"
	"fmt"
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
			router.Post("/login", r.userLoginHandler)
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

	token, err := u.gophermartService.AddUser(r.Context(), req.Login, req.Password)
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

	w.Header().Add("Authorization", token)
	http.SetCookie(w, &http.Cookie{Name: "Authorization", Value: fmt.Sprintf("Bearer %s", token)})
	w.WriteHeader(http.StatusOK)
}

func (u *userRoutes) userLoginHandler(w http.ResponseWriter, r *http.Request) {
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

	token, err := u.gophermartService.LoginUser(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(storage.ItemNotFound, err) || errors.Is(service.ErrorInvalidPassword, err) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Authorization", token)
	http.SetCookie(w, &http.Cookie{Name: "Authorization", Value: fmt.Sprintf("Bearer %s", token)})
	w.WriteHeader(http.StatusOK)

}
