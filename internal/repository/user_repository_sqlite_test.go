package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"go-postgres-api/internal/repository"

	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

func TestUserRepository_Search_SQLite(t *testing.T) {
	ctx := context.Background()

	sqliteDB, err := sql.Open("sqlite", ":memory:")
	assert.NoError(t, err)
	defer sqliteDB.Close()

	// テーブル作成
	_, err = sqliteDB.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`)
	assert.NoError(t, err)

	repo := repository.NewUserRepository(sqliteDB)

	t.Run("ILIKE と ANY 構文が SQLite で構文エラーになることの検証", func(t *testing.T) {
		emails := []string{"alice@example.com", "bob@example.com"}
		
		// PostgreSQL用のクエリを実行
		_, err := repo.SearchByEmailsAndPattern(ctx, emails, "%ali%")

		// SQLite は ANY 構文や ILIKE を解釈できずエラーを返す
		assert.Error(t, err)
		t.Logf("【SQLiteでの構文エラー検知】: %v", err)
	})
}