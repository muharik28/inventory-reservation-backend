package response

import "time"

type InfoDetailResponse struct {
	ServiceName string    `json:"service_name" example:"inventory-reservation-backend"`
	Version     string    `json:"version" example:"1.0.0"`
	Environment string    `json:"environment" example:"development"`
	Status      string    `json:"status" example:"healthy"`
	DBStatus    string    `json:"db_status" example:"connected"`
	ServerTime  time.Time `json:"server_time" example:"2026-10-10T08:30:00Z"`
}


type HealthResponse struct {
	Status   string `json:"status" example:"healthy"`
	Service  string `json:"service" example:"inventory-reservation-backend"`
	DBStatus string `json:"db_status" example:"connected"`
}

