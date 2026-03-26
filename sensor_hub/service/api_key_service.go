package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	database "example/sensorHub/db"
	"example/sensorHub/types"
	"fmt"
	"log"
	"time"
)

type ApiKeyServiceInterface interface {
	CreateApiKey(name string, userId int, expiresAt *time.Time) (fullKey string, err error)
	ListApiKeysForUser(userId int) ([]database.ApiKey, error)
	UpdateApiKeyExpiry(keyId int, userId int, expiresAt *time.Time) error
	RevokeApiKey(keyId int, userId int) error
	DeleteApiKey(keyId int, userId int) error
	ValidateApiKey(rawKey string) (*types.User, error)
}

type ApiKeyService struct {
	apiKeyRepo database.ApiKeyRepository
	userRepo   database.UserRepository
	roleRepo   database.RoleRepository
}

func NewApiKeyService(a database.ApiKeyRepository, u database.UserRepository, r database.RoleRepository) *ApiKeyService {
	return &ApiKeyService{apiKeyRepo: a, userRepo: u, roleRepo: r}
}

const apiKeyPrefix = "shk_"

func (s *ApiKeyService) CreateApiKey(name string, userId int, expiresAt *time.Time) (string, error) {
	rawBytes := make([]byte, 32)
	if _, err := rand.Read(rawBytes); err != nil {
		return "", fmt.Errorf("failed to generate random key: %w", err)
	}

	fullKey := apiKeyPrefix + hex.EncodeToString(rawBytes)
	keyPrefix := fullKey[:12]
	keyHash := hashKey(fullKey)

	_, err := s.apiKeyRepo.CreateApiKey(name, keyPrefix, keyHash, userId, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to store api key: %w", err)
	}

	return fullKey, nil
}

func (s *ApiKeyService) ListApiKeysForUser(userId int) ([]database.ApiKey, error) {
	return s.apiKeyRepo.ListApiKeysForUser(userId)
}

func (s *ApiKeyService) UpdateApiKeyExpiry(keyId int, userId int, expiresAt *time.Time) error {
	return s.apiKeyRepo.UpdateApiKeyExpiry(keyId, expiresAt)
}

func (s *ApiKeyService) RevokeApiKey(keyId int, userId int) error {
	return s.apiKeyRepo.RevokeApiKey(keyId)
}

func (s *ApiKeyService) DeleteApiKey(keyId int, userId int) error {
	return s.apiKeyRepo.DeleteApiKey(keyId)
}

func (s *ApiKeyService) ValidateApiKey(rawKey string) (*types.User, error) {
	keyHash := hashKey(rawKey)

	apiKey, err := s.apiKeyRepo.GetApiKeyByHash(keyHash)
	if err != nil {
		return nil, fmt.Errorf("failed to look up api key: %w", err)
	}
	if apiKey == nil {
		return nil, nil
	}

	user, err := s.userRepo.GetUserById(apiKey.UserId)
	if err != nil {
		return nil, fmt.Errorf("failed to look up user for api key: %w", err)
	}
	if user == nil {
		return nil, nil
	}

	perms, err := s.roleRepo.GetPermissionsForUser(user.Id)
	if err == nil {
		user.Permissions = perms
	}

	// Fire-and-forget last_used update
	go func() {
		if err := s.apiKeyRepo.UpdateLastUsed(apiKey.Id); err != nil {
			log.Printf("failed to update api key last_used_at: %v", err)
		}
	}()

	return user, nil
}

func hashKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}
