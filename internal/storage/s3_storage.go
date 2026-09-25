package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// ErrNotFound は S3 上にキーが存在しない場合に返却する共通エラー
var ErrNotFound = errors.New("object not found")

type S3Storage struct {
	client *s3.Client
	bucket string
}

func NewS3Storage(client *s3.Client, bucket string) *S3Storage {
	return &S3Storage{
		client: client,
		bucket: bucket,
	}
}

// EnsureBucket はバケットが存在しない場合に新規作成します
func (s *S3Storage) EnsureBucket(ctx context.Context) error {
	_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err == nil {
		return nil
	}

	_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(s.bucket),
	})
	if err != nil {
		var bfe *types.BucketAlreadyExists
		var boe *types.BucketAlreadyOwnedByYou
		if errors.As(err, &bfe) || errors.As(err, &boe) {
			return nil
		}
		return fmt.Errorf("failed to create bucket %s: %w", s.bucket, err)
	}

	return nil
}

// Upload は指定キーにバイトデータをアップロードします
func (s *S3Storage) Upload(ctx context.Context, key string, data []byte, contentType string) error {
	if len(data) == 0 {
		return errors.New("cannot upload empty data")
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	if _, err := s.client.PutObject(ctx, input); err != nil {
		return fmt.Errorf("failed to put object (key: %s): %w", key, err)
	}

	return nil
}

// Download は指定キーからオブジェクトをダウンロードしてバイト列を返却します
func (s *S3Storage) Download(ctx context.Context, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	output, err := s.client.GetObject(ctx, input)
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
			return nil, fmt.Errorf("%w: %s", ErrNotFound, key)
		}
		return nil, fmt.Errorf("failed to get object (key: %s): %w", key, err)
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	return data, nil
}

// Delete は指定キーのオブジェクトを削除します
func (s *S3Storage) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	if _, err := s.client.DeleteObject(ctx, input); err != nil {
		return fmt.Errorf("failed to delete object (key: %s): %w", key, err)
	}

	return nil
}