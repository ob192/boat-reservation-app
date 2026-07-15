package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/harbour-wave/harbour-wave-backend/internal/model"
	"github.com/harbour-wave/harbour-wave-backend/internal/repository"
)

// bookingFixture wires a bookingService whose collaborators all succeed for a
// happy-path create on Desna 2026-08-01 07:00. Individual tests then break one
// collaborator at a time.
type bookingFixture struct {
	bookings   *mockBookingRepo
	slots      *mockSlotRepo
	dateBlocks *mockDateBlockRepo
	system     *mockSystemRepo
	promocodes *mockPromocodeService
}

func newBookingFixture() *bookingFixture {
	return &bookingFixture{
		bookings: &mockBookingRepo{
			findByIdempotencyKey: func(context.Context, uuid.UUID, string, string, string, string) (*model.Booking, error) {
				return nil, repository.ErrNotFound
			},
			sumForSlot: func(context.Context, string, string, string) (int, int, int, error) {
				return 0, 0, 0, nil
			},
			create: func(context.Context, *model.Booking) error { return nil },
		},
		slots: &mockSlotRepo{
			lockForUpdate: func(_ context.Context, date, tm, route string) (*model.Slot, error) {
				return mkSlot(date, tm, route, 5, 5, 5), nil
			},
		},
		dateBlocks: noDateBlocks(),
		system:     bookingsEnabled(),
		promocodes: &mockPromocodeService{},
	}
}

func (f *bookingFixture) service(t *testing.T) BookingService {
	return NewBookingService(
		newTestDB(t),
		f.bookings,
		f.slots,
		f.dateBlocks,
		f.system,
		NewPricingService(),
		f.promocodes,
		fakeClock{t: testNow},
	)
}

func validInput() CreateBookingInput {
	return CreateBookingInput{
		UserID:         uuid.New(),
		UserEmail:      "user@example.com",
		IdempotencyKey: "idem-1",
		Date:           "2026-08-01",
		Time:           "07:00",
		RouteName:      RouteDesna,
		Quantities:     model.Quantities{Big: 1},
		FirstName:      "Olena",
		LastName:       "Kovalenko",
		Phone:          "+380501112233",
	}
}

func TestCreateBookingValidation(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*CreateBookingInput)
		wantErr error
	}{
		{"unknown route", func(in *CreateBookingInput) { in.RouteName = "Atlantis" }, ErrInvalidRoute},
		{"invalid time format", func(in *CreateBookingInput) { in.Time = "7:00" }, ErrInvalidInput},
		{"impossible time", func(in *CreateBookingInput) { in.Time = "25:00" }, ErrInvalidInput},
		{"garbage date", func(in *CreateBookingInput) { in.Date = "01-08-2026" }, ErrInvalidInput},
		{"date in the past", func(in *CreateBookingInput) { in.Date = "2026-07-14" }, ErrInvalidInput},
		{"negative quantity", func(in *CreateBookingInput) { in.Quantities.Medium = -1 }, ErrInvalidInput},
		{"no boats at all", func(in *CreateBookingInput) { in.Quantities = model.Quantities{} }, ErrValidationFailed},
		{"child seats only", func(in *CreateBookingInput) { in.Quantities = model.Quantities{Child: 2} }, ErrValidationFailed},
		{"child without big boat", func(in *CreateBookingInput) { in.Quantities = model.Quantities{Medium: 1, Child: 1} }, ErrValidationFailed},
		{"missing first name", func(in *CreateBookingInput) { in.FirstName = "" }, ErrInvalidInput},
		{"missing last name", func(in *CreateBookingInput) { in.LastName = "" }, ErrInvalidInput},
		{"missing phone", func(in *CreateBookingInput) { in.Phone = "" }, ErrInvalidInput},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// A zero-value fixture: validation must fail before ANY collaborator is touched,
			// and the nil-fn mocks panic if that contract is broken.
			f := &bookingFixture{
				bookings:   &mockBookingRepo{},
				slots:      &mockSlotRepo{},
				dateBlocks: &mockDateBlockRepo{},
				system:     &mockSystemRepo{},
				promocodes: &mockPromocodeService{},
			}
			in := validInput()
			tt.mutate(&in)
			_, err := f.service(t).Create(context.Background(), in)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("want %v, got %v", tt.wantErr, err)
			}
		})
	}

	t.Run("booking for today is allowed", func(t *testing.T) {
		f := newBookingFixture()
		in := validInput()
		in.Date = "2026-07-15" // == fake clock's today
		if _, err := f.service(t).Create(context.Background(), in); err != nil {
			t.Errorf("today must be bookable, got %v", err)
		}
	})
}

func TestCreateBookingKillSwitch(t *testing.T) {
	t.Run("disabled blocks creation", func(t *testing.T) {
		f := newBookingFixture()
		f.system = bookingsDisabled()
		_, err := f.service(t).Create(context.Background(), validInput())
		if !errors.Is(err, ErrBookingsDisabled) {
			t.Errorf("want ErrBookingsDisabled, got %v", err)
		}
	})

	t.Run("settings read failure propagates", func(t *testing.T) {
		f := newBookingFixture()
		boom := errors.New("db down")
		f.system = &mockSystemRepo{
			get: func(context.Context) (*model.SystemSettings, error) { return nil, boom },
		}
		if _, err := f.service(t).Create(context.Background(), validInput()); !errors.Is(err, boom) {
			t.Errorf("want wrapped settings error, got %v", err)
		}
	})
}

func TestCreateBookingDateBlock(t *testing.T) {
	t.Run("blocked date rejected", func(t *testing.T) {
		f := newBookingFixture()
		f.dateBlocks = &mockDateBlockRepo{
			find: func(_ context.Context, date string) (*model.DateBlock, error) {
				return &model.DateBlock{Date: date}, nil
			},
		}
		_, err := f.service(t).Create(context.Background(), validInput())
		if !errors.Is(err, ErrDateBlocked) {
			t.Errorf("want ErrDateBlocked, got %v", err)
		}
	})

	t.Run("date block lookup failure propagates", func(t *testing.T) {
		f := newBookingFixture()
		boom := errors.New("db down")
		f.dateBlocks = &mockDateBlockRepo{
			find: func(context.Context, string) (*model.DateBlock, error) { return nil, boom },
		}
		if _, err := f.service(t).Create(context.Background(), validInput()); !errors.Is(err, boom) {
			t.Errorf("want wrapped date-block error, got %v", err)
		}
	})
}

func TestCreateBookingIdempotency(t *testing.T) {
	t.Run("existing booking returned without creating another", func(t *testing.T) {
		f := newBookingFixture()
		existing := mkBooking()
		var gotKey string
		f.bookings.findByIdempotencyKey = func(_ context.Context, _ uuid.UUID, key, _, _, _ string) (*model.Booking, error) {
			gotKey = key
			return existing, nil
		}
		f.bookings.create = nil // must NOT be called

		got, err := f.service(t).Create(context.Background(), validInput())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != existing {
			t.Error("idempotent retry must return the original booking")
		}
		if gotKey != "idem-1" {
			t.Errorf("looked up key %q, want idem-1", gotKey)
		}
	})

	t.Run("idempotency lookup failure propagates", func(t *testing.T) {
		f := newBookingFixture()
		boom := errors.New("db down")
		f.bookings.findByIdempotencyKey = func(context.Context, uuid.UUID, string, string, string, string) (*model.Booking, error) {
			return nil, boom
		}
		if _, err := f.service(t).Create(context.Background(), validInput()); !errors.Is(err, boom) {
			t.Errorf("want wrapped idempotency error, got %v", err)
		}
	})
}

func TestCreateBookingPromocode(t *testing.T) {
	t.Run("invalid promo aborts before the capacity tx", func(t *testing.T) {
		for _, promoErr := range []error{ErrPromoNotFound, ErrPromoInactive, ErrPromoExhausted} {
			f := newBookingFixture()
			f.slots.lockForUpdate = nil // slot must never be locked
			f.promocodes.validate = func(context.Context, string) (*model.Promocode, error) {
				return nil, promoErr
			}
			in := validInput()
			in.PromoCode = "BAD"
			if _, err := f.service(t).Create(context.Background(), in); !errors.Is(err, promoErr) {
				t.Errorf("want %v, got %v", promoErr, err)
			}
		}
	})

	t.Run("discount is applied and snapshotted", func(t *testing.T) {
		f := newBookingFixture()
		var validated string
		f.promocodes.validate = func(_ context.Context, code string) (*model.Promocode, error) {
			validated = code
			return &model.Promocode{Code: "SUMMER", DiscountPercent: 10, MaxUses: 100, Active: true}, nil
		}
		var saved *model.Booking
		f.bookings.create = func(_ context.Context, b *model.Booking) error {
			saved = b
			return nil
		}

		in := validInput()
		in.PromoCode = "  summer " // service must normalise before validating
		in.Quantities = model.Quantities{Big: 2, Child: 1}

		got, err := f.service(t).Create(context.Background(), in)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if validated != "SUMMER" {
			t.Errorf("promo validated as %q, want normalised SUMMER", validated)
		}
		// list: 2×450 + 225 = 1125; 10% off → 1012.5 with 112.5 discount.
		if got.TotalAmount != 1012.5 {
			t.Errorf("total = %v, want 1012.5 (net of discount)", got.TotalAmount)
		}
		if saved.PromoCode == nil || *saved.PromoCode != "SUMMER" {
			t.Errorf("promo code not snapshotted: %v", saved.PromoCode)
		}
		if saved.DiscountPercent == nil || *saved.DiscountPercent != 10 {
			t.Errorf("discount percent not snapshotted: %v", saved.DiscountPercent)
		}
		if saved.DiscountAmount == nil || *saved.DiscountAmount != 112.5 {
			t.Errorf("discount amount not snapshotted: %v", saved.DiscountAmount)
		}
	})

	t.Run("blank promo code is ignored", func(t *testing.T) {
		f := newBookingFixture()
		f.promocodes.validate = nil // must not be called
		in := validInput()
		in.PromoCode = "   "
		got, err := f.service(t).Create(context.Background(), in)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.PromoCode != nil || got.DiscountAmount != nil {
			t.Errorf("no promo fields expected, got %+v", got)
		}
	})
}

func TestCreateBookingSlotChecks(t *testing.T) {
	slotErr := func(name string, setup func(*bookingFixture), want error) {
		t.Run(name, func(t *testing.T) {
			f := newBookingFixture()
			setup(f)
			if _, err := f.service(t).Create(context.Background(), validInput()); !errors.Is(err, want) {
				t.Errorf("want %v, got %v", want, err)
			}
		})
	}

	slotErr("slot missing", func(f *bookingFixture) {
		f.slots.lockForUpdate = func(context.Context, string, string, string) (*model.Slot, error) {
			return nil, repository.ErrNotFound
		}
	}, ErrSlotNotFound)

	slotErr("slot blocked", func(f *bookingFixture) {
		f.slots.lockForUpdate = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			s := mkSlot(d, tm, r, 5, 5, 5)
			s.Blocked = true
			return s, nil
		}
	}, ErrSlotBlocked)

	slotErr("slot cancelled", func(f *bookingFixture) {
		f.slots.lockForUpdate = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			s := mkSlot(d, tm, r, 5, 5, 5)
			s.Cancelled = true
			return s, nil
		}
	}, ErrSlotCancelled)

	boom := errors.New("lock timeout")
	slotErr("lock failure propagates", func(f *bookingFixture) {
		f.slots.lockForUpdate = func(context.Context, string, string, string) (*model.Slot, error) {
			return nil, boom
		}
	}, boom)

	slotErr("sum failure propagates", func(f *bookingFixture) {
		f.bookings.sumForSlot = func(context.Context, string, string, string) (int, int, int, error) {
			return 0, 0, 0, boom
		}
	}, boom)

	slotErr("insert failure propagates", func(f *bookingFixture) {
		f.bookings.create = func(context.Context, *model.Booking) error { return boom }
	}, boom)
}

func TestCreateBookingCapacity(t *testing.T) {
	// Slot: capacity 2/1/1, already booked 1/0/1.
	setup := func() *bookingFixture {
		f := newBookingFixture()
		f.slots.lockForUpdate = func(_ context.Context, d, tm, r string) (*model.Slot, error) {
			return mkSlot(d, tm, r, 2, 1, 1), nil
		}
		f.bookings.sumForSlot = func(context.Context, string, string, string) (int, int, int, error) {
			return 1, 0, 1, nil
		}
		return f
	}

	tests := []struct {
		name string
		q    model.Quantities
		ok   bool
	}{
		{"big exactly fits remaining capacity", model.Quantities{Big: 1}, true},
		{"big exceeds remaining capacity", model.Quantities{Big: 2}, false},
		{"medium exactly fits", model.Quantities{Medium: 1}, true},
		{"medium exceeds", model.Quantities{Medium: 2}, false},
		{"small fleet is full", model.Quantities{Small: 1}, false},
		{"fleets checked independently — big free, small full", model.Quantities{Big: 1, Small: 1}, false},
		{"child seats don't consume capacity", model.Quantities{Big: 1, Child: 5}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := setup()
			in := validInput()
			in.Quantities = tt.q
			_, err := f.service(t).Create(context.Background(), in)
			if tt.ok && err != nil {
				t.Errorf("want success, got %v", err)
			}
			if !tt.ok && !errors.Is(err, ErrSlotTaken) {
				t.Errorf("want ErrSlotTaken, got %v", err)
			}
		})
	}
}

func TestCreateBookingSuccess(t *testing.T) {
	f := newBookingFixture()
	var saved *model.Booking
	f.bookings.create = func(_ context.Context, b *model.Booking) error {
		saved = b
		return nil
	}

	in := validInput()
	in.Quantities = model.Quantities{Big: 1, Medium: 1, Child: 1}

	got, err := f.service(t).Create(context.Background(), in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != saved {
		t.Fatal("returned booking must be the persisted one")
	}

	if saved.ID == uuid.Nil {
		t.Error("booking must get a fresh id")
	}
	if saved.UserID != in.UserID || saved.UserEmail != in.UserEmail {
		t.Error("identity fields must come from the input (JWT), not defaults")
	}
	if saved.Status != model.StatusPending {
		t.Errorf("status = %s, want pending", saved.Status)
	}
	if saved.TotalAmount != 1125 { // 450 + 450 + 225
		t.Errorf("total = %v, want 1125", saved.TotalAmount)
	}
	if saved.DateFormatted() != "2026-08-01" || saved.Time != "07:00" || saved.RouteName != RouteDesna {
		t.Errorf("slot fields wrong: %s %s %s", saved.DateFormatted(), saved.Time, saved.RouteName)
	}
	if saved.QtyBig != 1 || saved.QtyMedium != 1 || saved.QtySmall != 0 || saved.QtyChild != 1 {
		t.Errorf("quantities wrong: %+v", saved.Quantities())
	}
	if saved.Phone == nil || *saved.Phone != in.Phone {
		t.Errorf("phone = %v, want %q", saved.Phone, in.Phone)
	}
	if saved.IdempotencyKey != "idem-1" {
		t.Errorf("idempotency key = %q", saved.IdempotencyKey)
	}
	if !saved.ExpiresAt.Equal(testNow.Add(HoldDuration)) {
		t.Errorf("expiresAt = %v, want now+HoldDuration %v", saved.ExpiresAt, testNow.Add(HoldDuration))
	}
	if !saved.CreatedAt.Equal(testNow) || !saved.UpdatedAt.Equal(testNow) {
		t.Errorf("timestamps must come from the injected clock: %v / %v", saved.CreatedAt, saved.UpdatedAt)
	}
	if saved.PromoCode != nil || saved.DiscountPercent != nil || saved.DiscountAmount != nil {
		t.Error("promo snapshot must stay nil without a promo code")
	}
}

func TestGetByID(t *testing.T) {
	id := uuid.New()

	t.Run("not found is typed", func(t *testing.T) {
		f := newBookingFixture()
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) {
			return nil, repository.ErrNotFound
		}
		if _, err := f.service(t).GetByID(context.Background(), id); !errors.Is(err, ErrBookingNotFound) {
			t.Errorf("want ErrBookingNotFound, got %v", err)
		}
	})

	t.Run("other errors pass through", func(t *testing.T) {
		f := newBookingFixture()
		boom := errors.New("db down")
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) { return nil, boom }
		if _, err := f.service(t).GetByID(context.Background(), id); !errors.Is(err, boom) {
			t.Errorf("want raw error, got %v", err)
		}
	})

	t.Run("found", func(t *testing.T) {
		f := newBookingFixture()
		b := mkBooking()
		f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) { return b, nil }
		got, err := f.service(t).GetByID(context.Background(), id)
		if err != nil || got != b {
			t.Errorf("got (%v, %v), want (%v, nil)", got, err, b)
		}
	})
}

func TestGetByIDForUser(t *testing.T) {
	f := newBookingFixture()
	owner := mkBooking()
	f.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) { return owner, nil }
	svc := f.service(t)

	t.Run("owner can read", func(t *testing.T) {
		got, err := svc.GetByIDForUser(context.Background(), owner.ID, owner.UserID)
		if err != nil || got != owner {
			t.Errorf("owner read failed: (%v, %v)", got, err)
		}
	})

	t.Run("stranger is forbidden", func(t *testing.T) {
		if _, err := svc.GetByIDForUser(context.Background(), owner.ID, uuid.New()); !errors.Is(err, ErrForbidden) {
			t.Errorf("want ErrForbidden, got %v", err)
		}
	})

	t.Run("not found propagates", func(t *testing.T) {
		f2 := newBookingFixture()
		f2.bookings.findByID = func(context.Context, uuid.UUID) (*model.Booking, error) {
			return nil, repository.ErrNotFound
		}
		if _, err := f2.service(t).GetByIDForUser(context.Background(), owner.ID, owner.UserID); !errors.Is(err, ErrBookingNotFound) {
			t.Errorf("want ErrBookingNotFound, got %v", err)
		}
	})
}

func TestGetBySession(t *testing.T) {
	owner := mkBooking()

	fixtureFor := func(find func(context.Context, string) (*model.Booking, error)) BookingService {
		f := newBookingFixture()
		f.bookings.findByPaymentSessionID = find
		return f.service(t)
	}

	t.Run("not found", func(t *testing.T) {
		svc := fixtureFor(func(context.Context, string) (*model.Booking, error) {
			return nil, repository.ErrNotFound
		})
		if _, err := svc.GetBySession(context.Background(), "sess_x", owner.UserID); !errors.Is(err, ErrBookingNotFound) {
			t.Errorf("want ErrBookingNotFound, got %v", err)
		}
	})

	t.Run("lookup failure", func(t *testing.T) {
		boom := errors.New("db down")
		svc := fixtureFor(func(context.Context, string) (*model.Booking, error) { return nil, boom })
		if _, err := svc.GetBySession(context.Background(), "sess_x", owner.UserID); !errors.Is(err, boom) {
			t.Errorf("want raw error, got %v", err)
		}
	})

	t.Run("stranger forbidden", func(t *testing.T) {
		svc := fixtureFor(func(context.Context, string) (*model.Booking, error) { return owner, nil })
		if _, err := svc.GetBySession(context.Background(), "sess_x", uuid.New()); !errors.Is(err, ErrForbidden) {
			t.Errorf("want ErrForbidden, got %v", err)
		}
	})

	t.Run("owner reads by session", func(t *testing.T) {
		var askedSession string
		svc := fixtureFor(func(_ context.Context, s string) (*model.Booking, error) {
			askedSession = s
			return owner, nil
		})
		got, err := svc.GetBySession(context.Background(), "sess_x", owner.UserID)
		if err != nil || got != owner {
			t.Errorf("got (%v, %v)", got, err)
		}
		if askedSession != "sess_x" {
			t.Errorf("queried session %q", askedSession)
		}
	})
}

func TestGetAllForUser(t *testing.T) {
	f := newBookingFixture()
	want := []model.Booking{*mkBooking()}
	var askedUser uuid.UUID
	f.bookings.findByUserID = func(_ context.Context, id uuid.UUID) ([]model.Booking, error) {
		askedUser = id
		return want, nil
	}
	userID := uuid.New()
	got, err := f.service(t).GetAllForUser(context.Background(), userID)
	if err != nil || len(got) != 1 {
		t.Fatalf("got (%v, %v)", got, err)
	}
	if askedUser != userID {
		t.Errorf("queried user %v, want %v", askedUser, userID)
	}
}

func TestExpireStale(t *testing.T) {
	f := newBookingFixture()
	var sweepTime time.Time
	f.bookings.expirePending = func(_ context.Context, now time.Time) (int64, error) {
		sweepTime = now
		return 3, nil
	}
	n, err := f.service(t).ExpireStale(context.Background())
	if err != nil || n != 3 {
		t.Fatalf("got (%d, %v), want (3, nil)", n, err)
	}
	if !sweepTime.Equal(testNow) {
		t.Errorf("sweep uses %v, want the injected clock %v", sweepTime, testNow)
	}
}
