package middlewares

import (
	"context"
	"net/http"
	"os"
	"strings"

	authstruct "github.com/Gamequic/LivePreviewBackend/pkg/features/auth/struct"
	"github.com/Gamequic/LivePreviewBackend/pkg/session"

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

/*
	Why this code is needed
	This code has the same purpose as the previous one, but is not a middleware, it can be use in any part of the code.
	This code is used to validate the users when they are connnecting from a websocket connection.
*/

func ValidateUser(authHeader string) int {
	var jwtKey = []byte(os.Getenv("JWTSECRET"))

	// Get token from Authorization header
	if authHeader == "" {
		logger.Error("Authorization header missing")
		return -1
	}

	// Remove Bearer prefix if present
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	tokenString = strings.Trim(tokenString, `"`) // Remove quotes if present

	// Ensure token is not empty
	if tokenString == "" {
		logger.Error("Token is missing after Bearer")
		return -1
	}

	// Validate JWT format
	if !strings.Contains(tokenString, ".") {
		logger.Error("Token is not in the correct JWT format")
		return -1
	}

	// Parse token with claims
	tokenData := &authstruct.TokenStruct{}
	token, err := jwt.ParseWithClaims(tokenString, tokenData, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			logger.Error("Stop hacking!")
			return -1
		}
		logger.Error("Invalid token")
		return -1
	}

	if !token.Valid {
		logger.Error("Invalid token")
		return -1
	}

	// Validate session in Redis
	err = session.ValidateSession(tokenData.SessionID, tokenData.Id)
	if err != nil {
		logger.Error("Session expired or invalid")
		return -1
	}

	return tokenData.Id
}
