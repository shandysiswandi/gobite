package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

// ErrMissingSigner indicates signed URL support is not configured.
var ErrMissingSigner = errors.New("storage: signed url signer not configured")

// Storage defines object storage operations.
type Storage interface {
	io.Closer

	// PutObject stores data and returns object metadata.
	PutObject(ctx context.Context, bucket, key string, r io.Reader, opts PutOptions) (ObjectInfo, error)
	// GetObject retrieves data and metadata for the object.
	GetObject(ctx context.Context, bucket, key string, opts GetOptions) (io.ReadCloser, ObjectInfo, error)
	// StatObject returns object metadata without reading its contents.
	StatObject(ctx context.Context, bucket, key string) (ObjectInfo, error)
	// DeleteObject removes the object.
	DeleteObject(ctx context.Context, bucket, key string) error
	// ListObjects lists objects in a bucket prefix.
	ListObjects(ctx context.Context, bucket, prefix string, opts ListOptions) ([]ObjectInfo, error)
	// PresignGet returns a signed URL for downloading.
	PresignGet(ctx context.Context, bucket, key string, expiry time.Duration) (string, error)
	// PresignPut returns a signed URL for uploading.
	PresignPut(ctx context.Context, bucket, key string, opts PutOptions, expiry time.Duration) (string, error)
}

// PutOptions configures upload behavior.
type PutOptions struct {
	// Size is the expected content length.
	Size int64
	// ContentType is the MIME type for the object.
	ContentType string
	// Metadata includes custom key/value metadata.
	Metadata map[string]string
}

// GetOptions configures download behavior.
type GetOptions struct {
	// Range requests a byte range when set.
	Range *ByteRange
}

// ListOptions configures listing behavior.
type ListOptions struct {
	// Limit caps the number of results.
	Limit int32
	// Token is a pagination token.
	Token string
}

// ByteRange represents an inclusive byte range.
type ByteRange struct {
	// Start is the starting byte offset.
	Start int64
	// End is the ending byte offset.
	End int64
}

// ObjectInfo describes object metadata.
type ObjectInfo struct {
	// Bucket is the bucket name.
	Bucket string
	// Key is the object key.
	Key string
	// Size is the object size in bytes.
	Size int64
	// ETag is the object ETag when provided.
	ETag string
	// ContentType is the object MIME type.
	ContentType string
	// Metadata is user-defined metadata.
	Metadata map[string]string
	// UpdatedAt is the last modified time.
	UpdatedAt time.Time
}
