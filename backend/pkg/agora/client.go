package agora

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client handles Agora API interactions
type Client struct {
	appID          string
	appCertificate string
	encodedKey     string
	httpClient     *http.Client
	s3Config       S3Config
}

// S3Config holds AWS S3 configuration for Agora recording storage
type S3Config struct {
	Vendor         int      `json:"vendor"`
	Region         int      `json:"region"`
	Bucket         string   `json:"bucket"`
	AccessKey      string   `json:"accessKey"`
	SecretKey      string   `json:"secretKey"`
	FileNamePrefix []string `json:"fileNamePrefix"`
}

// RecordingResult represents the result of a recording operation
type RecordingResult struct {
	Success    bool
	ResourceID string
	SID        string
	FileName   string
	S3Key      string
	S3URL      string
	FileList   []string
	FileSize   int64
	Duration   int
}

// NewClient creates a new Agora client
func NewClient(appID, appCertificate, encodedKey string) *Client {
	// If encodedKey is not provided, try to generate from customer credentials
	if encodedKey == "" {
		customerID := os.Getenv("AGORA_CUSTOMER_ID")
		customerSecret := os.Getenv("AGORA_CUSTOMER_SECRET")
		if customerID != "" && customerSecret != "" {
			// Generate Base64 encoded credentials: base64(customerID:customerSecret)
			credentials := customerID + ":" + customerSecret
			encodedKey = base64.StdEncoding.EncodeToString([]byte(credentials))
			log.Printf("Generated Agora encoded key from AGORA_CUSTOMER_ID and AGORA_CUSTOMER_SECRET")
		} else {
			log.Printf("Warning: AGORA_ENCODED_KEY not set and AGORA_CUSTOMER_ID/AGORA_CUSTOMER_SECRET not available. Recording will not work.")
		}
	}

	// Get S3 region code for Agora
	s3Region := getAgoraS3Region(os.Getenv("AWS_REGION"))

	return &Client{
		appID:          appID,
		appCertificate: appCertificate,
		encodedKey:     encodedKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		s3Config: S3Config{
			Vendor:         1, // AWS S3
			Region:         s3Region,
			Bucket:         os.Getenv("AWS_S3_BUCKET_NAME"),
			AccessKey:      os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretKey:      os.Getenv("AWS_SECRET_ACCESS_KEY"),
			FileNamePrefix: []string{"recordings"},
		},
	}
}

// GetAppID returns the Agora app ID
func (c *Client) GetAppID() string {
	return c.appID
}

// GetAppCertificate returns the Agora app certificate
func (c *Client) GetAppCertificate() string {
	return c.appCertificate
}

func (c *Client) getBaseURL() string {
	return fmt.Sprintf("https://api.agora.io/v1/apps/%s", c.appID)
}

func (c *Client) makeRequest(endpoint, method string, body interface{}) (map[string]interface{}, error) {
	// Check if we have credentials
	if c.encodedKey == "" {
		return nil, fmt.Errorf("Agora REST API credentials not configured. Set AGORA_ENCODED_KEY or both AGORA_CUSTOMER_ID and AGORA_CUSTOMER_SECRET environment variables")
	}

	url := c.getBaseURL() + endpoint

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
		log.Printf("Agora API Request to %s: %s", endpoint, string(jsonBody))
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Basic "+c.encodedKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("Agora API Response (status %d): %s", resp.StatusCode, string(respBody))

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		// If response is not JSON, return error with status code
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
		}
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode >= 400 {
		errMsg := ""
		if e, ok := result["error"].(string); ok {
			errMsg = e
		} else if m, ok := result["message"].(string); ok {
			errMsg = m
		} else if r, ok := result["reason"].(string); ok {
			errMsg = r
		}

		// Provide more helpful error messages
		if resp.StatusCode == 401 {
			return nil, fmt.Errorf("Invalid authentication credentials. Please verify your AGORA_CUSTOMER_ID and AGORA_CUSTOMER_SECRET (or AGORA_ENCODED_KEY). Get these from Agora Console > RESTful API")
		}

		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, errMsg)
	}

	return result, nil
}

// StartRecording starts cloud recording for a channel
func (c *Client) StartRecording(channelName, uid, token, loadNumber string) (*RecordingResult, error) {
	// Validate configuration
	if c.appID == "" {
		return nil, fmt.Errorf("AGORA_APP_ID is not configured")
	}
	if c.s3Config.Bucket == "" {
		return nil, fmt.Errorf("AWS_S3_BUCKET_NAME is not configured for recording storage")
	}
	if c.s3Config.AccessKey == "" || c.s3Config.SecretKey == "" {
		return nil, fmt.Errorf("AWS credentials (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY) are not configured for recording storage")
	}

	// Build file prefix - Agora expects an array of strings (folder levels)
	// Each string must be alphanumeric only (no special characters)
	// Format: ["recordings"] or ["recordings", "load123"]
	fileNamePrefix := []string{"recordings"}
	if loadNumber != "" {
		// Clean load number: remove all non-alphanumeric characters
		// Agora requires alphanumeric folder names only
		cleanLoadNumber := cleanAlphanumeric(loadNumber)
		if cleanLoadNumber != "" {
			fileNamePrefix = []string{"recordings", cleanLoadNumber}
		}
	}

	// Step 1: Acquire resource
	acquireBody := map[string]interface{}{
		"cname": channelName,
		"uid":   uid,
		"clientRequest": map[string]interface{}{
			"resourceExpiredHour": 24,
			"scene":               0,
		},
	}

	acquireResp, err := c.makeRequest("/cloud_recording/acquire", "POST", acquireBody)
	if err != nil {
		return nil, fmt.Errorf("acquire failed: %w", err)
	}

	resourceID, ok := acquireResp["resourceId"].(string)
	if !ok || resourceID == "" {
		return nil, fmt.Errorf("failed to get resourceId from acquire response")
	}

	// Step 2: Start recording
	storageConfig := map[string]interface{}{
		"vendor":         c.s3Config.Vendor,
		"region":         c.s3Config.Region,
		"bucket":         c.s3Config.Bucket,
		"accessKey":      c.s3Config.AccessKey,
		"secretKey":      c.s3Config.SecretKey,
		"fileNamePrefix": fileNamePrefix,
	}

	startBody := map[string]interface{}{
		"cname": channelName,
		"uid":   uid,
		"clientRequest": map[string]interface{}{
			"token": token,
			"recordingConfig": map[string]interface{}{
				"channelType":     0,
				"streamTypes":     2,
				"maxIdleTime":     30,
				"streamMode":      "standard",
				"videoStreamType": 0,
				"transcodingConfig": map[string]interface{}{
					"width":            640,
					"height":           480,
					"fps":              15,
					"bitrate":          500,
					"mixedVideoLayout": 1,
					"backgroundColor":  "#000000",
				},
			},
			"recordingFileConfig": map[string]interface{}{
				"avFileType": []string{"hls", "mp4"},
			},
			"storageConfig": storageConfig,
		},
	}

	startEndpoint := fmt.Sprintf("/cloud_recording/resourceid/%s/mode/mix/start", resourceID)
	startResp, err := c.makeRequest(startEndpoint, "POST", startBody)
	if err != nil {
		return nil, fmt.Errorf("start recording failed: %w", err)
	}

	sid, ok := startResp["sid"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to get sid from start response")
	}

	return &RecordingResult{
		Success:    true,
		ResourceID: resourceID,
		SID:        sid,
	}, nil
}

// StopRecording stops cloud recording
func (c *Client) StopRecording(resourceID, sid, uid, channelName string) (*RecordingResult, error) {
	return c.stopRecordingWithRetry(resourceID, sid, uid, channelName, 0)
}

func (c *Client) stopRecordingWithRetry(resourceID, sid, uid, channelName string, retryCount int) (*RecordingResult, error) {
	stopBody := map[string]interface{}{
		"cname":         channelName,
		"uid":           uid,
		"clientRequest": map[string]interface{}{},
	}

	stopEndpoint := fmt.Sprintf("/cloud_recording/resourceid/%s/sid/%s/mode/mix/stop", resourceID, sid)
	stopResp, err := c.makeRequest(stopEndpoint, "POST", stopBody)
	if err != nil {
		// Check for error code 65 (network jitter)
		if strings.Contains(err.Error(), "65") || strings.Contains(err.Error(), "request not completed") {
			if retryCount < 2 {
				delay := time.Duration((retryCount+1)*3) * time.Second
				time.Sleep(delay)
				return c.stopRecordingWithRetry(resourceID, sid, uid, channelName, retryCount+1)
			}
		}
		return nil, fmt.Errorf("stop recording failed: %w", err)
	}

	result := &RecordingResult{
		Success: true,
	}

	// Extract file information from response
	var fileList []interface{}
	if fl, ok := stopResp["fileList"].([]interface{}); ok {
		fileList = fl
	} else if sr, ok := stopResp["serverResponse"].(map[string]interface{}); ok {
		if fl, ok := sr["fileList"].([]interface{}); ok {
			fileList = fl
		}
	}

	if len(fileList) > 0 {
		result.FileList = make([]string, 0, len(fileList))
		var selectedFile map[string]interface{}

		// Prefer MP4 files
		for _, f := range fileList {
			if fileInfo, ok := f.(map[string]interface{}); ok {
				if fileName, ok := fileInfo["fileName"].(string); ok {
					result.FileList = append(result.FileList, fileName)
					if strings.HasSuffix(fileName, ".mp4") && selectedFile == nil {
						selectedFile = fileInfo
					}
				}
			}
		}

		// If no MP4 found, use first file
		if selectedFile == nil && len(fileList) > 0 {
			selectedFile, _ = fileList[0].(map[string]interface{})
		}

		if selectedFile != nil {
			if fileName, ok := selectedFile["fileName"].(string); ok {
				result.FileName = fileName
				result.S3Key = fileName
				result.S3URL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s",
					c.s3Config.Bucket,
					os.Getenv("AWS_REGION"),
					fileName)
			}
			if fileSize, ok := selectedFile["fileSize"].(float64); ok {
				result.FileSize = int64(fileSize)
			}
			if duration, ok := selectedFile["duration"].(float64); ok {
				result.Duration = int(duration)
			}
		}
	}

	return result, nil
}

// QueryRecording queries the status of a recording
func (c *Client) QueryRecording(resourceID, sid string) (map[string]interface{}, error) {
	queryEndpoint := fmt.Sprintf("/cloud_recording/resourceid/%s/sid/%s/mode/mix/query", resourceID, sid)
	return c.makeRequest(queryEndpoint, "GET", nil)
}

// getAgoraS3Region converts AWS region to Agora S3 region code
func getAgoraS3Region(awsRegion string) int {
	// Agora S3 region codes: https://docs.agora.io/en/cloud-recording/reference/region-config
	regionMap := map[string]int{
		"us-east-1":      0,  // US_EAST_1
		"us-east-2":      1,  // US_EAST_2
		"us-west-1":      2,  // US_WEST_1
		"us-west-2":      3,  // US_WEST_2
		"eu-west-1":      4,  // EU_WEST_1
		"eu-west-2":      5,  // EU_WEST_2
		"eu-west-3":      6,  // EU_WEST_3
		"eu-central-1":   7,  // EU_CENTRAL_1
		"ap-southeast-1": 8,  // AP_SOUTHEAST_1
		"ap-southeast-2": 9,  // AP_SOUTHEAST_2
		"ap-northeast-1": 10, // AP_NORTHEAST_1
		"ap-northeast-2": 11, // AP_NORTHEAST_2
		"sa-east-1":      12, // SA_EAST_1
		"ca-central-1":   13, // CA_CENTRAL_1
		"ap-south-1":     14, // AP_SOUTH_1
		"cn-north-1":     15, // CN_NORTH_1
		"cn-northwest-1": 16, // CN_NORTHWEST_1
		"us-gov-west-1":  17, // US_GOV_WEST_1
	}

	if code, ok := regionMap[awsRegion]; ok {
		return code
	}
	return 0 // Default to US_EAST_1
}

// cleanAlphanumeric removes all non-alphanumeric characters from a string
// Agora requires folder names to be alphanumeric only
func cleanAlphanumeric(s string) string {
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		}
	}
	return result.String()
}
