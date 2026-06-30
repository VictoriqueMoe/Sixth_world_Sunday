package dto

type (
	SaveWeatherLocationRequest struct {
		Label     string  `json:"label"`
		PlaceName string  `json:"place_name"`
		Country   string  `json:"country"`
		Admin1    string  `json:"admin1"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Timezone  string  `json:"timezone"`
	}

	RenameWeatherLocationRequest struct {
		Label string `json:"label"`
	}

	WeatherLocationResponse struct {
		ID        string  `json:"id"`
		Label     string  `json:"label"`
		PlaceName string  `json:"place_name"`
		Country   string  `json:"country"`
		Admin1    string  `json:"admin1"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Timezone  string  `json:"timezone"`
		IsDefault bool    `json:"is_default"`
		CreatedAt string  `json:"created_at"`
	}

	WeatherLocationListResponse struct {
		Locations []WeatherLocationResponse `json:"locations"`
	}
)
