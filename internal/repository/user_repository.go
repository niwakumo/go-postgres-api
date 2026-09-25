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

// SearchByEmailsAndPattern は複数のドメイン候補(= ANY($1))と名前に大文字小文字無視の部分一致(ILIKE)を適用する関数
func (r *UserRepository) SearchByEmailsAndPattern(ctx context.Context, emailList []string, namePattern string) ([]User, error) {
	// PostgreSQL 特有のイディオム:
	// 1. = ANY($1): 配列型スライスをそのまま渡せる構文
	// 2. ILIKE $2: 大文字・小文字を無視したパターンマッチング
	query := `
		SELECT id, name, email, created_at 
		FROM users 
		WHERE email = ANY($1) AND name ILIKE $2
		ORDER BY id ASC`

	rows, err := r.db.QueryContext(ctx, query, emailList, namePattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}