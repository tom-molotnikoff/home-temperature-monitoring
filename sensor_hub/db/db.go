package database

import (
	"database/sql"
	appProps "example/sensorHub/application_properties"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	sqlite_migrate "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "modernc.org/sqlite"
)

func InitialiseDatabase() (*sql.DB, error) {
	if appProps.AppConfig == nil {
		return nil, fmt.Errorf("application configuration not loaded")
	}

	dbPath := appProps.AppConfig.DatabasePath

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("could not create database directory: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	// SQLite performs best with a single writer connection
	db.SetMaxOpenConns(1)

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("could not run migrations: %w", err)
	}

	log.Println("Connected to database")
	return db, nil
}

func runMigrations(db *sql.DB) error {
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("could not create migration source: %w", err)
	}

	dbDriver, err := sqlite_migrate.WithInstance(db, &sqlite_migrate.Config{
		NoTxWrap: true,
	})
	if err != nil {
		return fmt.Errorf("could not create migration db driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "sqlite", dbDriver)
	if err != nil {
		return fmt.Errorf("could not create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration failed: %w", err)
	}

	version, dirty, _ := m.Version()
	if dirty {
		return fmt.Errorf("database migration state is dirty at version %d", version)
	}

	log.Printf("Database schema at version %d", version)
	return nil
}
