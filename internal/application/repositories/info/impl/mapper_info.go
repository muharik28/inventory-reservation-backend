package impl

import (
	"github.com/muharik28/inventory-reservation-backend/internal/domain/entities/info"
)

func ToInfoDomain(m *InfoModel) *info.ServiceInfo {
	if m == nil {
		return nil
	}
	return &info.ServiceInfo{
		ServiceName: m.ServiceName,
		Version:     m.Version,
		Environment: m.Environment,
		Status:      m.Status,
		DBStatus:    m.DBStatus,
		ServerTime:  m.ServerTime,
	}
}
