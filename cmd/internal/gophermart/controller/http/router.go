package http

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/apolsh/yapr-gophermart/cmd/internal/gophermart/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type AuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func RegisterRoutes(router *chi.Mux, s service.GophermartService) {
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(5))

	router.Use(middleware.Timeout(60 * time.Second))

	newUserRoutes(router, s)
	newOrdersRoutes(router, s)

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
