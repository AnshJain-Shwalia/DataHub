package s3

import (
	"context"
	"fmt"
	"time"

	"github.com/AnshJain-Shwalia/DataHub/backend/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
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
	
	// Use AWS default credential chain (environment variables, IAM roles, etc.)
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AWSRegion),
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

// convertToS3Metadata now expects a single "data" field containing pre-stringified JSON
func convertToS3Metadata(metadata map[string]any) (map[string]string, error) {
	s3Metadata := make(map[string]string)
	
	// Extract the pre-stringified JSON data field
	if data, ok := metadata["data"].(string); ok {
		s3Metadata["data"] = data
	}
	
	return s3Metadata, nil
}

func (s *S3Service) GenerateUploadURL(fileID, userID string, metadata map[string]any) (*UploadURLResponse, error) {
	// Validate required parameters
	if fileID == "" {
		return nil, fmt.Errorf("fileID cannot be empty")
	}
	if userID == "" {
		return nil, fmt.Errorf("userID cannot be empty")
	}
	
	key := fmt.Sprintf("uploads/%s/%s", userID, fileID)
	
	// Convert metadata to S3-compatible format (expects pre-stringified JSON in "data" field)
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