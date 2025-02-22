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
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"
)

// Build logs structure
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

// Get the file type
func getFileType(fileName string) string {
	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".log":
		return "doc"
	default:
		return "file"
	}
}

// Get the log file
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

// Dowlaod log or zip logs
func DownloadLogs(path string, w http.ResponseWriter, r *http.Request) error {
	fullPath := filepath.Join("./logs", path)

	// Vertify that the path is inside the logs directory
	cleanFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return err
	}
	logsBasePath, _ := filepath.Abs("./logs")
	if !strings.HasPrefix(cleanFullPath, logsBasePath) {
		panic(middlewares.GormError{
			Code:    http.StatusForbidden,
			Message: "File is outside of /logs directory, Stop hacking!",
			IsGorm:  false,
		})
	}

	info, err := os.Stat(cleanFullPath)
	if err != nil {
		if os.IsNotExist(err) {
			panic(middlewares.GormError{
				Code:    http.StatusNotFound,
				Message: "File not found",
				IsGorm:  false,
			})
		}
		return err
	}

	// If it is a directory, zip it
	if info.IsDir() {
		zipFileName := filepath.Base(fullPath) + ".zip"
		w.Header().Set("Content-Disposition", "attachment; filename="+zipFileName)
		w.Header().Set("Content-Type", "application/zip")

		zipWriter := zip.NewWriter(w)
		defer zipWriter.Close()

		return addDirToZip(zipWriter, fullPath, "")
	}

	// If it is a file, serve it
	w.Header().Set("Content-Disposition", "attachment; filename="+info.Name())
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, fullPath)

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
