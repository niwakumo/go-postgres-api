package service_test

import (
	"context"
	"testing"

	"go-postgres-api/internal/service"

	"github.com/stretchr/testify/assert"
)

func TestUserService_Validation(t *testing.T) {
	// Repositoryは一旦 nil のまま、バリデーションロジックのみを検証
	svc := service.NewUserService(nil)
	ctx := context.Background()

	t.Run("GetUser: バリデーション検証 (テーブル駆動テスト)", func(t *testing.T) {
		// テストケースのテーブル定義 (Spockの where: ブロックに相当)
		tests := []struct {
			name        string
			id          int64
			expectError bool
			errMessage  string
		}{
			{
				name:        "IDが0の場合はエラー",
				id:          0,
				expectError: true,
				errMessage:  "id must be greater than 0",
			},
			{
				name:        "負の数のIDの場合はエラー",
				id:          -1,
				expectError: true,
				errMessage:  "id must be greater than 0",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				_, err := svc.GetUser(ctx, tc.id)

				if tc.expectError {
					assert.Error(t, err)
					assert.EqualError(t, err, tc.errMessage)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("CreateUser: バリデーション検証 (テーブル駆動テスト)", func(t *testing.T) {
		tests := []struct {
			name        string
			userName    string
			email       string
			expectError bool
			errMessage  string
		}{
			{
				name:        "名前が空の場合はエラー",
				userName:    "",
				email:       "valid@example.com",
				expectError: true,
				errMessage:  "name and email cannot be empty",
			},
			{
				name:        "メールが空の場合はエラー",
				userName:    "Valid Name",
				email:       "",
				expectError: true,
				errMessage:  "name and email cannot be empty",
			},
			{
				name:        "両方空の場合はエラー",
				userName:    "",
				email:       "",
				expectError: true,
				errMessage:  "name and email cannot be empty",
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				_, err := svc.CreateUser(ctx, tc.userName, tc.email)

				if tc.expectError {
					assert.Error(t, err)
					assert.EqualError(t, err, tc.errMessage)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}