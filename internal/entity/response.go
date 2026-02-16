package entity

import "time"

type CreationTokenResponse struct {
	// JWT токен
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpX6CJ9g52d"`
	CreatedAt time.Time `json:"created_at" example:"2025-07-30T09:25:48.509957118+03:00"`
} // @name CreationTokenResponse

// SearchResponse представляет ответ с найденными адресами
type SearchResponse struct {
	Addresses []*Address `json:"addresses"`
} // @name SearchResponse

// ErrorResponse представляет ответ с ошибкой

type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid request"`
	Message string `json:"message,omitempty" example:"Error message"`
} // @name ErrorResponse

type HealthResponse struct {
	Status  string `json:"status" example:"OK"`
	Service string `json:"service" example:"GeoService"`
}
