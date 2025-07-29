package authservice

import (
	"context"
	"os"
	"time"

	"github.com/Gamequic/LivePreviewBackend/pkg/database"
	authstruct "github.com/Gamequic/LivePreviewBackend/pkg/features/auth/struct"
	userservice "github.com/Gamequic/LivePreviewBackend/pkg/features/users/service"
	"github.com/Gamequic/LivePreviewBackend/pkg/session"
	"github.com/Gamequic/LivePreviewBackend/utils"
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

var Logger *zap.Logger

// Initialize the auth service
func InitAuthService() {
	Logger = utils.NewLogger()
}

// Auth Operations

func LogIn(u *authstruct.LogIn) string {
	var jwtKey = []byte(os.Getenv("JWTSECRET"))

	// Check if user exists
	var user userservice.Users
	userservice.FindByEmail(&user, u.Email)

	// Check password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(u.Password))
	if err != nil {
		if err.Error() == "crypto/bcrypt: hashedPassword is not the hash of the given password" {
			panic(middlewares.GormError{Code: 401, Message: "Password is wrong", IsGorm: true})
		}
		panic(err.Error())
	}

	// Load user profiles
	var profiles []string
	database.DB.Raw("SELECT profile_id FROM user_profiles WHERE user_id = ?", user.ID).Scan(&profiles)

	// Create token
	expirationTime := time.Now().AddDate(0, 6, 0) // six months
	sessionID := uuid.New().String()              // Generate a new session ID
	TokenData := &authstruct.TokenStruct{
		Username:  user.Name,
		Email:     user.Email,
		Id:        int(user.ID),
		SessionID: sessionID,
		Profiles:  profiles,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, TokenData)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		panic(err)
	}

	// Store session in Redis using user ID as key
	sessionData := &authstruct.Session{
		UserID:    int(user.ID),
		Email:     user.Email,
		Username:  user.Name,
		Token:     tokenString,
		SessionID: sessionID,
	}

	err = session.StoreSession(int(user.ID), sessionData)
	if err != nil {
		panic(err)
	}

	return tokenString
}

func GetUserSessions(userId int) ([]authstruct.Session, error) {
	return session.GetUserSessions(userId)
}

func Logout(userId int, tokenString string) error {
	return session.RemoveSession(userId, tokenString)
}

func ValidateSession(sessionID string, userID int) error {
	return session.ValidateSession(sessionID, userID)
}

func ValidatePasswordResetToken(token string) (string, error) {
	ctx := context.Background()

	email, err := database.RedisClient.Get(ctx, "pwd_reset:"+token).Result()
	if err != nil {
		Logger.Error("Invalid or expired password reset token", zap.Error(err))
		return "", err
	}

	return email, nil
}

func ResetPassword(token string, newPassword string) error {
	email, err := ValidatePasswordResetToken(token)
	if err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		Logger.Error("Error hashing password", zap.Error(err))
		return err
	}

	// Update user password in database
	var user userservice.Users
	userservice.FindByEmail(&user, email)
	user.Password = string(hashedPassword)
	database.DB.Save(&user)

	// Delete reset token from Redis
	ctx := context.Background()
	database.RedisClient.Del(ctx, "pwd_reset:"+token)

	Logger.Info("Password reset successful", zap.String("email", email))
	return nil
}
