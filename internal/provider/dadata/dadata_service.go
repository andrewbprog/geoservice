package dadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"geoservice/internal/entity"
	"geoservice/pkg/metrics"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type DataProvider interface {
	AddressSearch(query string) ([]*entity.Address, error)
	GeoCode(lat, lng float64) ([]*entity.Address, error)
}

// DaDataService реализует service.DataProvider
type DaDataService struct {
	APIKey     string
	SecretKey  string
	BaseURL    string
	HttpClient *http.Client
}

func NewDaDataService(apiKey, secretKey string) DataProvider {
	return &DaDataService{
		APIKey:    apiKey,
		SecretKey: secretKey,
		BaseURL:   "https://suggestions.dadata.ru/suggestions/api/4_1/rs",
		HttpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (d *DaDataService) AddressSearch(query string) ([]*entity.Address, error) {
	start := time.Now()
	defer func() {
		metrics.APIDuration.WithLabelValues("GeoAddresses").Observe(time.Since(start).Seconds())
	}()
	requestBody := entity.DaDataSuggestRequest{
		Query: query,
		Count: 20,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", d.BaseURL+"/suggest/address", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+d.APIKey)
	if d.SecretKey != "" {
		req.Header.Set("X-Secret", d.SecretKey)
	}

	resp, err := d.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to DaData: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Ошибка при закрытии Body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("DaData API returned status %d: %s", resp.StatusCode, string(body))
	}

	var dadataResp entity.DaDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&dadataResp); err != nil {
		return nil, fmt.Errorf("failed to decode DaData response: %w", err)
	}

	addresses := make([]*entity.Address, 0, len(dadataResp.Suggestions))
	for _, suggestion := range dadataResp.Suggestions {
		addresses = append(addresses, d.convertDaDataToAddress(suggestion))
	}
	return addresses, nil
}

func (d *DaDataService) GeoCode(lat, lng float64) ([]*entity.Address, error) {
	start := time.Now()
	defer func() {
		metrics.APIDuration.WithLabelValues("GeocodeCoordinates").Observe(time.Since(start).Seconds())
	}()
	requestBody := entity.DaDataGeocodeRequest{
		Lat:   lat,
		Lon:   lng,
		Count: 10,
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", d.BaseURL+"/geolocate/address", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+d.APIKey)
	if d.SecretKey != "" {
		req.Header.Set("X-Secret", d.SecretKey)
	}

	resp, err := d.HttpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to DaData: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("Ошибка при закрытии Body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("DaData API returned status %d: %s", resp.StatusCode, string(body))
	}

	var dadataResp entity.DaDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&dadataResp); err != nil {
		return nil, fmt.Errorf("failed to decode DaData response: %w", err)
	}

	addresses := make([]*entity.Address, 0, len(dadataResp.Suggestions))
	for _, suggestion := range dadataResp.Suggestions {
		addresses = append(addresses, d.convertDaDataToAddress(suggestion))
	}
	return addresses, nil
}

func (d *DaDataService) convertDaDataToAddress(suggestion entity.DaDataSuggestion) *entity.Address {
	data := suggestion.Data
	var lat, lng float64
	if data.GeoLat != "" {
		if parsedLat, err := strconv.ParseFloat(data.GeoLat, 64); err == nil {
			lat = parsedLat
		}
	}
	if data.GeoLon != "" {
		if parsedLng, err := strconv.ParseFloat(data.GeoLon, 64); err == nil {
			lng = parsedLng
		}
	}
	city := data.City
	if city == "" {
		city = data.Settlement
	}
	if city == "" {
		city = data.Area
	}
	return &entity.Address{
		Result:         suggestion.UnrestrictedValue,
		City:           city,
		Street:         data.Street,
		House:          data.House,
		Lat:            lat,
		Lng:            lng,
		PostalCode:     data.PostalCode,
		Country:        data.Country,
		Region:         data.Region,
		Area:           data.Area,
		CityDistrict:   data.CityDistrict,
		Settlement:     data.Settlement,
		StreetWithType: data.StreetWithType,
		HouseWithType:  data.House,
	}
}
