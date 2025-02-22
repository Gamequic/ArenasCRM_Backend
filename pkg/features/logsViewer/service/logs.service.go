package logsservice

import (
	"archive/zip"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	logsstruct "github.com/Gamequic/LivePreviewBackend/pkg/features/logsViewer/struct"
)

// Construir estructura de logs
func BuildTree(basePath, parentID string) ([]logsstruct.TreeLogNode, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	var nodes []logsstruct.TreeLogNode
	for _, entry := range entries {
		node := logsstruct.TreeLogNode{
			ID:    filepath.Join(parentID, entry.Name()),
			Label: entry.Name(),
		}
		if entry.IsDir() {
			node.FileType = "folder"
			children, err := BuildTree(filepath.Join(basePath, entry.Name()), node.ID)
			if err != nil {
				return nil, err
			}
			node.Children = children
		} else {
			node.FileType = getFileType(entry.Name())
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// Obtener tipo de archivo
func getFileType(fileName string) string {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".log":
		return "doc"
	default:
		return "file"
	}
}

// Obtener archivo de log
func GetLogFile(date string) (*os.File, error) {
	logFilePath := filepath.Join("./logs", date[:4], date[5:7], date+".log")

	file, err := os.Open(logFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("Log file not found")
		}
		return nil, err
	}

	return file, nil
}

// Descargar logs (archivo o carpeta)
func DownloadLogs(path string, w http.ResponseWriter) error {
	fullPath := filepath.Join("./logs", path)
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return errors.New("File or directory not found")
		}
		return err
	}

	if info.IsDir() {
		zipFileName := filepath.Base(fullPath) + ".zip"
		w.Header().Set("Content-Disposition", "attachment; filename="+zipFileName)
		w.Header().Set("Content-Type", "application/zip")

		zipWriter := zip.NewWriter(w)
		defer zipWriter.Close()

		return addDirToZip(zipWriter, fullPath, "")
	}

	http.ServeFile(w, nil, fullPath)
	return nil
}

// Agregar directorio a ZIP
func addDirToZip(zipWriter *zip.Writer, dirPath, baseInZip string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			err := addDirToZip(zipWriter, fullPath, filepath.Join(baseInZip, entry.Name()))
			if err != nil {
				return err
			}
		} else {
			file, err := os.Open(fullPath)
			if err != nil {
				return err
			}
			defer file.Close()

			zipFileWriter, err := zipWriter.Create(filepath.Join(baseInZip, entry.Name()))
			if err != nil {
				return err
			}

			if _, err := io.Copy(zipFileWriter, file); err != nil {
				return err
			}
		}
	}
	return nil
}
