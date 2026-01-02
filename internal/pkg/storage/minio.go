package storage

import (
	"context"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOAdapter implements Storage using MinIO.
type MinIOAdapter struct {
	client *minio.Client
}

// MinIOOptions configures MinIO client initialization.
type MinIOOptions struct {
	// Endpoint is the MinIO server address.
	Endpoint string
	// AccessKey is the access key ID.
	AccessKey string
	// SecretKey is the secret access key.
	SecretKey string
	// SessionToken is the optional session token.
	SessionToken string
	// Region is the MinIO region.
	Region string
	// UseSSL toggles TLS for MinIO connections.
	UseSSL bool
}

// NewMinIO constructs a MinIO adapter with the provided options.
func NewMinIO(opts MinIOOptions) (*MinIOAdapter, error) {
	client, err := minio.New(opts.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(opts.AccessKey, opts.SecretKey, opts.SessionToken),
		Secure: opts.UseSSL,
		Region: opts.Region,
	})
	if err != nil {
		return nil, err
	}
	return &MinIOAdapter{client: client}, nil
}

// NewMinIOWithClient wraps an existing MinIO client.
func NewMinIOWithClient(client *minio.Client) *MinIOAdapter {
	return &MinIOAdapter{client: client}
}

// PutObject stores data in MinIO and returns metadata.
func (m *MinIOAdapter) PutObject(ctx context.Context, bucket, key string, r io.Reader, opts PutOptions) (ObjectInfo, error) {
	info, err := m.client.PutObject(ctx, bucket, key, r, opts.Size, minio.PutObjectOptions{
		ContentType:  opts.ContentType,
		UserMetadata: opts.Metadata,
	})
	if err != nil {
		return ObjectInfo{}, err
	}
	return ObjectInfo{
		Bucket:      bucket,
		Key:         key,
		Size:        info.Size,
		ETag:        info.ETag,
		ContentType: opts.ContentType,
		Metadata:    opts.Metadata,
	}, nil
}

// GetObject retrieves data and metadata from MinIO.
func (m *MinIOAdapter) GetObject(ctx context.Context, bucket, key string, opts GetOptions) (io.ReadCloser, ObjectInfo, error) {
	getOpts := minio.GetObjectOptions{}
	if opts.Range != nil {
		if opts.Range.End > 0 || opts.Range.End == opts.Range.Start {
			if err := getOpts.SetRange(opts.Range.Start, opts.Range.End); err != nil {
				return nil, ObjectInfo{}, err
			}
		} else {
			if err := getOpts.SetRange(opts.Range.Start, 0); err != nil {
				return nil, ObjectInfo{}, err
			}
		}
	}
	obj, err := m.client.GetObject(ctx, bucket, key, getOpts)
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	stat, err := obj.Stat()
	if err != nil {
		closeErr := obj.Close()
		if closeErr != nil {
			return nil, ObjectInfo{}, closeErr
		}
		return nil, ObjectInfo{}, err
	}
	return obj, minioStatToInfo(bucket, key, stat), nil
}

// StatObject returns metadata for a MinIO object.
func (m *MinIOAdapter) StatObject(ctx context.Context, bucket, key string) (ObjectInfo, error) {
	stat, err := m.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return ObjectInfo{}, err
	}
	return minioStatToInfo(bucket, key, stat), nil
}

// DeleteObject removes an object from MinIO.
func (m *MinIOAdapter) DeleteObject(ctx context.Context, bucket, key string) error {
	return m.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

// ListObjects lists objects from a MinIO bucket.
func (m *MinIOAdapter) ListObjects(ctx context.Context, bucket, prefix string, opts ListOptions) ([]ObjectInfo, error) {
	listOpts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}
	objects := make([]ObjectInfo, 0)
	for object := range m.client.ListObjects(ctx, bucket, listOpts) {
		if object.Err != nil {
			return nil, object.Err
		}
		objects = append(objects, ObjectInfo{
			Bucket:    bucket,
			Key:       object.Key,
			Size:      object.Size,
			ETag:      object.ETag,
			UpdatedAt: object.LastModified,
		})
		if opts.Limit > 0 && int32(len(objects)) >= opts.Limit {
			break
		}
	}
	return objects, nil
}

// PresignGet returns a signed URL for downloading from MinIO.
func (m *MinIOAdapter) PresignGet(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, bucket, key, expiry, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

// PresignPut returns a signed URL for uploading to MinIO.
func (m *MinIOAdapter) PresignPut(ctx context.Context, bucket, key string, _ PutOptions, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedPutObject(ctx, bucket, key, expiry)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

// Close releases MinIO adapter resources.
func (m *MinIOAdapter) Close() error {
	return nil
}

func minioStatToInfo(bucket, key string, stat minio.ObjectInfo) ObjectInfo {
	return ObjectInfo{
		Bucket:      bucket,
		Key:         key,
		Size:        stat.Size,
		ETag:        stat.ETag,
		ContentType: stat.ContentType,
		Metadata:    stat.UserMetadata,
		UpdatedAt:   stat.LastModified,
	}
}
