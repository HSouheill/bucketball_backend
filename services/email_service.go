package services

import (
	"fmt"
	"net/smtp"

	"github.com/HSouheil/bucketball_backend/config"
)

// EmailService handles email sending
type EmailService struct {
	config *config.EmailConfig
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.EmailConfig) *EmailService {
	return &EmailService{
		config: cfg,
	}
}

// SendOTPEmail sends an OTP verification email
func (s *EmailService) SendOTPEmail(toEmail, username, otp string) error {
	subject := "Verify Your Email - BucketBall"
	body := s.buildOTPEmailBody(username, otp)

	return s.sendEmail(toEmail, subject, body)
}

// buildOTPEmailBody builds the HTML body for OTP email
func (s *EmailService) buildOTPEmailBody(username, otp string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Email Verification</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333; max-width: 600px; margin: 0 auto; padding: 20px;">
    <div style="background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); padding: 30px; text-align: center; border-radius: 10px 10px 0 0;">
        <h1 style="color: white; margin: 0;">BucketBall</h1>
    </div>
    
    <div style="background-color: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; border: 1px solid #ddd;">
        <h2 style="color: #667eea; margin-top: 0;">Welcome, %s!</h2>
        
        <p>Thank you for registering with BucketBall. To complete your registration, please verify your email address using the OTP code below:</p>
        
        <div style="background-color: white; border: 2px dashed #667eea; border-radius: 8px; padding: 20px; text-align: center; margin: 30px 0;">
            <h1 style="color: #667eea; font-size: 42px; letter-spacing: 10px; margin: 0;">%s</h1>
        </div>
        
        <p style="color: #666; font-size: 14px; margin-top: 20px;">This OTP code will expire in <strong>10 minutes</strong>.</p>
        
        <p style="color: #666; font-size: 14px;">If you didn't request this verification, please ignore this email.</p>
        
        <hr style="border: none; border-top: 1px solid #ddd; margin: 30px 0;">
        
        <p style="color: #999; font-size: 12px; text-align: center;">
            This is an automated email. Please do not reply.<br>
            © 2024 BucketBall. All rights reserved.
        </p>
    </div>
</body>
</html>
`, username, otp)
}

// sendEmail sends an email using SMTP
func (s *EmailService) sendEmail(to, subject, body string) error {
	from := s.config.FromEmail

	// Set up authentication
	auth := smtp.PlainAuth("",
		s.config.SMTPUsername,
		s.config.SMTPPassword,
		s.config.SMTPHost,
	)

	// Build email message
	message := s.buildEmailMessage(from, to, subject, body)

	// Send email
	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)
	err := smtp.SendMail(addr, auth, from, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// buildEmailMessage builds the email message with headers
func (s *EmailService) buildEmailMessage(from, to, subject, body string) string {
	headers := make(map[string]string)
	headers["From"] = fmt.Sprintf("%s <%s>", s.config.FromName, from)
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=UTF-8"

	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	return message
}
