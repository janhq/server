package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/janhq/jan-server/services/memory-tools/internal/domain/memory"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Helper function to convert []float32 to pgvector format string
func embeddingToString(embedding []float32) string {
	if len(embedding) == 0 {
		return "[]"
	}

	parts := make([]string, len(embedding))
	for i, val := range embedding {
		parts[i] = fmt.Sprintf("%f", val)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// Helper function to parse pgvector string to []float32
func stringToEmbedding(s string) ([]float32, error) {
	// Remove brackets
	s = strings.TrimPrefix(s, "[")
	s = strings.TrimSuffix(s, "]")

	if s == "" {
		return []float32{}, nil
	}

	parts := strings.Split(s, ",")
	embedding := make([]float32, len(parts))

	for i, part := range parts {
		var val float32
		_, err := fmt.Sscanf(strings.TrimSpace(part), "%f", &val)
		if err != nil {
			return nil, fmt.Errorf("parse embedding value: %w", err)
		}
		embedding[i] = val
	}

	return embedding, nil
}

// User Memory Operations

func (r *PostgresRepository) GetUserMemoryItems(ctx context.Context, userID string) ([]memory.UserMemoryItem, error) {
	query := `
		SELECT id, user_id, scope, key, text, score, created_at, updated_at
		FROM user_memory_items
		WHERE user_id = $1 AND is_deleted = false
		ORDER BY score DESC, updated_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query user memory: %w", err)
	}
	defer rows.Close()

	var items []memory.UserMemoryItem
	for rows.Next() {
		var item memory.UserMemoryItem
		err := rows.Scan(
			&item.ID, &item.UserID, &item.Scope, &item.Key,
			&item.Text, &item.Score, &item.CreatedAt, &item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user memory item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

func (r *PostgresRepository) UpsertUserMemoryItem(ctx context.Context, item *memory.UserMemoryItem) (string, error) {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	now := time.Now()
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}
	item.UpdatedAt = now

	query := `
		INSERT INTO user_memory_items (
			id, user_id, scope, key, text, score, embedding, is_deleted, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7::vector, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
			scope = EXCLUDED.scope,
			key = EXCLUDED.key,
			text = EXCLUDED.text,
			score = EXCLUDED.score,
			embedding = EXCLUDED.embedding,
			is_deleted = EXCLUDED.is_deleted,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		item.ID, item.UserID, item.Scope, item.Key, item.Text,
		item.Score, embeddingToString(item.Embedding), item.IsDeleted,
		item.CreatedAt, item.UpdatedAt,
	)

	if err != nil {
		return "", fmt.Errorf("upsert user memory item: %w", err)
	}

	return item.ID, nil
}

func (r *PostgresRepository) DeleteUserMemoryItem(ctx context.Context, id string) error {
	query := `UPDATE user_memory_items SET is_deleted = true, updated_at = $1 WHERE id = $2`
	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	if rows := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("user memory item not found")
	}

	return nil
}

func (r *PostgresRepository) SearchUserMemory(
	ctx context.Context,
	userID string,
	queryEmbedding []float32,
	limit int,
	minSimilarity float32,
) ([]memory.UserMemoryItem, error) {
	query := `
		SELECT 
			id, user_id, scope, key, text, score, created_at, updated_at,
			1 - (embedding <=> $1::vector) AS similarity
		FROM user_memory_items
		WHERE user_id = $2 
		  AND is_deleted = false
		  AND score >= 2
		  AND 1 - (embedding <=> $1::vector) >= $3
		ORDER BY embedding <=> $1::vector
		LIMIT $4
	`

	rows, err := r.db.Query(ctx, query,
		embeddingToString(queryEmbedding),
		userID,
		minSimilarity,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search user memory: %w", err)
	}
	defer rows.Close()

	var items []memory.UserMemoryItem
	for rows.Next() {
		var item memory.UserMemoryItem
		err := rows.Scan(
			&item.ID, &item.UserID, &item.Scope, &item.Key,
			&item.Text, &item.Score, &item.CreatedAt, &item.UpdatedAt,
			&item.Similarity,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user memory item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// Project Facts Operations

func (r *PostgresRepository) GetProjectFacts(ctx context.Context, projectID string) ([]memory.ProjectFact, error) {
	query := `
		SELECT id, project_id, kind, title, text, confidence, 
		       source_conversation_id, created_at, updated_at
		FROM project_facts
		WHERE project_id = $1 AND is_deleted = false
		ORDER BY confidence DESC, updated_at DESC
	`

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, fmt.Errorf("query project facts: %w", err)
	}
	defer rows.Close()

	var facts []memory.ProjectFact
	for rows.Next() {
		var fact memory.ProjectFact
		err := rows.Scan(
			&fact.ID, &fact.ProjectID, &fact.Kind, &fact.Title,
			&fact.Text, &fact.Confidence, &fact.SourceConversationID,
			&fact.CreatedAt, &fact.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan project fact: %w", err)
		}
		facts = append(facts, fact)
	}

	return facts, nil
}

func (r *PostgresRepository) UpsertProjectFact(ctx context.Context, fact *memory.ProjectFact) (string, error) {
	if fact.ID == "" {
		fact.ID = uuid.New().String()
	}

	now := time.Now()
	if fact.CreatedAt.IsZero() {
		fact.CreatedAt = now
	}
	fact.UpdatedAt = now

	query := `
		INSERT INTO project_facts (
			id, project_id, kind, title, text, confidence, embedding,
			source_conversation_id, is_deleted, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7::vector, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			kind = EXCLUDED.kind,
			title = EXCLUDED.title,
			text = EXCLUDED.text,
			confidence = EXCLUDED.confidence,
			embedding = EXCLUDED.embedding,
			is_deleted = EXCLUDED.is_deleted,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(ctx, query,
		fact.ID, fact.ProjectID, fact.Kind, fact.Title, fact.Text,
		fact.Confidence, embeddingToString(fact.Embedding),
		fact.SourceConversationID, fact.IsDeleted, fact.CreatedAt, fact.UpdatedAt,
	)

	if err != nil {
		return "", fmt.Errorf("upsert project fact: %w", err)
	}

	return fact.ID, nil
}

func (r *PostgresRepository) DeleteProjectFact(ctx context.Context, id string) error {
	query := `UPDATE project_facts SET is_deleted = true, updated_at = $1 WHERE id = $2`
	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	if rows := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("project fact not found")
	}

	return nil
}

func (r *PostgresRepository) SearchProjectFacts(
	ctx context.Context,
	projectID string,
	queryEmbedding []float32,
	limit int,
	minSimilarity float32,
) ([]memory.ProjectFact, error) {
	query := `
		SELECT 
			id, project_id, kind, title, text, confidence,
			source_conversation_id, created_at, updated_at,
			1 - (embedding <=> $1::vector) AS similarity
		FROM project_facts
		WHERE project_id = $2 
		  AND is_deleted = false
		  AND confidence >= 0.7
		  AND 1 - (embedding <=> $1::vector) >= $3
		ORDER BY embedding <=> $1::vector
		LIMIT $4
	`

	rows, err := r.db.Query(ctx, query,
		embeddingToString(queryEmbedding),
		projectID,
		minSimilarity,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search project facts: %w", err)
	}
	defer rows.Close()

	var facts []memory.ProjectFact
	for rows.Next() {
		var fact memory.ProjectFact
		err := rows.Scan(
			&fact.ID, &fact.ProjectID, &fact.Kind, &fact.Title,
			&fact.Text, &fact.Confidence, &fact.SourceConversationID,
			&fact.CreatedAt, &fact.UpdatedAt, &fact.Similarity,
		)
		if err != nil {
			return nil, fmt.Errorf("scan project fact: %w", err)
		}
		facts = append(facts, fact)
	}

	return facts, nil
}

// Episodic Events Operations

func (r *PostgresRepository) GetEpisodicEvents(ctx context.Context, userID string, limit int) ([]memory.EpisodicEvent, error) {
	query := `
		SELECT id, user_id, project_id, conversation_id, time, text, kind, created_at
		FROM episodic_events
		WHERE user_id = $1 AND is_deleted = false
		ORDER BY time DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("query episodic events: %w", err)
	}
	defer rows.Close()

	var events []memory.EpisodicEvent
	for rows.Next() {
		var event memory.EpisodicEvent
		err := rows.Scan(
			&event.ID, &event.UserID, &event.ProjectID, &event.ConversationID,
			&event.Time, &event.Text, &event.Kind, &event.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan episodic event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

func (r *PostgresRepository) CreateEpisodicEvent(ctx context.Context, event *memory.EpisodicEvent) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	if event.CreatedAt.IsZero() {
		event.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO episodic_events (
			id, user_id, project_id, conversation_id, time, text, kind,
			embedding, is_deleted, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8::vector, $9, $10)
	`

	_, err := r.db.Exec(ctx, query,
		event.ID, event.UserID, event.ProjectID, event.ConversationID,
		event.Time, event.Text, event.Kind,
		embeddingToString(event.Embedding), event.IsDeleted, event.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("create episodic event: %w", err)
	}

	return nil
}

func (r *PostgresRepository) DeleteEpisodicEvent(ctx context.Context, id string) error {
	query := `UPDATE episodic_events SET is_deleted = true WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if rows := result.RowsAffected(); rows == 0 {
		return fmt.Errorf("episodic event not found")
	}

	return nil
}

func (r *PostgresRepository) SearchEpisodicEvents(
	ctx context.Context,
	userID string,
	queryEmbedding []float32,
	limit int,
	minSimilarity float32,
) ([]memory.EpisodicEvent, error) {
	query := `
		SELECT 
			id, user_id, project_id, conversation_id, time, text, kind, created_at,
			1 - (embedding <=> $1::vector) AS similarity
		FROM episodic_events
		WHERE user_id = $2 
		  AND is_deleted = false
		  AND time > NOW() - INTERVAL '2 weeks'
		  AND 1 - (embedding <=> $1::vector) >= $3
		ORDER BY embedding <=> $1::vector
		LIMIT $4
	`

	rows, err := r.db.Query(ctx, query,
		embeddingToString(queryEmbedding),
		userID,
		minSimilarity,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("search episodic events: %w", err)
	}
	defer rows.Close()

	var events []memory.EpisodicEvent
	for rows.Next() {
		var event memory.EpisodicEvent
		err := rows.Scan(
			&event.ID, &event.UserID, &event.ProjectID, &event.ConversationID,
			&event.Time, &event.Text, &event.Kind, &event.CreatedAt,
			&event.Similarity,
		)
		if err != nil {
			return nil, fmt.Errorf("scan episodic event: %w", err)
		}
		events = append(events, event)
	}

	return events, nil
}

// Conversation Items Operations

func (r *PostgresRepository) CreateConversationItem(ctx context.Context, item *memory.ConversationItem) error {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO conversation_items (
			id, conversation_id, role, content, tool_calls, created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query,
		item.ID, item.ConversationID, item.Role, item.Content,
		item.ToolCalls, item.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("create conversation item: %w", err)
	}

	return nil
}

func (r *PostgresRepository) GetConversationItems(ctx context.Context, conversationID string) ([]memory.ConversationItem, error) {
	query := `
		SELECT id, conversation_id, role, content, tool_calls, created_at
		FROM conversation_items
		WHERE conversation_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, conversationID)
	if err != nil {
		return nil, fmt.Errorf("query conversation items: %w", err)
	}
	defer rows.Close()

	var items []memory.ConversationItem
	for rows.Next() {
		var item memory.ConversationItem
		err := rows.Scan(
			&item.ID, &item.ConversationID, &item.Role,
			&item.Content, &item.ToolCalls, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan conversation item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}
