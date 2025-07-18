package pieceservice

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Gamequic/LivePreviewBackend/pkg/database"
	"github.com/Gamequic/LivePreviewBackend/utils"
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Logger *zap.Logger

// Update Pieces struct to use UserProfile
// Dentro del struct
type Pieces struct {
	gorm.Model

	PublicId     uint      `gorm:"unique;not null" json:"PublicId"`
	Hospital     string    `json:"Hospital"`
	Medico       string    `json:"Medico"`
	Paciente     string    `json:"Paciente"`
	Pieza        string    `json:"Pieza"`
	Price        float64   `json:"Price"`
	IsPaid       bool      `json:"IsPaid"`
	IsFactura    bool      `json:"IsFactura"`
	IsAseguranza bool      `json:"IsAseguranza"`
	PaidWithCard bool      `json:"PaidWithCard"`
	Status       string    `json:"Status"`
	Date         time.Time `json:"Date"`
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

func Find(Piece *[]Pieces) int {
	if err := database.DB.Select("id, created_at, updated_at, deleted_at").Find(Piece).Error; err != nil {
		panic(middlewares.GormError{
			Code:    http.StatusInternalServerError,
			Message: "Error retrieving pieces",
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

	// Filtro por publicId
	if idStr := strings.TrimSpace(filters["publicId"]); idStr != "" && idStr != "null" {
		if id, err := strconv.Atoi(idStr); err == nil {
			query = query.Where("public_id = ?", id)
		}
	}

	// Filtros por texto (LIKE insensible a mayúsculas)
	textFields := []string{"hospital", "medico", "paciente", "pieza", "status"}
	for _, field := range textFields {
		if value := strings.TrimSpace(filters[field]); value != "" && value != "null" {
			query = query.Where(fmt.Sprintf("%s ILIKE ?", field), "%"+value+"%")
		}
	}

	// Filtro por precio exacto
	if priceStr := strings.TrimSpace(filters["price"]); priceStr != "" && priceStr != "null" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
			query = query.Where("price = ?", price)
		}
	}

	// Filtros booleanos
	// Filtros booleanos con mapeo de nombre struct → columna SQL
	fieldMap := map[string]string{
		"IsPaid":       "is_paid",
		"IsFactura":    "is_factura",
		"IsAseguranza": "is_aseguranza",
		"PaidWithCard": "paid_with_card",
	}

	for key, column := range fieldMap {
		if value, ok := filters[key]; ok && value != "" && value != "null" {
			boolVal, err := strconv.ParseBool(value)
			if err == nil {
				query = query.Where(fmt.Sprintf("%s = ?", column), boolVal)
			}
		}
	}

	// Filtro por fecha exacta
	if date := strings.TrimSpace(filters["date"]); date != "" && date != "null" {
		query = query.Where("DATE(date) = ?", date)
	}

	// Filtro por rango de fechas de date
	startDate := strings.TrimSpace(filters["startDate"])
	endDate := strings.TrimSpace(filters["endDate"])
	if startDate != "" && startDate != "null" && endDate != "" && endDate != "null" {
		layout := "2006-01-02"
		start, errStart := time.Parse(layout, startDate)
		end, errEnd := time.Parse(layout, endDate)
		if errStart == nil && errEnd == nil {
			end = end.Add(24 * time.Hour)
			query = query.Where("date >= ? AND date < ?", start, end)
		}
	}

	// Filtro por fecha de creación (RegisteredAt) exacta
	if regDate := strings.TrimSpace(filters["registeredAt"]); regDate != "" && regDate != "null" {
		query = query.Where("DATE(registered_at) = ?", regDate)
	}

	// Rango de fechas para RegisteredAt
	startReg := strings.TrimSpace(filters["startRegisteredAt"])
	endReg := strings.TrimSpace(filters["endRegisteredAt"])
	if startReg != "" && startReg != "null" && endReg != "" && endReg != "null" {
		layout := "2006-01-02"
		start, errStart := time.Parse(layout, startReg)
		end, errEnd := time.Parse(layout, endReg)
		if errStart == nil && errEnd == nil {
			end = end.Add(24 * time.Hour)
			query = query.Where("registered_at >= ? AND registered_at < ?", start, end)
		}
	}

	// Ejecutar query
	if err := query.Find(&results).Error; err != nil {
		return nil, http.StatusInternalServerError
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
