package services

import (
	"context"
	"fmt"
	"io"
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
	isLocal    bool
}

func NewStorageService() *StorageService {
	endpoint := os.Getenv("IDRIVE_ENDPOINT")
	accessKey := os.Getenv("IDRIVE_ACCESS_KEY")
	secretKey := os.Getenv("IDRIVE_SECRET_KEY")
	bucketName := os.Getenv("IDRIVE_BUCKET")
	region := os.Getenv("IDRIVE_REGION")
	publicUrl := os.Getenv("IDRIVE_PUBLIC_URL")

	forceLocal := true

	if forceLocal || endpoint == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		fmt.Println("Warning: Storage configuration missing, using local storage")
		if _, err := os.Stat("./uploads"); os.IsNotExist(err) {
			os.Mkdir("./uploads", 0755)
		}
		
		baseUrl := os.Getenv("PUBLIC_URL")
		if baseUrl == "" {
			baseUrl = "http://52.20.206.74:8080" // Default to current server IP
		}

		return &StorageService{
			isLocal:   true,
			publicUrl: baseUrl,
		}
	}

	if region == "" {
		region = "us-east-1"
	}

	// Cargar configuración de AWS
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		fmt.Printf("Error loading AWS config: %v\n", err)
		return nil
	}

	// Cliente S3 con endpoint personalizado para IDrive
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true 
	})

	if publicUrl == "" {
		publicUrl = endpoint
	}

	return &StorageService{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,
		publicUrl:  publicUrl,
		isLocal:    false,
	}
}

func (s *StorageService) UploadFile(file multipart.File, fileHeader *multipart.FileHeader) (string, error) {
	// Generar nombre de archivo único con UUID
	ext := filepath.Ext(fileHeader.Filename)
	uniqueFilename := uuid.New().String() + ext

	// Manejo de almacenamiento Local
	if s.isLocal {
		dst := filepath.Join("./uploads", uniqueFilename)
		out, err := os.Create(dst)
		if err != nil {
			return "", fmt.Errorf("failed to create local file: %w", err)
		}
		defer out.Close()

		// Resetear puntero del archivo por seguridad
		if seeker, ok := file.(io.Seeker); ok {
			seeker.Seek(0, 0)
		}

		_, err = io.Copy(out, file)
		if err != nil {
			return "", fmt.Errorf("failed to save local file: %w", err)
		}

		return fmt.Sprintf("%s/uploads/%s", s.publicUrl, uniqueFilename), nil
	}

	// Manejo de almacenamiento en S3/IDrive
	if s.client == nil {
		return "", fmt.Errorf("storage service not configured")
	}

	ctx := context.Background()

	// Subir archivo al bucket
	uploader := manager.NewUploader(s.client)
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(uniqueFilename),
		Body:        file,
		ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
		ACL:         types.ObjectCannedACLPublicRead, 
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3/IDrive: %w", err)
	}

	// Construir URL pública final
	fileURL := fmt.Sprintf("%s/%s/%s", s.publicUrl, s.bucketName, uniqueFilename)

	return fileURL, nil
}
