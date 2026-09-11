package s3

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Config holds the settings needed to construct an [S3Client].
type Config struct {
	// Region is the AWS region to sign requests for. Most S3-compatible
	// providers accept an arbitrary non-empty value here even though they
	// aren't region-partitioned.
	Region string
	// Endpoint is the base URL of the S3-compatible service, e.g.
	// "https://s3.us-east-1.amazonaws.com" or a self-hosted provider's URL.
	Endpoint string
	// AccessKey and SecretKey are static credentials used to sign requests.
	AccessKey string
	SecretKey string
	// Bucket is the name of the bucket every [S3Client] operation is scoped to.
	Bucket string
	// UsePathStyle selects path-style addressing (endpoint/bucket/key) instead
	// of virtual-hosted-style (bucket.endpoint/key). Most non-AWS S3-compatible
	// providers require this to be true.
	UsePathStyle bool
}

// S3Client is a thin wrapper around an [s3.Client] scoped to a single bucket.
// Every method call addresses an object by key within that bucket.
type S3Client struct {
	client *s3.Client
	bucket string
}

// NewS3Client builds an [S3Client] from cfg. It does not perform any network
// calls or verify that the bucket exists or is reachable.
func NewS3Client(cfg Config) *S3Client {
	client := s3.New(s3.Options{
		Region:       cfg.Region,
		BaseEndpoint: aws.String(cfg.Endpoint),
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		UsePathStyle: cfg.UsePathStyle,
	})

	return &S3Client{
		client: client,
		bucket: cfg.Bucket,
	}
}
