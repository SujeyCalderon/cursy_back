package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

type StorageService struct {
	client     *s3.Client
	bucketName string
	endpoint   string
	publicUrl  string
}

func NewStorageService() *StorageService {
	endpoint := os.Getenv("IDRIVE_ENDPOINT")
	accessKey := os.Getenv("IDRIVE_ACCESS_KEY")
	secretKey := os.Getenv("IDRIVE_SECRET_KEY")
	bucketName := os.Getenv("IDRIVE_BUCKET")
	region := os.Getenv("IDRIVE_REGION")
	publicUrl := os.Getenv("IDRIVE_PUBLIC_URL") // Optional: Custom domain or public endpoint

	if endpoint == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		fmt.Println("Warning: Storage configuration missing")
		return nil
	}

	if region == "" {
		region = "us-east-1"
	}

	// Load custom configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		fmt.Printf("Error loading AWS config: %v\n", err)
		return nil
	}

	// Create S3 client with custom endpoint resolver for IDrive
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // IDrive usually works better with path style
	})

	if publicUrl == "" {
		publicUrl = endpoint
	}

	return &StorageService{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,
		publicUrl:  publicUrl,
	}
}

func (s *StorageService) UploadFile(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("storage service not configured")
	}

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	uniqueFilename := uuid.New().String() + ext

	ctx := context.Background()

	// Upload file
	uploader := manager.NewUploader(s.client)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(uniqueFilename),
		Body:        file,
		ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
		ACL:         types.ObjectCannedACLPublicRead, // Set file as publicly readable
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3/IDrive: %w", err)
	}

	// Construct public URL
	// Usually IDrive public URL: https://<endpoint>/<bucket>/<filename>
	// Or sometimes: https://<bucket>.<endpoint>/<filename>
	// We'll use the configured PublicURL or Endpoint + Bucket + Filename (Path Style)
	fileURL := fmt.Sprintf("%s/%s/%s", s.publicUrl, s.bucketName, uniqueFilename)

	return fileURL, nil
}
