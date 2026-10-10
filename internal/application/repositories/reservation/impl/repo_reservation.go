package impl

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/muharik28/inventory-reservation-backend/internal/application/repositories/reservation"
	domain "github.com/muharik28/inventory-reservation-backend/internal/domain/entities/reservation"
	dbsql "github.com/muharik28/inventory-reservation-backend/internal/pkg/database/sql"
	"github.com/uptrace/bun"
)

type ReservationRepositoryImpl struct {
	db *bun.DB
}

func NewReservationRepository(db *bun.DB) reservation.ReservationRepository {
	return &ReservationRepositoryImpl{
		db: db,
	}
}

func (r *ReservationRepositoryImpl) Create(ctx context.Context, res *domain.Reservation) error {
	if res.CreatedAt.IsZero() {
		res.CreatedAt = time.Now()
	}
	res.UpdatedAt = time.Now()

	model := ToReservationModel(res)
	_, err := dbsql.TxOrDB(ctx, r.db).NewInsert().
		Model(model).
		Exec(ctx)

	return err
}

func (r *ReservationRepositoryImpl) GetByID(ctx context.Context, reservationID string) (*domain.Reservation, error) {
	model := new(ReservationModel)
	err := r.db.NewSelect().
		Model(model).
		Where("reservation_id = ?", reservationID).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReservationNotFound
		}
		return nil, err
	}

	return ToReservationDomain(model), nil
}

// GetByIDForUpdate locks the reservation row (SELECT ... FOR UPDATE) on the
// executor carried by ctx, so it participates in the ambient transaction.
func (r *ReservationRepositoryImpl) GetByIDForUpdate(ctx context.Context, reservationID string) (*domain.Reservation, error) {
	model := new(ReservationModel)
	err := dbsql.TxOrDB(ctx, r.db).NewSelect().
		Model(model).
		Where("reservation_id = ?", reservationID).
		For("UPDATE").
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrReservationNotFound
		}
		return nil, err
	}

	return ToReservationDomain(model), nil
}

func (r *ReservationRepositoryImpl) Update(ctx context.Context, res *domain.Reservation) error {
	res.UpdatedAt = time.Now()
	model := ToReservationModel(res)

	_, err := dbsql.TxOrDB(ctx, r.db).NewUpdate().
		Model(model).
		Column("status", "confirmed_at", "updated_at").
		Where("reservation_id = ?", res.ReservationID).
		Exec(ctx)

	return err
}

func (r *ReservationRepositoryImpl) FindExpiredPending(ctx context.Context, now time.Time, limit int) ([]*domain.Reservation, error) {
	var models []ReservationModel

	q := r.db.NewSelect().
		Model(&models).
		Where("status = ?", domain.StatusPending).
		Where("expires_at <= ?", now)

	if limit > 0 {
		q = q.Limit(int64(limit))
	}

	err := q.Scan(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]*domain.Reservation, len(models))
	for i := range models {
		results[i] = ToReservationDomain(&models[i])
	}

	return results, nil
}

func (r *ReservationRepositoryImpl) SumActiveByItem(ctx context.Context, itemID string, now time.Time) (int, error) {
	var models []ReservationModel
	var total int

	err := dbsql.TxOrDB(ctx, r.db).NewSelect().
		Model(&models).
		ColumnExpr("COALESCE(SUM(quantity), 0) AS active_reserved").
		Where("item_id = ?", itemID).
		Where("status = ?", domain.StatusPending).
		Where("expires_at > ?", now).
		Scan(ctx, &total)
	if err != nil {
		return 0, err
	}

	return total, nil
}
