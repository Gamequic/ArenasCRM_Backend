package pieceservice

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Gamequic/LivePreviewBackend/pkg/database"
	doctorservice "github.com/Gamequic/LivePreviewBackend/pkg/features/doctor/service"
	hospitalservice "github.com/Gamequic/LivePreviewBackend/pkg/features/hospital/service"
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

	PublicId uint `gorm:"unique;not null" json:"PublicId"`

	HospitalID uint                     `gorm:"not null" json:"HospitalId"`
	Hospital   hospitalservice.Hospital `gorm:"foreignKey:HospitalID" json:"Hospital"`

	DoctorID uint                 `gorm:"not null" json:"DoctorID"`
	Doctor   doctorservice.Doctor `gorm:"foreignKey:DoctorID" json:"Doctor"`

	PatientName  string    `json:"PatientName"`
	PatientAge   uint      `json:"PatientAge"`
	Pieza        string    `gorm:"not null" json:"Pieza"`
	PriceTotal   float64   `gorm:"not null" json:"PriceTotal"`
	PricePaid    float64   `gorm:"not null" json:"PricePaid"`
	IsPaid       bool      `gorm:"not null" json:"IsPaid"`
	IsFactura    bool      `gorm:"not null" json:"IsFactura"`
	IsAseguranza bool      `gorm:"not null" json:"IsAseguranza"`
	PaidWithCard bool      `gorm:"not null" json:"PaidWithCard"`
	Date         time.Time `gorm:"not null" json:"Date"`
	Description  string    `json:"Description"`
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

func Create(piece *Pieces) int {
	// Verifica que el PublicId sea único
	var count int64
	database.DB.Model(&Pieces{}).Where("public_id = ?", piece.PublicId).Count(&count)
	if count > 0 {
		panic(middlewares.GormError{
			Code:    http.StatusBadRequest,
			Message: "PublicId must be unique",
			IsGorm:  true,
		})
	}

	// Buscar o crear hospital
	var hospital hospitalservice.Hospital
	err := database.DB.Where("name = ?", piece.Hospital.Name).First(&hospital).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			hospital = hospitalservice.Hospital{Name: piece.Hospital.Name}
			if err := database.DB.Create(&hospital).Error; err != nil {
				panic(fmt.Errorf("failed to create hospital: %w", err))
			}
		} else {
			panic(err)
		}
	}

	// Buscar o crear doctor
	var doctor doctorservice.Doctor
	err = database.DB.Where("name = ?", piece.Doctor.Name).First(&doctor).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			doctor = doctorservice.Doctor{Name: piece.Doctor.Name}
			if err := database.DB.Create(&doctor).Error; err != nil {
				panic(fmt.Errorf("failed to create doctor: %w", err))
			}
		} else {
			panic(err)
		}
	}

	// Asociar hospital con el doctor (tabla many2many)
	if err := database.DB.Model(&doctor).Association("Hospitals").Append(&hospital); err != nil {
		panic(fmt.Errorf("failed to associate doctor and hospital: %w", err))
	}

	// Asignar IDs
	piece.HospitalID = hospital.ID
	piece.DoctorID = doctor.ID

	// Limpiar structs embebidos para evitar error al insertar
	piece.Hospital = hospitalservice.Hospital{}
	piece.Doctor = doctorservice.Doctor{}

	// Crear pieza en DB
	if err := database.DB.Create(piece).Error; err != nil {
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
	// Query con preloads para traer relaciones
	if err := database.DB.Model(&Pieces{}).
		Preload("Doctor").
		Preload("Hospital").
		First(Piece, id).Error; err != nil {

		// Manejo de errores: registro no encontrado
		if err.Error() == "record not found" {
			panic(middlewares.GormError{
				Code:    404,
				Message: "Piece not found",
				IsGorm:  true,
			})
		} else {
			// Otro tipo de error
			panic(err)
		}
	}

	return http.StatusOK
}

func FindByFilters(filters map[string]string) []Pieces {
	loc, _ := time.LoadLocation("America/Mexico_City")
	dateLayout := "2006-01-02"

	var results []Pieces

	// Base query con Preloads válidos y orden descendente
	query := database.DB.Model(&Pieces{}).
		Preload("Doctor").
		Preload("Hospital").
		Order("created_at DESC")

	// --- FILTROS ---

	// Filtrar por publicId (string -> uint)
	if publicIdStr := filters["publicId"]; publicIdStr != "" && publicIdStr != "null" {
		publicId, err := strconv.ParseUint(publicIdStr, 10, 64)
		if err == nil {
			query = query.Where("public_id = ?", publicId)
		}
	}

	// Filtrar por paciente (campo string patient_name en Pieces)
	if paciente := filters["paciente"]; paciente != "" && paciente != "null" {
		query = query.Where("LOWER(patient_name) LIKE ?", "%"+strings.ToLower(paciente)+"%")
	}

	// Filtrar por medico (JOIN con tabla doctors)
	if medico := filters["medico"]; medico != "" && medico != "null" {
		query = query.Joins("JOIN doctors ON pieces.doctor_id = doctors.id").
			Where("LOWER(doctors.name) LIKE ?", "%"+strings.ToLower(medico)+"%")
	}

	// Filtrar por hospital (JOIN con tabla hospitals)
	if hospital := filters["hospital"]; hospital != "" && hospital != "null" {
		query = query.Joins("JOIN hospitals ON pieces.hospital_id = hospitals.id").
			Where("LOWER(hospitals.name) LIKE ?", "%"+strings.ToLower(hospital)+"%")
	}

	// Filtrar por pieza (string)
	if pieza := filters["pieza"]; pieza != "" && pieza != "null" {
		query = query.Where("LOWER(pieza) LIKE ?", "%"+strings.ToLower(pieza)+"%")
	}

	// Filtrar por price (float)
	if priceStr := filters["price"]; priceStr != "" && priceStr != "null" {
		price, err := strconv.ParseFloat(priceStr, 64)
		if err == nil {
			query = query.Where("price = ?", price)
		}
	}

	boolFilters := map[string]string{
		"IsPaid":       "is_paid",
		"IsFactura":    "is_factura",
		"IsAseguranza": "is_aseguranza",
		"PaidWithCard": "paid_with_card",
	}

	for paramKey, dbColumn := range boolFilters {
		if valStr := filters[paramKey]; valStr != "" && valStr != "null" {
			valBool, err := strconv.ParseBool(valStr)
			if err == nil {
				query = query.Where(dbColumn+" = ?", valBool)
			}
		}
	}

	// Filtrar por status (string)
	if status := filters["status"]; status != "" && status != "null" {
		query = query.Where("status = ?", status)
	}

	// Filtrar por fecha exacta campo 'date'
	if dateStr := strings.TrimSpace(filters["date"]); dateStr != "" && dateStr != "null" {
		dateParsed, err := time.ParseInLocation(dateLayout, dateStr, loc)
		if err == nil {
			// Tomamos todo el día de dateParsed
			query = query.Where("pieces.date >= ? AND pieces.date < ?", dateParsed, dateParsed.Add(24*time.Hour))
		}
	}

	// Filtrar por rango fechas campo 'date'
	startDate, endDate := strings.TrimSpace(filters["startDate"]), strings.TrimSpace(filters["endDate"])
	if startDate != "" && startDate != "null" && endDate != "" && endDate != "null" {
		startParsed, errStart := time.ParseInLocation(dateLayout, startDate, loc)
		endParsed, errEnd := time.ParseInLocation(dateLayout, endDate, loc)
		if errStart == nil && errEnd == nil {
			if startParsed.After(endParsed) {
				startParsed, endParsed = endParsed, startParsed
			}
			// Incluye todo el día final
			endParsed = endParsed.Add(24 * time.Hour)
			query = query.Where("pieces.date >= ? AND pieces.date < ?", startParsed, endParsed)
		}
	} else if startDate != "" && startDate != "null" {
		startParsed, err := time.ParseInLocation(dateLayout, startDate, loc)
		if err == nil {
			query = query.Where("pieces.date >= ?", startParsed)
		}
	} else if endDate != "" && endDate != "null" {
		endParsed, err := time.ParseInLocation(dateLayout, endDate, loc)
		if err == nil {
			query = query.Where("pieces.date < ?", endParsed.Add(24*time.Hour))
		}
	}

	// Filtrar por rango fechas campo 'created_at'
	startReg, endReg := strings.TrimSpace(filters["startRegisteredAt"]), strings.TrimSpace(filters["endRegisteredAt"])
	if startReg != "" && startReg != "null" && endReg != "" && endReg != "null" {
		startParsed, errStart := time.ParseInLocation(dateLayout, startReg, loc)
		endParsed, errEnd := time.ParseInLocation(dateLayout, endReg, loc)
		if errStart == nil && errEnd == nil {
			if startParsed.After(endParsed) {
				startParsed, endParsed = endParsed, startParsed
			}
			endParsed = endParsed.Add(24 * time.Hour)
			query = query.Where("pieces.created_at >= ? AND pieces.created_at < ?", startParsed, endParsed)
		}
	} else if startReg != "" && startReg != "null" {
		startParsed, err := time.ParseInLocation(dateLayout, startReg, loc)
		if err == nil {
			query = query.Where("pieces.created_at >= ?", startParsed)
		}
	} else if endReg != "" && endReg != "null" {
		endParsed, err := time.ParseInLocation(dateLayout, endReg, loc)
		if err == nil {
			query = query.Where("pieces.created_at < ?", endParsed.Add(24*time.Hour))
		}
	}

	// Ejecutar consulta
	query.Find(&results)

	return results
}

func Update(piece *Pieces, id uint) int {
	// Verificar que la pieza exista
	var existingPiece Pieces
	if err := database.DB.First(&existingPiece, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			panic(middlewares.GormError{
				Code:    http.StatusNotFound,
				Message: "Piece not found",
				IsGorm:  true,
			})
		}
		panic(err)
	}

	// Validar que PublicId sea único (excepto si es el mismo de la pieza actual)
	var count int64
	database.DB.Model(&Pieces{}).
		Where("public_id = ? AND id != ?", piece.PublicId, id).
		Count(&count)
	if count > 0 {
		panic(middlewares.GormError{
			Code:    http.StatusBadRequest,
			Message: "PublicId must be unique",
			IsGorm:  true,
		})
	}

	// Buscar o crear hospital
	var hospital hospitalservice.Hospital
	err := database.DB.Where("name = ?", piece.Hospital.Name).First(&hospital).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			hospital = hospitalservice.Hospital{Name: piece.Hospital.Name}
			if err := database.DB.Create(&hospital).Error; err != nil {
				panic(fmt.Errorf("failed to create hospital: %w", err))
			}
		} else {
			panic(err)
		}
	}

	// Buscar o crear doctor
	var doctor doctorservice.Doctor
	err = database.DB.Where("name = ?", piece.Doctor.Name).First(&doctor).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			doctor = doctorservice.Doctor{Name: piece.Doctor.Name}
			if err := database.DB.Create(&doctor).Error; err != nil {
				panic(fmt.Errorf("failed to create doctor: %w", err))
			}
		} else {
			panic(err)
		}
	}

	// Asociar hospital con doctor (many2many)
	if err := database.DB.Model(&doctor).Association("Hospitals").Append(&hospital); err != nil {
		panic(fmt.Errorf("failed to associate doctor and hospital: %w", err))
	}

	// Asignar IDs
	piece.HospitalID = hospital.ID
	piece.DoctorID = doctor.ID

	// Limpiar structs embebidos para evitar problemas en Save
	piece.Hospital = hospitalservice.Hospital{}
	piece.Doctor = doctorservice.Doctor{}

	// Forzar que el ID sea el correcto
	piece.ID = existingPiece.ID

	// Actualizar con mapa para incluir booleanos aunque sean false
	updates := map[string]interface{}{
		"public_id":      piece.PublicId,
		"date":           piece.Date,
		"hospital_id":    piece.HospitalID,
		"doctor_id":      piece.DoctorID,
		"patient_name":   piece.PatientName,
		"patient_age":    piece.PatientAge,
		"pieza":          piece.Pieza,
		"price_total":    piece.PriceTotal,
		"price_paid":     piece.PricePaid,
		"is_paid":        piece.IsPaid,
		"is_factura":     piece.IsFactura,
		"is_aseguranza":  piece.IsAseguranza,
		"paid_with_card": piece.PaidWithCard,
		"description":    piece.Description,
	}

	if err := database.DB.Model(&existingPiece).Updates(updates).Error; err != nil {
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
