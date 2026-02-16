package test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"geoservice/internal/entity"
	"geoservice/internal/repository"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
)

// fakeRepo — моковый UserRepository для теста
type fakeUserRepo struct {
	calls int
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id string) (*entity.User, error) {
	f.calls++
	return &entity.User{ID: id, Name: "John"}, nil
}
func (f *fakeUserRepo) Create(ctx context.Context, u *entity.User) error { return nil }
func (f *fakeUserRepo) Update(ctx context.Context, u *entity.User) error { return nil }
func (f *fakeUserRepo) Delete(ctx context.Context, id string) error      { return nil }
func (f *fakeUserRepo) List(ctx context.Context, c entity.Conditions) ([]entity.User, int64, error) {
	return nil, 0, nil
}

func TestProxyUserRepository_Cache(t *testing.T) {
	// запускаем in-memory redis
	s, _ := miniredis.Run()
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	ctx := context.Background()

	repo := &fakeUserRepo{}
	proxy := repository.NewProxyUserRepository(repo, rdb, 1*time.Minute)

	// 1. Первый вызов — идёт в fakeRepo
	u1, err := proxy.GetByID(ctx, "123")
	assert.NoError(t, err)
	assert.Equal(t, "John", u1.Name)
	assert.Equal(t, 1, repo.calls)

	// 2. Второй вызов — берём из кэша, calls не увеличился
	u2, err := proxy.GetByID(ctx, "123")
	assert.NoError(t, err)
	assert.Equal(t, "John", u2.Name)
	assert.Equal(t, 1, repo.calls) // всё ещё 1

	// 3. Проверяем, что в Redis реально что-то лежит
	val, _ := rdb.Get("user:123").Result()
	var cached entity.User
	_ = json.Unmarshal([]byte(val), &cached)
	assert.Equal(t, "John", cached.Name)
}
