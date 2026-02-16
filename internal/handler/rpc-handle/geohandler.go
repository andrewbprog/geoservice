package rpc_handle

import (
	"errors"
	"geoservice/internal/entity"
	"geoservice/internal/service"
	"strconv"
)

// GeoRPCHandler обрабатывает RPC-запросы для гео-сервиса
type GeoRPCHandler struct {
	geoService service.GeoServicer
}

func NewGeoRPCHandler(geoService service.GeoServicer) *GeoRPCHandler {
	return &GeoRPCHandler{geoService: geoService}
}

// RPC-запросы/ответы

type SearchAddressRequest struct {
	Query string
}

type SearchAddressResponse struct {
	Addresses []*entity.Address
}

type ReverseGeocodeRequest struct {
	Lat string
	Lng string
}

type ReverseGeocodeResponse struct {
	Addresses []*entity.Address
}

// RPC метод поиска адресов

func (h *GeoRPCHandler) SearchAddress(req SearchAddressRequest, resp *SearchAddressResponse) error {
	if req.Query == "" {
		return errors.New("query parameter is required")
	}
	addresses, err := h.geoService.SearchAddress(req.Query)
	if err != nil {
		return err
	}
	*resp = SearchAddressResponse{Addresses: addresses}
	return nil
}

// RPC метод обратного геокодирования

func (h *GeoRPCHandler) ReverseGeocode(req ReverseGeocodeRequest, resp *ReverseGeocodeResponse) error {
	lat, err := strconv.ParseFloat(req.Lat, 64)
	if err != nil {
		return errors.New("invalid lat parameter")
	}
	lng, err := strconv.ParseFloat(req.Lng, 64)
	if err != nil {
		return errors.New("invalid lng parameter")
	}
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return errors.New("invalid coordinates range")
	}
	addresses, err := h.geoService.ReverseGeocode(lat, lng)
	if err != nil {
		return err
	}
	*resp = ReverseGeocodeResponse{Addresses: addresses}
	return nil
}
