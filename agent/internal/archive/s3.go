package archive

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Uploader struct {
	client *s3.Client
}

func NewS3Uploader(ctx context.Context) (*S3Uploader, error) {
	awsConfig, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return &S3Uploader{client: s3.NewFromConfig(awsConfig)}, nil
}

func (u *S3Uploader) UploadFile(ctx context.Context, bucket string, key string, path string) error {
	root, err := os.OpenRoot(filepath.Dir(path))
	if err != nil {
		return fmt.Errorf("open archive batch root: %w", err)
	}
	defer func() {
		_ = root.Close()
	}()

	file, err := root.Open(filepath.Base(path))
	if err != nil {
		return fmt.Errorf("open archive batch: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	_, err = u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("upload archive batch to s3: %w", err)
	}
	return nil
}
