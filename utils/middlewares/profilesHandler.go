package middlewares

import (
	"net/http"
	"strconv"

	authstruct "github.com/Gamequic/LivePreviewBackend/pkg/features/auth/struct"
)

func ProfilesHandler(profileIDs []uint) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := r.Context().Value(UserKey).(*authstruct.TokenStruct)

			// If profiles if 0, then all users has access
			for _, requiredProfile := range profileIDs {
				if requiredProfile == 0 {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Verify if the user has the required profile
			hasAccess := false
			for _, userProfile := range user.Profiles {
				userProfileID, err := strconv.Atoi(userProfile)
				if err != nil {
					continue
				}
				for _, requiredProfile := range profileIDs {
					if uint(userProfileID) == requiredProfile {
						hasAccess = true
						break
					}
				}
				if hasAccess {
					break
				}
			}

			if !hasAccess {
				http.Error(w, "Access denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
