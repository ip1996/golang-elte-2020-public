package app

import (
	"context"
	"errors"
	"net/http"
)

type contextKey string

func (c contextKey) String() string {
	return "feladat.context." + string(c)
}

var nameContextKey = contextKey("name")

func (a *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO(feladat): kerd le az "Authorization" header erteket a `code` valtozoba
		code := r.Header.Get("Authorization") // <???>
		if code == "" {
			sendError(a.logger, w, errors.New("empty code"), http.StatusUnauthorized)
			return
		}

		var name string
		err := a.db.Get(&name, "SELECT name FROM user WHERE code=?", code)
		if err != nil {
			sendError(a.logger, w, err, http.StatusUnauthorized)
			return
		}

		context.WithValue(r.Context(), nameContextKey, name)
		// TODO(feladat): fejezd be a http handlert, hogy tovabb adja a vezerlest a megfelelo context-el
		// <???>
		next.ServeHTTP(w, r)
	})
}

func getNameFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(nameContextKey).(string); ok {
		return val
	}
	return ""
}
