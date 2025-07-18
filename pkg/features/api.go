package featuresApi

import (
	"github.com/Gamequic/LivePreviewBackend/pkg/features/auth"
	authservice "github.com/Gamequic/LivePreviewBackend/pkg/features/auth/service"
	logs "github.com/Gamequic/LivePreviewBackend/pkg/features/logsViewer"
	"github.com/Gamequic/LivePreviewBackend/pkg/features/notifications"
	notificationservice "github.com/Gamequic/LivePreviewBackend/pkg/features/notifications/service"
	"github.com/Gamequic/LivePreviewBackend/pkg/features/pieces"
	pieceservice "github.com/Gamequic/LivePreviewBackend/pkg/features/pieces/service"
	"github.com/Gamequic/LivePreviewBackend/pkg/features/profiles"
	profileservice "github.com/Gamequic/LivePreviewBackend/pkg/features/profiles/service"
	"github.com/Gamequic/LivePreviewBackend/pkg/features/system"
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
	notificationservice.InitNotificationsService()
	pieceservice.InitPiecesService()

	apiRouter := router.PathPrefix("/api").Subrouter()

	users.RegisterSubRoutes(apiRouter)
	auth.RegisterSubRoutes(apiRouter)
	profiles.RegisterSubRoutes(apiRouter)
	logs.RegisterSubRoutes(apiRouter)
	system.RegisterSubRoutes(apiRouter)
	notifications.RegisterSubRoutes(apiRouter)
	pieces.RegisterSubRoutes(apiRouter)
}
