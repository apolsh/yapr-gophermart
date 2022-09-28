package http

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/service"
	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type controller struct {
	gophermartService service.GophermartService
}

func RegisterRoutes(r *chi.Mux, s service.GophermartService) {
	c := &controller{gophermartService: s}

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	r.Use(middleware.Timeout(60 * time.Second))

	r.Route("/api/user", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/register", c.userRegisterHandler)
			r.Post("/login", c.userLoginHandler)
		})
		r.With(AuthMiddleware(s.ParseJWTToken)).Group(func(r chi.Router) {
			r.Route("/order", func(r chi.Router) {
				r.Post("/", c.createOrder)
				r.Get("/", c.getOrders)
			})
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", c.getBalance)
				r.Post("/withdraw", c.createWithdraw)
			})
			r.Get("/withdrawals", c.getWithdrawals)
		})
		r.Mount("/", newOrdersRoutes(s))
	})
}

func (c *controller) userRegisterHandler(w http.ResponseWriter, r *http.Request) {
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

	token, err := c.gophermartService.AddUser(r.Context(), req.Login, req.Password)
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

func (c *controller) userLoginHandler(w http.ResponseWriter, r *http.Request) {
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

	token, err := c.gophermartService.LoginUser(r.Context(), req.Login, req.Password)
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

func (c *controller) createOrder(w http.ResponseWriter, r *http.Request) {
	if !isValidContentType(r, "text/plain", "application/x-gzip") {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	reader, err := getBodyReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	num, err := strconv.Atoi(string(body))
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	err = c.gophermartService.AddOrder(r.Context(), num)
	if err != nil {
		if errors.Is(service.ErrorInvalidOrderNumberFormat, err) {
			http.Error(w, "", http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(storage.OrderAlreadyStored, err) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(storage.OrderAlreadyStoredByOtherUser, err) {
			http.Error(w, "", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (c *controller) getOrders(w http.ResponseWriter, r *http.Request) {
	//TODO: implement me
}

func (c *controller) getBalance(w http.ResponseWriter, r *http.Request) {
	//TODO: implement me
}

func (c *controller) createWithdraw(w http.ResponseWriter, r *http.Request) {
	//TODO: implement me
}

func (c *controller) getWithdrawals(w http.ResponseWriter, r *http.Request) {
	//TODO: implement me
}

func isValidContentType(r *http.Request, allowedTypes ...string) bool {
	actualContentType := r.Header.Get("Content-Type")

	for _, allowed := range allowedTypes {
		if strings.Contains(actualContentType, allowed) {
			return true
		}
	}
	return false
}

func getBodyReader(r *http.Request) (io.ReadCloser, error) {
	if r.Header.Get(`Content-Encoding`) == `gzip` {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return nil, err
		}
		return gz, nil
	}
	return r.Body, nil
}
