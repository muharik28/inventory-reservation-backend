package impl

import (
	"github.com/muharik28/inventory-reservation-backend/internal/domain/entities/reservation"
)

func ToReservationDomain(m *ReservationModel) *reservation.Reservation {
	if m == nil {
		return nil
	}
	return &reservation.Reservation{
		ReservationID: m.ReservationID,
		UserID:        m.UserID,
		ItemID:        m.ItemID,
		Quantity:      m.Quantity,
		Status:        reservation.Status(m.Status),
		ExpiresAt:     m.ExpiresAt,
		ConfirmedAt:   m.ConfirmedAt,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func ToReservationModel(d *reservation.Reservation) *ReservationModel {
	if d == nil {
		return nil
	}
	return &ReservationModel{
		ReservationID: d.ReservationID,
		UserID:        d.UserID,
		ItemID:        d.ItemID,
		Quantity:      d.Quantity,
		Status:        string(d.Status),
		ExpiresAt:     d.ExpiresAt,
		ConfirmedAt:   d.ConfirmedAt,
		CreatedAt:     d.CreatedAt,
		UpdatedAt:     d.UpdatedAt,
	}
}
