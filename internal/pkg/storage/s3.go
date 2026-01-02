package storage

import (
	"context"
	"io"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Adapter implements Storage using AWS S3.
type S3Adapter struct {
	client  *s3.Client
	presign *s3.PresignClient
}

// S3Options configures S3 client initialization.
type S3Options struct {
	// Region is the AWS region.
	Region string
	// Endpoint overrides the AWS endpoint.
	Endpoint string
	// AccessKey is the static access key ID.
	AccessKey string
	// SecretKey is the static secret access key.
	SecretKey string
	// SessionToken is the optional session token.
	SessionToken string
	// UsePathStyle forces path-style addressing.
	UsePathStyle bool
}

// NewS3 constructs an S3 adapter with the provided options.
func NewS3(ctx context.Context, opts S3Options) (*S3Adapter, error) {
	cfgOpts := []func(*config.LoadOptions) error{}
	if opts.Region != "" {
		cfgOpts = append(cfgOpts, config.WithRegion(opts.Region))
	} else if opts.Endpoint != "" {
		cfgOpts = append(cfgOpts, config.WithRegion("us-east-1"))
	}
	if opts.AccessKey != "" || opts.SecretKey != "" {
		cfgOpts = append(cfgOpts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(opts.AccessKey, opts.SecretKey, opts.SessionToken),
		))
	}
	cfg, err := config.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		return nil, err
	}
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = opts.UsePathStyle
		if opts.Endpoint != "" {
			o.BaseEndpoint = aws.String(opts.Endpoint)
		}
	})
	return NewS3WithClient(client), nil
}

// NewS3WithClient wraps an existing S3 client.
func NewS3WithClient(client *s3.Client) *S3Adapter {
	return &S3Adapter{
		client:  client,
		presign: s3.NewPresignClient(client),
	}
}

// PutObject stores data in S3 and returns metadata.
func (s *S3Adapter) PutObject(ctx context.Context, bucket, key string, r io.Reader, opts PutOptions) (ObjectInfo, error) {
	input := &s3.PutObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		Body:     r,
		Metadata: opts.Metadata,
	}
	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}
	if opts.Size > 0 {
		input.ContentLength = aws.Int64(opts.Size)
	}
	out, err := s.client.PutObject(ctx, input)
	if err != nil {
		return ObjectInfo{}, err
	}
	return ObjectInfo{
		Bucket:      bucket,
		Key:         key,
		Size:        opts.Size,
		ETag:        aws.ToString(out.ETag),
		ContentType: opts.ContentType,
		Metadata:    opts.Metadata,
	}, nil
}

// GetObject retrieves data and metadata from S3.
func (s *S3Adapter) GetObject(ctx context.Context, bucket, key string, opts GetOptions) (io.ReadCloser, ObjectInfo, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	if header := s3RangeHeader(opts.Range); header != nil {
		input.Range = header
	}
	out, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, ObjectInfo{}, err
	}
	info := ObjectInfo{
		Bucket:      bucket,
		Key:         key,
		Size:        aws.ToInt64(out.ContentLength),
		ETag:        aws.ToString(out.ETag),
		ContentType: aws.ToString(out.ContentType),
		Metadata:    out.Metadata,
	}
	if out.LastModified != nil {
		info.UpdatedAt = *out.LastModified
	}
	return out.Body, info, nil
}

// StatObject returns metadata for an S3 object.
func (s *S3Adapter) StatObject(ctx context.Context, bucket, key string) (ObjectInfo, error) {
	out, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return ObjectInfo{}, err
	}
	info := ObjectInfo{
		Bucket:      bucket,
		Key:         key,
		Size:        aws.ToInt64(out.ContentLength),
		ETag:        aws.ToString(out.ETag),
		ContentType: aws.ToString(out.ContentType),
		Metadata:    out.Metadata,
	}
	if out.LastModified != nil {
		info.UpdatedAt = *out.LastModified
	}
	return info, nil
}

// DeleteObject removes an object from S3.
func (s *S3Adapter) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// ListObjects lists objects from an S3 bucket.
func (s *S3Adapter) ListObjects(ctx context.Context, bucket, prefix string, opts ListOptions) ([]ObjectInfo, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	}
	if opts.Limit > 0 {
		input.MaxKeys = aws.Int32(opts.Limit)
	}
	if opts.Token != "" {
		input.ContinuationToken = aws.String(opts.Token)
	}
	pager := s3.NewListObjectsV2Paginator(s.client, input)
	objects := make([]ObjectInfo, 0)
	for pager.HasMorePages() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, obj := range page.Contents {
			objects = append(objects, ObjectInfo{
				Bucket:    bucket,
				Key:       aws.ToString(obj.Key),
				Size:      aws.ToInt64(obj.Size),
				ETag:      aws.ToString(obj.ETag),
				UpdatedAt: aws.ToTime(obj.LastModified),
			})
			if opts.Limit > 0 && int32(len(objects)) >= opts.Limit {
				return objects, nil
			}
		}
		if opts.Limit > 0 {
			break
		}
	}
	return objects, nil
}

// PresignGet returns a signed URL for downloading from S3.
func (s *S3Adapter) PresignGet(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	out, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

// PresignPut returns a signed URL for uploading to S3.
func (s *S3Adapter) PresignPut(ctx context.Context, bucket, key string, opts PutOptions, expiry time.Duration) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		Metadata: opts.Metadata,
	}
	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}
	if opts.Size > 0 {
		input.ContentLength = aws.Int64(opts.Size)
	}
	out, err := s.presign.PresignPutObject(ctx, input, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

// Close releases the S3 adapter resources.
func (s *S3Adapter) Close() error {
	return nil
}

func s3RangeHeader(rng *ByteRange) *string {
	if rng == nil {
		return nil
	}
	if rng.End >= 0 && rng.End < rng.Start {
		return nil
	}
	start := strconv.FormatInt(rng.Start, 10)
	if rng.End > 0 || rng.End == rng.Start {
		end := strconv.FormatInt(rng.End, 10)
		return aws.String("bytes=" + start + "-" + end)
	}
	return aws.String("bytes=" + start + "-")
}
