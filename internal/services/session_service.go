package services

import (
	"encoding/json"
	"fmt"
	"keizer-auth-api/internal/models"
	"keizer-auth-api/internal/repositories"
	"keizer-auth-api/internal/utils"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type SessionService struct {
	redisRepo *repositories.RedisRepository
	userRepo  *repositories.UserRepository
}

func NewSessionService(redisRepo *repositories.RedisRepository) *SessionService {
	return &SessionService{redisRepo: redisRepo}
}

func (ss *SessionService) CreateSession(uuid uuid.UUID) (string, error) {
	sessionId, err := utils.GenerateSessionID()
	if err != nil {
		return "", fmt.Errorf("error in generating session %w", err)
	}

	var user *models.User
	user, err = ss.userRepo.GetUser(uuid)
	if err != nil {
		return "", fmt.Errorf("error in getting user %w", err)
	}
	userJson, err := json.Marshal(user)
	if err != nil {
		return "", fmt.Errorf("error occured %w", err)
	}

	err = ss.redisRepo.Set("dashboard-user-session-"+sessionId, string(userJson), utils.SessionExpiresIn)
	if err != nil {
		return "", fmt.Errorf("error in setting session %w", err)
	}

	return sessionId, nil
}

func (ss *SessionService) GetSession(sessionId string) (*models.User, error) {
	userSession, err := ss.redisRepo.Get("dashboard-user-session-" + sessionId)
	if err != nil {
		return nil, fmt.Errorf("no session found")
	}
	var userData *models.User
	err = json.Unmarshal([]byte(userSession), userData)
	if err != nil {
		return nil, fmt.Errorf("error in unmarshalling")
	}
	return userData, nil
}

func (ss *SessionService) UpdateSession(sessionId string) error {
	val, err := ss.redisRepo.Get("dashboard-user-session-" + sessionId)
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("session not found")
		}
		return err
	}
	err = ss.redisRepo.Set("dashboard-user-session-"+sessionId, val, utils.SessionExpiresIn)
	if err != nil {
		return fmt.Errorf("error in updating session %w", err)
	}
	return nil
}

func (ss *SessionService) TTL(sessionId string) (time.Duration, error) {
	ttl, err := ss.redisRepo.TTL("dashboard-user-session-" + sessionId)
	return ttl, err
}
