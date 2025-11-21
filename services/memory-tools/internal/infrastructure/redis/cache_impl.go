package redis

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheImpl implements Redis-based caching
type CacheImpl struct {
	client    *redis.Client
	keyPrefix string
	ttl       time.Duration
}

// NewCacheImpl creates a new Redis cache implementation
func NewCacheImpl(redisURL, keyPrefix string, ttl time.Duration) (*CacheImpl, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}

	client := redis.NewClient(opts)
	
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("connect to redis: %w", err)
	}

	return &CacheImpl{
		client:    client,
		keyPrefix: keyPrefix,
		ttl:       ttl,
	}, nil
}

// Get retrieves an embedding from cache
func (c *CacheImpl) Get(key string) ([]float32, bool) {
	ctx := context.Background()
	data, err := c.client.Get(ctx, c.keyPrefix+key).Bytes()
	if err != nil {
		return nil, false
	}

	// Deserialize float32 array
	embedding := make([]float32, len(data)/4)
	for i := range embedding {
		bits := binary.LittleEndian.Uint32(data[i*4:])
		embedding[i] = math.Float32frombits(bits)
	}

	return embedding, true
}

// Set stores an embedding in cache
func (c *CacheImpl) Set(key string, value []float32, ttl time.Duration) {
	ctx := context.Background()

	// Serialize float32 array
	data := make([]byte, len(value)*4)
	for i, f := range value {
		bits := math.Float32bits(f)
		binary.LittleEndian.PutUint32(data[i*4:], bits)
	}

	if ttl == 0 {
		ttl = c.ttl
	}

	c.client.Set(ctx, c.keyPrefix+key, data, ttl)
}

// Delete removes an embedding from cache
func (c *CacheImpl) Delete(key string) error {
	ctx := context.Background()
	return c.client.Del(ctx, c.keyPrefix+key).Err()
}

// Clear removes all cached embeddings with the prefix
func (c *CacheImpl) Clear() error {
	ctx := context.Background()
	
	iter := c.client.Scan(ctx, 0, c.keyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	
	return iter.Err()
}

// Stats returns cache statistics
func (c *CacheImpl) Stats() (map[string]interface{}, error) {
	ctx := context.Background()
	
	info, err := c.client.Info(ctx, "stats").Result()
	if err != nil {
		return nil, err
	}
	
	// Count keys with our prefix
	var count int64
	iter := c.client.Scan(ctx, 0, c.keyPrefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		count++
	}
	
	return map[string]interface{}{
		"type":       "redis",
		"key_prefix": c.keyPrefix,
		"key_count":  count,
		"ttl":        c.ttl.String(),
		"info":       info,
	}, nil
}

// Close closes the Redis connection
func (c *CacheImpl) Close() error {
	return c.client.Close()
}
