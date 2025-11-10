package apikey

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"jan-server/services/llm-api/internal/domain/user"
	"jan-server/services/llm-api/internal/infrastructure/keycloak"
)

// ErrLimitExceeded indicates the user hit the maximum number of active API keys.
var ErrLimitExceeded = errors.New("api key limit exceeded")

// ErrNotFound indicates the API key does not exist or does not belong to user.
var ErrNotFound = errors.New("api key not found")

// Service orchestrates API key lifecycle operations.
type Service struct {
	repo       Repository
	keycloak   *keycloak.Client
	logger     zerolog.Logger
	defaultTTL time.Duration
	maxTTL     time.Duration
	maxPerUser int
	keyPrefix  string
}

// Config configures the Service.
type Config struct {
	DefaultTTL time.Duration
	MaxTTL     time.Duration
	MaxPerUser int
	KeyPrefix  string
}

// NewService constructs an API key service.
func NewService(repo Repository, keycloakClient *keycloak.Client, cfg Config, logger zerolog.Logger) *Service {
	return &Service{
		repo:       repo,
		keycloak:   keycloakClient,
		logger:     logger.With().Str("component", "api-key-service").Logger(),
		defaultTTL: cfg.DefaultTTL,
		maxTTL:     cfg.MaxTTL,
		maxPerUser: cfg.MaxPerUser,
		keyPrefix:  cfg.KeyPrefix,
	}
}

// CreateKey generates a new API key for the given user and persists metadata.
func (s *Service) CreateKey(ctx context.Context, usr *user.User, name string, requestedTTL time.Duration) (*APIKey, string, error) {
	if usr == nil || usr.ID == 0 {
		return nil, "", fmt.Errorf("user is required")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, "", fmt.Errorf("name is required")
	}

	count, err := s.repo.CountActiveByUser(ctx, usr.ID)
	if err != nil {
		return nil, "", err
	}
	if s.maxPerUser > 0 && count >= int64(s.maxPerUser) {
		return nil, "", ErrLimitExceeded
	}

	ttl := s.defaultTTL
	if requestedTTL > 0 && requestedTTL < s.maxTTL {
		ttl = requestedTTL
	} else if requestedTTL > s.maxTTL {
		ttl = s.maxTTL
	}
	expiresAt := time.Now().Add(ttl)

	rawKey, err := s.generateKeySecret()
	if err != nil {
		return nil, "", err
	}
	displaySuffix := ""
	if len(rawKey) >= 4 {
		displaySuffix = rawKey[len(rawKey)-4:]
	}

	// Hash the API key for storage
	keyHash := hashKey(rawKey)

	record := &APIKey{
		ID:        uuid.NewString(),
		UserID:    usr.ID,
		Name:      name,
		Prefix:    s.keyPrefix,
		Suffix:    displaySuffix,
		Hash:      keyHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	persisted, err := s.repo.Create(ctx, record)
	if err != nil {
		return nil, "", err
	}

	// Store API key hash in Keycloak user attributes
	if s.keycloak != nil {
		if err := s.keycloak.StoreAPIKeyHash(ctx, usr.Subject, record.ID, keyHash); err != nil {
			s.logger.Warn().Err(err).Str("user_id", usr.Subject).Msg("failed to store api key in keycloak")
			// Continue - we have it in database
		}
	}

	return persisted, rawKey, nil
}

// ListKeys returns API keys for the provided user.
func (s *Service) ListKeys(ctx context.Context, userID uint) ([]APIKey, error) {
	items, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// RevokeKey marks the API key as revoked and removes it from Keycloak.
func (s *Service) RevokeKey(ctx context.Context, usr *user.User, keyID string) error {
	if usr == nil {
		return fmt.Errorf("user is required")
	}

	key, err := s.repo.FindByID(ctx, keyID)
	if err != nil {
		return err
	}
	if key == nil || key.UserID != usr.ID {
		return ErrNotFound
	}

	// Mark as revoked in database
	if err := s.repo.MarkRevoked(ctx, key.ID, time.Now()); err != nil {
		return fmt.Errorf("mark revoked: %w", err)
	}

	if s.keycloak != nil {
		if usr.Subject != "" {
			if err := s.keycloak.RemoveAPIKeyHash(ctx, usr.Subject, key.ID); err != nil {
				s.logger.Warn().
					Err(err).
					Str("key_id", key.ID).
					Str("user_subject", usr.Subject).
					Msg("failed to remove api key hash from keycloak")
			}
		} else {
			s.logger.Warn().
				Str("key_id", key.ID).
				Uint("user_id", usr.ID).
				Msg("api key revoked but user subject missing; unable to remove from keycloak")
		}
	}

	return nil
}

// ValidateAPIKey validates an API key and returns user information
func (s *Service) ValidateAPIKey(ctx context.Context, apiKey string) (*keycloak.APIKeyUserInfo, error) {
	// Hash the API key
	keyHash := hashKey(apiKey)

	// Validate via Keycloak
	if s.keycloak == nil {
		return nil, errors.New("keycloak client not configured")
	}

	userInfo, err := s.keycloak.ValidateAPIKeyHash(ctx, keyHash)
	if err != nil {
		return nil, fmt.Errorf("validate api key: %w", err)
	}

	return userInfo, nil
}

func (s *Service) generateKeySecret() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate key: %w", err)
	}
	randomPart := hex.EncodeToString(buf)
	prefix := strings.TrimSpace(s.keyPrefix)
	if prefix == "" {
		prefix = "sk"
	}
	return fmt.Sprintf("%s_%s", prefix, randomPart), nil
}

func hashKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}
