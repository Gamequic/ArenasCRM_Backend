package auth

import (
	"encoding/json"
	"net/http"
	"reflect"

	authservice "github.com/Gamequic/LivePreviewBackend/pkg/features/auth/service"
	authstruct "github.com/Gamequic/LivePreviewBackend/pkg/features/auth/struct"
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"

	"github.com/gorilla/mux"
)

// CRUD

func LogIn(w http.ResponseWriter, r *http.Request) {
	var user authstruct.LogIn

	json.NewDecoder(r.Body).Decode(&user)

	var token string = authservice.LogIn(&user)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(token)
}

// ValidateToken checks if the token is valid and the session is active
func ValidateToken(w http.ResponseWriter, r *http.Request) {
	// Get the token from the context (set by the middleware)
	userClaims := r.Context().Value(middlewares.UserKey).(*authstruct.TokenStruct)

	// Validate the session in Redis
	err := authservice.ValidateSession(userClaims.SessionID, userClaims.Id)
	if err != nil {
		panic(middlewares.GormError{
			Code:    http.StatusUnauthorized,
			Message: "Session expired or invalid",
			IsGorm:  false,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"valid":    true,
		"user_id":  userClaims.Id,
		"email":    userClaims.Email,
		"username": userClaims.Username,
	})
}

// New handlers for protected routes
func Logout(w http.ResponseWriter, r *http.Request) {
	userClaims := r.Context().Value(middlewares.UserKey).(*authstruct.TokenStruct)
	tokenString := r.Context().Value(middlewares.UserTokenKey).(string)

	err := authservice.Logout(userClaims.Id, tokenString)
	if err != nil {
		panic(middlewares.GormError{
			Code:    http.StatusInternalServerError,
			Message: "Error logging out",
			IsGorm:  false,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

func GetSessions(w http.ResponseWriter, r *http.Request) {
	userClaims := r.Context().Value(middlewares.UserKey).(*authstruct.TokenStruct)

	sessions, err := authservice.GetUserSessions(userClaims.Id)
	if err != nil {
		panic(middlewares.GormError{
			Code:    http.StatusInternalServerError,
			Message: "Error getting sessions",
			IsGorm:  false,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(sessions)
}

// Register function

func RegisterSubRoutes(router *mux.Router) {
	authRouter := router.PathPrefix("/auth").Subrouter()

	// ValidatorHandler for login
	authLogInValidator := authRouter.NewRoute().Subrouter()
	authLogInValidator.Use(middlewares.ValidatorHandler(reflect.TypeOf(authstruct.LogIn{})))

	// Protected routes with AuthHandler
	protectedRoutes := authRouter.NewRoute().Subrouter()
	protectedRoutes.Use(middlewares.AuthHandler)

	// Public endpoints (login)
	authLogInValidator.HandleFunc("/login", LogIn).Methods("POST")

	// Protected endpoints
	protectedRoutes.HandleFunc("/validate", ValidateToken).Methods("GET")
	protectedRoutes.HandleFunc("/logout", Logout).Methods("POST")
	protectedRoutes.HandleFunc("/sessions", GetSessions).Methods("GET")
}
