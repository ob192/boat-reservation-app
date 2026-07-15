package service

// Shared test doubles for the service package.
//
// Mocks are handwritten function-field fakes: a nil field panics with the
// method name, so any unexpected repository call fails the test loudly and
// doubles as an implicit "this path must not be reached" assertion.

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/provider"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
)

// ----------------------------------------------------------------------------
// Clock
// ----------------------------------------------------------------------------

type fakeClock struct{ t time.Time }

func (c fakeClock) Now() time.Time { return c.t }

// testNow is the frozen "current time" used across service tests.
var testNow = time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)

// ----------------------------------------------------------------------------
// Fake *gorm.DB — supports Begin/Commit/Rollback only, no queries.
// platform.WithTx works against it while all data access goes through mocks.
// ----------------------------------------------------------------------------

type fakeConnector struct{}

func (fakeConnector) Connect(context.Context) (driver.Conn, error) { return fakeConn{}, nil }
func (fakeConnector) Driver() driver.Driver                        { return fakeDriver{} }

type fakeDriver struct{}

func (fakeDriver) Open(string) (driver.Conn, error) { return fakeConn{}, nil }

type fakeConn struct{}

func (fakeConn) Prepare(string) (driver.Stmt, error) {
	return nil, fmt.Errorf("fake driver: queries not supported")
}
func (fakeConn) Close() error              { return nil }
func (fakeConn) Begin() (driver.Tx, error) { return fakeTx{}, nil }

type fakeTx struct{}

func (fakeTx) Commit() error   { return nil }
func (fakeTx) Rollback() error { return nil }

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sql.OpenDB(fakeConnector{})}), &gorm.Config{
		Logger: logger.Discard,
	})
	if err != nil {
		t.Fatalf("open fake gorm db: %v", err)
	}
	return db
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// ----------------------------------------------------------------------------
// Pointer / model helpers
// ----------------------------------------------------------------------------

func sptr(s string) *string    { return &s }
func iptr(i int) *int          { return &i }
func fptr(f float64) *float64  { return &f }
func bptr(b bool) *bool        { return &b }
func pgDate(s string) pgtype.Date {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return pgtype.Date{Time: t, Valid: true}
}

var _ = bptr // used by some test files only

// mkBooking returns a minimal pending booking on Desna 2026-08-01 07:00.
func mkBooking() *model.Booking {
	return &model.Booking{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		UserEmail:   "user@example.com",
		Date:        pgDate("2026-08-01"),
		Time:        "07:00",
		RouteName:   RouteDesna,
		QtyBig:      1,
		FirstName:   "Olena",
		LastName:    "Kovalenko",
		Phone:       sptr("+380501112233"),
		TotalAmount: 450,
		Status:      model.StatusPending,
		ExpiresAt:   testNow.Add(HoldDuration),
		CreatedAt:   testNow,
		UpdatedAt:   testNow,
	}
}

func mkSlot(date, tm, route string, big, medium, small int) *model.Slot {
	return &model.Slot{
		Date:           pgDate(date),
		Time:           tm,
		RouteName:      route,
		CapacityBig:    big,
		CapacityMedium: medium,
		CapacitySmall:  small,
	}
}

// ----------------------------------------------------------------------------
// repository.BookingRepository mock
// ----------------------------------------------------------------------------

type mockBookingRepo struct {
	create                 func(ctx context.Context, b *model.Booking) error
	findByUserID           func(ctx context.Context, userID uuid.UUID) ([]model.Booking, error)
	findByID               func(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	findByPaymentSessionID func(ctx context.Context, sessionID string) (*model.Booking, error)
	findByIdempotencyKey   func(ctx context.Context, userID uuid.UUID, key, date, time, route string) (*model.Booking, error)
	setStatus              func(ctx context.Context, id uuid.UUID, status model.BookingStatus) error
	setPaymentSessionID    func(ctx context.Context, id uuid.UUID, sessionID string) error
	setPriceOverride       func(ctx context.Context, id uuid.UUID, override *float64, reason *string, adminID *uuid.UUID) error
	cancel                 func(ctx context.Context, id uuid.UUID, adminID uuid.UUID, reason string) error
	sumForSlot             func(ctx context.Context, date, time, route string) (int, int, int, error)
	sumForDate             func(ctx context.Context, date string) (map[repository.SlotKey]repository.BookedQty, error)
	sumForRange            func(ctx context.Context, startDate, endDate string) (map[string]map[repository.SlotKey]repository.BookedQty, error)
	expirePending          func(ctx context.Context, now time.Time) (int64, error)
	findBySlot             func(ctx context.Context, date, time, route string) ([]model.Booking, error)
	setPosterIDs           func(ctx context.Context, id uuid.UUID, orderID, txID int64) error
	findAllForAdmin        func(ctx context.Context, f repository.BookingHistoryFilter) ([]model.Booking, error)
	cancelBySlot           func(ctx context.Context, date, time, route string, adminID uuid.UUID, reason string) (int64, error)
	move                   func(ctx context.Context, id uuid.UUID, date, time, route string) error
}

func (m *mockBookingRepo) Create(ctx context.Context, b *model.Booking) error {
	if m.create == nil {
		panic("unexpected call: BookingRepository.Create")
	}
	return m.create(ctx, b)
}

func (m *mockBookingRepo) FindByUserID(ctx context.Context, userID uuid.UUID) ([]model.Booking, error) {
	if m.findByUserID == nil {
		panic("unexpected call: BookingRepository.FindByUserID")
	}
	return m.findByUserID(ctx, userID)
}

func (m *mockBookingRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	if m.findByID == nil {
		panic("unexpected call: BookingRepository.FindByID")
	}
	return m.findByID(ctx, id)
}

func (m *mockBookingRepo) FindByPaymentSessionID(ctx context.Context, sessionID string) (*model.Booking, error) {
	if m.findByPaymentSessionID == nil {
		panic("unexpected call: BookingRepository.FindByPaymentSessionID")
	}
	return m.findByPaymentSessionID(ctx, sessionID)
}

func (m *mockBookingRepo) FindByIdempotencyKey(ctx context.Context, userID uuid.UUID, key, date, time, route string) (*model.Booking, error) {
	if m.findByIdempotencyKey == nil {
		panic("unexpected call: BookingRepository.FindByIdempotencyKey")
	}
	return m.findByIdempotencyKey(ctx, userID, key, date, time, route)
}

func (m *mockBookingRepo) SetStatus(ctx context.Context, id uuid.UUID, status model.BookingStatus) error {
	if m.setStatus == nil {
		panic("unexpected call: BookingRepository.SetStatus")
	}
	return m.setStatus(ctx, id, status)
}

func (m *mockBookingRepo) SetPaymentSessionID(ctx context.Context, id uuid.UUID, sessionID string) error {
	if m.setPaymentSessionID == nil {
		panic("unexpected call: BookingRepository.SetPaymentSessionID")
	}
	return m.setPaymentSessionID(ctx, id, sessionID)
}

func (m *mockBookingRepo) SetPriceOverride(ctx context.Context, id uuid.UUID, override *float64, reason *string, adminID *uuid.UUID) error {
	if m.setPriceOverride == nil {
		panic("unexpected call: BookingRepository.SetPriceOverride")
	}
	return m.setPriceOverride(ctx, id, override, reason, adminID)
}

func (m *mockBookingRepo) Cancel(ctx context.Context, id uuid.UUID, adminID uuid.UUID, reason string) error {
	if m.cancel == nil {
		panic("unexpected call: BookingRepository.Cancel")
	}
	return m.cancel(ctx, id, adminID, reason)
}

func (m *mockBookingRepo) SumActiveQuantitiesForSlot(ctx context.Context, date, time, route string) (int, int, int, error) {
	if m.sumForSlot == nil {
		panic("unexpected call: BookingRepository.SumActiveQuantitiesForSlot")
	}
	return m.sumForSlot(ctx, date, time, route)
}

func (m *mockBookingRepo) SumActiveQuantitiesForDate(ctx context.Context, date string) (map[repository.SlotKey]repository.BookedQty, error) {
	if m.sumForDate == nil {
		panic("unexpected call: BookingRepository.SumActiveQuantitiesForDate")
	}
	return m.sumForDate(ctx, date)
}

func (m *mockBookingRepo) SumActiveQuantitiesForRange(ctx context.Context, startDate, endDate string) (map[string]map[repository.SlotKey]repository.BookedQty, error) {
	if m.sumForRange == nil {
		panic("unexpected call: BookingRepository.SumActiveQuantitiesForRange")
	}
	return m.sumForRange(ctx, startDate, endDate)
}

func (m *mockBookingRepo) ExpirePending(ctx context.Context, now time.Time) (int64, error) {
	if m.expirePending == nil {
		panic("unexpected call: BookingRepository.ExpirePending")
	}
	return m.expirePending(ctx, now)
}

func (m *mockBookingRepo) FindBySlot(ctx context.Context, date, time, route string) ([]model.Booking, error) {
	if m.findBySlot == nil {
		panic("unexpected call: BookingRepository.FindBySlot")
	}
	return m.findBySlot(ctx, date, time, route)
}

func (m *mockBookingRepo) SetPosterIDs(ctx context.Context, id uuid.UUID, orderID, txID int64) error {
	if m.setPosterIDs == nil {
		panic("unexpected call: BookingRepository.SetPosterIDs")
	}
	return m.setPosterIDs(ctx, id, orderID, txID)
}

func (m *mockBookingRepo) FindAllForAdmin(ctx context.Context, f repository.BookingHistoryFilter) ([]model.Booking, error) {
	if m.findAllForAdmin == nil {
		panic("unexpected call: BookingRepository.FindAllForAdmin")
	}
	return m.findAllForAdmin(ctx, f)
}

func (m *mockBookingRepo) CancelBySlot(ctx context.Context, date, time, route string, adminID uuid.UUID, reason string) (int64, error) {
	if m.cancelBySlot == nil {
		panic("unexpected call: BookingRepository.CancelBySlot")
	}
	return m.cancelBySlot(ctx, date, time, route, adminID, reason)
}

func (m *mockBookingRepo) Move(ctx context.Context, id uuid.UUID, date, time, route string) error {
	if m.move == nil {
		panic("unexpected call: BookingRepository.Move")
	}
	return m.move(ctx, id, date, time, route)
}

// ----------------------------------------------------------------------------
// repository.SlotRepository mock
// ----------------------------------------------------------------------------

type mockSlotRepo struct {
	findByDateTime func(ctx context.Context, date, time, route string) (*model.Slot, error)
	findByDate     func(ctx context.Context, date, route string) ([]model.Slot, error)
	findByMonth    func(ctx context.Context, monthStart, monthEnd time.Time, route string) ([]model.Slot, error)
	lockForUpdate  func(ctx context.Context, date, time, route string) (*model.Slot, error)
	block          func(ctx context.Context, date, time, route string, adminID uuid.UUID, reason *string) error
	unblock        func(ctx context.Context, date, time, route string) error
	cancel         func(ctx context.Context, date, time, route string, adminID uuid.UUID, reason *string) error
	uncancel       func(ctx context.Context, date, time, route string) error
	upsert         func(ctx context.Context, s model.Slot) (*model.Slot, bool, error)
	del            func(ctx context.Context, date, time, route string) error
}

func (m *mockSlotRepo) FindByDateTime(ctx context.Context, date, time, route string) (*model.Slot, error) {
	if m.findByDateTime == nil {
		panic("unexpected call: SlotRepository.FindByDateTime")
	}
	return m.findByDateTime(ctx, date, time, route)
}

func (m *mockSlotRepo) FindByDate(ctx context.Context, date, route string) ([]model.Slot, error) {
	if m.findByDate == nil {
		panic("unexpected call: SlotRepository.FindByDate")
	}
	return m.findByDate(ctx, date, route)
}

func (m *mockSlotRepo) FindByMonth(ctx context.Context, monthStart, monthEnd time.Time, route string) ([]model.Slot, error) {
	if m.findByMonth == nil {
		panic("unexpected call: SlotRepository.FindByMonth")
	}
	return m.findByMonth(ctx, monthStart, monthEnd, route)
}

func (m *mockSlotRepo) LockForUpdate(ctx context.Context, date, time, route string) (*model.Slot, error) {
	if m.lockForUpdate == nil {
		panic("unexpected call: SlotRepository.LockForUpdate")
	}
	return m.lockForUpdate(ctx, date, time, route)
}

func (m *mockSlotRepo) Block(ctx context.Context, date, time, route string, adminID uuid.UUID, reason *string) error {
	if m.block == nil {
		panic("unexpected call: SlotRepository.Block")
	}
	return m.block(ctx, date, time, route, adminID, reason)
}

func (m *mockSlotRepo) Unblock(ctx context.Context, date, time, route string) error {
	if m.unblock == nil {
		panic("unexpected call: SlotRepository.Unblock")
	}
	return m.unblock(ctx, date, time, route)
}

func (m *mockSlotRepo) Cancel(ctx context.Context, date, time, route string, adminID uuid.UUID, reason *string) error {
	if m.cancel == nil {
		panic("unexpected call: SlotRepository.Cancel")
	}
	return m.cancel(ctx, date, time, route, adminID, reason)
}

func (m *mockSlotRepo) Uncancel(ctx context.Context, date, time, route string) error {
	if m.uncancel == nil {
		panic("unexpected call: SlotRepository.Uncancel")
	}
	return m.uncancel(ctx, date, time, route)
}

func (m *mockSlotRepo) Upsert(ctx context.Context, s model.Slot) (*model.Slot, bool, error) {
	if m.upsert == nil {
		panic("unexpected call: SlotRepository.Upsert")
	}
	return m.upsert(ctx, s)
}

func (m *mockSlotRepo) Delete(ctx context.Context, date, time, route string) error {
	if m.del == nil {
		panic("unexpected call: SlotRepository.Delete")
	}
	return m.del(ctx, date, time, route)
}

// ----------------------------------------------------------------------------
// repository.DateBlockRepository mock
// ----------------------------------------------------------------------------

type mockDateBlockRepo struct {
	find            func(ctx context.Context, date string) (*model.DateBlock, error)
	create          func(ctx context.Context, db *model.DateBlock) error
	del             func(ctx context.Context, date string) error
	findManyInRange func(ctx context.Context, startDate, endDate string) ([]model.DateBlock, error)
}

func (m *mockDateBlockRepo) Find(ctx context.Context, date string) (*model.DateBlock, error) {
	if m.find == nil {
		panic("unexpected call: DateBlockRepository.Find")
	}
	return m.find(ctx, date)
}

func (m *mockDateBlockRepo) Create(ctx context.Context, db *model.DateBlock) error {
	if m.create == nil {
		panic("unexpected call: DateBlockRepository.Create")
	}
	return m.create(ctx, db)
}

func (m *mockDateBlockRepo) Delete(ctx context.Context, date string) error {
	if m.del == nil {
		panic("unexpected call: DateBlockRepository.Delete")
	}
	return m.del(ctx, date)
}

func (m *mockDateBlockRepo) FindManyInRange(ctx context.Context, startDate, endDate string) ([]model.DateBlock, error) {
	if m.findManyInRange == nil {
		panic("unexpected call: DateBlockRepository.FindManyInRange")
	}
	return m.findManyInRange(ctx, startDate, endDate)
}

// noDateBlocks is the common "nothing blocked" stub.
func noDateBlocks() *mockDateBlockRepo {
	return &mockDateBlockRepo{
		find: func(context.Context, string) (*model.DateBlock, error) {
			return nil, repository.ErrNotFound
		},
	}
}

// ----------------------------------------------------------------------------
// repository.SystemRepository mock
// ----------------------------------------------------------------------------

type mockSystemRepo struct {
	get                func(ctx context.Context) (*model.SystemSettings, error)
	setBookingsEnabled func(ctx context.Context, enabled bool, reason *string, adminID uuid.UUID) (*model.SystemSettings, error)
}

func (m *mockSystemRepo) Get(ctx context.Context) (*model.SystemSettings, error) {
	if m.get == nil {
		panic("unexpected call: SystemRepository.Get")
	}
	return m.get(ctx)
}

func (m *mockSystemRepo) SetBookingsEnabled(ctx context.Context, enabled bool, reason *string, adminID uuid.UUID) (*model.SystemSettings, error) {
	if m.setBookingsEnabled == nil {
		panic("unexpected call: SystemRepository.SetBookingsEnabled")
	}
	return m.setBookingsEnabled(ctx, enabled, reason, adminID)
}

func bookingsEnabled() *mockSystemRepo {
	return &mockSystemRepo{
		get: func(context.Context) (*model.SystemSettings, error) {
			return &model.SystemSettings{ID: 1, BookingsEnabled: true}, nil
		},
	}
}

func bookingsDisabled() *mockSystemRepo {
	return &mockSystemRepo{
		get: func(context.Context) (*model.SystemSettings, error) {
			return &model.SystemSettings{ID: 1, BookingsEnabled: false}, nil
		},
	}
}

// ----------------------------------------------------------------------------
// repository.PromocodeRepository mock
// ----------------------------------------------------------------------------

type mockPromoRepo struct {
	create         func(ctx context.Context, p *model.Promocode) error
	findByCode     func(ctx context.Context, code string) (*model.Promocode, error)
	findAll        func(ctx context.Context, createdBy *uuid.UUID) ([]model.Promocode, error)
	incrementUsage func(ctx context.Context, code string) (bool, error)
}

func (m *mockPromoRepo) Create(ctx context.Context, p *model.Promocode) error {
	if m.create == nil {
		panic("unexpected call: PromocodeRepository.Create")
	}
	return m.create(ctx, p)
}

func (m *mockPromoRepo) FindByCode(ctx context.Context, code string) (*model.Promocode, error) {
	if m.findByCode == nil {
		panic("unexpected call: PromocodeRepository.FindByCode")
	}
	return m.findByCode(ctx, code)
}

func (m *mockPromoRepo) FindAll(ctx context.Context, createdBy *uuid.UUID) ([]model.Promocode, error) {
	if m.findAll == nil {
		panic("unexpected call: PromocodeRepository.FindAll")
	}
	return m.findAll(ctx, createdBy)
}

func (m *mockPromoRepo) IncrementUsage(ctx context.Context, code string) (bool, error) {
	if m.incrementUsage == nil {
		panic("unexpected call: PromocodeRepository.IncrementUsage")
	}
	return m.incrementUsage(ctx, code)
}

// ----------------------------------------------------------------------------
// service.PromocodeService mock (for booking service tests)
// ----------------------------------------------------------------------------

type mockPromocodeService struct {
	create   func(ctx context.Context, code string, discountPercent, maxUses int, adminID uuid.UUID) (*model.AdminPromocodeResponse, error)
	list     func(ctx context.Context, createdBy *uuid.UUID) (*model.AdminPromocodeListResponse, error)
	validate func(ctx context.Context, code string) (*model.Promocode, error)
}

func (m *mockPromocodeService) Create(ctx context.Context, code string, discountPercent, maxUses int, adminID uuid.UUID) (*model.AdminPromocodeResponse, error) {
	if m.create == nil {
		panic("unexpected call: PromocodeService.Create")
	}
	return m.create(ctx, code, discountPercent, maxUses, adminID)
}

func (m *mockPromocodeService) List(ctx context.Context, createdBy *uuid.UUID) (*model.AdminPromocodeListResponse, error) {
	if m.list == nil {
		panic("unexpected call: PromocodeService.List")
	}
	return m.list(ctx, createdBy)
}

func (m *mockPromocodeService) Validate(ctx context.Context, code string) (*model.Promocode, error) {
	if m.validate == nil {
		panic("unexpected call: PromocodeService.Validate")
	}
	return m.validate(ctx, code)
}

// ----------------------------------------------------------------------------
// provider.PaymentGateway mock
// ----------------------------------------------------------------------------

type mockGateway struct {
	createSession func(ctx context.Context, req provider.CreateSessionRequest) (provider.CreateSessionResponse, error)
	parseWebhook  func(ctx context.Context, rawBody []byte, headers map[string]string) (provider.WebhookEvent, error)
}

func (m *mockGateway) CreateSession(ctx context.Context, req provider.CreateSessionRequest) (provider.CreateSessionResponse, error) {
	if m.createSession == nil {
		panic("unexpected call: PaymentGateway.CreateSession")
	}
	return m.createSession(ctx, req)
}

func (m *mockGateway) ParseWebhook(ctx context.Context, rawBody []byte, headers map[string]string) (provider.WebhookEvent, error) {
	if m.parseWebhook == nil {
		panic("unexpected call: PaymentGateway.ParseWebhook")
	}
	return m.parseWebhook(ctx, rawBody, headers)
}

// ----------------------------------------------------------------------------
// provider.PosterClient mock
// ----------------------------------------------------------------------------

type mockPosterClient struct {
	createIncomingOrder func(ctx context.Context, order provider.PosterOrder) (provider.PosterOrderResult, error)
}

func (m *mockPosterClient) CreateIncomingOrder(ctx context.Context, order provider.PosterOrder) (provider.PosterOrderResult, error) {
	if m.createIncomingOrder == nil {
		panic("unexpected call: PosterClient.CreateIncomingOrder")
	}
	return m.createIncomingOrder(ctx, order)
}

// ----------------------------------------------------------------------------
// service.EmailService mock
// ----------------------------------------------------------------------------

type mockEmailService struct {
	sendConfirmation func(b *model.Booking)
}

func (m *mockEmailService) SendConfirmation(b *model.Booking) {
	if m.sendConfirmation == nil {
		panic("unexpected call: EmailService.SendConfirmation")
	}
	m.sendConfirmation(b)
}
