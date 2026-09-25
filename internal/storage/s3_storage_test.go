package storage_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-postgres-api/internal/storage"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	tclocalstack "github.com/testcontainers/testcontainers-go/modules/localstack"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupLocalStackS3(t *testing.T) (*s3.Client, func()) {
	t.Helper()
	ctx := context.Background()

	// 1. LocalStack コンテナの起動 (S3 サービスのみを有効化)
	localstackContainer, err := tclocalstack.Run(ctx,
		"localstack/localstack:3.0",
		testcontainers.WithEnv(map[string]string{
			"SERVICES": "s3",
		}),
		testcontainers.WithWaitStrategy(
			wait.ForListeningPort("4566/tcp").WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start localstack container: %v", err)
	}

	// 2. LocalStack の動的エンドポイント URL を取得 (例: http://127.0.0.1:49153)
	endpoint, err := localstackContainer.PortEndpoint(ctx, "4566/tcp", "http")
	if err != nil {
		t.Fatalf("failed to get localstack endpoint: %v", err)
	}

	// 3. AWS SDK v2 の構成（ダミー認証情報と LocalStack エンドポイント）
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	if err != nil {
		t.Fatalf("failed to load aws config: %v", err)
	}

	// S3 クライアント生成:
	// - BaseEndpoint で LocalStack の動的ポートを指定
	// - UsePathStyle: true で http://localhost:port/bucket-name の形式を強制 (LocalStack 必須)
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	cleanup := func() {
		_ = localstackContainer.Terminate(ctx)
	}

	return s3Client, cleanup
}

func TestS3Storage_CRUD_LocalStack(t *testing.T) {
	ctx := context.Background()
	s3Client, cleanup := setupLocalStackS3(t)
	defer cleanup()

	bucketName := "test-app-bucket"
	s3Storage := storage.NewS3Storage(s3Client, bucketName)

	t.Run("EnsureBucket で新規バケットが作成され、複数回呼んでも冪等であること", func(t *testing.T) {
		err := s3Storage.EnsureBucket(ctx)
		assert.NoError(t, err)

		// 2回目もエラーにならない（冪等性）
		err = s3Storage.EnsureBucket(ctx)
		assert.NoError(t, err)
	})

	t.Run("オブジェクトのアップロードとダウンロードが正常に一致すること", func(t *testing.T) {
		key := "uploads/sample.txt"
		content := []byte("Hello LocalStack and Testcontainers!")

		// アップロード
		err := s3Storage.Upload(ctx, key, content, "text/plain")
		assert.NoError(t, err)

		// ダウンロード
		downloaded, err := s3Storage.Download(ctx, key)
		assert.NoError(t, err)
		assert.Equal(t, content, downloaded)
	})

	t.Run("存在しないキーを取得した場合は ErrNotFound が返ること", func(t *testing.T) {
		_, err := s3Storage.Download(ctx, "not-exist-key.txt")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrNotFound), "ErrNotFound のラップであること")
	})

	t.Run("空データをアップロードしようとするとバリデーションエラーになること", func(t *testing.T) {
		err := s3Storage.Upload(ctx, "empty.txt", []byte{}, "text/plain")
		assert.Error(t, err)
		assert.EqualError(t, err, "cannot upload empty data")
	})

	t.Run("オブジェクトを削除後、再取得すると ErrNotFound になること", func(t *testing.T) {
		key := "uploads/to-be-deleted.txt"
		err := s3Storage.Upload(ctx, key, []byte("temp"), "text/plain")
		assert.NoError(t, err)

		// 削除
		err = s3Storage.Delete(ctx, key)
		assert.NoError(t, err)

		// 削除後の確認
		_, err = s3Storage.Download(ctx, key)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, storage.ErrNotFound))
	})
}