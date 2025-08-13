package files

import (
	filescontroller "github.com/Gamequic/LivePreviewBackend/pkg/features/files/controller"
	fileservice "github.com/Gamequic/LivePreviewBackend/pkg/features/files/service"
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"
	"github.com/gorilla/mux"
)

func RegisterFileRoutes(router *mux.Router) {
	// Crear carpeta al inicializar
	fileservice.CreateUploadsFolder()

	// Subrouter protegido
	filesRouter := router.PathPrefix("/files").Subrouter()
	filesRouter.Use(middlewares.AuthHandler)

	filesRouter.HandleFunc("/{filename}", filescontroller.GetFile).Methods("GET")
	filesRouter.HandleFunc("/upload", filescontroller.UploadFile).Methods("POST")
}
