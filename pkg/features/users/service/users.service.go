package userservice

import (
	"fmt"
	"net/http"
	"storegestserver/pkg/database"
	"storegestserver/utils"
	"storegestserver/utils/middlewares"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var Logger *zap.Logger

// Define profile structure for response
type UserProfile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Update Users struct to use UserProfile
type Users struct {
	gorm.Model
	Name     string `gorm:"not null"`
	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	Profiles []uint `gorm:"-"` // Not a database field
}

// Initialize the user service
func InitUsersService() {
	Logger = utils.NewLogger()
	err := database.DB.AutoMigrate(&Users{})
	if err != nil {
		Logger.Error(fmt.Sprint("Failed to migrate database:", err))
	}
}

// CRUD Operations

func loadUserProfiles(userID uint) []uint {
	var profiles []uint
	database.DB.Raw(`
        SELECT profile_id 
        FROM user_profiles 
        WHERE user_id = ?
    `, userID).Scan(&profiles)
	return profiles
}

func Create(user *Users) int {
	// Encrypt password
	bytes, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(bytes)

	// Create users in DB
	if err := database.DB.Create(user).Error; err != nil {
		if err.Error() == `ERROR: duplicate key value violates unique constraint "uni_users_email" (SQLSTATE 23505)` {
			panic(middlewares.GormError{Code: 409, Message: "Email is on use", IsGorm: true})
		} else {
			panic(err)
		}
	}

	// Assign default profile (ID 3 - guest)
	err := database.DB.Exec(
		"INSERT INTO user_profiles (user_id, profile_id) VALUES (?, ?)",
		user.ID,
		3, // guest profile ID
	).Error
	if err != nil {
		panic(middlewares.GormError{
			Code:    500,
			Message: "Failed to assign default profile",
			IsGorm:  true,
		})
	}

	// Exclude password from response
	user.Password = ""

	// Load profiles
	user.Profiles = loadUserProfiles(user.ID)
	return http.StatusOK
}

func Find(u *[]Users) int {
	// Find all users and select all fields except password
	if err := database.DB.Select("id, name, email, created_at, updated_at, deleted_at").Find(u).Error; err != nil {
		panic(middlewares.GormError{
			Code:    http.StatusInternalServerError,
			Message: "Error retrieving users",
			IsGorm:  true,
		})
	}

	// Load profiles for each user
	for i := range *u {
		(*u)[i].Password = "" // Ensure password is empty
		(*u)[i].Profiles = loadUserProfiles((*u)[i].ID)
	}

	return http.StatusOK
}

func FindOne(user *Users, id uint) int {
	if err := database.DB.First(user, id).Error; err != nil {
		if err.Error() == "record not found" {
			panic(middlewares.GormError{Code: 404, Message: "Users not found", IsGorm: true})
		} else {
			panic(err)
		}
	}

	// Exclude password from response
	user.Password = ""

	user.Profiles = loadUserProfiles(user.ID)
	return http.StatusOK
}

func FindByEmail(user *Users, email string) int {
	if err := database.DB.Where("email = ?", email).First(user).Error; err != nil {
		if err.Error() == "record not found" {
			panic(middlewares.GormError{Code: 404, Message: "User not found", IsGorm: true})
		} else {
			panic(err)
		}
	}
	return http.StatusOK
}

func Update(user *Users, userId uint) int {
	// No autorize editing no existing users
	var previousUsers Users
	FindOne(&previousUsers, uint(user.ID))

	// Is the same user?
	if user.ID != userId {
		panic(middlewares.GormError{Code: http.StatusNotAcceptable, Message: "Is not allow to modify others users", IsGorm: true})
	}

	// Encrypt password
	bytes, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(bytes)

	if err := database.DB.Save(user).Error; err != nil {
		if err.Error() == `ERROR: duplicate key value violates unique constraint "uni_users_email" (SQLSTATE 23505)` {
			panic(middlewares.GormError{Code: 409, Message: "Email is on use", IsGorm: true})
		} else {
			panic(err)
		}
	}

	// Update profiles
	database.DB.Exec("DELETE FROM user_profiles WHERE user_id = ?", userId)
	for _, profileID := range user.Profiles {
		database.DB.Exec(
			"INSERT INTO user_profiles (user_id, profile_id) VALUES (?, ?)",
			userId,
			profileID,
		)
	}

	// Exclude password from response
	user.Password = ""

	// Reload profiles
	user.Profiles = loadUserProfiles(user.ID)
	return http.StatusOK
}

func Delete(id int) {
	Logger = utils.NewLogger()

	// No autorize deleting no existing users
	var previousUsers Users
	FindOne(&previousUsers, uint(id))

	if err := database.DB.Delete(&Users{}, id).Error; err != nil {
		panic(err)
	}
}
