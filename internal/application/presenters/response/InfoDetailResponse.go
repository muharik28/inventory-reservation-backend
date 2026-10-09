package response

import "time"

type InfoDetailResponse struct {
	ServiceName string    `json:"service_name"`
	Version     string    `json:"version"`
	Environment string    `json:"environment"`
	Status      string    `json:"status"`
	DBStatus    string    `json:"db_status"`
	ServerTime  time.Time `json:"server_time"`
}
