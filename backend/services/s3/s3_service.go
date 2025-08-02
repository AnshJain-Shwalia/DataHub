package s3

import (
	"context"
	"encoding/json"
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

func convertToS3Metadata(metadata map[string]any) (map[string]string, error) {
	s3Metadata := make(map[string]string)
	
	for key, value := range metadata {
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("failed to JSON encode value for key %s: %w", key, err)
		}
		s3Metadata[key] = string(jsonBytes)
	}
	
	return s3Metadata, nil
}

func (s *S3Service) GenerateUploadURL(metadata map[string]any) (*UploadURLResponse, error) {
	// Extract and validate required metadata fields
	fileID, ok := metadata["fileId"].(string)
	if !ok || fileID == "" {
		return nil, fmt.Errorf("missing or invalid fileId in metadata")
	}
	
	userID, ok := metadata["userId"].(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("missing or invalid userId in metadata")
	}
	
	chunkID, ok := metadata["chunkId"].(string)
	if !ok || chunkID == "" {
		return nil, fmt.Errorf("missing or invalid chunkId in metadata")
	}
	
	key := fmt.Sprintf("uploads/%s/%s/%s", userID, fileID, chunkID)
	
	// Convert metadata to S3-compatible format
	s3Metadata, err := convertToS3Metadata(metadata)
	if err != nil {
		return nil, err
	}
	
	expirationTime := 15 * time.Minute
	expiresAt := time.Now().Add(expirationTime)

	presigner := s3.NewPresignClient(s.client)
	
	request, err := presigner.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:   aws.String(s.bucketName),
		Key:      aws.String(key),
		Metadata: s3Metadata,
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