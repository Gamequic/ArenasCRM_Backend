package middlewares

import (
	"context"
	"net/http"
	"os"
	authstruct "storegestserver/pkg/features/auth/struct"
	"storegestserver/pkg/session"
	"strings"

	"github.com/golang-jwt/jwt"
)

type contextKey string

const (
	UserIDKey    contextKey = "userId"
	UserTokenKey contextKey = "userToken"
	UserKey      contextKey = "user"
)

func AuthHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var jwtKey = []byte(os.Getenv("JWTSECRET"))

		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			panic(GormError{
				Code:    http.StatusBadRequest,
				Message: "Authorization header missing",
				IsGorm:  true,
			})
		}

		// Remove Bearer prefix if present
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse token with claims
		tokenData := &authstruct.TokenStruct{}
		token, err := jwt.ParseWithClaims(tokenString, tokenData, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				panic(GormError{
					Code:    http.StatusUnauthorized,
					Message: "Stop hacking!",
					IsGorm:  true,
				})
			}
			panic(GormError{
				Code:    http.StatusUnauthorized,
				Message: "Invalid token",
				IsGorm:  true,
			})
		}

		if !token.Valid {
			panic(GormError{
				Code:    http.StatusUnauthorized,
				Message: "Invalid token",
				IsGorm:  true,
			})
		}

		// Validate session in Redis
		err = session.ValidateSession(tokenData.SessionID, tokenData.Id)
		if err != nil {
			panic(GormError{
				Code:    http.StatusUnauthorized,
				Message: "Session expired or invalid",
				IsGorm:  true,
			})
		}

		// Create context with both userId and full tokenData
		ctx := context.WithValue(r.Context(), UserIDKey, tokenData.Id)
		ctx = context.WithValue(ctx, UserTokenKey, tokenString)
		ctx = context.WithValue(ctx, UserKey, tokenData)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
