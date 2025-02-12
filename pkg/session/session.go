package session

import (
	"context"
	"encoding/json"
	"fmt"
	"storegestserver/pkg/database"
	authstruct "storegestserver/pkg/features/auth/struct"
	"storegestserver/utils"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func InitSessionService() {
	Logger = utils.NewLogger()
}

func StoreSession(userID int, sessionData *authstruct.Session) error {
	ctx := context.Background()
	key := fmt.Sprintf("user:%d:sessions", userID)

	sessionJSON, err := json.Marshal(sessionData)
	if err != nil {
		Logger.Error("Error marshaling session data", zap.Error(err))
		return err
	}

	err = database.RedisClient.SAdd(ctx, key, sessionJSON).Err()
	if err != nil {
		Logger.Error("Error storing session", zap.Error(err))
		return err
	}

	return nil
}

func GetUserSessions(userID int) ([]authstruct.Session, error) {
	ctx := context.Background()
	key := fmt.Sprintf("user:%d:sessions", userID)

	sessionJSONs, err := database.RedisClient.SMembers(ctx, key).Result()
	if err != nil {
		Logger.Error("Error getting sessions", zap.Error(err))
		return nil, err
	}

	sessions := make([]authstruct.Session, 0)
	for _, sessionJSON := range sessionJSONs {
		var session authstruct.Session
		if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
			Logger.Error("Error unmarshaling session", zap.Error(err))
			continue
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

func RemoveSession(userID int, tokenString string) error {
	ctx := context.Background()
	key := fmt.Sprintf("user:%d:sessions", userID)

	sessions, err := GetUserSessions(userID)
	if err != nil {
		return err
	}

	for _, session := range sessions {
		if session.Token == tokenString {
			sessionJSON, err := json.Marshal(session)
			if err != nil {
				Logger.Error("Error marshaling session", zap.Error(err))
				continue
			}

			err = database.RedisClient.SRem(ctx, key, sessionJSON).Err()
			if err != nil {
				Logger.Error("Error removing session", zap.Error(err))
				return err
			}
			break
		}
	}

	return nil
}

func ValidateSession(sessionID string, userID int) error {
	ctx := context.Background()
	key := fmt.Sprintf("user:%d:sessions", userID)

	sessionJSONs, err := database.RedisClient.SMembers(ctx, key).Result()
	if err != nil {
		Logger.Error("Error getting sessions", zap.Error(err))
		return err
	}

	for _, sessionJSON := range sessionJSONs {
		var session authstruct.Session
		if err := json.Unmarshal([]byte(sessionJSON), &session); err != nil {
			Logger.Error("Error unmarshaling session", zap.Error(err))
			continue
		}
		if session.SessionID == sessionID {
			return nil
		}
	}

	return fmt.Errorf("session not found or expired")
}
