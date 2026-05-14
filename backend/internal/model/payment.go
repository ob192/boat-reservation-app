package model

// PaymentStatus mirrors provider.PaymentStatus at the model layer
// so services can refer to canonical constants without importing provider.
// Kept as separate string type to avoid cycles; conversion is trivial.
type PaymentStatus string

const (
	PaymentPaid    PaymentStatus = "paid"
	PaymentFailed  PaymentStatus = "failed"
	PaymentExpired PaymentStatus = "expired"
)
