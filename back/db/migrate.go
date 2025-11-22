package db

import (
	"embed"
	"fmt"

	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(dbPool *pgxpool.Pool) error {
	slog.Info("Running database migrations...")

	config := dbPool.Config()
	db := stdlib.OpenDB(*config.ConnConfig)
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create source driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	migrateErr := m.Up()
	if migrateErr != nil && migrateErr != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", migrateErr)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if dirty {
		slog.Warn("Database is in dirty state, forcing version", "version", version)
		if err := m.Force(int(version)); err != nil {
			return fmt.Errorf("failed to force migration version: %w", err)
		}
	}

	if migrateErr == migrate.ErrNoChange {
		slog.Info("Database migrations are up to date")
	} else {
		slog.Info("Database migrations completed", "version", version)
	}

	return nil
}
