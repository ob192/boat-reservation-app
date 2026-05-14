package platform

import (
	"context"

	"gorm.io/gorm"
)

// txContextKey is used to attach an in-flight *gorm.DB transaction to a context.
type txContextKey struct{}

// WithTx runs fn inside a database transaction. If fn returns an error,
// the transaction is rolled back. Otherwise, it is committed.
//
// Inside fn, repositories should call DBFromContext(ctx) to get the
// transactional handle. This way the same repository code works both
// inside and outside transactions.
func WithTx(ctx context.Context, db *gorm.DB, fn func(ctx context.Context) error) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		ctxWithTx := context.WithValue(ctx, txContextKey{}, tx)
		return fn(ctxWithTx)
	})
}

// DBFromContext returns the transactional *gorm.DB if one is attached to ctx,
// otherwise it returns the supplied default db. Always pass through .WithContext(ctx)
// at the call site to honour cancellation.
func DBFromContext(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return defaultDB.WithContext(ctx)
}
