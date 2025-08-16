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

	PatientName   string `json:"PatientName"`
	PatientAge    uint   `json:"PatientAge"`
	PatientGender string `json:"PatientGender"`

	Pieza        string    `gorm:"not null" json:"Pieza"`
	PriceTotal   float64   `gorm:"not null" json:"PriceTotal"`
	PricePaid    float64   `gorm:"not null" json:"PricePaid"`
	IsPaid       bool      `gorm:"not null" json:"IsPaid"`
	IsFactura    bool      `gorm:"not null" json:"IsFactura"`
	IsAseguranza bool      `gorm:"not null" json:"IsAseguranza"`
	PaidWithCard bool      `gorm:"not null" json:"PaidWithCard"`
	ReceivedAt   time.Time `json:"receivedAt"`
	DeliveredAt  time.Time `json:"deliveredAt"`
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

	query := database.DB.Model(&Pieces{}).
		Preload("Doctor").
		Preload("Hospital").
		Order("created_at DESC")

	// --- FILTROS BÁSICOS ---
	if publicIdStr := filters["publicId"]; publicIdStr != "" && publicIdStr != "null" {
		if publicId, err := strconv.ParseUint(publicIdStr, 10, 64); err == nil {
			query = query.Where("public_id = ?", publicId)
		}
	}

	if paciente := filters["paciente"]; paciente != "" && paciente != "null" {
		query = query.Where("LOWER(patient_name) LIKE ?", "%"+strings.ToLower(paciente)+"%")
	}

	if medico := filters["medico"]; medico != "" && medico != "null" {
		query = query.Joins("JOIN doctors ON pieces.doctor_id = doctors.id").
			Where("LOWER(doctors.name) LIKE ?", "%"+strings.ToLower(medico)+"%")
	}

	if hospital := filters["hospital"]; hospital != "" && hospital != "null" {
		query = query.Joins("JOIN hospitals ON pieces.hospital_id = hospitals.id").
			Where("LOWER(hospitals.name) LIKE ?", "%"+strings.ToLower(hospital)+"%")
	}

	if pieza := filters["pieza"]; pieza != "" && pieza != "null" {
		query = query.Where("LOWER(pieza) LIKE ?", "%"+strings.ToLower(pieza)+"%")
	}

	if priceStr := filters["price"]; priceStr != "" && priceStr != "null" {
		if price, err := strconv.ParseFloat(priceStr, 64); err == nil {
			query = query.Where("price_total = ?", price)
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
			if valBool, err := strconv.ParseBool(valStr); err == nil {
				query = query.Where(dbColumn+" = ?", valBool)
			}
		}
	}

	if status := filters["status"]; status != "" && status != "null" {
		query = query.Where("status = ?", status)
	}

	// --- FILTROS DE FECHAS ---
	addDateFilter := func(field string, exact, start, end string) {
		if exact != "" && exact != "null" {
			if parsed, err := time.ParseInLocation(dateLayout, exact, loc); err == nil {
				query = query.Where(fmt.Sprintf("pieces.%s >= ? AND pieces.%s < ?", field, field), parsed, parsed.Add(24*time.Hour))
			}
		}
		if start != "" && start != "null" && end != "" && end != "null" {
			if startParsed, errStart := time.ParseInLocation(dateLayout, start, loc); errStart == nil {
				if endParsed, errEnd := time.ParseInLocation(dateLayout, end, loc); errEnd == nil {
					if startParsed.After(endParsed) {
						startParsed, endParsed = endParsed, startParsed
					}
					query = query.Where(fmt.Sprintf("pieces.%s >= ? AND pieces.%s < ?", field, field), startParsed, endParsed.Add(24*time.Hour))
				}
			}
		} else if start != "" && start != "null" {
			if startParsed, err := time.ParseInLocation(dateLayout, start, loc); err == nil {
				query = query.Where(fmt.Sprintf("pieces.%s >= ?", field), startParsed)
			}
		} else if end != "" && end != "null" {
			if endParsed, err := time.ParseInLocation(dateLayout, end, loc); err == nil {
				query = query.Where(fmt.Sprintf("pieces.%s < ?", field), endParsed.Add(24*time.Hour))
			}
		}
	}

	// Fecha principal (date)
	addDateFilter("date", filters["date"], filters["startDate"], filters["endDate"])

	// Fecha registro (created_at)
	addDateFilter("created_at", filters["registeredAt"], filters["startRegisteredAt"], filters["endRegisteredAt"])

	// Nuevas fechas
	addDateFilter("received_at", filters["receivedAt"], filters["startReceivedAt"], filters["endReceivedAt"])
	addDateFilter("delivered_at", filters["deliveredAt"], filters["startDeliveredAt"], filters["endDeliveredAt"])

	// Ejecutar consulta
	query.Find(&results)
	return results
}

func Update(piece *Pieces, id uint) int {
	// 1. Verificar que la pieza exista
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

	// 2. Validar que PublicId sea único
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

	// 3. Buscar o crear hospital
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

	// 4. Buscar o crear doctor
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

	// 5. Asociar doctor y hospital
	if err := database.DB.Model(&doctor).Association("Hospitals").Append(&hospital); err != nil {
		panic(fmt.Errorf("failed to associate doctor and hospital: %w", err))
	}

	// 6. Asignar IDs y limpiar structs embebidos
	piece.HospitalID = hospital.ID
	piece.DoctorID = doctor.ID
	piece.Hospital = hospitalservice.Hospital{}
	piece.Doctor = doctorservice.Doctor{}
	piece.ID = existingPiece.ID

	// 7. Actualizar incluyendo los nuevos campos de fechas
	updates := map[string]interface{}{
		"public_id":      piece.PublicId,
		"hospital_id":    piece.HospitalID,
		"doctor_id":      piece.DoctorID,
		"patient_name":   piece.PatientName,
		"patient_age":    piece.PatientAge,
		"patient_gender": piece.PatientGender,
		"pieza":          piece.Pieza,
		"price_total":    piece.PriceTotal,
		"price_paid":     piece.PricePaid,
		"is_paid":        piece.IsPaid,
		"is_factura":     piece.IsFactura,
		"is_aseguranza":  piece.IsAseguranza,
		"paid_with_card": piece.PaidWithCard,
		"description":    piece.Description,
		"received_at":    piece.ReceivedAt,
		"delivered_at":   piece.DeliveredAt,
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
