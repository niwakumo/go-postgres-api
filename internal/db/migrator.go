package db

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed schema/migrations/*.sql
var migrationFiles embed.FS

// RunMigrations は渡された *sql.DB に対して埋め込まれたマイグレーションを最新まで適用します
func RunMigrations(db *sql.DB) error {
	// 1. embed.FS を golang-migrate 用のドライバーとしてラップ
	d, err := iofs.New(migrationFiles, "schema/migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver: %w", err)
	}

	// 2. pgx ドライバーのインスタンスを生成
	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return fmt.Errorf("failed to create pgx migrate driver: %w", err)
	}

	// 3. migrator の初期化
	m, err := migrate.NewWithInstance("iofs", d, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// 4. 最新バージョンまでマイグレーションを実行 (Up)
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}