package middlewares

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func ProfileHandler(requiredProfile string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userProfiles := r.Context().Value("profiles").([]string)
			for _, profile := range userProfiles {
				if strings.EqualFold(profile, requiredProfile) {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}
}
