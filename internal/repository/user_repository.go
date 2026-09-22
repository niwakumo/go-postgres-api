package repository

import (
	"context"
	"database/sql"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRepository struct {
	db *sql.DB
}

// コンストラクタ関数（SpringのBean定義に相当）
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// IDによるユーザー1件取得
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*User, error) {
	query := `SELECT id, name, email, created_at FROM users WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)

	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

// ユーザー新規作成
func (r *UserRepository) Create(ctx context.Context, name, email string) (*User, error) {
	query := `INSERT INTO users (name, email) VALUES ($1, $2) RETURNING id, name, email, created_at`
	row := r.db.QueryRowContext(ctx, query, name, email)

	var u User
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}