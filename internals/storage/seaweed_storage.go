package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type SeaweedS3 struct {
	Client *s3.Client
}

func NewSeaweedS3(client *s3.Client) ObjectS3 {
	n := &SeaweedS3{
		Client: client,
	}

	return n
}

func (s *SeaweedS3) EnsureBucket(ctx context.Context, bucket string) error {
	if _, err := s.Client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)}); err == nil {
		return nil
	}
	_, err := s.Client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	return err
}

// create new unique URL with expired, nanti di call UI supaya bisa langsung upload ke S3 tanpa lewat server
func (s *SeaweedS3) PresignObject(
	ctx context.Context,
	bucketName,
	subDir,
	key,
	contentType,
	fileName string,
	optFns ...func(*s3.PresignOptions),
) (*v4.PresignedHTTPRequest, string, error) {

	// /subDir/key/filename
	objectKey := fmt.Sprintf(
		"%s/%s/%s",
		subDir,
		uuid.NewString(),
		fileName,
	)

	// implement io.Reader

	input := &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(objectKey), // location/key-unique/file-name
		ContentType: aws.String(contentType),
	}

	presignClient := s3.NewPresignClient(s.Client, optFns...)

	preSignedHttp, err := presignClient.PresignPutObject(
		ctx,
		input,
		s3.WithPresignExpires(10*time.Minute),
	)

	if err != nil {
		return nil, "", err
	}

	return preSignedHttp, objectKey, nil
}

// upload via server
func (s *SeaweedS3) PutObject(ctx context.Context, bucket, subDir, fileName string, file io.Reader) (*s3.PutObjectOutput, string, error) {
	objectKey := fmt.Sprintf(
		"%s/%s/%s",
		subDir,
		uuid.NewString(), // key must be unique
		fileName,
	)

	output, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(objectKey),
		Body:   file, // io.Reader // multipart.File type
	})
	if err != nil {
		return nil, "", err
	}

	return output, fmt.Sprintf("%s/%s", bucket, objectKey), nil

}

func (s *SeaweedS3) GetObject(ctx context.Context, bucket, subDir, key, fileName string) (*s3.GetObjectOutput, error) {

	objectKey := fmt.Sprintf(
		"%s/%s/%s",
		subDir,
		key,
		fileName,
	)

	return s.Client.GetObject(
		ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(objectKey),
		},
	)

}
