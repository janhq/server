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
	"jan-server/services/llm-api/internal/infrastructure/kong"
)

// ErrLimitExceeded indicates the user hit the maximum number of active API keys.
var ErrLimitExceeded = errors.New("api key limit exceeded")

// ErrNotFound indicates the API key does not exist or does not belong to user.
var ErrNotFound = errors.New("api key not found")

// Service orchestrates API key lifecycle operations.
type Service struct {
	repo       Repository
	kong       *kong.Client
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
func NewService(repo Repository, kongClient *kong.Client, cfg Config, logger zerolog.Logger) *Service {
	return &Service{
		repo:       repo,
		kong:       kongClient,
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

	consumerUsername := usr.Subject
	if consumerUsername == "" {
		consumerUsername = fmt.Sprintf("user-%d", usr.ID)
	}

	if s.kong == nil {
		return nil, "", errors.New("kong admin client not configured")
	}

	if _, err := s.kong.EnsureConsumer(ctx, consumerUsername, usr.Subject, []string{"jan-user"}); err != nil {
		return nil, "", fmt.Errorf("ensure kong consumer: %w", err)
	}

	cred, err := s.kong.CreateKeyCredential(ctx, consumerUsername, rawKey, []string{"jan-user"})
	if err != nil {
		return nil, "", fmt.Errorf("create kong credential: %w", err)
	}

	record := &APIKey{
		ID:               uuid.NewString(),
		UserID:           usr.ID,
		Name:             name,
		Prefix:           s.keyPrefix,
		Suffix:           displaySuffix,
		Hash:             hashKey(rawKey),
		KongCredentialID: cred.ID,
		ExpiresAt:        expiresAt,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	persisted, err := s.repo.Create(ctx, record)
	if err != nil {
		return nil, "", err
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

// RevokeKey deletes the key credential and marks record revoked.
func (s *Service) RevokeKey(ctx context.Context, userID uint, keyID string) error {
	key, err := s.repo.FindByID(ctx, keyID)
	if err != nil {
		return err
	}
	if key == nil || key.UserID != userID {
		return ErrNotFound
	}

	if s.kong != nil && key.KongCredentialID != "" {
		if err := s.kong.DeleteKeyCredential(ctx, key.KongCredentialID); err != nil {
			return fmt.Errorf("delete kong credential: %w", err)
		}
	}

	return s.repo.MarkRevoked(ctx, key.ID, time.Now())
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
