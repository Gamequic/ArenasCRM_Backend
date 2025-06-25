package pieceservice

import (
	"fmt"
	"net/http"

	"github.com/Gamequic/LivePreviewBackend/pkg/database"
	"github.com/Gamequic/LivePreviewBackend/utils"
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Logger *zap.Logger

// Update Pieces struct to use UserProfile
type Pieces struct {
	gorm.Model
	PublicId     string  `gorm:"unique;not null"`
	Hospital     string  `gorm:"string"`
	Medico       string  `gorm:"string"`
	Paciente     string  `gorm:"string"`
	Pieza        string  `gorm:"string"`
	Price        float64 `gorm:"type:decimal(10,2)"`
	IsPaid       bool    `gorm:"default:false"`
	IsFactura    bool    `gorm:"default:true"`
	IsAseguranza bool    `gorm:"default:true"`
	Status       string  `gorm:"string"`
}

// Initialize the user service
func InitPiecesService() {
	Logger = utils.NewLogger()
	err := database.DB.AutoMigrate(&Pieces{})
	if err != nil {
		Logger.Error(fmt.Sprint("Failed to migrate database:", err))
	}
}

// CRUD Operations

func Create(Piece *Pieces) int {
	// Create users in DB
	if err := database.DB.Create(Piece).Error; err != nil {
		panic(err)
	}

	return http.StatusOK
}

func Find(Piece *[]Pieces) int {
	// Find all users and select all fields except password
	if err := database.DB.Select("id, name, created_at, updated_at, deleted_at").Find(Piece).Error; err != nil {
		panic(middlewares.GormError{
			Code:    http.StatusInternalServerError,
			Message: "Error retrieving users",
			IsGorm:  true,
		})
	}

	return http.StatusOK
}

func FindOne(Piece *Pieces, id uint) int {
	if err := database.DB.First(Piece, id).Error; err != nil {
		if err.Error() == "record not found" {
			panic(middlewares.GormError{Code: 404, Message: "Piece not found", IsGorm: true})
		} else {
			panic(err)
		}
	}
	return http.StatusOK
}

func FindByFilters(filters map[string]string) ([]Pieces, int) {
	var results []Pieces
	query := database.DB

	if id, ok := filters["identifier"]; ok && id != "" && id != "null" {
		query = query.Where("public_id = ?", id)
	}
	if date, ok := filters["date"]; ok && date != "" && date != "null" {
		query = query.Where("DATE(created_at) = ?", date)
	}
	if hospital, ok := filters["hospital"]; ok && hospital != "" && hospital != "null" {
		query = query.Where("hospital ILIKE ?", "%"+hospital+"%")
	}
	if medico, ok := filters["medico"]; ok && medico != "" && medico != "null" {
		query = query.Where("medico ILIKE ?", "%"+medico+"%")
	}
	if paciente, ok := filters["paciente"]; ok && paciente != "" && paciente != "null" {
		query = query.Where("paciente ILIKE ?", "%"+paciente+"%")
	}

	if err := query.Find(&results).Error; err != nil {
		panic(err)
	}

	return results, http.StatusOK
}

func Update(Piece *Pieces, id uint) int {
	// No autorize editing no existing pieces
	var previousPiece Pieces
	FindOne(&previousPiece, uint(Piece.ID))

	if err := database.DB.Save(Piece).Error; err != nil {
		panic(err)
	}

	return http.StatusOK
}

func Delete(id int) {
	Logger = utils.NewLogger()

	// No autorize deleting no existing pieces
	var previousPiece Pieces
	FindOne(&previousPiece, uint(id))

	if err := database.DB.Delete(&Pieces{}, id).Error; err != nil {
		panic(err)
	}
}
