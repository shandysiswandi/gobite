package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

const (
	// DriverS3 selects the AWS S3 backend.
	DriverS3 = "s3"
	// DriverGCS selects the Google Cloud Storage backend.
	DriverGCS = "gcs"
	// DriverMinIO selects the MinIO backend.
	DriverMinIO = "minio"
)

// ErrUnknownDriver indicates an unsupported storage driver.
var ErrUnknownDriver = errors.New("storage: unknown driver")

// FactoryOptions groups configuration for storage drivers.
type FactoryOptions struct {
	// S3 configures the S3 backend.
	S3 S3Options
	// GCS configures the GCS backend.
	GCS GCSOptions
	// MinIO configures the MinIO backend.
	MinIO MinIOOptions
}

// NewFromDriver constructs a Storage implementation by driver name.
func NewFromDriver(ctx context.Context, driver string, opts FactoryOptions) (Storage, error) {
	switch strings.ToLower(driver) {
	case DriverS3:
		return NewS3(ctx, opts.S3)
	case DriverGCS:
		return NewGCS(ctx, opts.GCS)
	case DriverMinIO:
		return NewMinIO(opts.MinIO)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownDriver, driver)
	}
}
