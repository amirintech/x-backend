package api

import (
	"context"
	"net/http"

	"github.com/aimrintech/x-backend/constants"
	"github.com/aimrintech/x-backend/utils"
)

func chain(middlewares ...func(http.HandlerFunc) http.HandlerFunc) func(http.HandlerFunc) http.HandlerFunc {
	return func(final http.HandlerFunc) http.HandlerFunc {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			if utils.HandleCORSPreflight(w, r) {
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := utils.GetAuthCookie(r)
		if err != nil {
			http.Error(w, "Missing authentication token", http.StatusUnauthorized)
			return
		}
		userID, err := utils.ValidateJWT(tokenString)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), constants.USER_ID_KEY, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
