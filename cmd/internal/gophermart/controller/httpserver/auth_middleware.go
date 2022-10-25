package httpserver

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	authCookieName = "Authorization"
	//UserID         = "UserID"
)

type ContextKey string

var UserID ContextKey = "UserID"

func AuthMiddleware(parseCallback func(string) (string, error)) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(authCookieName)
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
				log.Error().Err(err).Msg(err.Error())
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			ctxWithUserID := context.WithValue(r.Context(), UserID, id)
			rWithUserID := r.WithContext(ctxWithUserID)
			next.ServeHTTP(w, rWithUserID)
		})
	}
}
