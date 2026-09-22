package repository_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"go-postgres-api/internal/repository"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	// 1. スキーマ定義SQLのパスを取得 (プロジェクトルートの docker/init/01_schema.sql を参照)
	initScriptPath, err := filepath.Abs("../../docker/init/01_schema.sql")
	if err != nil {
		t.Fatalf("failed to get init script path: %v", err)
	}

	// 2. Testcontainers で PostgreSQL コンテナを起動定義
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithInitScripts(initScriptPath),
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("postgres"),
		tcpostgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	// 3. 起動したコンテナの動的接続文字列 (DSN) を取得
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("failed to connect to test db: %v", err)
	}

	// 4. テスト終了時のクリーンアップ処理（後片付け）
	cleanup := func() {
		db.Close()
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return db, cleanup
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