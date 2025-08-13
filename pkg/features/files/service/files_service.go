package fileservice

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Gamequic/LivePreviewBackend/utils"
	"go.uber.org/zap"
)

var Logger *zap.Logger

const UploadsPath = "./uploads"

// Initialize the auth service
func InitAuthService() {
	Logger = utils.NewLogger()
}

// CreateUploadsFolder ensures the uploads folder exists
func CreateUploadsFolder() {
	if _, err := os.Stat(UploadsPath); os.IsNotExist(err) {
		err := os.MkdirAll(UploadsPath, 0755)
		if err != nil {
			Logger.Error(fmt.Sprintf("❌ Error creating uploads folder: %v", err))
		}
		Logger.Info("✅ Uploads folder created")
	} else {
		Logger.Info("📂 Uploads folder already exists")
	}
}

// GetFilePath returns the full path for a given filename, ensuring it's safe
func GetFilePath(filename string) (string, error) {
	cleanName := filepath.Base(filename)
	fullPath := filepath.Join(UploadsPath, cleanName)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", os.ErrNotExist
	}
	return fullPath, nil
}
