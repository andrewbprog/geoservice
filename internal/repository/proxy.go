package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"geoservice/pkg/metrics"
	"time"

	"geoservice/internal/entity"
	"github.com/go-redis/redis"
)

// ProxyUserRepository кэширует запросы к UserRepository
type ProxyUserRepository struct {
	repo  UserRepository
	cache *redis.Client
	ttl   time.Duration
}

func NewProxyUserRepository(repo UserRepository, cache *redis.Client, ttl time.Duration) UserRepository {
	return &ProxyUserRepository{repo: repo, cache: cache, ttl: ttl}
}

func (p *ProxyUserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	cacheKey := fmt.Sprintf("user:%s", id)

	// --- Cache GET ---
	start := time.Now()
	val, err := p.cache.Get(cacheKey).Result()
	metrics.CacheDuration.WithLabelValues("GetUserCache").Observe(time.Since(start).Seconds())

	if err == nil {
		var user entity.User
		if jsonErr := json.Unmarshal([]byte(val), &user); jsonErr == nil {
			fmt.Println("User from cache:", id)
			return &user, nil
		}
	}

	user, err := p.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Сохраняем в Redis
	data, _ := json.Marshal(user)
	start = time.Now()
	_ = p.cache.Set(cacheKey, data, p.ttl).Err()
	metrics.CacheDuration.WithLabelValues("SetUserCache").Observe(time.Since(start).Seconds())

	fmt.Println("User from Postgres:", id)
	return user, nil
}

func (p *ProxyUserRepository) Create(ctx context.Context, user *entity.User) error {
	return p.repo.Create(ctx, user)
}

func (p *ProxyUserRepository) Update(ctx context.Context, user *entity.User) error {
	// Сбросим кэш, чтобы не было устаревших данных
	start := time.Now()
	_ = p.cache.Del(fmt.Sprintf("user:%s", user.ID)).Err()
	metrics.CacheDuration.WithLabelValues("DelUserCache").Observe(time.Since(start).Seconds())
	return p.repo.Update(ctx, user)
}

func (p *ProxyUserRepository) Delete(ctx context.Context, id string) error {
	start := time.Now()
	_ = p.cache.Del(fmt.Sprintf("user:%s", id)).Err()
	metrics.CacheDuration.WithLabelValues("DelUserCache").Observe(time.Since(start).Seconds())
	return p.repo.Delete(ctx, id)
}

func (p *ProxyUserRepository) List(ctx context.Context, cond entity.Conditions) ([]entity.User, int64, error) {
	return p.repo.List(ctx, cond)
}
