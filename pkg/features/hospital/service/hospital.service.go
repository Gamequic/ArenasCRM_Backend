package hospitalservice

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Gamequic/LivePreviewBackend/pkg/database"
	"github.com/Gamequic/LivePreviewBackend/utils"
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Logger *zap.Logger

// Update Pieces struct to use UserProfile
// Dentro del struct
type Hospital struct {
	gorm.Model
	Name string `json:"name" gorm:"not null"`
}

// Initialize the user service
func InitHospitalService() {
	Logger = utils.NewLogger()
	err := database.DB.AutoMigrate(&Hospital{})
	if err != nil {
		Logger.Error(fmt.Sprint("Failed to migrate database:", err))
	}
}

// CRUD Operations

func Create(Piece *Hospital) int {
	// Create users in DB
	if err := database.DB.Create(Piece).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			panic(middlewares.GormError{
				Code:    http.StatusBadRequest,
				Message: "PublicId must be unique",
				IsGorm:  true,
			})
		}
		panic(err)
	}

	return http.StatusOK
}

func Find(Piece *[]Hospital) int {
	if err := database.DB.Find(Piece).Error; err != nil {
		panic(middlewares.GormError{
			Code:    http.StatusInternalServerError,
			Message: "Error retrieving pieces",
			IsGorm:  true,
		})
	}
	return http.StatusOK
}

func FindOne(Piece *Hospital, id uint) int {
	if err := database.DB.First(Piece, id).Error; err != nil {
		if err.Error() == "record not found" {
			panic(middlewares.GormError{Code: 404, Message: "Piece not found", IsGorm: true})
		} else {
			panic(err)
		}
	}
	return http.StatusOK
}

func Update(Piece *Hospital, id uint) int {
	// No autorize editing no existing pieces
	var previousPiece Hospital
	FindOne(&previousPiece, uint(Piece.ID))

	if err := database.DB.Save(Piece).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			panic(middlewares.GormError{
				Code:    http.StatusBadRequest,
				Message: "PublicId must be unique",
				IsGorm:  true,
			})
		}
		panic(err)
	}

	return http.StatusOK
}

func Delete(id int) {
	Logger = utils.NewLogger()

	// No autorize deleting no existing pieces
	var previousPiece Hospital
	FindOne(&previousPiece, uint(id))

	if err := database.DB.Delete(&Hospital{}, id).Error; err != nil {
		panic(err)
	}
}
