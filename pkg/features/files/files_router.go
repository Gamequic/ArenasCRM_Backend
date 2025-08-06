package files

import (
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"
	"github.com/gorilla/mux"
)

// RegisterFileRoutes agrega las rutas protegidas de archivos
func RegisterFileRoutes(r *mux.Router) {
	fileRouter := r.PathPrefix("/files").Subrouter()

	// Aplica middleware de autorización
	fileRouter.Use(middlewares.AuthHandler)

	// Ruta para servir archivos: /files/{filename}
	fileRouter.HandleFunc("/{filename:.*}", ServeProtectedFile).Methods("GET")
}
