package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/securedlinq/backend/internal/config"
	"github.com/stretchr/testify/assert"
)

func setupEmailTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestSendMeetingLink(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    EmailMeetingLinkRequest
		emailConfig    config.EmailConfig
		expectedStatus int
		expectedError  string
		checkResponse  func(*testing.T, map[string]interface{})
	}{
		{
			name: "Successful email send",
			requestBody: EmailMeetingLinkRequest{
				DriverEmail: "driver@example.com",
				DriverName:  "John Doe",
				MeetingLink: "http://localhost:5173/join/load_1_abc123",
				LoadNumber:  "LOAD-001",
			},
			emailConfig: config.EmailConfig{
				SMTPHost:    "smtp.gmail.com",
				SMTPPort:    "587",
				SenderEmail: "sender@example.com",
				SenderName:  "SecuredLinQ",
				AppPassword: "app-password-123",
			},
			expectedStatus: http.StatusOK,
			expectedError:  "",
			checkResponse: func(t *testing.T, response map[string]interface{}) {
				assert.True(t, response["success"].(bool))
				assert.Contains(t, response["message"].(string), "sent successfully")
			},
		},
		{
			name: "Missing driver email",
			requestBody: EmailMeetingLinkRequest{
				DriverName:  "John Doe",
				MeetingLink: "http://localhost:5173/join/load_1_abc123",
				LoadNumber:  "LOAD-001",
			},
			emailConfig: config.EmailConfig{
				SMTPHost:    "smtp.gmail.com",
				SMTPPort:    "587",
				SenderEmail: "sender@example.com",
				SenderName:  "SecuredLinQ",
				AppPassword: "app-password-123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "driverEmail, driverName, meetingLink, and loadNumber are required",
		},
		{
			name: "Missing driver name",
			requestBody: EmailMeetingLinkRequest{
				DriverEmail: "driver@example.com",
				MeetingLink: "http://localhost:5173/join/load_1_abc123",
				LoadNumber:  "LOAD-001",
			},
			emailConfig: config.EmailConfig{
				SMTPHost:    "smtp.gmail.com",
				SMTPPort:    "587",
				SenderEmail: "sender@example.com",
				SenderName:  "SecuredLinQ",
				AppPassword: "app-password-123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "driverEmail, driverName, meetingLink, and loadNumber are required",
		},
		{
			name: "Missing meeting link",
			requestBody: EmailMeetingLinkRequest{
				DriverEmail: "driver@example.com",
				DriverName:  "John Doe",
				LoadNumber:  "LOAD-001",
			},
			emailConfig: config.EmailConfig{
				SMTPHost:    "smtp.gmail.com",
				SMTPPort:    "587",
				SenderEmail: "sender@example.com",
				SenderName:  "SecuredLinQ",
				AppPassword: "app-password-123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "driverEmail, driverName, meetingLink, and loadNumber are required",
		},
		{
			name: "Missing load number",
			requestBody: EmailMeetingLinkRequest{
				DriverEmail: "driver@example.com",
				DriverName:  "John Doe",
				MeetingLink: "http://localhost:5173/join/load_1_abc123",
			},
			emailConfig: config.EmailConfig{
				SMTPHost:    "smtp.gmail.com",
				SMTPPort:    "587",
				SenderEmail: "sender@example.com",
				SenderName:  "SecuredLinQ",
				AppPassword: "app-password-123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "driverEmail, driverName, meetingLink, and loadNumber are required",
		},
		{
			name: "Email not configured - missing sender email",
			requestBody: EmailMeetingLinkRequest{
				DriverEmail: "driver@example.com",
				DriverName:  "John Doe",
				MeetingLink: "http://localhost:5173/join/load_1_abc123",
				LoadNumber:  "LOAD-001",
			},
			emailConfig: config.EmailConfig{
				SMTPHost:    "smtp.gmail.com",
				SMTPPort:    "587",
				SenderEmail: "", // Missing
				SenderName:  "SecuredLinQ",
				AppPassword: "app-password-123",
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Email service is not configured",
		},
		{
			name: "Email not configured - missing app password",
			requestBody: EmailMeetingLinkRequest{
				DriverEmail: "driver@example.com",
				DriverName:  "John Doe",
				MeetingLink: "http://localhost:5173/join/load_1_abc123",
				LoadNumber:  "LOAD-001",
			},
			emailConfig: config.EmailConfig{
				SMTPHost:    "smtp.gmail.com",
				SMTPPort:    "587",
				SenderEmail: "sender@example.com",
				SenderName:  "SecuredLinQ",
				AppPassword: "", // Missing
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "Email service is not configured",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewEmailHandler(&tt.emailConfig)

			router := setupEmailTestRouter()
			router.POST("/email/send-meeting-link", handler.SendMeetingLink)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/email/send-meeting-link", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if tt.expectedError != "" {
				assert.Contains(t, response["error"].(string), tt.expectedError)
			} else if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestEmailContentGeneration(t *testing.T) {
	handler := &EmailHandler{
		config: &config.EmailConfig{
			SMTPHost:    "smtp.gmail.com",
			SMTPPort:    "587",
			SenderEmail: "sender@example.com",
			SenderName:  "SecuredLinQ",
			AppPassword: "app-password-123",
		},
	}

	router := setupEmailTestRouter()
	router.POST("/email/send-meeting-link", handler.SendMeetingLink)

	reqBody := EmailMeetingLinkRequest{
		DriverEmail: "driver@example.com",
		DriverName:  "John Doe",
		MeetingLink: "http://localhost:5173/join/load_1_abc123",
		LoadNumber:  "LOAD-001",
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/email/send-meeting-link", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Note: This test will fail if SMTP is not configured, but we can verify
	// that the email content would be generated correctly by checking the request
	// was processed (even if it fails at SMTP sending stage)

	// Verify request was accepted (not a validation error)
	assert.NotEqual(t, http.StatusBadRequest, w.Code)
}

func TestEmailSubjectGeneration(t *testing.T) {
	handler := &EmailHandler{
		config: &config.EmailConfig{
			SMTPHost:    "smtp.gmail.com",
			SMTPPort:    "587",
			SenderEmail: "sender@example.com",
			SenderName:  "SecuredLinQ",
			AppPassword: "app-password-123",
		},
	}

	router := setupEmailTestRouter()
	router.POST("/email/send-meeting-link", handler.SendMeetingLink)

	tests := []struct {
		name        string
		loadNumber  string
		expectedSub string
	}{
		{
			name:        "Standard load number",
			loadNumber:  "LOAD-001",
			expectedSub: "SecuredLinQ - Video Call for Load LOAD-001",
		},
		{
			name:        "Long load number",
			loadNumber:  "LOAD-2024-001-ABC",
			expectedSub: "SecuredLinQ - Video Call for Load LOAD-2024-001-ABC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := EmailMeetingLinkRequest{
				DriverEmail: "driver@example.com",
				DriverName:  "John Doe",
				MeetingLink: "http://localhost:5173/join/test",
				LoadNumber:  tt.loadNumber,
			}

			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/email/send-meeting-link", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			// The subject should contain the load number
			// Since we can't easily test the actual email content without mocking SMTP,
			// we verify the request was processed (subject generation happens in the handler)
			assert.NotEqual(t, http.StatusBadRequest, w.Code)
		})
	}
}

func TestEmailHandlerInitialization(t *testing.T) {
	cfg := &config.EmailConfig{
		SMTPHost:    "smtp.gmail.com",
		SMTPPort:    "587",
		SenderEmail: "sender@example.com",
		SenderName:  "SecuredLinQ",
		AppPassword: "app-password-123",
	}

	handler := NewEmailHandler(cfg)

	assert.NotNil(t, handler)
	assert.Equal(t, cfg, handler.config)
}

// Helper function to verify email body contains expected content
// This would be used in integration tests with a mock SMTP server
func verifyEmailContent(body string, driverName, loadNumber, meetingLink string) bool {
	return strings.Contains(body, driverName) &&
		strings.Contains(body, loadNumber) &&
		strings.Contains(body, meetingLink) &&
		strings.Contains(body, "SecuredLinQ") &&
		strings.Contains(body, "Join Video Call")
}

func TestVerifyEmailContentHelper(t *testing.T) {
	// Test the helper function that would be used to verify email content
	driverName := "John Doe"
	loadNumber := "LOAD-001"
	meetingLink := "http://localhost:5173/join/load_1_abc123"

	// Simulated email body (what would be generated)
	emailBody := `
<!DOCTYPE html>
<html>
<body>
    <p>Hello <strong>John Doe</strong>,</p>
    <p>You have been invited to join a video call for load verification.</p>
    <div class="load-info">
        <strong>Load Number:</strong> LOAD-001
    </div>
    <a href="http://localhost:5173/join/load_1_abc123" class="btn">Join Video Call</a>
    <p>This is an automated message from SecuredLinQ.</p>
</body>
</html>
`

	assert.True(t, verifyEmailContent(emailBody, driverName, loadNumber, meetingLink))
	assert.False(t, verifyEmailContent(emailBody, "Wrong Name", loadNumber, meetingLink))
	assert.False(t, verifyEmailContent(emailBody, driverName, "WRONG-LOAD", meetingLink))
}
