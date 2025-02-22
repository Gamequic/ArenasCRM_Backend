package logs

import (
	"encoding/json"
	"net/http"
	"net/url"

	logsservice "github.com/Gamequic/LivePreviewBackend/pkg/features/logsViewer/service"
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"
	"github.com/gorilla/mux"
)

// Obtener estructura de logs
func getLogsStructure(w http.ResponseWriter, r *http.Request) {
	tree, err := logsservice.BuildTree("./logs", "")
	if err != nil {
		http.Error(w, "Error reading logs directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tree)
}

// Obtener archivo de log por fecha
func getLogFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	date := vars["date"]

	if date == "" || len(date) != 10 || date[4] != '-' || date[7] != '-' {
		http.Error(w, "Date must be in the format YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	file, err := logsservice.GetLogFile(date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Type", "text/plain")
	http.ServeFile(w, r, file.Name())
}

// Descargar logs (archivo o carpeta)
func downloadLogs(w http.ResponseWriter, r *http.Request) {
	queryPath := r.URL.Query().Get("path")
	if queryPath == "" {
		http.Error(w, "Path parameter is required", http.StatusBadRequest)
		return
	}

	decodedPath, err := url.QueryUnescape(queryPath)
	if err != nil {
		http.Error(w, "Invalid path parameter", http.StatusBadRequest)
		return
	}

	err = logsservice.DownloadLogs(decodedPath, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Registrar rutas
func RegisterSubRoutes(router *mux.Router) {
	logsRouter := router.PathPrefix("/logs").Subrouter()
	logsRouter.Use(middlewares.AuthHandler)
	logsRouter.Use(middlewares.ProfilesHandler([]uint{1, 2, 5}))

	logsRouter.HandleFunc("/structure", getLogsStructure).Methods("GET")
	logsRouter.HandleFunc("/view/{date}", getLogFile).Methods("GET")
	logsRouter.HandleFunc("/download", downloadLogs).Methods("GET")
}
