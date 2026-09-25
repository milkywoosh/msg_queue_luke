package storage

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestNewSeaweedS3(t *testing.T) {

	t.Run("returns configured SeaweedS3 instance", func(t *testing.T) {
		cfg := aws.Config{
			Region: "us-east-1",
		}
		client := s3.NewFromConfig(cfg)

		got := NewSeaweedS3(client)

		if got == nil {
			t.Fatal("NewSeaweedS3() returned nil")
		}

		seaweed, ok := got.(*SeaweedS3)
		if !ok {
			t.Fatalf("NewSeaweedS3() returned %T, want *SeaweedS3", got)
		}

		if seaweed.Client == nil {
			t.Fatal("NewSeaweedS3() did not set Client")
		}

		if seaweed.Client != client {
			t.Fatal("NewSeaweedS3() did not preserve the input client")
		}
	})

	t.Run("accepts nil client", func(t *testing.T) {
		got := NewSeaweedS3(nil)

		if got == nil {
			t.Fatal("NewSeaweedS3(nil) returned nil")
		}

		seaweed, ok := got.(*SeaweedS3)
		if !ok {
			t.Fatalf("NewSeaweedS3(nil) returned %T, want *SeaweedS3", got)
		}

		if seaweed.Client != nil {
			t.Fatal("NewSeaweedS3(nil) must get nil")
		}
	})
}
