package http_handle

import (
	"geoservice/internal/entity"
	"geoservice/internal/service"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type GeoHandler struct {
	geoService service.GeoServicer
}

func NewGeoHandler(geoService service.GeoServicer) *GeoHandler {
	return &GeoHandler{geoService: geoService}
}

// SearchAddress godoc
// @Security BearerAuth
// @Summary Поиск адресов
// @Description Поиск адресов по текстовому запросу через DaData API
// @Tags addresses
// @Accept json
// @Produce json
// @Param request body entity.SearchRequest true "Запрос для поиска адресов"
// @Success 200 {object} entity.SearchResponse "OK"
// @Failure 400 {object} entity.ErrorResponse "Bad Request"
// @Failure 401 {object} entity.ErrorResponse "Unauthorized"
// @Failure 403 {object} entity.ErrorResponse "Forbidden"
// @Failure 500 {object} entity.ErrorResponse "Internal Server Error"
// @Failure 503 {object} entity.ErrorResponse "Service Unavailable"
// @Router /address/search [post]
func (gc *GeoHandler) SearchAddress(c *gin.Context) {
	var req entity.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON format"})
		return
	}

	if strings.TrimSpace(req.Query) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter is required"})
		return
	}

	addresses, err := gc.geoService.SearchAddress(req.Query)
	if err != nil {
		log.Printf("Error searching addresses: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service temporarily unavailable"})
		return
	}

	c.JSON(http.StatusOK, entity.SearchResponse{Addresses: addresses})
}

// ReverseGeocode godoc
// @Security BearerAuth
// @Summary Обратное геокодирование
// @Description Получение адреса по координатам через DaData API
// @Tags addresses
// @Accept json
// @Produce json
// @Param request body entity.GeocodeRequest true "Координаты для обратного геокодирования"
// @Success 200 {object} entity.SearchResponse "OK"
// @Failure 400 {object} entity.ErrorResponse "Bad Request"
// @Failure 401 {object} entity.ErrorResponse "Unauthorized"
// @Failure 403 {object} entity.ErrorResponse "Forbidden"
// @Failure 500 {object} entity.ErrorResponse "Internal Server Error"
// @Failure 503 {object} entity.ErrorResponse "Service Unavailable"
// @Router /address/geocode [post]
func (gc *GeoHandler) ReverseGeocode(c *gin.Context) {
	var req entity.GeocodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON format"})
		return
	}

	if strings.TrimSpace(req.Lat) == "" || strings.TrimSpace(req.Lng) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat and lng parameters are required"})
		return
	}

	lat, err := strconv.ParseFloat(req.Lat, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lat parameter"})
		return
	}

	lng, err := strconv.ParseFloat(req.Lng, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid lng parameter"})
		return
	}

	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coordinates range"})
		return
	}

	addresses, err := gc.geoService.ReverseGeocode(lat, lng)
	if err != nil {
		log.Printf("Error geocoding coordinates: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service temporarily unavailable"})
		return
	}

	c.JSON(http.StatusOK, entity.SearchResponse{Addresses: addresses})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, entity.HealthResponse{
		Status:  "OK",
		Service: "GeoService",
	})
}
