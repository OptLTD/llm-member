package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"

	"swiflow-auth/internal/config"
	"swiflow-auth/internal/models"
)

type LLMService struct {
	config *config.Config
}

func NewLLMService(cfg *config.Config) *LLMService {
	return &LLMService{
		config: cfg,
	}
}

func (s *LLMService) ChatCompletions(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	provider := s.getProviderFromModel(req.Model)
	
	switch provider {
	case "openai":
		return s.callOpenAI(ctx, req)
	case "claude":
		return s.callClaude(ctx, req)
	case "qwen":
		return s.callQwen(ctx, req)
	case "doubao":
		return s.callDoubao(ctx, req)
	case "bigmodel":
		return s.callBigModel(ctx, req)
	case "grok":
		return s.callGrok(ctx, req)
	case "gemini":
		return s.callGemini(ctx, req)
	case "openrouter":
		return s.callOpenRouter(ctx, req)
	case "siliconflow":
		return s.callSiliconFlow(ctx, req)
	case "openai-like":
		return s.callOpenAILike(ctx, req)
	default:
		return nil, fmt.Errorf("unsupported model: %s", req.Model)
	}
}

func (s *LLMService) GetModels() []models.ModelInfo {
	var modelList []models.ModelInfo

	// OpenAI 模型
	if s.config.OpenAIAPIKey != "" {
		modelList = append(modelList, []models.ModelInfo{
			{ID: "gpt-4", Object: "model", Provider: "openai", Name: "GPT-4"},
			{ID: "gpt-4-turbo", Object: "model", Provider: "openai", Name: "GPT-4 Turbo"},
			{ID: "gpt-3.5-turbo", Object: "model", Provider: "openai", Name: "GPT-3.5 Turbo"},
			{ID: "gpt-4o", Object: "model", Provider: "openai", Name: "GPT-4o"},
			{ID: "gpt-4o-mini", Object: "model", Provider: "openai", Name: "GPT-4o Mini"},
		}...)
	}

	// Claude 模型
	if s.config.ClaudeAPIKey != "" {
		modelList = append(modelList, []models.ModelInfo{
			{ID: "claude-3-5-sonnet-20241022", Object: "model", Provider: "claude", Name: "Claude 3.5 Sonnet"},
			{ID: "claude-3-5-haiku-20241022", Object: "model", Provider: "claude", Name: "Claude 3.5 Haiku"},
			{ID: "claude-3-opus-20240229", Object: "model", Provider: "claude", Name: "Claude 3 Opus"},
		}...)
	}

	// 通义千问模型
	if s.config.QwenAPIKey != "" {
		modelList = append(modelList, []models.ModelInfo{
			{ID: "qwen-turbo", Object: "model", Provider: "qwen", Name: "通义千问 Turbo"},
			{ID: "qwen-plus", Object: "model", Provider: "qwen", Name: "通义千问 Plus"},
			{ID: "qwen-max", Object: "model", Provider: "qwen", Name: "通义千问 Max"},
			{ID: "qwen2.5-72b-instruct", Object: "model", Provider: "qwen", Name: "通义千问 2.5 72B"},
		}...)
	}

	// 豆包模型
	if s.config.DoubaoAPIKey != "" {
		modelList = append(modelList, []models.ModelInfo{
			{ID: "doubao-pro-4k", Object: "model", Provider: "doubao", Name: "豆包 Pro 4K"},
			{ID: "doubao-pro-32k", Object: "model", Provider: "doubao", Name: "豆包 Pro 32K"},
			{ID: "doubao-lite-4k", Object: "model", Provider: "doubao", Name: "豆包 Lite 4K"},
		}...)
	}

	// 智谱清言模型
	if s.config.BigModelAPIKey != "" {
		modelList = append(modelList, []models.ModelInfo{
			{ID: "glm-4", Object: "model", Provider: "bigmodel", Name: "GLM-4"},
			{ID: "glm-4-plus", Object: "model", Provider: "bigmodel", Name: "GLM-4 Plus"},
			{ID: "glm-4-air", Object: "model", Provider: "bigmodel", Name: "GLM-4 Air"},
			{ID: "glm-4-flash", Object: "model", Provider: "bigmodel", Name: "GLM-4 Flash"},
		}...)
	}

	// Grok 模型
	if s.config.GrokAPIKey != "" {
		modelList = append(modelList, []models.ModelInfo{
			{ID: "grok-beta", Object: "model", Provider: "grok", Name: "Grok Beta"},
			{ID: "grok-vision-beta", Object: "model", Provider: "grok", Name: "Grok Vision Beta"},
		}...)
	}

	// Gemini 模型
	if s.config.GeminiAPIKey != "" {
		modelList = append(modelList, []models.ModelInfo{
			{ID: "gemini-1.5-pro", Object: "model", Provider: "gemini", Name: "Gemini 1.5 Pro"},
			{ID: "gemini-1.5-flash", Object: "model", Provider: "gemini", Name: "Gemini 1.5 Flash"},
			{ID: "gemini-pro", Object: "model", Provider: "gemini", Name: "Gemini Pro"},
		}...)
	}

	// OpenRouter 模型
	if s.config.OpenRouterAPIKey != "" {
		modelList = append(modelList, []models.ModelInfo{
			{ID: "openai/gpt-4o", Object: "model", Provider: "openrouter", Name: "GPT-4o (OpenRouter)"},
			{ID: "anthropic/claude-3.5-sonnet", Object: "model", Provider: "openrouter", Name: "Claude 3.5 Sonnet (OpenRouter)"},
			{ID: "google/gemini-pro-1.5", Object: "model", Provider: "openrouter", Name: "Gemini Pro 1.5 (OpenRouter)"},
		}...)
	}

	// SiliconFlow 模型
	if s.config.SiliconFlowAPIKey != "" {
		modelList = append(modelList, []models.ModelInfo{
			{ID: "deepseek-chat", Object: "model", Provider: "siliconflow", Name: "DeepSeek Chat"},
			{ID: "qwen/qwen2.5-72b-instruct", Object: "model", Provider: "siliconflow", Name: "Qwen2.5 72B"},
			{ID: "meta-llama/llama-3.1-405b-instruct", Object: "model", Provider: "siliconflow", Name: "Llama 3.1 405B"},
		}...)
	}

	// OpenAI-Like 模型
	if s.config.OpenAILikeAPIKey != "" {
		modelList = append(modelList, []models.ModelInfo{
			{ID: "custom-model", Object: "model", Provider: "openai-like", Name: "Custom Model"},
		}...)
	}

	return modelList
}

func (s *LLMService) getProviderFromModel(model string) string {
	// OpenAI 模型
	if strings.HasPrefix(model, "gpt-") {
		return "openai"
	}
	
	// Claude 模型
	if strings.HasPrefix(model, "claude-") {
		return "claude"
	}
	
	// 通义千问模型
	if strings.HasPrefix(model, "qwen-") || strings.HasPrefix(model, "qwen2") {
		return "qwen"
	}
	
	// 豆包模型
	if strings.HasPrefix(model, "doubao-") {
		return "doubao"
	}
	
	// 智谱清言模型
	if strings.HasPrefix(model, "glm-") {
		return "bigmodel"
	}
	
	// Grok 模型
	if strings.HasPrefix(model, "grok-") {
		return "grok"
	}
	
	// Gemini 模型
	if strings.HasPrefix(model, "gemini-") {
		return "gemini"
	}
	
	// OpenRouter 模型（包含斜杠的模型名）
	if strings.Contains(model, "/") {
		return "openrouter"
	}
	
	// SiliconFlow 模型
	if model == "deepseek-chat" || strings.Contains(model, "llama") || strings.Contains(model, "qwen/") {
		return "siliconflow"
	}
	
	// OpenAI-Like 模型
	if model == "custom-model" {
		return "openai-like"
	}
	
	return "unknown"
}

func (s *LLMService) callOpenAI(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	config := openai.DefaultConfig(s.config.OpenAIAPIKey)
	config.BaseURL = s.config.OpenAIBaseURL
	client := openai.NewClientWithConfig(config)

	// 转换消息格式
	var messages []openai.ChatCompletionMessage
	for _, msg := range req.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	// 构建请求
	chatReq := openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: messages,
	}

	if req.Temperature != nil {
		chatReq.Temperature = *req.Temperature
	}
	if req.MaxTokens != nil {
		chatReq.MaxTokens = *req.MaxTokens
	}
	if req.TopP != nil {
		chatReq.TopP = *req.TopP
	}

	// 调用 API
	resp, err := client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, err
	}

	// 转换响应格式
	var choices []models.ChatChoice
	for i, choice := range resp.Choices {
		choices = append(choices, models.ChatChoice{
			Index: i,
			Message: models.ChatMessage{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
			FinishReason: string(choice.FinishReason),
		})
	}

	return &models.ChatResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage: models.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (s *LLMService) callClaude(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// Claude API 调用实现
	// 这里简化实现，实际需要根据 Claude API 文档调整
	return s.callGenericAPI(ctx, req, s.config.ClaudeBaseURL, s.config.ClaudeAPIKey, "claude")
}

func (s *LLMService) callQwen(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// 通义千问 API 调用实现
	return s.callGenericAPI(ctx, req, s.config.QwenBaseURL, s.config.QwenAPIKey, "qwen")
}

func (s *LLMService) callGenericAPI(ctx context.Context, req *models.ChatRequest, baseURL, apiKey, provider string) (*models.ChatResponse, error) {
	// 通用 API 调用实现
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s", string(body))
	}

	var chatResp models.ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}

	// 如果响应中没有 ID，生成一个
	if chatResp.ID == "" {
		chatResp.ID = "chatcmpl-" + uuid.New().String()
	}

	// 设置时间戳
	if chatResp.Created == 0 {
		chatResp.Created = time.Now().Unix()
	}

	return &chatResp, nil
}

// 豆包 API 调用
func (s *LLMService) callDoubao(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	return s.callGenericAPI(ctx, req, s.config.DoubaoBaseURL, s.config.DoubaoAPIKey, "doubao")
}

// 智谱清言 API 调用
func (s *LLMService) callBigModel(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	return s.callGenericAPI(ctx, req, s.config.BigModelBaseURL, s.config.BigModelAPIKey, "bigmodel")
}

// Grok API 调用
func (s *LLMService) callGrok(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	return s.callGenericAPI(ctx, req, s.config.GrokBaseURL, s.config.GrokAPIKey, "grok")
}

// Gemini API 调用
func (s *LLMService) callGemini(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// Gemini 使用不同的 API 格式，这里简化处理
	return s.callGenericAPI(ctx, req, s.config.GeminiBaseURL, s.config.GeminiAPIKey, "gemini")
}

// OpenRouter API 调用
func (s *LLMService) callOpenRouter(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	return s.callGenericAPI(ctx, req, s.config.OpenRouterBaseURL, s.config.OpenRouterAPIKey, "openrouter")
}

// SiliconFlow API 调用
func (s *LLMService) callSiliconFlow(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	return s.callGenericAPI(ctx, req, s.config.SiliconFlowBaseURL, s.config.SiliconFlowAPIKey, "siliconflow")
}

// OpenAI-Like API 调用
func (s *LLMService) callOpenAILike(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	return s.callGenericAPI(ctx, req, s.config.OpenAILikeBaseURL, s.config.OpenAILikeAPIKey, "openai-like")
}