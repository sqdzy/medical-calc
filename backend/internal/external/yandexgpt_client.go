package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// YandexGPTClient provides access to YandexGPT API for text generation.
type YandexGPTClient struct {
	httpClient *http.Client
	apiKey     string
	iamToken   string
	folderID   string
	limiter    *rate.Limiter
	baseURL    string
	model      string
}

// NewYandexGPTClient creates a new YandexGPT client.
func NewYandexGPTClient(apiKey, folderID string) *YandexGPTClient {
	return &YandexGPTClient{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		apiKey:     apiKey,
		folderID:   folderID,
		limiter:    rate.NewLimiter(rate.Limit(1), 1), // 1 req/s (conservative)
		baseURL:    "https://llm.api.cloud.yandex.net/foundationModels/v1",
		model:      "yandexgpt-lite",
	}
}

// NewYandexGPTClientWithIAMToken creates a new YandexGPT client using an IAM token (Bearer).
func NewYandexGPTClientWithIAMToken(iamToken, folderID string) *YandexGPTClient {
	return &YandexGPTClient{
		httpClient: &http.Client{Timeout: 60 * time.Second},
		iamToken:   iamToken,
		folderID:   folderID,
		limiter:    rate.NewLimiter(rate.Limit(1), 1), // 1 req/s (conservative)
		baseURL:    "https://llm.api.cloud.yandex.net/foundationModels/v1",
		model:      "yandexgpt-lite",
	}
}

// SetModel sets the model to use (yandexgpt, yandexgpt-lite, etc.)
func (c *YandexGPTClient) SetModel(model string) {
	c.model = model
}

// Message represents a chat message.
type Message struct {
	Role string `json:"role"` // "system", "user", "assistant"
	Text string `json:"text"`
}

// CompletionRequest is the request to YandexGPT completion API.
type CompletionRequest struct {
	ModelURI          string            `json:"modelUri"`
	CompletionOptions CompletionOptions `json:"completionOptions"`
	Messages          []Message         `json:"messages"`
}

// CompletionOptions controls generation parameters.
type CompletionOptions struct {
	Stream      bool    `json:"stream"`
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"maxTokens"`
}

// CompletionResponse is the response from YandexGPT.
type CompletionResponse struct {
	Result struct {
		Alternatives []struct {
			Message struct {
				Role string `json:"role"`
				Text string `json:"text"`
			} `json:"message"`
			Status string `json:"status"`
		} `json:"alternatives"`
		Usage map[string]any `json:"usage"`
	} `json:"result"`
}

// Complete sends a completion request to YandexGPT.
func (c *YandexGPTClient) Complete(ctx context.Context, messages []Message, temperature float64, maxTokens int) (*CompletionResponse, error) {
	if c.folderID == "" {
		return nil, fmt.Errorf("yandexgpt: folder_id required")
	}
	if c.apiKey == "" && c.iamToken == "" {
		return nil, fmt.Errorf("yandexgpt: api_key or iam_token required")
	}

	if err := c.limiter.Wait(ctx); err != nil {
		return nil, err
	}

	modelURI := fmt.Sprintf("gpt://%s/%s", c.folderID, c.model)

	reqBody := CompletionRequest{
		ModelURI: modelURI,
		CompletionOptions: CompletionOptions{
			Stream:      false,
			Temperature: temperature,
			MaxTokens:   maxTokens,
		},
		Messages: messages,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/completion", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Api-Key "+c.apiKey)
	} else {
		req.Header.Set("Authorization", "Bearer "+c.iamToken)
	}
	req.Header.Set("x-folder-id", c.folderID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody map[string]any
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("yandexgpt: status %d: %v", resp.StatusCode, errBody)
	}

	var result CompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GenerateText is a convenience method for simple text generation.
func (c *YandexGPTClient) GenerateText(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	messages := []Message{
		{Role: "system", Text: systemPrompt},
		{Role: "user", Text: userPrompt},
	}

	resp, err := c.Complete(ctx, messages, 0.3, 2000)
	if err != nil {
		return "", err
	}

	if len(resp.Result.Alternatives) == 0 {
		return "", fmt.Errorf("yandexgpt: no alternatives returned")
	}

	return resp.Result.Alternatives[0].Message.Text, nil
}

// MedicalSummaryPrompts contains prompts for medical use cases.
var MedicalSummaryPrompts = struct {
	SurveyInterpretation  string
	TherapyRecommendation string
	DrugInteraction       string
}{
	SurveyInterpretation: `Ты — медицинский AI-ассистент. Твоя задача — интерпретировать результаты медицинских опросников (BVAS, DAS28, BASDAI и др.) для врача.
Отвечай кратко и по делу. Используй медицинскую терминологию. Укажи степень активности заболевания и возможные рекомендации по дальнейшему обследованию.`,

	TherapyRecommendation: `Ты — медицинский AI-ассистент, помогающий врачу с подбором биологической терапии (ГИБП).
На основе предоставленных данных о пациенте (диагноз, индексы активности, предыдущая терапия) предложи возможные варианты ГИБП-терапии.
Укажи механизм действия препаратов и возможные противопоказания. Это рекомендация для врача, не для пациента.`,

	DrugInteraction: `Ты — фармацевт-консультант. Проанализируй возможные лекарственные взаимодействия между указанными препаратами.
Укажи клинически значимые взаимодействия и рекомендации по их предотвращению.`,
}

// InterpretSurvey generates AI interpretation of survey results.
func (c *YandexGPTClient) InterpretSurvey(ctx context.Context, surveyType string, score float64, breakdown map[string]any) (string, error) {
	userPrompt := fmt.Sprintf(
		"Опросник: %s\nИтоговый балл: %.2f\nДетали: %v\n\nДай краткую интерпретацию результатов для врача.",
		surveyType, score, breakdown,
	)
	return c.GenerateText(ctx, MedicalSummaryPrompts.SurveyInterpretation, userPrompt)
}

// RecommendTherapy generates therapy recommendations based on patient data.
func (c *YandexGPTClient) RecommendTherapy(ctx context.Context, diagnosis string, activityScore float64, previousTherapy []string) (string, error) {
	userPrompt := fmt.Sprintf(
		"Диагноз: %s\nИндекс активности: %.2f\nПредыдущая терапия: %v\n\nПредложи варианты ГИБП-терапии.",
		diagnosis, activityScore, previousTherapy,
	)
	return c.GenerateText(ctx, MedicalSummaryPrompts.TherapyRecommendation, userPrompt)
}
