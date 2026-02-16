package dadata

import (
	"encoding/json"
	"fmt"
	"geoservice/pkg/metrics"
	"time"

	"geoservice/internal/entity"
	"github.com/go-redis/redis"
)

// ProxyDaDataService кэширует ответы DaData
type ProxyDaDataService struct {
	service DataProvider
	cache   *redis.Client
	ttl     time.Duration
}

func NewProxyDaDataService(s DataProvider, cache *redis.Client, ttl time.Duration) DataProvider {
	return &ProxyDaDataService{service: s, cache: cache, ttl: ttl}
}

func (p *ProxyDaDataService) AddressSearch(query string) ([]*entity.Address, error) {

	cacheKey := fmt.Sprintf("dadata:geo:%s", query)

	// --- Cache GET ---
	start := time.Now()
	if val, err := p.cache.Get(cacheKey).Result(); err == nil {
		metrics.CacheDuration.WithLabelValues("GetGeoAddressesCache").Observe(time.Since(start).Seconds())
		var result []*entity.Address
		if jsonErr := json.Unmarshal([]byte(val), &result); jsonErr == nil {
			fmt.Println("DaData Geo from cache:", query)
			return result, nil
		}
	} else {
		metrics.CacheDuration.WithLabelValues("GetGeoAddressesCache").Observe(time.Since(start).Seconds())
	}

	// --- API ---
	start = time.Now()
	result, err := p.service.AddressSearch(query)
	metrics.APIDuration.WithLabelValues("GeoAddresses").Observe(time.Since(start).Seconds())
	if err != nil {
		return nil, err
	}

	// --- Cache SET ---
	data, _ := json.Marshal(result)
	start = time.Now()
	_ = p.cache.Set(cacheKey, data, p.ttl).Err()
	metrics.CacheDuration.WithLabelValues("SetGeoAddressesCache").Observe(time.Since(start).Seconds())

	fmt.Println("DaData Geo from API:", query)
	return result, nil
}

func (p *ProxyDaDataService) GeoCode(lat, lng float64) ([]*entity.Address, error) {
	cacheKey := fmt.Sprintf("dadata:coords:%f:%f", lat, lng)

	// --- Cache GET ---
	start := time.Now()
	if val, err := p.cache.Get(cacheKey).Result(); err == nil {
		metrics.CacheDuration.WithLabelValues("GetCoordsCache").Observe(time.Since(start).Seconds())
		var result []*entity.Address
		if jsonErr := json.Unmarshal([]byte(val), &result); jsonErr == nil {
			fmt.Println("DaData Coords from cache:", lat, lng)
			return result, nil
		}
	} else {
		metrics.CacheDuration.WithLabelValues("GetCoordsCache").Observe(time.Since(start).Seconds())
	}

	// --- API ---
	start = time.Now()
	result, err := p.service.GeoCode(lat, lng)
	metrics.APIDuration.WithLabelValues("GeocodeCoordinates").Observe(time.Since(start).Seconds())
	if err != nil {
		return nil, err
	}

	// --- Cache SET ---
	data, _ := json.Marshal(result)
	start = time.Now()
	_ = p.cache.Set(cacheKey, data, p.ttl).Err()
	metrics.CacheDuration.WithLabelValues("SetCoordsCache").Observe(time.Since(start).Seconds())

	fmt.Println("DaData Coords from API:", lat, lng)
	return result, nil
}
