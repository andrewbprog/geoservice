package entity

// SearchRequest представляет запрос для поиска адресов
type SearchRequest struct {
	Query string `json:"query" validate:"required" example:"Москва, ул. Тверская, 1"`
} // @name SearchRequest

// GeocodeRequest представляет запрос для обратного геокодирования
type GeocodeRequest struct {
	Lat string `json:"lat" validate:"required" example:"55.7558"`
	Lng string `json:"lng" validate:"required" example:"37.6176"`
} // @name GeocodeRequest
