package storage

import (
	"context"
	"io"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// bucket/key-unique/csvFileName
type ObjectS3 interface {
	PresignObject(
		ctx context.Context,
		bucketName,
		key,
		contentType,
		fileName string,
		optFns ...func(*s3.PresignOptions),
	) (*v4.PresignedHTTPRequest, string, error)
	EnsureBucket(ctx context.Context, bucket string) error
	// note directory yg akan di-generate bucket/subdir/key/filename.ext
	PutObject(ctx context.Context, bucket, key, fileName string, file io.Reader) (*s3.PutObjectOutput, string, error)
	// normally info bucket/subDir/key, fileName disimpan di DB
	GetObject(ctx context.Context, bucket, key, fileName string) (*s3.GetObjectOutput, error)
	DeleteObject(ctx context.Context, bucket, key, fileName string) (*s3.DeleteObjectOutput, error)
}
