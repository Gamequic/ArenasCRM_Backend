package featuresApi

import (
	"storegestserver/pkg/features/auth"
	authservice "storegestserver/pkg/features/auth/service"
	"storegestserver/pkg/features/profiles"
	profileservice "storegestserver/pkg/features/profiles/service"
	"storegestserver/pkg/features/users"
	userservice "storegestserver/pkg/features/users/service"
	"storegestserver/pkg/session"

	"github.com/gorilla/mux"
)

func RegisterSubRoutes(router *mux.Router) {
	userservice.InitUsersService()
	authservice.InitAuthService()
	profileservice.InitProfileService()
	session.InitSessionService()

	apiRouter := router.PathPrefix("/api").Subrouter()

	users.RegisterSubRoutes(apiRouter)
	auth.RegisterSubRoutes(apiRouter)
	profiles.RegisterSubRoutes(apiRouter)
}
