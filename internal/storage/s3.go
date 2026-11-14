// Package storage implements domain.AudioStorage on any S3-compatible store.
package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/raymuos/survoyce/internal/domain"
)

type S3 struct {
	client *s3.Client
	bucket string
}

var _ domain.AudioStorage = (*S3)(nil)

func NewS3(endpoint, bucket, accessKey, secretKey string) *S3 {
	client := s3.New(s3.Options{
		Region:       "us-east-1",
		BaseEndpoint: aws.String(endpoint),
		UsePathStyle: true, // required for most non-AWS S3 stores
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		// Newer SDK versions add checksums that some S3-compatible stores reject.
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	})
	return &S3{client: client, bucket: bucket}
}

// EnsureBucket creates the bucket if it does not exist yet. Call once at startup.
func (s *S3) EnsureBucket(ctx context.Context) error {
	if _, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: &s.bucket}); err == nil {
		return nil
	}
	if _, err := s.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: &s.bucket}); err != nil {
		return fmt.Errorf("create bucket %q: %w", s.bucket, err)
	}
	return nil
}

func (s *S3) Put(ctx context.Context, key string, r io.Reader) error {
	// Clips are ~5 seconds, and the handler caps upload size, so buffering is fine.
	// A bytes.Reader is seekable, which the SDK needs for signing over plain HTTP.
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("read audio: %w", err)
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("put object %q: %w", key, err)
	}
	return nil
}

func (s *S3) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: &s.bucket, Key: &key})
	if err != nil {
		return nil, fmt.Errorf("get object %q: %w", key, err)
	}
	return out.Body, nil
}
