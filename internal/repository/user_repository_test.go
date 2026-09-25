package repository_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	appdb "go-postgres-api/internal/db"
	"go-postgres-api/internal/repository"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	// 1. PostgreSQL コンテナの起動
	// Postgres モジュール推奨の wait.ForLog で起動完了ログを待機
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60 * time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// 2. 接続文字列 (DSN) の取得
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// 3. CI環境向け: マイグレーション実行前に実際にクエリが通るか Ping リトライ
	var pingErr error
	for i := 0; i < 15; i++ {
		pingCtx, pingCancel := context.WithTimeout(ctx, 2*time.Second)
		pingErr = sqlDB.PingContext(pingCtx)
		pingCancel()
		if pingErr == nil {
			break
		}
		t.Logf("waiting for db to be pingable (attempt %d/15): %v", i+1, pingErr)
		time.Sleep(1 * time.Second)
	}
	if pingErr != nil {
		t.Fatalf("failed to ping postgres before migration: %v", pingErr)
	}

	// 4. golang-migrate によるスキーマ適用
	if err := appdb.RunMigrations(sqlDB); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	cleanup := func() {
		sqlDB.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return sqlDB, cleanup
}

func TestUserRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewUserRepository(db)

	t.Run("初期データ (Alice) が取得できること", func(t *testing.T) {
		user, err := repo.FindByID(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if user.Name != "Alice" {
			t.Errorf("expected name 'Alice', got '%s'", user.Name)
		}
	})

	t.Run("新規ユーザーが作成でき、取得できること", func(t *testing.T) {
		created, err := repo.Create(ctx, "Dave", "dave@example.com")
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
		}
		if created.ID == 0 {
			t.Errorf("expected non-zero ID")
		}

		found, err := repo.FindByID(ctx, created.ID)
		if err != nil {
			t.Fatalf("failed to find created user: %v", err)
		}
		if found.Name != "Dave" || found.Email != "dave@example.com" {
			t.Errorf("data mismatch: got %+v", found)
		}
	})
}

func TestUserRepository_SearchByEmailsAndPattern(t *testing.T) {
	ctx := context.Background()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewUserRepository(db)

	t.Run("ILIKE と ANY を使った検索が実PostgreSQLで正常に機能すること", func(t *testing.T) {
		// 初期データ: Alice (alice@example.com), Bob (bob@example.com)
		// 検索条件: 
		// email が alice@example.com または bob@example.com のいずれか (= ANY)
		// name に大文字小文字を無視して 'ali' が含まれる (ILIKE '%ali%')
		targets := []string{"alice@example.com", "bob@example.com"}
		users, err := repo.SearchByEmailsAndPattern(ctx, targets, "%aLi%")

		assert.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, "Alice", users[0].Name)
		assert.Equal(t, "alice@example.com", users[0].Email)
	})
}