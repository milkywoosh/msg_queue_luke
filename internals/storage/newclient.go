package storage

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	configS3 "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewClientObjectS3(ctx context.Context, accessKey, secretKey, addr string) (*s3.Client, error) {
	cfg, err := configS3.LoadDefaultConfig(ctx,
		configS3.WithRegion("us-east-1"),
		configS3.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				accessKey, secretKey, "",
			),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if ep := addr; ep != "" {
			o.BaseEndpoint = aws.String(ep)
			o.UsePathStyle = true
		}
	})

	return client, nil
}
