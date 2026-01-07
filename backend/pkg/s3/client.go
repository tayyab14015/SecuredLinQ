package s3

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Client handles AWS S3 interactions
type Client struct {
	s3Client   *s3.S3
	bucketName string
	region     string
}

// MediaObject represents an S3 media object
type MediaObject struct {
	Key          string `json:"key"`
	Type         string `json:"type"`
	Step         int    `json:"step,omitempty"`
	LastModified string `json:"lastModified"`
	FileName     string `json:"fileName"`
	Size         int64  `json:"size"`
	LoadNumber   string `json:"loadNumber"`
	SignedURL    string `json:"signedUrl"`
}

// UploadResult represents the result of an upload operation
type UploadResult struct {
	Success bool   `json:"success"`
	Key     string `json:"key,omitempty"`
	URL     string `json:"url,omitempty"`
	Error   string `json:"error,omitempty"`
}

// NewClient creates a new S3 client
func NewClient(accessKeyID, secretAccessKey, region, bucketName string) (*Client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	return &Client{
		s3Client:   s3.New(sess),
		bucketName: bucketName,
		region:     region,
	}, nil
}

// ListLoadMedia lists all media files for a specific load
func (c *Client) ListLoadMedia(loadNumber string) ([]MediaObject, error) {
	prefix := loadNumber + "_"

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucketName),
		Prefix: aws.String(prefix),
	}

	result, err := c.s3Client.ListObjectsV2(input)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	media := make([]MediaObject, 0)
	for _, obj := range result.Contents {
		if obj.Key == nil {
			continue
		}

		key := *obj.Key
		fileName := key

		// Only include screenshot files
		if !strings.Contains(fileName, "screenshot") {
			continue
		}

		// Determine media type
		mediaType := "unknown"
		ext := strings.ToLower(getExtension(fileName))
		if isImageExtension(ext) || strings.Contains(fileName, "screenshot") {
			mediaType = "image"
		} else if isVideoExtension(ext) {
			mediaType = "video"
		}

		// Generate signed URL
		signedURL, err := c.GetSignedURL(key, 3600)
		if err != nil {
			continue
		}

		var lastModified string
		if obj.LastModified != nil {
			lastModified = obj.LastModified.Format(time.RFC3339)
		}

		var size int64
		if obj.Size != nil {
			size = *obj.Size
		}

		media = append(media, MediaObject{
			Key:          key,
			Type:         mediaType,
			LastModified: lastModified,
			FileName:     fileName,
			Size:         size,
			LoadNumber:   loadNumber,
			SignedURL:    signedURL,
		})
	}

	return media, nil
}

// UploadScreenshot uploads a screenshot to S3
func (c *Client) UploadScreenshot(loadNumber string, imageData []byte, contentType string) (*UploadResult, error) {
	timestamp := time.Now().UnixMilli()
	key := fmt.Sprintf("%s_%d.screenshot.png", loadNumber, timestamp)

	return c.uploadFile(key, imageData, contentType)
}

// UploadScreenshotWithTimestamp uploads a screenshot with a specific timestamp
func (c *Client) UploadScreenshotWithTimestamp(loadNumber string, imageData []byte, contentType string, timestamp int64) (*UploadResult, error) {
	key := fmt.Sprintf("%s_%d.screenshot.png", loadNumber, timestamp)
	return c.uploadFile(key, imageData, contentType)
}

// UploadBase64Image uploads a base64 encoded image
func (c *Client) UploadBase64Image(loadNumber string, base64Data string) (*UploadResult, error) {
	// Remove data URL prefix if present
	base64Data = strings.TrimPrefix(base64Data, "data:image/png;base64,")
	base64Data = strings.TrimPrefix(base64Data, "data:image/jpeg;base64,")
	base64Data = strings.TrimPrefix(base64Data, "data:image/jpg;base64,")

	// Decode base64
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to decode base64: %v", err),
		}, nil
	}

	return c.UploadScreenshot(loadNumber, imageData, "image/png")
}

func (c *Client) uploadFile(key string, data []byte, contentType string) (*UploadResult, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(c.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	_, err := c.s3Client.PutObject(input)
	if err != nil {
		return &UploadResult{
			Success: false,
			Error:   fmt.Sprintf("failed to upload to S3: %v", err),
		}, nil
	}

	// Generate signed URL
	signedURL, err := c.GetSignedURL(key, 3600)
	if err != nil {
		signedURL = ""
	}

	return &UploadResult{
		Success: true,
		Key:     key,
		URL:     signedURL,
	}, nil
}

// GetSignedURL generates a signed URL for an S3 object
func (c *Client) GetSignedURL(key string, expiresInSeconds int64) (string, error) {
	req, _ := c.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	})

	url, err := req.Presign(time.Duration(expiresInSeconds) * time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return url, nil
}

// DeleteObject deletes an object from S3
func (c *Client) DeleteObject(key string) error {
	_, err := c.s3Client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	})
	return err
}

// Helper functions

func getExtension(fileName string) string {
	parts := strings.Split(fileName, ".")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return ""
}

func isImageExtension(ext string) bool {
	switch ext {
	case "jpg", "jpeg", "png", "gif", "webp":
		return true
	}
	return false
}

func isVideoExtension(ext string) bool {
	switch ext {
	case "mp4", "mov", "avi", "mkv", "webm":
		return true
	}
	return false
}
