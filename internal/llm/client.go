package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ielts_bot/internal/config"
	"ielts_bot/internal/models"
)

type Client interface {
	AssessEssay(ctx context.Context, topic, essay string) (*models.AssessmentResult, error)
}

type LLMClient struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *LLMClient {
	return &LLMClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (c *LLMClient) AssessEssay(ctx context.Context, topic, essay string) (*models.AssessmentResult, error) {
	userPrompt := BuildUserPrompt(topic, essay)
	var rawResponse string
	var err error

	if strings.ToLower(c.cfg.LLMProvider) == "openai" {
		rawResponse, err = c.callOpenAI(ctx, userPrompt)
	} else {
		rawResponse, err = c.callGemini(ctx, userPrompt)
	}

	if err != nil {
		return nil, fmt.Errorf("llm request failed: %w", err)
	}

	result := parseAssessmentResult(rawResponse)
	return result, nil
}

// callGemini communicates with Google Gemini API
func (c *LLMClient) callGemini(ctx context.Context, userPrompt string) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.cfg.LLMModel, c.cfg.LLMAPIKey)

	payload := map[string]interface{}{
		"system_instruction": map[string]interface{}{
			"parts": []map[string]string{
				{"text": SystemPrompt},
			},
		},
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": userPrompt},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature": 0.2,
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini api error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response candidates returned by Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// callOpenAI communicates with OpenAI Chat Completions API
func (c *LLMClient) callOpenAI(ctx context.Context, userPrompt string) (string, error) {
	baseURL := "https://api.openai.com/v1"
	if c.cfg.LLMBaseURL != "" {
		baseURL = strings.TrimRight(c.cfg.LLMBaseURL, "/")
	}
	url := baseURL + "/chat/completions"

	payload := map[string]interface{}{
		"model": c.cfg.LLMModel,
		"messages": []map[string]string{
			{"role": "system", "content": SystemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.2,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.cfg.LLMAPIKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("openai api error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	var openAIResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBytes, &openAIResp); err != nil {
		return "", err
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned by OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// Regex patterns to parse scores from text
var (
	overallRegex = regexp.MustCompile(`(?i)OVERALL(?:\s*SCORE)?\*?:\s*\*?(\d+)\s*/\s*75\s*(?:[—\-–]\s*\*?([A-Za-z0-9]+))?`)
	trRegex      = regexp.MustCompile(`(?i)T/R[^\d]*?(\d+)\s*/\s*75`)
	ccRegex      = regexp.MustCompile(`(?i)C/C[^\d]*?(\d+)\s*/\s*75`)
	gaRegex      = regexp.MustCompile(`(?i)G/A[^\d]*?(\d+)\s*/\s*75`)
	lrRegex      = regexp.MustCompile(`(?i)L/R[^\d]*?(\d+)\s*/\s*75`)
)

func parseAssessmentResult(text string) *models.AssessmentResult {
	res := &models.AssessmentResult{
		Feedback: text,
	}

	// Overall & Level
	if match := overallRegex.FindStringSubmatch(text); len(match) > 1 {
		if s, err := strconv.Atoi(match[1]); err == nil {
			res.OverallScore = s
		}
		if len(match) > 2 && match[2] != "" {
			res.CEFRLevel = strings.ToUpper(match[2])
		}
	}

	// T/R
	if match := trRegex.FindStringSubmatch(text); len(match) > 1 {
		if s, err := strconv.Atoi(match[1]); err == nil {
			res.TRScore = s
		}
	}

	// C/C
	if match := ccRegex.FindStringSubmatch(text); len(match) > 1 {
		if s, err := strconv.Atoi(match[1]); err == nil {
			res.CCScore = s
		}
	}

	// G/A
	if match := gaRegex.FindStringSubmatch(text); len(match) > 1 {
		if s, err := strconv.Atoi(match[1]); err == nil {
			res.GAScore = s
		}
	}

	// L/R
	if match := lrRegex.FindStringSubmatch(text); len(match) > 1 {
		if s, err := strconv.Atoi(match[1]); err == nil {
			res.LRScore = s
		}
	}

	// Fallback CEFR level determination if not matched
	if res.CEFRLevel == "" && res.OverallScore > 0 {
		switch {
		case res.OverallScore >= 65:
			res.CEFRLevel = "C1"
		case res.OverallScore >= 51:
			res.CEFRLevel = "B2"
		case res.OverallScore >= 41:
			res.CEFRLevel = "B1"
		default:
			res.CEFRLevel = "A2"
		}
	}

	return res
}
