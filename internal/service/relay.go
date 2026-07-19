package service

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/consts"
	"llm-member/internal/model"

	"github.com/sashabaranov/go-openai"
)

// APIConfig 通用 API 配置（统一 OpenAI-compatible）
type APIConfig struct {
	BaseURL  string
	APIKey   string
	Provider string
	ReqModel string
}

type RelayService struct{}

func NewRelayService() *RelayService {
	return &RelayService{}
}

// 共享 HTTP 客户端，复用连接
var sharedHTTPClient = &http.Client{
	Timeout: 0, // 由 context 控制超时
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ForceAttemptHTTP2:     true,
	},
}

func (s *RelayService) ChatCompletions(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	apiConfig, err := s.GetAPIConfig(req.Model)
	if err != nil {
		return nil, err
	}
	if req.Model == "" || req.Model == "auto-match" {
		req.Model = apiConfig.ReqModel
	}
	return s.callWithClient(ctx, req, apiConfig)
}

// ChatCompletionsStream 流式聊天完成
func (s *RelayService) ChatCompletionsStream(ctx context.Context, req *model.ChatRequest) (<-chan *model.ChatStreamResponse, <-chan error) {
	respChan := make(chan *model.ChatStreamResponse, 100)
	errorChan := make(chan error, 1)

	go func() {
		defer close(respChan)
		defer close(errorChan)

		apiConfig, err := s.GetAPIConfig(req.Model)
		if err != nil {
			errorChan <- err
			return
		}

		if req.Model == "" || req.Model == "auto-match" {
			req.Model = apiConfig.ReqModel
		}
		s.streamWithClient(ctx, req, apiConfig, respChan, errorChan)
	}()

	return respChan, errorChan
}

func isNativeNonCompatibleURL(provider, baseURL string) bool {
	u := strings.ToLower(baseURL)
	switch provider {
	case "claude":
		return strings.Contains(u, "api.anthropic.com")
	case "gemini":
		return strings.Contains(u, "generativelanguage.googleapis.com")
	default:
		return false
	}
}

func (s *RelayService) GetAPIConfig(modelName string) (*APIConfig, error) {
	apiConfig := &APIConfig{}
	apiConfig.Provider = s.GetProvider(modelName)
	if modelName == "" || modelName == "auto-match" {
		if models := s.GetModels(); len(models) > 0 {
			apiConfig.Provider = models[0].Provider
			apiConfig.ReqModel = models[0].ID
		}
	}
	if apiConfig.Provider == "unknown" {
		return nil, fmt.Errorf("%w: %s", consts.ErrUnsupportedModel, modelName)
	}

	providerName := apiConfig.Provider
	provider := config.GetProvider(providerName)
	if provider == nil {
		log.Printf("[LLM] Provider %s not found in config", providerName)
		return nil, fmt.Errorf("%w: %s", consts.ErrProviderNotConfigured, providerName)
	}
	apiConfig.APIKey = provider.APIKey
	apiConfig.BaseURL = provider.BaseURL

	if isNativeNonCompatibleURL(providerName, provider.BaseURL) {
		return nil, fmt.Errorf(
			"%w: provider %s BaseURL %q is not OpenAI-compatible; use a compatible gateway (e.g. OpenRouter) or set an OpenAI-compatible BaseURL",
			consts.ErrProviderNotConfigured, providerName, provider.BaseURL,
		)
	}

	return apiConfig, nil
}

func (s *RelayService) GetModels() []model.LLModelInfo {
	var modelList []model.LLModelInfo

	if config.HasProvider("openai") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "gpt-4", Object: "model", Provider: "openai", Name: "GPT-4"},
			{ID: "gpt-4-turbo", Object: "model", Provider: "openai", Name: "GPT-4 Turbo"},
			{ID: "gpt-3.5-turbo", Object: "model", Provider: "openai", Name: "GPT-3.5 Turbo"},
			{ID: "gpt-4o", Object: "model", Provider: "openai", Name: "GPT-4o"},
			{ID: "gpt-4o-mini", Object: "model", Provider: "openai", Name: "GPT-4o Mini"},
		}...)
	}

	if config.HasProvider("claude") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "claude-3-5-sonnet-20241022", Object: "model", Provider: "claude", Name: "Claude 3.5 Sonnet"},
			{ID: "claude-3-5-haiku-20241022", Object: "model", Provider: "claude", Name: "Claude 3.5 Haiku"},
			{ID: "claude-3-opus-20240229", Object: "model", Provider: "claude", Name: "Claude 3 Opus"},
		}...)
	}

	if config.HasProvider("qwen") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "qwen-turbo", Object: "model", Provider: "qwen", Name: "通义千问 Turbo"},
			{ID: "qwen-plus", Object: "model", Provider: "qwen", Name: "通义千问 Plus"},
			{ID: "qwen-max", Object: "model", Provider: "qwen", Name: "通义千问 Max"},
			{ID: "qwen2.5-72b-instruct", Object: "model", Provider: "qwen", Name: "通义千问 2.5 72B"},
		}...)
	}

	if config.HasProvider("doubao") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "doubao-seed-1-6-250615", Object: "model", Provider: "doubao", Name: "豆包 Seed 1.6"},
		}...)
	}

	if config.HasProvider("bigmodel") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "glm-4", Object: "model", Provider: "bigmodel", Name: "GLM-4"},
			{ID: "glm-4-plus", Object: "model", Provider: "bigmodel", Name: "GLM-4 Plus"},
			{ID: "glm-4-air", Object: "model", Provider: "bigmodel", Name: "GLM-4 Air"},
			{ID: "glm-4-flash", Object: "model", Provider: "bigmodel", Name: "GLM-4 Flash"},
		}...)
	}

	if config.HasProvider("grok") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "grok-beta", Object: "model", Provider: "grok", Name: "Grok Beta"},
			{ID: "grok-vision-beta", Object: "model", Provider: "grok", Name: "Grok Vision Beta"},
		}...)
	}

	if config.HasProvider("gemini") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "gemini-1.5-pro", Object: "model", Provider: "gemini", Name: "Gemini 1.5 Pro"},
			{ID: "gemini-1.5-flash", Object: "model", Provider: "gemini", Name: "Gemini 1.5 Flash"},
			{ID: "gemini-pro", Object: "model", Provider: "gemini", Name: "Gemini Pro"},
		}...)
	}

	if config.HasProvider("openrouter") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "openai/gpt-4o", Object: "model", Provider: "openrouter", Name: "GPT-4o (OpenRouter)"},
			{ID: "anthropic/claude-3.5-sonnet", Object: "model", Provider: "openrouter", Name: "Claude 3.5 Sonnet (OpenRouter)"},
			{ID: "google/gemini-pro-1.5", Object: "model", Provider: "openrouter", Name: "Gemini Pro 1.5 (OpenRouter)"},
		}...)
	}

	if config.HasProvider("siliconflow") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "qwen/qwen2.5-72b-instruct", Object: "model", Provider: "siliconflow", Name: "Qwen2.5 72B"},
			{ID: "meta-llama/llama-3.1-405b-instruct", Object: "model", Provider: "siliconflow", Name: "Llama 3.1 405B"},
		}...)
	}

	if config.HasProvider("deepseek") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "deepseek-chat", Object: "model", Provider: "deepseek", Name: "DeepSeek Chat"},
			{ID: "deepseek-coder", Object: "model", Provider: "deepseek", Name: "DeepSeek Coder"},
			{ID: "deepseek-reasoner", Object: "model", Provider: "deepseek", Name: "DeepSeek Reasoner"},
		}...)
	}

	if config.HasProvider("openai-like") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "custom-model", Object: "model", Provider: "openai-like", Name: "Custom Model"},
		}...)
	}

	return modelList
}

func (s *RelayService) GetProvider(modelName string) string {
	// 含 provider 前缀的斜杠模型：先按已知前缀区分
	if strings.Contains(modelName, "/") {
		lower := strings.ToLower(modelName)
		switch {
		case strings.HasPrefix(lower, "qwen/"),
			strings.HasPrefix(lower, "meta-llama/"),
			strings.HasPrefix(lower, "deepseek/"):
			if config.HasProvider("siliconflow") {
				return "siliconflow"
			}
		case strings.HasPrefix(lower, "openai/"),
			strings.HasPrefix(lower, "anthropic/"),
			strings.HasPrefix(lower, "google/"):
			if config.HasProvider("openrouter") {
				return "openrouter"
			}
		}
		if config.HasProvider("openrouter") {
			return "openrouter"
		}
		if config.HasProvider("siliconflow") {
			return "siliconflow"
		}
	}

	if strings.HasPrefix(modelName, "gpt-") {
		return "openai"
	}
	if strings.HasPrefix(modelName, "claude-") {
		return "claude"
	}
	if strings.HasPrefix(modelName, "qwen-") || strings.HasPrefix(modelName, "qwen2") {
		return "qwen"
	}
	if strings.HasPrefix(modelName, "doubao-") {
		return "doubao"
	}
	if strings.HasPrefix(modelName, "glm-") {
		return "bigmodel"
	}
	if strings.HasPrefix(modelName, "deepseek-") {
		return "deepseek"
	}
	if strings.HasPrefix(modelName, "grok-") {
		return "grok"
	}
	if strings.HasPrefix(modelName, "gemini-") {
		return "gemini"
	}
	if modelName == "custom-model" {
		return "openai-like"
	}
	return "unknown"
}

func (s *RelayService) callWithClient(ctx context.Context, req *model.ChatRequest, apiConfig *APIConfig) (*model.ChatResponse, error) {
	log.Printf("[LLM] model=%s base=%s", req.Model, apiConfig.BaseURL)

	cfg := openai.DefaultConfig(apiConfig.APIKey)
	cfg.BaseURL = apiConfig.BaseURL
	cfg.HTTPClient = sharedHTTPClient
	client := openai.NewClientWithConfig(cfg)

	var messages []openai.ChatCompletionMessage
	for _, msg := range req.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role: msg.Role, Content: msg.Content,
		})
	}

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

	resp, err := client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		log.Printf("[LLM] client error model=%s: %v", req.Model, err)
		return nil, err
	}

	var choices []model.ChatChoice
	for i, choice := range resp.Choices {
		choices = append(choices, model.ChatChoice{
			Index: i,
			Message: model.ChatMessage{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
			FinishReason: string(choice.FinishReason),
		})
	}

	return &model.ChatResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage: model.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (s *RelayService) streamWithClient(ctx context.Context, req *model.ChatRequest, apiConfig *APIConfig, responseChan chan<- *model.ChatStreamResponse, errorChan chan<- error) {
	log.Printf("[LLM] stream model=%s base=%s", req.Model, apiConfig.BaseURL)

	cfg := openai.DefaultConfig(apiConfig.APIKey)
	cfg.BaseURL = apiConfig.BaseURL
	cfg.HTTPClient = sharedHTTPClient
	client := openai.NewClientWithConfig(cfg)

	var messages []openai.ChatCompletionMessage
	for _, msg := range req.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role: msg.Role, Content: msg.Content,
		})
	}

	chatReq := openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   true,
		StreamOptions: &openai.StreamOptions{
			IncludeUsage: true,
		},
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

	stream, err := client.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		log.Printf("[LLM] stream error model=%s: %v", req.Model, err)
		errorChan <- err
		return
	}
	defer stream.Close()

	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("[LLM] stream recv error: %v", err)
			errorChan <- err
			return
		}

		streamResp := &model.ChatStreamResponse{
			ID:      response.ID,
			Object:  response.Object,
			Created: response.Created,
			Model:   response.Model,
		}

		if response.Usage != nil {
			streamResp.Usage = &model.Usage{
				PromptTokens:     response.Usage.PromptTokens,
				CompletionTokens: response.Usage.CompletionTokens,
				TotalTokens:      response.Usage.TotalTokens,
			}
		}

		for i, choice := range response.Choices {
			finishReason := (*string)(nil)
			if choice.FinishReason != "" {
				reason := string(choice.FinishReason)
				finishReason = &reason
			}
			streamResp.Choices = append(streamResp.Choices, model.ChatStreamChoice{
				Index: i, FinishReason: finishReason,
				Delta: model.ChatStreamDelta{
					Role:    choice.Delta.Role,
					Content: choice.Delta.Content,
				},
			})
		}

		select {
		case responseChan <- streamResp:
		case <-ctx.Done():
			return
		}
	}
}
