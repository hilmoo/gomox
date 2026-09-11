// Package tests3 spins up a disposable S3-compatible object store (via testcontainers'
// s3mock) for use in tests.
package tests3

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/s3mock"
)

// Config carries the endpoint and credentials of the disposable s3mock container, for
// building an S3 client against it.
type Config struct {
	Bucket    string
	Region    string
	Endpoint  string
	AccessKey string
	SecretKey string
}

// New starts a disposable s3mock container running image with an initial bucket named
// bucketName, and builds a client of type C from it via newClient. The container is
// torn down automatically via t.Cleanup.
func New[C any](t *testing.T, image, bucketName string, newClient func(Config) C) C {
	t.Helper()

	ctx := context.Background()

	container, err := s3mock.Run(ctx, image, s3mock.WithInitialBuckets(bucketName))
	if err != nil {
		t.Fatalf("tests3: start s3mock container: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(container); err != nil {
			t.Logf("tests3: terminate container: %v", err)
		}
	})

	endpoint, err := container.EndpointURL(ctx)
	if err != nil {
		t.Fatalf("tests3: endpoint url: %v", err)
	}

	return newClient(Config{
		Bucket:    bucketName,
		Region:    "us-east-1",
		Endpoint:  endpoint,
		AccessKey: "test",
		SecretKey: "test",
	})
}
