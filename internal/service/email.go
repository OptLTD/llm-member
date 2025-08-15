package service

import (
	"fmt"
	"llm-member/internal/model"
)

type EmailService struct {
}

func NewEmailService() *EmailService {
	return &EmailService{}
}

// SendSigninCodeEmail logs the signin verification email to the console.
func (s *EmailService) SendSigninCodeEmail(token *model.TokenModel) error {
	verificationCode := token.Token
	fmt.Printf("---- Sending Signin Verification Email ----\n")
	fmt.Printf("To: %s\n", token.Email)
	fmt.Printf("Subject: Your signin verification code\n")
	fmt.Printf("Body: Your signin verification code is: %s\n", verificationCode)
	fmt.Printf("-------------------------------------\n")
	return nil
}

// SendSignupCodeEmail logs the signup verification email to the console.
func (s *EmailService) SendSignupCodeEmail(token *model.TokenModel) error {
	link := fmt.Sprintf("http://localhost:8080/verify-email?code=%s", token.Token)
	fmt.Printf("---- Sending Signup Verification Email ----\n")
	fmt.Printf("To: %s\n", token.Email)
	fmt.Printf("Subject: Please verify your email address for signup\n")
	fmt.Printf("Body: Click the following link to verify your email: %s\n", link)
	fmt.Printf("-------------------------------------\n")
	return nil
}

// SendResetCodeEmail logs the password reset verification email to the console.
func (s *EmailService) SendResetCodeEmail(token *model.TokenModel) error {
	link := fmt.Sprintf("http://localhost:8080/reset-password?code=%s", token.Token)
	fmt.Printf("---- Sending Reset Verification Email ----\n")
	fmt.Printf("To: %s\n", token.Email)
	fmt.Printf("Subject: Reset your password\n")
	fmt.Printf("Body: Click the following link to reset your password: %s\n", link)
	fmt.Printf("-------------------------------------\n")
	return nil
}
