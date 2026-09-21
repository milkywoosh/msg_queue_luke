package storage

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewClient(ctx context.Context, accessKey, secretKey, addr string) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				accessKey, secretKey, "",
			),
		),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		if ep := addr; ep != "" {
			o.BaseEndpoint = aws.String(ep)
			o.UsePathStyle = true
		}
	}), nil
}

func EnsureBucket(ctx context.Context, c *s3.Client, bucket string) error {
	if _, err := c.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)}); err == nil {
		return nil
	}
	_, err := c.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})
	return err
}

// func main() {
// 	ctx := context.Background()

// 	_, err = client.PutObject(ctx, &s3.PutObjectInput{
// 		Bucket: aws.String(bucket),
// 		Key:    aws.String("test/hello.txt"),
// 		Body:   strings.NewReader("halo dari S3 lokal"),
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	log.Println("upload ok")
// }
