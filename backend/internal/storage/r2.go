package storage

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

const defaultPresignExpiry = 15 * time.Minute

type R2Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Endpoint        string
	PublicBaseURL   string
}

type R2Driver struct {
	bucket        string
	publicBaseURL string
	client        *s3.Client
	presign       *s3.PresignClient
}

func NewR2Driver(ctx context.Context, cfg R2Config) (*R2Driver, error) {
	if strings.TrimSpace(cfg.AccessKeyID) == "" {
		return nil, newValidationError("R2_ACCESS_KEY_ID is required when STORAGE_DRIVER=r2")
	}
	if strings.TrimSpace(cfg.SecretAccessKey) == "" {
		return nil, newValidationError("R2_SECRET_ACCESS_KEY is required when STORAGE_DRIVER=r2")
	}
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, newValidationError("R2_BUCKET is required when STORAGE_DRIVER=r2")
	}
	if strings.TrimSpace(cfg.Endpoint) == "" {
		return nil, newValidationError("R2_ENDPOINT is required when STORAGE_DRIVER=r2")
	}
	if strings.TrimSpace(cfg.PublicBaseURL) == "" {
		return nil, newValidationError("R2_PUBLIC_BASE_URL is required when STORAGE_DRIVER=r2")
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion("auto"),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(opts *s3.Options) {
		opts.UsePathStyle = true
		opts.BaseEndpoint = aws.String(strings.TrimRight(cfg.Endpoint, "/"))
	})

	return &R2Driver{
		bucket:        cfg.Bucket,
		publicBaseURL: strings.TrimRight(cfg.PublicBaseURL, "/"),
		client:        client,
		presign: s3.NewPresignClient(client, func(opts *s3.PresignOptions) {
			opts.Expires = defaultPresignExpiry
		}),
	}, nil
}

func (d *R2Driver) Name() string {
	return "r2"
}

func (d *R2Driver) GenerateObjectKey(req UploadRequest) (string, error) {
	return NewObjectKey(req.ContentType, time.Now().UTC())
}

func (d *R2Driver) CreatePresign(ctx context.Context, req UploadRequest, objectKey string) (*PresignResult, error) {
	presigned, err := d.presign.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      aws.String(d.bucket),
			Key:         aws.String(objectKey),
			ContentType: aws.String(NormalizeContentType(req.ContentType)),
		},
		func(opts *s3.PresignOptions) {
			opts.Expires = defaultPresignExpiry
		},
	)
	if err != nil {
		return nil, err
	}

	return &PresignResult{
		ObjectKey:    objectKey,
		UploadMode:   d.Name(),
		UploadURL:    presigned.URL,
		UploadMethod: presigned.Method,
		PublicURL:    d.BuildPublicURL(objectKey),
		Headers: map[string]string{
			"Content-Type": NormalizeContentType(req.ContentType),
		},
	}, nil
}

func (d *R2Driver) BuildPublicURL(objectKey string) string {
	key := NormalizeObjectKey(objectKey)
	if key == "" {
		return ""
	}
	return d.publicBaseURL + "/" + key
}

func (d *R2Driver) Upload(ctx context.Context, objectKey, contentType string, body []byte) error {
	_, err := d.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(d.bucket),
		Key:           aws.String(NormalizeObjectKey(objectKey)),
		Body:          bytes.NewReader(body),
		ContentLength: aws.Int64(int64(len(body))),
		ContentType:   aws.String(NormalizeContentType(contentType)),
	})
	return err
}

func (d *R2Driver) Stat(ctx context.Context, objectKey string) (*ObjectInfo, error) {
	result, err := d.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(NormalizeObjectKey(objectKey)),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "NotFound" {
			return nil, nil
		}
		return nil, err
	}
	return &ObjectInfo{
		SizeBytes:   aws.ToInt64(result.ContentLength),
		ContentType: NormalizeContentType(aws.ToString(result.ContentType)),
	}, nil
}

func (d *R2Driver) Delete(ctx context.Context, objectKey string) error {
	_, err := d.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(d.bucket),
		Key:    aws.String(NormalizeObjectKey(objectKey)),
	})
	return err
}
