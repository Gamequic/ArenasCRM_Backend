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
	if err := database.DB.Find(Piece).Error; err != nil {
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

	const dateLayout = "2006-01-02"

	// Load local timezone (e.g., Mexico City)
	loc, err := time.LoadLocation("America/Mexico_City")
	if err != nil {
		loc = time.UTC
	}

	// Helper para parsear booleanos de forma flexible (acepta "true", "1", "yes", etc.)
	parseFlexibleBool := func(val string) (bool, error) {
		val = strings.ToLower(strings.TrimSpace(val))
		switch val {
		case "true", "1", "yes":
			return true, nil
		case "false", "0", "no":
			return false, nil
		default:
			return false, fmt.Errorf("invalid bool value: %s", val)
		}
	}

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
	fieldMap := map[string]string{
		"IsPaid":       "is_paid",
		"IsFactura":    "is_factura",
		"IsAseguranza": "is_aseguranza",
		"PaidWithCard": "paid_with_card",
	}

	for key, column := range fieldMap {
		if value, ok := filters[key]; ok && value != "" && value != "null" {
			if boolVal, err := parseFlexibleBool(value); err == nil {
				query = query.Where(fmt.Sprintf("%s = ?", column), boolVal)
			}
		}
	}

	// Filtro por fecha exacta (campo date)
	if date := strings.TrimSpace(filters["date"]); date != "" && date != "null" {
		if _, err := time.ParseInLocation(dateLayout, date, loc); err == nil {
			query = query.Where("DATE(date) = ?", date)
		}
	}

	// Rango de fechas para campo date (con validación e intercambio y zona horaria)
	startDate := strings.TrimSpace(filters["startDate"])
	endDate := strings.TrimSpace(filters["endDate"])

	if startDate != "" && startDate != "null" && endDate != "" && endDate != "null" {
		start, errStart := time.ParseInLocation(dateLayout, startDate, loc)
		end, errEnd := time.ParseInLocation(dateLayout, endDate, loc)
		if errStart == nil && errEnd == nil {
			if start.After(end) {
				start, end = end, start // intercambiar para que rango tenga sentido
			}
			end = end.Add(24 * time.Hour) // incluir todo el día final
			query = query.Where("date >= ? AND date < ?", start, end)
		}
	} else if startDate != "" && startDate != "null" {
		if start, err := time.ParseInLocation(dateLayout, startDate, loc); err == nil {
			query = query.Where("date >= ?", start)
		}
	} else if endDate != "" && endDate != "null" {
		if end, err := time.ParseInLocation(dateLayout, endDate, loc); err == nil {
			end = end.Add(24 * time.Hour)
			query = query.Where("date < ?", end)
		}
	}

	// Filtro por fecha de creación (created_at) exacta
	if regDate := strings.TrimSpace(filters["registeredAt"]); regDate != "" && regDate != "null" {
		if _, err := time.ParseInLocation(dateLayout, regDate, loc); err == nil {
			query = query.Where("DATE(created_at) = ?", regDate)
		}
	}

	// Rango de fechas para created_at
	startReg := strings.TrimSpace(filters["startRegisteredAt"])
	endReg := strings.TrimSpace(filters["endRegisteredAt"])
	if startReg != "" && startReg != "null" && endReg != "" && endReg != "null" {
		start, errStart := time.ParseInLocation(dateLayout, startReg, loc)
		end, errEnd := time.ParseInLocation(dateLayout, endReg, loc)
		if errStart == nil && errEnd == nil {
			end = end.Add(24 * time.Hour)
			query = query.Where("created_at >= ? AND created_at < ?", start, end)
		}
	}

	// Ejecutar la consulta final y devolver resultados o error
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
	var previousPiece Pieces
	FindOne(&previousPiece, uint(id))

	if err := database.DB.Delete(&Pieces{}, id).Error; err != nil {
		panic(err)
	}
}
