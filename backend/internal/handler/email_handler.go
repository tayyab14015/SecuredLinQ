package handler

import (
	"fmt"
	"net/http"
	"net/smtp"

	"github.com/gin-gonic/gin"
	"github.com/securedlinq/backend/internal/config"
)

// EmailHandler handles email HTTP requests
type EmailHandler struct {
	config *config.EmailConfig
}

// NewEmailHandler creates a new email handler
func NewEmailHandler(cfg *config.EmailConfig) *EmailHandler {
	return &EmailHandler{
		config: cfg,
	}
}

// EmailMeetingLinkRequest represents a meeting link email request
type EmailMeetingLinkRequest struct {
	DriverEmail   string `json:"driverEmail" binding:"required"`
	DriverName    string `json:"driverName" binding:"required"`
	MeetingLink   string `json:"meetingLink" binding:"required"`
	LoadNumber    string `json:"loadNumber" binding:"required"`
}

// SendMeetingLink sends a meeting link via email
func (h *EmailHandler) SendMeetingLink(c *gin.Context) {
	var req EmailMeetingLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "driverEmail, driverName, meetingLink, and loadNumber are required"})
		return
	}

	// Check if email is configured
	if h.config.SenderEmail == "" || h.config.AppPassword == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Email service is not configured"})
		return
	}

	// Create email content
	subject := fmt.Sprintf("SecuredLinQ - Video Call for Load %s", req.LoadNumber)
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #1e3a5f 0%%, #2e5a8f 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f8f9fa; padding: 30px; border-radius: 0 0 10px 10px; }
        .btn { display: inline-block; background: #f59e0b; color: white; padding: 15px 30px; text-decoration: none; border-radius: 8px; font-weight: bold; margin: 20px 0; }
        .btn:hover { background: #d97706; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #666; }
        .load-info { background: white; padding: 15px; border-radius: 8px; margin: 15px 0; border-left: 4px solid #f59e0b; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🎥 SecuredLinQ Video Call</h1>
        </div>
        <div class="content">
            <p>Hello <strong>%s</strong>,</p>
            <p>You have been invited to join a video call for load verification.</p>
            
            <div class="load-info">
                <strong>Load Number:</strong> %s
            </div>
            
            <p>Click the button below to join the video call:</p>
            
            <center>
                <a href="%s" class="btn">Join Video Call</a>
            </center>
            
            <p style="font-size: 12px; color: #666; margin-top: 20px;">
                If the button doesn't work, copy and paste this link into your browser:<br>
                <a href="%s">%s</a>
            </p>
        </div>
        <div class="footer">
            <p>This is an automated message from SecuredLinQ.</p>
        </div>
    </div>
</body>
</html>
`, req.DriverName, req.LoadNumber, req.MeetingLink, req.MeetingLink, req.MeetingLink)

	// Send email
	err := h.sendEmail(req.DriverEmail, subject, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to send email: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Meeting link sent successfully",
	})
}

// sendEmail sends an email using SMTP
func (h *EmailHandler) sendEmail(to, subject, body string) error {
	from := h.config.SenderEmail
	password := h.config.AppPassword
	smtpHost := h.config.SMTPHost
	smtpPort := h.config.SMTPPort

	// Email headers
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", h.config.SenderName, from)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	// Build message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Auth
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("smtp error: %w", err)
	}

	return nil
}

