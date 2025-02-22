package featuresApi

import (
	"github.com/Gamequic/LivePreviewBackend/pkg/features/auth"
	authservice "github.com/Gamequic/LivePreviewBackend/pkg/features/auth/service"
	"github.com/Gamequic/LivePreviewBackend/pkg/features/profiles"
	profileservice "github.com/Gamequic/LivePreviewBackend/pkg/features/profiles/service"
	"github.com/Gamequic/LivePreviewBackend/pkg/features/users"
	userservice "github.com/Gamequic/LivePreviewBackend/pkg/features/users/service"
	"github.com/Gamequic/LivePreviewBackend/pkg/session"

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
