package files

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// ServeProtectedFile sirve un archivo si el usuario está autorizado
func ServeProtectedFile(w http.ResponseWriter, r *http.Request) {
	// Obtiene el nombre del archivo desde el path
	fileName := strings.TrimPrefix(r.URL.Path, "/files/")

	// Define la ruta segura del archivo
	filePath := filepath.Join("storage", "files", fileName)

	// Verifica si el archivo existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// Sirve el archivo
	http.ServeFile(w, r, filePath)
}
