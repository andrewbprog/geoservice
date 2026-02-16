package service

import (
	"errors"
	"geoservice/internal/entity"
	da "geoservice/internal/provider/dadata"
	"strings"
)

// GeoServicer определяет интерфейс для работы с геосервисом
type GeoServicer interface {
	// SearchAddress получает адреса по текстовому запросу
	SearchAddress(query string) ([]*entity.Address, error)

	// ReverseGeocode получает адреса по географическим координатам
	ReverseGeocode(lat, lng float64) ([]*entity.Address, error)
}

type geoService struct {
	provider da.DataProvider
}

func NewGeoService(provider da.DataProvider) GeoServicer {
	return &geoService{provider: provider}
}

func (s *geoService) SearchAddress(query string) ([]*entity.Address, error) {
	if strings.TrimSpace(query) == "" {
		return nil, errors.New("query cannot be empty")
	}
	return s.provider.AddressSearch(query)
}

func (s *geoService) ReverseGeocode(lat, lng float64) ([]*entity.Address, error) {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return nil, errors.New("invalid coordinates")
	}
	return s.provider.GeoCode(lat, lng)
}
