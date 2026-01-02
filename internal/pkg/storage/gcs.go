package storage

import (
	"context"
	"errors"
	"io"
	"time"

	gcs "cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

// GCSAdapter implements Storage using Google Cloud Storage.
type GCSAdapter struct {
	client *gcs.Client
	signer *GCSSigner
}

// GCSOptions configures GCS client initialization.
type GCSOptions struct {
	// Client provides an existing GCS client.
	Client *gcs.Client
	// GoogleAccessID is the service account access ID for signing.
	GoogleAccessID string
	// PrivateKey is the service account private key for signing.
	PrivateKey []byte
}

// GCSSigner holds credentials for signed URL generation.
type GCSSigner struct {
	// GoogleAccessID is the service account access ID.
	GoogleAccessID string
	// PrivateKey is the service account private key.
	PrivateKey []byte
}

// NewGCS constructs a GCS adapter with optional signing support.
func NewGCS(ctx context.Context, opts GCSOptions) (*GCSAdapter, error) {
	client := opts.Client
	if client == nil {
		created, err := gcs.NewClient(ctx)
		if err != nil {
			return nil, err
		}
		client = created
	}
	var signer *GCSSigner
	if opts.GoogleAccessID != "" && len(opts.PrivateKey) > 0 {
		signer = &GCSSigner{
			GoogleAccessID: opts.GoogleAccessID,
			PrivateKey:     opts.PrivateKey,
		}
	}
	return &GCSAdapter{
		client: client,
		signer: signer,
	}, nil
}

// PutObject stores data in GCS and returns metadata.
func (g *GCSAdapter) PutObject(ctx context.Context, bucket, key string, r io.Reader, opts PutOptions) (ObjectInfo, error) {
	obj := g.client.Bucket(bucket).Object(key)
	writer := obj.NewWriter(ctx)
	if opts.ContentType != "" {
		writer.ContentType = opts.ContentType
	}
	if len(opts.Metadata) > 0 {
		writer.Metadata = opts.Metadata
	}
	_, err := io.Copy(writer, r)
	if err != nil {
		closeErr := writer.Close()
		if closeErr != nil {
			return ObjectInfo{}, closeErr
		}
		return ObjectInfo{}, err
	}
	if err := writer.Close(); err != nil {
		return ObjectInfo{}, err
	}
	attrs := writer.Attrs()
	if attrs == nil {
		return ObjectInfo{
			Bucket:      bucket,
			Key:         key,
			Size:        opts.Size,
			ContentType: opts.ContentType,
			Metadata:    opts.Metadata,
		}, nil
	}
	return gcsAttrsToInfo(attrs), nil
}

// GetObject retrieves data and metadata from GCS.
func (g *GCSAdapter) GetObject(ctx context.Context, bucket, key string, opts GetOptions) (io.ReadCloser, ObjectInfo, error) {
	obj := g.client.Bucket(bucket).Object(key)
	var reader *gcs.Reader
	var err error
	if opts.Range != nil {
		length := int64(-1)
		if opts.Range.End > 0 || opts.Range.End == opts.Range.Start {
			length = opts.Range.End - opts.Range.Start + 1
		}
		reader, err = obj.NewRangeReader(ctx, opts.Range.Start, length)
	} else {
		reader, err = obj.NewReader(ctx)
	}
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		closeErr := reader.Close()
		if closeErr != nil {
			return nil, ObjectInfo{}, closeErr
		}
		return nil, ObjectInfo{}, err
	}
	return reader, gcsAttrsToInfo(attrs), nil
}

// StatObject returns metadata for a GCS object.
func (g *GCSAdapter) StatObject(ctx context.Context, bucket, key string) (ObjectInfo, error) {
	attrs, err := g.client.Bucket(bucket).Object(key).Attrs(ctx)
	if err != nil {
		return ObjectInfo{}, err
	}
	return gcsAttrsToInfo(attrs), nil
}

// DeleteObject removes an object from GCS.
func (g *GCSAdapter) DeleteObject(ctx context.Context, bucket, key string) error {
	return g.client.Bucket(bucket).Object(key).Delete(ctx)
}

// ListObjects lists objects from a GCS bucket.
func (g *GCSAdapter) ListObjects(ctx context.Context, bucket, prefix string, opts ListOptions) ([]ObjectInfo, error) {
	query := &gcs.Query{Prefix: prefix}
	it := g.client.Bucket(bucket).Objects(ctx, query)
	if opts.Token != "" {
		it.PageInfo().Token = opts.Token
	}
	if opts.Limit > 0 {
		it.PageInfo().MaxSize = int(opts.Limit)
	}
	objects := make([]ObjectInfo, 0)
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, err
		}
		objects = append(objects, gcsAttrsToInfo(attrs))
		if opts.Limit > 0 && int32(len(objects)) >= opts.Limit {
			break
		}
	}
	return objects, nil
}

// PresignGet returns a signed URL for downloading from GCS.
func (g *GCSAdapter) PresignGet(_ context.Context, bucket, key string, expiry time.Duration) (string, error) {
	if g.signer == nil {
		return "", ErrMissingSigner
	}
	return gcs.SignedURL(bucket, key, &gcs.SignedURLOptions{
		Method:         "GET",
		Expires:        time.Now().Add(expiry),
		GoogleAccessID: g.signer.GoogleAccessID,
		PrivateKey:     g.signer.PrivateKey,
	})
}

// PresignPut returns a signed URL for uploading to GCS.
func (g *GCSAdapter) PresignPut(_ context.Context, bucket, key string, opts PutOptions, expiry time.Duration) (string, error) {
	if g.signer == nil {
		return "", ErrMissingSigner
	}
	return gcs.SignedURL(bucket, key, &gcs.SignedURLOptions{
		Method:         "PUT",
		Expires:        time.Now().Add(expiry),
		GoogleAccessID: g.signer.GoogleAccessID,
		PrivateKey:     g.signer.PrivateKey,
		ContentType:    opts.ContentType,
	})
}

// Close closes the GCS client.
func (g *GCSAdapter) Close() error {
	return g.client.Close()
}

func gcsAttrsToInfo(attrs *gcs.ObjectAttrs) ObjectInfo {
	if attrs == nil {
		return ObjectInfo{}
	}
	return ObjectInfo{
		Bucket:      attrs.Bucket,
		Key:         attrs.Name,
		Size:        attrs.Size,
		ETag:        attrs.Etag,
		ContentType: attrs.ContentType,
		Metadata:    attrs.Metadata,
		UpdatedAt:   attrs.Updated,
	}
}
