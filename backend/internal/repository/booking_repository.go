package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/platform"
)

// ErrNotFound is returned when a row lookup yields zero rows.
var ErrNotFound = errors.New("not found")

// BookedQty is the per-(date,time) aggregate of active booked quantities.
type BookedQty struct {
	Big    int
	Medium int
	Small  int
}

// SlotKey identifies a slot within a date by (time, route).
type SlotKey struct {
	Time  string
	Route string
}

// BookingRepository is the persistence boundary for bookings.
type BookingRepository interface {
	Create(ctx context.Context, b *model.Booking) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Booking, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	FindByPaymentSessionID(ctx context.Context, sessionID string) (*model.Booking, error)
	FindByIdempotencyKey(ctx context.Context, userID uuid.UUID, key, date, time, route string) (*model.Booking, error)
	SetStatus(ctx context.Context, id uuid.UUID, status model.BookingStatus) error
	SetPaymentSessionID(ctx context.Context, id uuid.UUID, sessionID string) error
	SetPriceOverride(ctx context.Context, id uuid.UUID, override *float64, reason *string, adminID *uuid.UUID) error
	Cancel(ctx context.Context, id uuid.UUID, adminID uuid.UUID, reason string) error

	// SumActiveQuantitiesForSlot returns booked_big and booked_medium for a slot,
	// counting only confirmed bookings or pending bookings whose hold has not expired.
	// MUST be called inside a transaction that has already SELECT...FOR UPDATE'd the slot row.
	SumActiveQuantitiesForSlot(ctx context.Context, date, time, route string) (bookedBig, bookedMedium, bookedSmall int, err error)

	// SumActiveQuantitiesForDate returns total booked counts grouped by time
	// across an entire date — used by the per-date availability endpoint.
	SumActiveQuantitiesForDate(ctx context.Context, date string) (map[SlotKey]BookedQty, error)

	// SumActiveQuantitiesForRange aggregates booked quantities for every (date, time)
	// in the half-open interval [startDate, endDate) — used by the monthly availability
	// endpoint to avoid the N+1 of one query per day.
	//
	// Returns a nested map: out[date][time] = {Big, Medium}. Dates / times with no
	// active bookings are simply absent from the map (caller must treat as zero).
	SumActiveQuantitiesForRange(ctx context.Context, startDate, endDate string) (map[string]map[SlotKey]BookedQty, error)

	// ExpirePending sets all stale pending bookings to expired and returns how many rows changed.
	ExpirePending(ctx context.Context, now time.Time) (int64, error)

	// FindBySlot returns all bookings for a given (date, time) pair, ordered by creation time.
	// No status filter — admin sees everything.
	FindBySlot(ctx context.Context, date, time, route string) ([]model.Booking, error)

	SetPosterIDs(ctx context.Context, id uuid.UUID, orderID, txID int64) error
}

type bookingRepo struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepo{db: db}
}

func (r *bookingRepo) tx(ctx context.Context) *gorm.DB {
	return platform.DBFromContext(ctx, r.db)
}

func (r *bookingRepo) Create(ctx context.Context, b *model.Booking) error {
	return r.tx(ctx).Create(b).Error
}

func (r *bookingRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	var b model.Booking
	err := r.tx(ctx).First(&b, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *bookingRepo) FindByPaymentSessionID(ctx context.Context, sessionID string) (*model.Booking, error) {
	var b model.Booking
	err := r.tx(ctx).First(&b, "payment_session_id = ?", sessionID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *bookingRepo) FindByIdempotencyKey(
	ctx context.Context,
	userID uuid.UUID,
	key, date, time, route string,
) (*model.Booking, error) {
	var b model.Booking
	err := r.tx(ctx).
		Where("user_id = ? AND idempotency_key = ? AND date = ? AND time = ? AND route_name = ?",
			userID, key, date, time, route).
		First(&b).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *bookingRepo) SetStatus(ctx context.Context, id uuid.UUID, status model.BookingStatus) error {
	return r.tx(ctx).
		Model(&model.Booking{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *bookingRepo) SetPaymentSessionID(ctx context.Context, id uuid.UUID, sessionID string) error {
	return r.tx(ctx).
		Model(&model.Booking{}).
		Where("id = ?", id).
		Update("payment_session_id", sessionID).Error
}

func (r *bookingRepo) SetPriceOverride(
	ctx context.Context,
	id uuid.UUID,
	override *float64,
	reason *string,
	adminID *uuid.UUID,
) error {
	updates := map[string]any{
		"price_override":  override,
		"override_reason": reason,
	}
	if override != nil {
		updates["overridden_by"] = adminID
		updates["overridden_at"] = time.Now().UTC()
	} else {
		updates["overridden_by"] = nil
		updates["overridden_at"] = nil
	}
	return r.tx(ctx).Model(&model.Booking{}).Where("id = ?", id).Updates(updates).Error
}

func (r *bookingRepo) Cancel(ctx context.Context, id uuid.UUID, adminID uuid.UUID, reason string) error {
	now := time.Now().UTC()
	return r.tx(ctx).Model(&model.Booking{}).Where("id = ?", id).Updates(map[string]any{
		"status":        model.StatusCancelled,
		"cancelled_by":  adminID,
		"cancelled_at":  now,
		"cancel_reason": reason,
	}).Error
}

func (r *bookingRepo) SumActiveQuantitiesForSlot(
	ctx context.Context,
	date, time_, route string,
) (int, int, int, error) {
	var result struct {
		BookedBig    int
		BookedMedium int
		BookedSmall  int
	}
	err := r.tx(ctx).
		Table("bookings").
		Select(`COALESCE(SUM(qty_big), 0)    AS booked_big,
		        COALESCE(SUM(qty_medium), 0) AS booked_medium,
		        COALESCE(SUM(qty_small), 0)  AS booked_small`).
		Where("date = ? AND time = ? AND route_name = ?", date, time_, route).
		Where("status = ? OR (status = ? AND expires_at > NOW())",
			model.StatusConfirmed, model.StatusPending).
		Scan(&result).Error
	if err != nil {
		return 0, 0, 0, err
	}
	return result.BookedBig, result.BookedMedium, result.BookedSmall, nil
}

func (r *bookingRepo) SumActiveQuantitiesForDate(
	ctx context.Context,
	date string,
) (map[SlotKey]BookedQty, error) {
	type row struct {
		Time         string
		RouteName    string
		BookedBig    int
		BookedMedium int
		BookedSmall  int
	}
	var rows []row
	err := r.tx(ctx).
		Table("bookings").
		Select(`time, route_name,
		        COALESCE(SUM(qty_big), 0)    AS booked_big,
		        COALESCE(SUM(qty_medium), 0) AS booked_medium,
		        COALESCE(SUM(qty_small), 0)  AS booked_small`).
		Where("date = ?", date).
		Where("status = ? OR (status = ? AND expires_at > NOW())",
			model.StatusConfirmed, model.StatusPending).
		Group("time, route_name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[SlotKey]BookedQty, len(rows))
	for _, rw := range rows {
		out[SlotKey{Time: rw.Time, Route: rw.RouteName}] =
			BookedQty{Big: rw.BookedBig, Medium: rw.BookedMedium, Small: rw.BookedSmall}
	}
	return out, nil
}

// SumActiveQuantitiesForRange aggregates booked qty for every (date, time) in
// the half-open interval [startDate, endDate). One query, grouped on the DB side.
// Indexes used: idx_bookings_date_time and idx_bookings_status_expires.
func (r *bookingRepo) SumActiveQuantitiesForRange(
	ctx context.Context,
	startDate, endDate string,
) (map[string]map[SlotKey]BookedQty, error) {
	type row struct {
		Date         time.Time
		Time         string
		RouteName    string
		BookedBig    int
		BookedMedium int
		BookedSmall  int
	}
	var rows []row
	err := r.tx(ctx).
		Table("bookings").
		Select(`date, time, route_name,
		        COALESCE(SUM(qty_big), 0)    AS booked_big,
		        COALESCE(SUM(qty_medium), 0) AS booked_medium,
		        COALESCE(SUM(qty_small), 0)  AS booked_small`).
		Where("date >= ? AND date < ?", startDate, endDate).
		Where("status = ? OR (status = ? AND expires_at > NOW())",
			model.StatusConfirmed, model.StatusPending).
		Group("date, time, route_name").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]map[SlotKey]BookedQty, len(rows))
	for _, rw := range rows {
		dateStr := rw.Date.UTC().Format("2006-01-02")
		inner, ok := out[dateStr]
		if !ok {
			inner = make(map[SlotKey]BookedQty, 8)
			out[dateStr] = inner
		}
		inner[SlotKey{Time: rw.Time, Route: rw.RouteName}] =
			BookedQty{Big: rw.BookedBig, Medium: rw.BookedMedium, Small: rw.BookedSmall}
	}
	return out, nil
}

func (r *bookingRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Booking, error) {
	var bookings []model.Booking
	err := r.tx(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepo) ExpirePending(ctx context.Context, now time.Time) (int64, error) {
	res := r.tx(ctx).
		Model(&model.Booking{}).
		Where("status = ? AND expires_at < ?", model.StatusPending, now).
		Updates(map[string]any{"status": model.StatusExpired})
	return res.RowsAffected, res.Error
}

func (r *bookingRepo) FindBySlot(ctx context.Context, date, time_, route string) ([]model.Booking, error) {
	var bookings []model.Booking
	err := r.tx(ctx).
		Where("date = ? AND time = ? AND route_name = ?", date, time_, route).
		Order("created_at ASC").
		Find(&bookings).Error
	return bookings, err
}

func (r *bookingRepo) SetPosterIDs(ctx context.Context, id uuid.UUID, orderID, txID int64) error {
	return r.tx(ctx).
		Model(&model.Booking{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"poster_incoming_order_id":       orderID,
			"poster_incoming_transaction_id": txID,
		}).Error
}
