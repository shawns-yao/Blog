package media

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	appconfig "github.com/shawns-yao/shawn-blog/server/internal/config"
)

type remoteStorage interface {
	Exists(context.Context, string) (bool, error)
	PutFile(context.Context, string, string) error
	ReadURL(context.Context, string) (string, error)
	Delete(context.Context, string) error
}

// R2Storage mirrors media into a private Cloudflare R2 bucket. Stable public
// paths remain local (/uploads/...), while reads are redirected to short-lived
// signed URLs only when the remote object is confirmed to exist.
type R2Storage struct {
	client    *s3.Client
	presigner *s3.PresignClient
	bucket    string
	prefix    string
	readTTL   time.Duration
	timeout   time.Duration
	uploadTTL time.Duration
}

// NewR2Storage returns nil when R2 is not configured. Partial configuration is
// rejected so a deployment cannot silently run with a malformed remote setup.
func NewR2Storage(cfg appconfig.MediaConfig) (*R2Storage, error) {
	endpoint := strings.TrimSpace(cfg.R2Endpoint)
	bucket := strings.TrimSpace(cfg.R2Bucket)
	accessKeyID := strings.TrimSpace(cfg.R2AccessKeyID)
	secretAccessKey := strings.TrimSpace(cfg.R2SecretAccessKey)
	configured := endpoint != "" || bucket != "" || accessKeyID != "" || secretAccessKey != ""
	if !configured {
		return nil, nil
	}
	if endpoint == "" || bucket == "" || accessKeyID == "" || secretAccessKey == "" {
		return nil, errors.New("r2: endpoint, bucket, access key and secret access key are required together")
	}
	if !strings.HasPrefix(endpoint, "https://") {
		return nil, errors.New("r2: endpoint must use HTTPS")
	}

	prefix := strings.Trim(strings.TrimSpace(cfg.R2Prefix), "/")
	if prefix != "" {
		prefix += "/"
	}
	readTTL := cfg.R2ReadURLTTL
	if readTTL <= 0 || readTTL > 24*time.Hour {
		readTTL = 15 * time.Minute
	}
	timeout := cfg.R2RequestTimeout
	if timeout <= 0 || timeout > 30*time.Second {
		timeout = 4 * time.Second
	}
	uploadTTL := cfg.R2UploadTimeout
	if uploadTTL <= 0 || uploadTTL > 30*time.Minute {
		uploadTTL = 2 * time.Minute
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("r2: load client configuration: %w", err)
	}
	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(strings.TrimRight(endpoint, "/"))
		options.UsePathStyle = true
	})
	return &R2Storage{
		client:    client,
		presigner: s3.NewPresignClient(client),
		bucket:    bucket,
		prefix:    prefix,
		readTTL:   readTTL,
		timeout:   timeout,
		uploadTTL: uploadTTL,
	}, nil
}

func (s *R2Storage) objectKey(storedPath string) string {
	return s.prefix + strings.TrimLeft(filepath.ToSlash(strings.TrimSpace(storedPath)), "/")
}

func (s *R2Storage) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, s.timeout)
}

func (s *R2Storage) withUploadTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, s.uploadTTL)
}

func (s *R2Storage) Exists(ctx context.Context, storedPath string) (bool, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.objectKey(storedPath)),
	})
	if err == nil {
		return true, nil
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch strings.ToLower(apiErr.ErrorCode()) {
		case "notfound", "nosuchkey", "404":
			return false, nil
		}
	}
	return false, fmt.Errorf("r2: head object %s: %w", storedPath, err)
}

func (s *R2Storage) PutFile(ctx context.Context, storedPath string, diskPath string) error {
	f, err := os.Open(diskPath)
	if err != nil {
		return err
	}
	defer f.Close()

	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(diskPath)))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	ctx, cancel := s.withUploadTimeout(ctx)
	defer cancel()
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(s.objectKey(storedPath)),
		Body:        f,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return fmt.Errorf("r2: put object %s: %w", storedPath, err)
	}
	return nil
}

func (s *R2Storage) ReadURL(ctx context.Context, storedPath string) (string, error) {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	presigned, err := s.presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.objectKey(storedPath)),
	}, s3.WithPresignExpires(s.readTTL))
	if err != nil {
		return "", fmt.Errorf("r2: presign object %s: %w", storedPath, err)
	}
	return presigned.URL, nil
}

func (s *R2Storage) Delete(ctx context.Context, storedPath string) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.objectKey(storedPath)),
	})
	if err != nil {
		return fmt.Errorf("r2: delete object %s: %w", storedPath, err)
	}
	return nil
}
