package main

import (
	"bytes"
	"context"
	"fmt"
	"go-tutuplapak-file/config"
	"go-tutuplapak-file/db"
	"go-tutuplapak-file/repository"
	"go-tutuplapak-file/utils"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gin-gonic/gin"
	"github.com/h2non/bimg"
)

type File struct {
	S3Client   *s3.Client
	BucketName string
}

func CompressImage(fileBuffer *bytes.Buffer) (*bytes.Buffer, error) {
	imageData := fileBuffer.Bytes()
	img := bimg.NewImage(imageData)

	newImg, err := img.Resize(50, 50)
	if err != nil {
		return nil, fmt.Errorf("failed to resizing file: %v", err)
	}

	outputBuffer := bytes.NewBuffer(newImg)
	return outputBuffer, nil
}

func UploadToS3(s3Client *s3.Client, file *bytes.Buffer, bucketName string, resultChan chan<- string, fileName string, contentType string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key := fmt.Sprintf("uploads/%d_%s", time.Now().UnixNano(), fileName)

	// tuning soon
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucketName),
		Key:         aws.String(key),
		Body:        file,
		ACL:         types.ObjectCannedACLPublicRead,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		fmt.Print(err)
		resultChan <- ""
		return
	}

	resultChan <- fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucketName, key)
}

func main() {
	cfg := config.LoadConfig()
	router := gin.Default()

	// Init DB
	db.InitDB(cfg)
	defer func() {
		if err := db.DB.Close(); err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
		log.Println("Database connection closed.")
	}()

	fileHandler := repository.NewFileRepository(db.DB)

	// Init S3
	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithRegion(cfg.AwsRegion),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AwsAccessKeyId, cfg.AwsSecretAccessKey, "")),
	)
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}

	s3File := File{s3.NewFromConfig(awsCfg), cfg.S3Bucket}

	router.POST("/v1/file", func(c *gin.Context) {
		// auth := c.GetHeader("Authorization")

		// if auth == "" {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		// 	return
		// }

		// if !strings.HasPrefix(auth, "Bearer ") {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
		// 	return
		// }

		// auth = auth[7:]
		// _, err = utils.ValidateJWT(auth)
		// if err != nil {
		// 	c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		// 	return
		// }

		fileHeader, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
			return
		}

		if err := utils.ValidateFile(fileHeader); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		openedFile, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}

		fileBuffer := bytes.NewBuffer(nil)
		_, err = fileBuffer.ReadFrom(openedFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		}

		compressedFileBuffer, err := CompressImage(fileBuffer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		}

		var wg sync.WaitGroup
		resultChan := make(chan string, 2)

		wg.Add(2)
		go func() {
			defer wg.Done()
			UploadToS3(s3File.S3Client, fileBuffer, s3File.BucketName, resultChan, fileHeader.Filename, fileHeader.Header.Get("Content-Type"))
		}()

		compressedFileName := fmt.Sprintf("compressed_%s", fileHeader.Filename)

		go func() {
			defer wg.Done()
			UploadToS3(s3File.S3Client, compressedFileBuffer, s3File.BucketName, resultChan, compressedFileName, fileHeader.Header.Get("Content-Type"))
		}()

		go func() {
			wg.Wait()
			close(resultChan)
		}()

		var originalURI, compressedURI string
		for result := range resultChan {
			if strings.Contains(result, compressedFileName) {
				compressedURI = result
			} else {
				originalURI = result
			}
		}

		if originalURI == "" || compressedURI == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
			return
		}

		createdFile, err := fileHandler.Create(originalURI, compressedURI)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Create file failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"fileId":        createdFile.ID,
			"fileUri":       originalURI,
			"compressedUri": compressedURI,
		})
	})

	fmt.Printf("Starting server on port %s...\n", cfg.AppPort)
	router.Run(":" + cfg.AppPort)
}
