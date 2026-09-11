// Package s3 provides a small, bucket-scoped client for S3-compatible object
// storage, built on top of the AWS SDK for Go v2.
//
// Construct a client with [NewS3Client], giving it a [Config] describing the
// region, endpoint, credentials, and bucket to use:
//
//	client := s3.NewS3Client(s3.Config{
//		Region:       "us-east-1",
//		Endpoint:     "https://s3.example.com",
//		AccessKey:    accessKey,
//		SecretKey:    secretKey,
//		Bucket:       "my-bucket",
//		UsePathStyle: true, // required by most non-AWS S3-compatible providers
//	})
//
// The resulting [S3Client] is bound to a single bucket for the lifetime of the
// client; every operation ([S3Client.Upload], [S3Client.Delete]) addresses an
// object by key within that bucket.
package s3
