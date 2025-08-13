package filescontroller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	fileservice "github.com/Gamequic/LivePreviewBackend/pkg/features/files/service"
	"github.com/gorilla/mux"
)

func GetFile(w http.ResponseWriter, r *http.Request) {
	// Gorilla mux vars
	vars := mux.Vars(r)
	filename := vars["filename"]

	filepath, err := fileservice.GetFilePath(filename)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, filepath)
}

// POST /files/upload
func UploadFile(w http.ResponseWriter, r *http.Request) {
	// Limitar tamaño del archivo (ej: 10MB)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "File too big", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error reading file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := filepath.Base(handler.Filename)
	filepath := filepath.Join(fileservice.UploadsPath, filename)

	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := dst.ReadFrom(file); err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "File uploaded successfully: %s\n", filename)
}
