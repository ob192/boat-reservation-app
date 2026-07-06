package platform

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDB opens a connection pool to PostgreSQL via GORM.
// Configures sensible pool defaults; caller is responsible for AutoMigrate.
func NewDB(databaseURL string, debug bool) (*gorm.DB, error) {
	logLevel := logger.Warn
	if debug {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}

// NewMigrationDB opens a short-lived, single-connection handle intended only for
// schema-changing operations (AutoMigrate) — never for serving requests. Callers
// must Close() the returned *gorm.DB (via its sql.DB) once migration finishes.
//
// This must point at a direct (non-pooled) database URL. Pooled endpoints
// (e.g. Neon/PgBouncer in transaction mode) can hand a DDL statement and a later
// query to different physical connections that share a stale server-side cached
// plan, which Postgres then rejects with "cached plan must not change result
// type" (SQLSTATE 0A000) as soon as AutoMigrate alters a table's columns.
func NewMigrationDB(directDatabaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(directDatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres (migration): %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB (migration): %w", err)
	}
	sqlDB.SetMaxOpenConns(1)

	return db, nil
}
