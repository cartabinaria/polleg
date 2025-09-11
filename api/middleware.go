package api

import (
	"net/http"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/util"
)

func BanMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.MustGetUser(r).ID
		user, err := util.GetUserByID(util.GetDb(), userID)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "Could not get user from database")
			return
		}

		if user.Banned {
			httputil.WriteError(w, http.StatusForbidden, "You are banned from using this service")
			return
		}
		next.ServeHTTP(w, r)
	})
}
