package handler

// Shared gin/httptest plumbing and service mocks for handler tests.
//
// Handlers are exercised through a real gin router so path params, JSON
// binding, and middleware ordering behave exactly as in production. Service
// mocks follow the function-field pattern: a nil field panics with the method
// name, turning any unexpected downstream call into a loud test failure.

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/provider"
	"github.com/harbour-wave/harbour-wave-backend/internal/service"
	"github.com/harbour-wave/harbour-wave-backend/pkg/httpx"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func testUser() *model.AuthUser {
	return &model.AuthUser{ID: uuid.New(), Email: "user@example.com", AppMetadata: map[string]any{}}
}

// withUser is middleware that plants an authenticated user like AuthMiddleware would.
func withUser(u *model.AuthUser) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(ContextUserDetails, u)
		c.Next()
	}
}

// doRequest performs one request against r and returns the recorder.
func doRequest(t *testing.T, r http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// decodeError unpacks the standard error body.
func decodeError(t *testing.T, w *httptest.ResponseRecorder) httpx.ErrorBody {
	t.Helper()
	var e httpx.ErrorBody
	if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil {
		t.Fatalf("error body is not JSON: %v (%s)", err, w.Body.String())
	}
	return e
}

// wantError asserts status code + error message code in one shot.
func wantError(t *testing.T, w *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if w.Code != status {
		t.Errorf("status = %d, want %d (body %s)", w.Code, status, w.Body.String())
	}
	if e := decodeError(t, w); e.Message != code {
		t.Errorf("error code = %q, want %q", e.Message, code)
	}
}

func decodeJSON[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("body is not %T: %v (%s)", v, err, w.Body.String())
	}
	return v
}

func sptr(s string) *string   { return &s }
func iptr(i int) *int         { return &i }
func fptr(f float64) *float64 { return &f }

// ----------------------------------------------------------------------------
// service.BookingService mock
// ----------------------------------------------------------------------------

type mockBookingService struct {
	create         func(ctx context.Context, in service.CreateBookingInput) (*model.Booking, error)
	getByID        func(ctx context.Context, id uuid.UUID) (*model.Booking, error)
	getByIDForUser func(ctx context.Context, id, userID uuid.UUID) (*model.Booking, error)
	getBySession   func(ctx context.Context, sessionID string, userID uuid.UUID) (*model.Booking, error)
	getAllForUser  func(ctx context.Context, userID uuid.UUID) ([]model.Booking, error)
	expireStale    func(ctx context.Context) (int64, error)
}

func (m *mockBookingService) Create(ctx context.Context, in service.CreateBookingInput) (*model.Booking, error) {
	if m.create == nil {
		panic("unexpected call: BookingService.Create")
	}
	return m.create(ctx, in)
}

func (m *mockBookingService) GetByID(ctx context.Context, id uuid.UUID) (*model.Booking, error) {
	if m.getByID == nil {
		panic("unexpected call: BookingService.GetByID")
	}
	return m.getByID(ctx, id)
}

func (m *mockBookingService) GetByIDForUser(ctx context.Context, id, userID uuid.UUID) (*model.Booking, error) {
	if m.getByIDForUser == nil {
		panic("unexpected call: BookingService.GetByIDForUser")
	}
	return m.getByIDForUser(ctx, id, userID)
}

func (m *mockBookingService) GetBySession(ctx context.Context, sessionID string, userID uuid.UUID) (*model.Booking, error) {
	if m.getBySession == nil {
		panic("unexpected call: BookingService.GetBySession")
	}
	return m.getBySession(ctx, sessionID, userID)
}

func (m *mockBookingService) GetAllForUser(ctx context.Context, userID uuid.UUID) ([]model.Booking, error) {
	if m.getAllForUser == nil {
		panic("unexpected call: BookingService.GetAllForUser")
	}
	return m.getAllForUser(ctx, userID)
}

func (m *mockBookingService) ExpireStale(ctx context.Context) (int64, error) {
	if m.expireStale == nil {
		panic("unexpected call: BookingService.ExpireStale")
	}
	return m.expireStale(ctx)
}

// ----------------------------------------------------------------------------
// service.CheckoutService mock
// ----------------------------------------------------------------------------

type mockCheckoutService struct {
	createCheckout func(ctx context.Context, bookingID, userID uuid.UUID, successURL string) (*model.CreateCheckoutResponse, error)
}

func (m *mockCheckoutService) CreateCheckout(ctx context.Context, bookingID, userID uuid.UUID, successURL string) (*model.CreateCheckoutResponse, error) {
	if m.createCheckout == nil {
		panic("unexpected call: CheckoutService.CreateCheckout")
	}
	return m.createCheckout(ctx, bookingID, userID, successURL)
}

// ----------------------------------------------------------------------------
// service.AvailabilityService mock
// ----------------------------------------------------------------------------

type mockAvailabilityService struct {
	getStatus           func(ctx context.Context) (*model.BookingStatusResponse, error)
	getMonth            func(ctx context.Context, month, route string) (*model.AvailabilityMonthResponse, error)
	getDate             func(ctx context.Context, date, route string) (*model.SlotsForDateResponse, error)
	getSlotAvailability func(ctx context.Context, date, time, route string) (*model.SlotAvailability, error)
}

func (m *mockAvailabilityService) GetStatus(ctx context.Context) (*model.BookingStatusResponse, error) {
	if m.getStatus == nil {
		panic("unexpected call: AvailabilityService.GetStatus")
	}
	return m.getStatus(ctx)
}

func (m *mockAvailabilityService) GetMonth(ctx context.Context, month, route string) (*model.AvailabilityMonthResponse, error) {
	if m.getMonth == nil {
		panic("unexpected call: AvailabilityService.GetMonth")
	}
	return m.getMonth(ctx, month, route)
}

func (m *mockAvailabilityService) GetDate(ctx context.Context, date, route string) (*model.SlotsForDateResponse, error) {
	if m.getDate == nil {
		panic("unexpected call: AvailabilityService.GetDate")
	}
	return m.getDate(ctx, date, route)
}

func (m *mockAvailabilityService) GetSlotAvailability(ctx context.Context, date, time, route string) (*model.SlotAvailability, error) {
	if m.getSlotAvailability == nil {
		panic("unexpected call: AvailabilityService.GetSlotAvailability")
	}
	return m.getSlotAvailability(ctx, date, time, route)
}

// ----------------------------------------------------------------------------
// service.PromocodeService mock
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
// service.AdminService mock
// ----------------------------------------------------------------------------

type mockAdminService struct {
	overridePrice      func(ctx context.Context, bookingID, adminID uuid.UUID, amount *float64, reason string) (*model.AdminPriceOverrideResponse, error)
	blockSlot          func(ctx context.Context, date, time, route string, adminID uuid.UUID, reason string) (*model.AdminBlockSlotResponse, error)
	unblockSlot        func(ctx context.Context, date, time, route string) (*model.AdminUnblockSlotResponse, error)
	blockDate          func(ctx context.Context, date string, adminID uuid.UUID, reason string) (*model.AdminBlockDateResponse, error)
	unblockDate        func(ctx context.Context, date string) (*model.AdminUnblockDateResponse, error)
	setBookingsEnabled func(ctx context.Context, enabled bool, reason string, adminID uuid.UUID) (*model.AdminSetBookingsEnabledResponse, error)
	cancelBooking      func(ctx context.Context, bookingID, adminID uuid.UUID, reason string) (*model.AdminCancelBookingResponse, error)
	cancelSlot         func(ctx context.Context, date, time, route string, adminID uuid.UUID, reason string) (*model.AdminCancelSlotResponse, error)
	uncancelSlot       func(ctx context.Context, date, time, route string) (*model.AdminUncancelSlotResponse, error)
	upsertSlot         func(ctx context.Context, date, time, route string, capacityBig, capacityMedium, capacitySmall int, adminID uuid.UUID) (*model.AdminUpsertSlotResponse, error)
	getSlotBookings    func(ctx context.Context, date, time, route string) (*model.AdminSlotBookingsResponse, error)
	listBookings       func(ctx context.Context, date, status string, limit, offset int) (*model.AdminBookingHistoryResponse, error)
	moveBooking        func(ctx context.Context, bookingID uuid.UUID, date, slotTime, route string) (*model.AdminMoveBookingResponse, error)
	deleteSlot         func(ctx context.Context, date, time, route string) error
}

func (m *mockAdminService) OverridePrice(ctx context.Context, bookingID, adminID uuid.UUID, amount *float64, reason string) (*model.AdminPriceOverrideResponse, error) {
	if m.overridePrice == nil {
		panic("unexpected call: AdminService.OverridePrice")
	}
	return m.overridePrice(ctx, bookingID, adminID, amount, reason)
}

func (m *mockAdminService) BlockSlot(ctx context.Context, date, time, route string, adminID uuid.UUID, reason string) (*model.AdminBlockSlotResponse, error) {
	if m.blockSlot == nil {
		panic("unexpected call: AdminService.BlockSlot")
	}
	return m.blockSlot(ctx, date, time, route, adminID, reason)
}

func (m *mockAdminService) UnblockSlot(ctx context.Context, date, time, route string) (*model.AdminUnblockSlotResponse, error) {
	if m.unblockSlot == nil {
		panic("unexpected call: AdminService.UnblockSlot")
	}
	return m.unblockSlot(ctx, date, time, route)
}

func (m *mockAdminService) BlockDate(ctx context.Context, date string, adminID uuid.UUID, reason string) (*model.AdminBlockDateResponse, error) {
	if m.blockDate == nil {
		panic("unexpected call: AdminService.BlockDate")
	}
	return m.blockDate(ctx, date, adminID, reason)
}

func (m *mockAdminService) UnblockDate(ctx context.Context, date string) (*model.AdminUnblockDateResponse, error) {
	if m.unblockDate == nil {
		panic("unexpected call: AdminService.UnblockDate")
	}
	return m.unblockDate(ctx, date)
}

func (m *mockAdminService) SetBookingsEnabled(ctx context.Context, enabled bool, reason string, adminID uuid.UUID) (*model.AdminSetBookingsEnabledResponse, error) {
	if m.setBookingsEnabled == nil {
		panic("unexpected call: AdminService.SetBookingsEnabled")
	}
	return m.setBookingsEnabled(ctx, enabled, reason, adminID)
}

func (m *mockAdminService) CancelBooking(ctx context.Context, bookingID, adminID uuid.UUID, reason string) (*model.AdminCancelBookingResponse, error) {
	if m.cancelBooking == nil {
		panic("unexpected call: AdminService.CancelBooking")
	}
	return m.cancelBooking(ctx, bookingID, adminID, reason)
}

func (m *mockAdminService) CancelSlot(ctx context.Context, date, time, route string, adminID uuid.UUID, reason string) (*model.AdminCancelSlotResponse, error) {
	if m.cancelSlot == nil {
		panic("unexpected call: AdminService.CancelSlot")
	}
	return m.cancelSlot(ctx, date, time, route, adminID, reason)
}

func (m *mockAdminService) UncancelSlot(ctx context.Context, date, time, route string) (*model.AdminUncancelSlotResponse, error) {
	if m.uncancelSlot == nil {
		panic("unexpected call: AdminService.UncancelSlot")
	}
	return m.uncancelSlot(ctx, date, time, route)
}

func (m *mockAdminService) UpsertSlot(ctx context.Context, date, time, route string, capacityBig, capacityMedium, capacitySmall int, adminID uuid.UUID) (*model.AdminUpsertSlotResponse, error) {
	if m.upsertSlot == nil {
		panic("unexpected call: AdminService.UpsertSlot")
	}
	return m.upsertSlot(ctx, date, time, route, capacityBig, capacityMedium, capacitySmall, adminID)
}

func (m *mockAdminService) GetSlotBookings(ctx context.Context, date, time, route string) (*model.AdminSlotBookingsResponse, error) {
	if m.getSlotBookings == nil {
		panic("unexpected call: AdminService.GetSlotBookings")
	}
	return m.getSlotBookings(ctx, date, time, route)
}

func (m *mockAdminService) ListBookings(ctx context.Context, date, status string, limit, offset int) (*model.AdminBookingHistoryResponse, error) {
	if m.listBookings == nil {
		panic("unexpected call: AdminService.ListBookings")
	}
	return m.listBookings(ctx, date, status, limit, offset)
}

func (m *mockAdminService) MoveBooking(ctx context.Context, bookingID uuid.UUID, date, slotTime, route string) (*model.AdminMoveBookingResponse, error) {
	if m.moveBooking == nil {
		panic("unexpected call: AdminService.MoveBooking")
	}
	return m.moveBooking(ctx, bookingID, date, slotTime, route)
}

func (m *mockAdminService) DeleteSlot(ctx context.Context, date, time, route string) error {
	if m.deleteSlot == nil {
		panic("unexpected call: AdminService.DeleteSlot")
	}
	return m.deleteSlot(ctx, date, time, route)
}

// ----------------------------------------------------------------------------
// service.WebhookService, service.AuthService, provider.PaymentGateway,
// repository.AdminRepository mocks
// ----------------------------------------------------------------------------

type mockWebhookService struct {
	handle func(ctx context.Context, event provider.WebhookEvent) error
}

func (m *mockWebhookService) Handle(ctx context.Context, event provider.WebhookEvent) error {
	if m.handle == nil {
		panic("unexpected call: WebhookService.Handle")
	}
	return m.handle(ctx, event)
}

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

type mockAuthService struct {
	getUserByToken func(token string) (*model.AuthUser, error)
	getUserByID    func(userID uuid.UUID) (*model.AuthUser, error)
}

func (m *mockAuthService) GetUserByToken(token string) (*model.AuthUser, error) {
	if m.getUserByToken == nil {
		panic("unexpected call: AuthService.GetUserByToken")
	}
	return m.getUserByToken(token)
}

func (m *mockAuthService) GetUserByID(userID uuid.UUID) (*model.AuthUser, error) {
	if m.getUserByID == nil {
		panic("unexpected call: AuthService.GetUserByID")
	}
	return m.getUserByID(userID)
}

type mockAdminRepo struct {
	isAdmin func(ctx context.Context, userID uuid.UUID) (bool, error)
}

func (m *mockAdminRepo) IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	if m.isAdmin == nil {
		panic("unexpected call: AdminRepository.IsAdmin")
	}
	return m.isAdmin(ctx, userID)
}
