package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmailService struct {
	apiKey string
	fromEmail string
}

func NewEmailService(apiKey string) *EmailService {
	return &EmailService{
		apiKey: apiKey,
		fromEmail: "Witness <onboarding@resend.dev>",
	}
}

type resendRequest struct {
	From string `json:"from"`
	To []string `json:"to"`
	Subject string `json:"subject"`
	HTML string `json:"html"`
}

func (s *EmailService) SendVerificationEmail(toEmail, rawToken string) error {
	verifyURL := fmt.Sprintf("http://localhost:8081/auth/verify-email?token=%s", rawToken)

	htmlContent := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; padding: 20px; color: #333;">
		<h2>Ласкаво просимо у Witness!</h2>
		<p>Дякуємо за реєстрацію. Будь ласка, підтвердіть свою електронну адресу, натиснувши на кнопку нижче:</p>
			<a href="%s" style="display: inline-block; padding: 12px 24px; background-color: #4F46E5; color: white; text-decoration: none; border-radius: 6px; font-weight: bold; margin: 20px 0;">Підтвердити пошту</a>
		<p>Або скопіюйте це посилання у браузер:</p>
			<p><a href="%s">%s</a></p>
			<p>Термін дії посилання — 24 години.</p>
			<p>Якщо ви не реєструвалися у Witness, просто проігноруйте цей лист.</p>
		</div>
	`, verifyURL, verifyURL, verifyURL)

	reqBody := resendRequest{
		From: s.fromEmail,
		To: []string{toEmail},
		Subject: "Підтвердження реєстрації у Witness",
		HTML: htmlContent,
	}

	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("помилка мфршалізації JSON для email: %w, err")
	}

	req, err := http.NewRequest("POST", "http://api.resend.com/emails", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return fmt.Errorf("помилка створення запиту до Resend: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("помилка відправки запиту до Resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("resend повернув помилку, статус код: %d", resp.StatusCode)
	}

	return nil
}