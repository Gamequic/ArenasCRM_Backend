package notificationservice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Gamequic/LivePreviewBackend/pkg/database"
	userservice "github.com/Gamequic/LivePreviewBackend/pkg/features/users/service"
	"github.com/Gamequic/LivePreviewBackend/utils"
	"github.com/Gamequic/LivePreviewBackend/utils/middlewares"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Logger *zap.Logger

// Update Users struct to use UserProfile
type Notifications struct {
	gorm.Model
	UserId  int               `gorm:"not null"`
	Message string            `gorm:"not null"`
	Seen    bool              `gorm:"default:false"`
	User    userservice.Users `gorm:"foreignKey:UserId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

// Initialize the user service
func InitNotificationsService() {
	Logger = utils.NewLogger()
	err := database.DB.AutoMigrate(&Notifications{})
	if err != nil {
		Logger.Error(fmt.Sprint("Failed to migrate database:", err))
	}
}

// CRUD Operations

func FindOne(notification *Notifications, id int) error {
	if err := database.DB.First(notification, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			panic(middlewares.GormError{Code: 404, Message: "Message not found", IsGorm: true})
		} else {
			panic(err)
		}
	}
	return nil
}

func Find(notification *[]Notifications) int {
	if err := database.DB.Find(notification).Error; err != nil {
		panic(middlewares.GormError{
			Code:    http.StatusInternalServerError,
			Message: "Error retrieving notifications",
			IsGorm:  true,
		})
	}

	return http.StatusOK
}

func MarkAsSeen(notificationId int) int {
	var notification Notifications
	FindOne(&notification, notificationId)

	notification.Seen = true
	database.DB.Save(&notification)

	return http.StatusOK
}

func Create(notification *Notifications, userId int) int {
	notification.Seen = false

	// Create the notification on the database
	if err := database.DB.Create(notification).Error; err != nil {
		if err.Error() == "Error 1062: Duplicate entry" {
			panic(middlewares.GormError{Code: 409, Message: "Message already exists", IsGorm: true})
		} else {
			panic(err)
		}
	}

	// Publish the notification to the user on redis
	ctx := context.Background()
	channel := "user_notifications:" + strconv.Itoa(userId)

	payload, err := json.Marshal(notification)
	if err != nil {
		Logger.Error("Failed to serialize payload", zap.Error(err))
		panic(middlewares.GormError{Code: 500, Message: "Failed to serialize payload", IsGorm: false})
	}

	if err := database.RedisClient.Publish(ctx, channel, payload).Err(); err != nil {
		Logger.Error("Failed to publish notification to Redis", zap.Error(err))
		panic(middlewares.GormError{Code: 500, Message: "Failed to publish notification to Redis", IsGorm: false})
	}

	Logger.Info(fmt.Sprintf("Notification sent to user %d", userId))
	return http.StatusCreated
}

func NotificationWebSocketEndpoint(w http.ResponseWriter, r *http.Request) {
	userId := r.URL.Query().Get("userID")

	conn, err := utils.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		Logger.Error(fmt.Sprintf("No se pudo actualizar a WebSocket: %v", err))
		return
	}
	defer conn.Close()

	// Timeout for discont the client
	timeoutDuration := 5 * time.Second
	timeout := time.NewTimer(timeoutDuration)
	defer timeout.Stop()

	// Validate if the userId can hear notifications
	firstMessage := true

	// variables for auth
	// Variables para autenticación
	var (
		mu            sync.Mutex
		authenticated bool
	)

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if firstMessage {
				firstMessage = false
				userID := middlewares.ValidateUser(string(message))
				if userID == -1 {
					Logger.Info("User not valid for notifications")
					conn.Close()
					return
				} else {
					Logger.Info(fmt.Sprintf("User %v connected in notifications", userID))

					// Lock to modify authenticated safely
					mu.Lock()
					authenticated = true
					mu.Unlock()
				}
			}
			if err != nil {
				Logger.Info(fmt.Sprintf("Client desconection from notifications: %v", err))
				return
			}
		}
	}()

	ctx := context.Background()
	channel := "user_notifications:" + userId

	pubsub := database.RedisClient.Subscribe(ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()

	for {
		select {
		case msg := <-ch:
			if err := conn.WriteJSON(msg.Payload); err != nil {
				Logger.Error(fmt.Sprint("Error writing to WebSocket ", err))
				conn.Close()
				return
			}
		// Disconect the client if the timeout is reached
		case <-timeout.C:
			mu.Lock()
			// Disconnect the client if it is not authenticated
			if !authenticated {
				Logger.Info(fmt.Sprintf("Cliente %s desconectado por falta de autenticación", userId))
				conn.Close()
				mu.Unlock()
				return
			}
			mu.Unlock()
		}
	}

	return
}
