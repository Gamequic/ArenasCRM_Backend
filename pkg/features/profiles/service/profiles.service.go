package profileservice

import (
	"errors"
	"os"
	"path/filepath"
	"storegestserver/pkg/database"
	profilestruct "storegestserver/pkg/features/profiles/struct"
	"storegestserver/utils"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Logger *zap.Logger

// Initialize the profile service
func InitProfileService() {
	Logger = utils.NewLogger()

	// AutoMigrate the profiles table
	err := database.DB.AutoMigrate(&profilestruct.Profile{})
	if err != nil {
		Logger.Error("Failed to migrate database:", zap.Error(err))
		return
	}

	// Check if profiles exist
	var count int64
	database.DB.Model(&profilestruct.Profile{}).Count(&count)
	if count > 0 {
		Logger.Info("Profiles already initialized")
		return
	}

	// Read and execute SQL file
	// The SQL file do the next thins
	// 1. Create the users_profiles
	// 2. Create the default profiles
	sqlFile, err := os.ReadFile(filepath.Join("sql", "init.sql"))
	if err != nil {
		Logger.Fatal("Failed to read SQL file:", zap.Error(err))
		return
	}

	// Execute SQL statements
	if err := database.DB.Exec(string(sqlFile)).Error; err != nil {
		Logger.Fatal("Failed to execute SQL:", zap.Error(err))
		return
	}

	Logger.Info("Successfully initialized profiles from SQL")
}

// CRUD Operations

func Create(profile *profilestruct.Profile) error {
	// Check if profile already exists
	var existingProfile profilestruct.Profile
	err := database.DB.Where("name = ?", profile.Name).First(&existingProfile).Error
	if err == nil {
		return errors.New("profile already exists with the provided name")
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	// Set timestamps
	profile.CreatedAt = time.Now()
	profile.UpdatedAt = time.Now()

	// Create profile
	if err := database.DB.Create(profile).Error; err != nil {
		return err
	}

	return nil
}

func Update(profile *profilestruct.Profile, id int) error {
	var existingProfile profilestruct.Profile
	if err := database.DB.First(&existingProfile, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("profile not found")
		}
		return err
	}

	profile.ID = id
	profile.UpdatedAt = time.Now()

	if err := database.DB.Save(profile).Error; err != nil {
		return err
	}

	return nil
}

func FindOne(profile *profilestruct.Profile, id int) error {
	if err := database.DB.First(profile, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("profile not found")
		}
		return err
	}
	return nil
}

func Find(profiles *[]profilestruct.Profile) error {
	if err := database.DB.Find(profiles).Error; err != nil {
		return err
	}
	return nil
}

func Delete(id int) error {
	var profile profilestruct.Profile
	if err := database.DB.First(&profile, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("profile not found")
		}
		return err
	}

	if err := database.DB.Delete(&profile).Error; err != nil {
		return err
	}

	return nil
}
