package entity

// Address представляет полученный адрес
type Address struct {
	City           string  `json:"city"`
	Street         string  `json:"street,omitempty"`
	House          string  `json:"house,omitempty"`
	Lat            float64 `json:"lat,omitempty"`
	Lng            float64 `json:"lng,omitempty"`
	Result         string  `json:"result"`
	PostalCode     string  `json:"postal_code,omitempty"`
	Country        string  `json:"country,omitempty"`
	Region         string  `json:"region,omitempty"`
	Area           string  `json:"area,omitempty"`
	CityDistrict   string  `json:"city_district,omitempty"`
	Settlement     string  `json:"settlement,omitempty"`
	StreetWithType string  `json:"street_with_type,omitempty"`
	HouseWithType  string  `json:"house_with_type,omitempty"`
} // @name Address
