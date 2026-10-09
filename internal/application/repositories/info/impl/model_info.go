package impl

import "time"

type InfoModel struct {
	ServiceName string    `bun:"service_name"`
	Version     string    `bun:"version"`
	Environment string    `bun:"environment"`
	Status      string    `bun:"status"`
	DBStatus    string    `bun:"db_status"`
	ServerTime  time.Time `bun:"server_time"`
}
