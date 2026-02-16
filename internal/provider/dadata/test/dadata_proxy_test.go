package test

import (
	"geoservice/internal/entity"
	"geoservice/internal/provider/dadata"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type fakeDaData struct {
	calls int
}

func (f *fakeDaData) GeoCode(lat, lng float64) ([]*entity.Address, error) {
	return []*entity.Address{}, nil
}

func (f *fakeDaData) AddressSearch(query string) ([]*entity.Address, error) {
	f.calls++
	return []*entity.Address{{City: "Москва"}}, nil
}

func TestProxyDaDataService_Cache(t *testing.T) {
	s, _ := miniredis.Run()
	defer s.Close()
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})

	// Чистим кэш перед тестом
	require.NoError(t, rdb.FlushDB().Err())

	orig := &fakeDaData{}

	// Оборачиваем в прокси
	proxy := dadata.NewProxyDaDataService(orig, rdb, 10*time.Minute)

	// 1. Первый вызов должен пойти в "API"
	res1, err := proxy.AddressSearch("Москва")
	require.NoError(t, err)
	require.NotEmpty(t, res1, "результат из API пустой")

	// 2. Второй вызов — должен идти из кэша
	res2, err := proxy.AddressSearch("Москва")
	require.NoError(t, err)
	require.NotEmpty(t, res2, "результат из кэша пустой")

	// Проверяем, что данные одинаковые
	require.Equal(t, res1, res2)
}
