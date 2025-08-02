package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Service struct {
	client     *s3.Client
	bucketName string
}

type UploadURLResponse struct {
	UploadURL string    `json:"uploadUrl"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func NewS3Service() (*S3Service, error) {
	cfg := config.LoadConfig()
	
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AWSRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AWSAccessKeyID,
			cfg.AWSSecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg)

	return &S3Service{
		client:     client,
		bucketName: cfg.S3BucketName,
	}, nil
}


func (s *S3Service) GenerateUploadURL(metadata map[string]string) (*UploadURLResponse, error) {
	// Extract and validate required metadata fields
	fileID := metadata["fileId"]
	if fileID == "" {
		return nil, fmt.Errorf("missing or invalid fileId in metadata")
	}
	
	userID := metadata["userId"]
	if userID == "" {
		return nil, fmt.Errorf("missing or invalid userId in metadata")
	}
	
	chunkID := metadata["chunkId"]
	if chunkID == "" {
		return nil, fmt.Errorf("missing or invalid chunkId in metadata")
	}
	
	key := fmt.Sprintf("uploads/%s/%s/%s", userID, fileID, chunkID)
	
	expirationTime := 15 * time.Minute
	expiresAt := time.Now().Add(expirationTime)

	presigner := s3.NewPresignClient(s.client)
	
	request, err := presigner.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:   aws.String(s.bucketName),
		Key:      aws.String(key),
		Metadata: metadata,
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expirationTime
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &UploadURLResponse{
		UploadURL: request.URL,
		ExpiresAt: expiresAt,
	}, nil
}