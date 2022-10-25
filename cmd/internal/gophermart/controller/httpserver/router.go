package httpserver

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

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/entity/dto"
	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/service"
	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type controller struct {
	gophermartService service.GophermartService
}

const (
	applicationJsonContentType  = "application/json"
	textPlainContentType        = "text/plain"
	applicationXGzipContentType = "application/x-gzip"
)

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
			r.Route("/orders", func(r chi.Router) {
				r.Post("/", c.createOrder)
				r.Get("/", c.getOrders)
			})
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", c.getBalance)
				r.Post("/withdraw", c.createWithdraw)
			})
			r.Get("/withdrawals", c.getWithdrawals)
		})
	})
}

func (c *controller) userRegisterHandler(w http.ResponseWriter, r *http.Request) {
	if !isValidContentType(r, applicationJsonContentType, applicationXGzipContentType) {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	req := &AuthRequest{}
	err := extractJSONBody(r, &req)
	if err != nil {
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

	w.Header().Add(authorizationHeaderKey, token)
	http.SetCookie(w, &http.Cookie{Name: authorizationHeaderKey, Value: fmt.Sprintf("Bearer %s", token)})
	w.WriteHeader(http.StatusOK)
}

func (c *controller) userLoginHandler(w http.ResponseWriter, r *http.Request) {

	if !isValidContentType(r, applicationJsonContentType, applicationXGzipContentType) {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	req := &AuthRequest{}
	err := extractJSONBody(r, &req)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	token, err := c.gophermartService.LoginUser(r.Context(), req.Login, req.Password)
	if err != nil {
		if errors.Is(storage.ErrItemNotFound, err) || errors.Is(service.ErrorInvalidPassword, err) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add(authorizationHeaderKey, token)
	http.SetCookie(w, &http.Cookie{Name: authorizationHeaderKey, Value: fmt.Sprintf("Bearer %s", token)})
	w.WriteHeader(http.StatusOK)
}

func (c *controller) createOrder(w http.ResponseWriter, r *http.Request) {
	if !isValidContentType(r, textPlainContentType, applicationXGzipContentType) {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	var orderNum int
	err := extractJSONBody(r, &orderNum)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(UserID).(string)

	err = c.gophermartService.AddOrder(r.Context(), strconv.Itoa(orderNum), userID)
	if err != nil {
		if errors.Is(service.ErrorInvalidOrderNumberFormat, err) {
			http.Error(w, "", http.StatusUnprocessableEntity)
			return
		}
		if errors.Is(storage.ErrOrderAlreadyStored, err) {
			w.WriteHeader(http.StatusOK)
			return
		}
		if errors.Is(storage.ErrOrderAlreadyStoredByOtherUser, err) {
			http.Error(w, "", http.StatusConflict)
			return
		}
		if errors.Is(storage.ErrUnknownDatabase, err) {
			http.Error(w, "", http.StatusHTTPVersionNotSupported)
		}
		http.Error(w, err.Error(), http.StatusGone)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

func (c *controller) getOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserID).(string)
	orders, err := c.gophermartService.GetOrdersByUser(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *controller) getBalance(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserID).(string)
	balance, err := c.gophermartService.GetBalanceByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(balance); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *controller) createWithdraw(w http.ResponseWriter, r *http.Request) {
	if !isValidContentType(r, applicationJsonContentType, applicationXGzipContentType) {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	var withdraw dto.Withdraw
	err := extractJSONBody(r, &withdraw)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(UserID).(string)

	err = c.gophermartService.CreateWithdraw(r.Context(), userID, withdraw)
	if err != nil {
		if errors.Is(service.ErrorInvalidOrderNumberFormat, err) {
			http.Error(w, "", http.StatusUnprocessableEntity)
			return
		}

		if errors.Is(storage.ErrInsufficientFunds, err) {
			http.Error(w, "", http.StatusPaymentRequired)
			return
		}

		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *controller) getWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserID).(string)

	withdrawals, err := c.gophermartService.GetWithdrawalsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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

func extractJSONBody(r *http.Request, v interface{}) error {
	var reader io.ReadCloser
	if r.Header.Get(`Content-Encoding`) == `gzip` {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			return err
		}
		reader = gz
	}
	reader = r.Body
	defer func(reader io.ReadCloser) {
		err := reader.Close()
		if err != nil {
			log.Err(err).Msg(err.Error())
		}
	}(reader)
	if err := json.NewDecoder(reader).Decode(v); err != nil {
		return err
	}
	return nil
}
