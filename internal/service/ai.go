package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AIService struct {
	apiKey string
}

func NewAIService(apiKey string) *AIService {
	return &AIService{apiKey: apiKey}
}

type AIAnalysisResult struct {
	IsSafe  bool     `json:"is_safe"`
	Tags    []string `json:"tags"`
	Summary string   `json:"summary"`
}

func (s *AIService) AnalyzeAndSummarize(ctx context.Context, text string) (*AIAnalysisResult, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(s.apiKey))
	if err != nil {
		return nil, fmt.Errorf("не вдалося створити клієнт Gemini: %w", err)
	}
	defer client.Close()


	model := client.GenerativeModel("gemini-flash-latest")
	model.ResponseMIMEType = "application/json"

	prompt := fmt.Sprintf(`Проаналізуй наступний текст для платформи духовних свідчень "Witness".
		Поверни виключно JSON-об'єкт із трьома полями:
		1. "is_safe" (boolean): false, якщо текст містить жорстоку нецензурну лексику, явну агресію, погрози або відвертий спам. Інакше true.
		2. "tags" (array of strings): від 1 до 3 релевантних тематичних тегів англійською мовою, маленькими літерами, БЕЗ знаку "#" (наприклад: ["faith", "hope", "trial"]).
		3. "summary" (string): дуже короткий витяг/анотація (TL;DR, максимум 1-2 речення) англійською мовою. Збережи автентичний та емоційний тон.

		Текст для аналізу: %s`, text)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		errStr := err.Error()

		log.Printf("🔴 [AIService] Помилка від Gemini API під час аналізу: %s\n", errStr)

		if strings.Contains(errStr, "429") || strings.Contains(strings.ToLower(errStr), "quota exceeded") {
			log.Println("⚠️ [AIService] Спрацював реальний Rate Limit (429). Повертаємо заглушку...")
			return &AIAnalysisResult{
				IsSafe:  true,
				Tags:    []string{"faith", "hope", "development-mode"},
				Summary: "This is a temporary fallback summary because Google Gemini API rate limit (429) was reached. Your backend flow is fully working!",
			}, nil
		}
		return nil, fmt.Errorf("помилка аналізу вмісту: %w", err)

	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("отримано порожню відповідь від Gemini під час аналізу")
	}

	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		sb.WriteString(fmt.Sprintf("%v", part))
	}
	jsonText := sb.String()

	var result AIAnalysisResult
	if err := json.Unmarshal([]byte(jsonText), &result); err != nil {
		return nil, fmt.Errorf("помилка десеріалізації відповіді ШІ: %w", err)
	}

	return &result, nil
}

func (s *AIService) TranslateToEnglish(ctx context.Context, text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(s.apiKey))
	if err != nil {
		return "", fmt.Errorf("не вдалося створити клієнт Gemini: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-flash-latest")

	prompt := fmt.Sprintf(`Переклади наступне духовне свідчення англійською мовою для платформи "Witness".
Збережи щирий, емоційний та автентичний тон автора.
Вимоги:
1. Поверни ВИКЛЮЧНО текст перекладу.
2. Не додавай вступних слів, коментарів чи пояснень (наприклад, НЕ пиши "Here is the translation:").
3. Не обгортай текст у подвійні лапки.

Текст для перекладу: %s`, text)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		errStr := err.Error()
		log.Printf("🔴 [AIService] Помилка перекладу від Gemini API: %s\n", errStr)

		if strings.Contains(errStr, "429") || strings.Contains(strings.ToLower(errStr), "quota exceeded") {
			log.Println("⚠️ [AIService] Rate Limit (429) під час перекладу. Повертаємо оригінал...")
			return text, nil
		}

		return "", fmt.Errorf("помилка виконання перекладу: %w", err)
	}

	if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("отримано порожню відповідь від Gemini під час перекладу")
	}

	var sb strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		sb.WriteString(fmt.Sprintf("%v", part))
	}

	translatedText := strings.TrimSpace(sb.String())
	return translatedText, nil
}