package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
)

// redisCache implements Redis-based caching
type redisCache struct {
	client *redis.Client
	config CacheConfig
}

// newRedisCache creates a new Redis cache client
func newRedisCache(config CacheConfig) (*redisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &redisCache{
		client: client,
		config: config,
	}, nil
}

// getActiveCampaigns retrieves active campaigns from Redis
func (rc *redisCache) getActiveCampaigns(ctx context.Context) ([]models.CampaignWithRules, error) {
	key := "adbeacon:campaigns:active"

	data, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("Redis get error: %w", err)
	}

	var campaigns []models.CampaignWithRules
	if err := json.Unmarshal([]byte(data), &campaigns); err != nil {
		return nil, fmt.Errorf("JSON unmarshal error: %w", err)
	}

	return campaigns, nil
}

// setActiveCampaigns stores active campaigns in Redis
func (rc *redisCache) setActiveCampaigns(ctx context.Context, campaigns []models.CampaignWithRules, ttl time.Duration) error {
	key := "adbeacon:campaigns:active"

	data, err := json.Marshal(campaigns)
	if err != nil {
		return fmt.Errorf("JSON marshal error: %w", err)
	}

	if err := rc.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("Redis set error: %w", err)
	}

	return nil
}

// getCampaignIndex retrieves campaign index from Redis
func (rc *redisCache) getCampaignIndex(ctx context.Context, key string) ([]string, error) {
	redisKey := fmt.Sprintf("adbeacon:index:%s", key)

	data, err := rc.client.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrCacheMiss
		}
		return nil, fmt.Errorf("Redis get error: %w", err)
	}

	var campaignIDs []string
	if err := json.Unmarshal([]byte(data), &campaignIDs); err != nil {
		return nil, fmt.Errorf("JSON unmarshal error: %w", err)
	}

	return campaignIDs, nil
}

// setCampaignIndex stores campaign index in Redis
func (rc *redisCache) setCampaignIndex(ctx context.Context, key string, campaignIDs []string, ttl time.Duration) error {
	redisKey := fmt.Sprintf("adbeacon:index:%s", key)

	data, err := json.Marshal(campaignIDs)
	if err != nil {
		return fmt.Errorf("JSON marshal error: %w", err)
	}

	if err := rc.client.Set(ctx, redisKey, data, ttl).Err(); err != nil {
		return fmt.Errorf("Redis set error: %w", err)
	}

	return nil
}

// clear removes all adbeacon cache keys from Redis
func (rc *redisCache) clear(ctx context.Context) error {
	// Get all keys matching our pattern
	keys, err := rc.client.Keys(ctx, "adbeacon:*").Result()
	if err != nil {
		return fmt.Errorf("Redis keys error: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	// Delete all keys
	if err := rc.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("Redis delete error: %w", err)
	}

	return nil
}

// publishCacheInvalidation publishes cache invalidation event
func (rc *redisCache) publishCacheInvalidation(ctx context.Context, event string) error {
	channel := "adbeacon:cache:invalidate"
	return rc.client.Publish(ctx, channel, event).Err()
}

// subscribeCacheInvalidation subscribes to cache invalidation events
func (rc *redisCache) subscribeCacheInvalidation(ctx context.Context, handler func(string)) error {
	channel := "adbeacon:cache:invalidate"
	pubsub := rc.client.Subscribe(ctx, channel)
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		handler(msg.Payload)
	}

	return nil
}

// close closes the Redis connection
func (rc *redisCache) close() error {
	return rc.client.Close()
}

// healthCheck checks Redis connection health
func (rc *redisCache) healthCheck(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}
