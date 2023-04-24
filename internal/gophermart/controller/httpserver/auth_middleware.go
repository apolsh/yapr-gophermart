package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/apolsh/yapr-gophermart/internal/logger"
)

type ContextKey string

var UserID ContextKey = "UserID"

var authMiddlewareLogger = logger.LoggerOfComponent("authMiddleware")

func AuthMiddleware(parseCallback func(string) (string, error)) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(authorizationHeaderKey)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			cookieParts := strings.Split(cookie.Value, " ")
			if len(cookieParts) != 2 {
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			id, err := parseCallback(cookieParts[1])
			if err != nil {
				authMiddlewareLogger.Error(fmt.Errorf("authorization error: %w", err))
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			ctxWithUserID := context.WithValue(r.Context(), UserID, id)
			rWithUserID := r.WithContext(ctxWithUserID)
			next.ServeHTTP(w, rWithUserID)
		})
	}
}
