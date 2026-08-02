package relay

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

var (
	// ErrProviderNotConfigured 上游未配置
	ErrProviderNotConfigured = errors.New("provider not configured")
	// ErrUnsupportedModel 不支持的模型
	ErrUnsupportedModel = errors.New("unsupported model")
)

// APIConfig 通用 API 配置（统一 OpenAI-compatible）
type APIConfig struct {
	BaseURL  string
	APIKey   string
	Provider string
	ReqModel string
}

// Service LLM 中继服务，负责转发 OpenAI-compatible 请求。
type Service struct {
	getConfig func() *LLMConfig
}

// NewService 创建中继服务；getConfig 返回当前上游配置（例如按项目动态解析），为空时使用环境变量。
func NewService(getConfig func() *LLMConfig) *Service {
	if getConfig == nil {
		getConfig = ConfigFromEnv
	}
	return &Service{getConfig: getConfig}
}

// WithConfig 返回一个使用固定配置的 Service 副本（不修改原实例，并发安全）。
// 适用于按请求解析上游配置的场景。
func (s *Service) WithConfig(cfg *LLMConfig) *Service {
	return &Service{getConfig: func() *LLMConfig { return cfg }}
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

// ChatCompletions 非流式聊天完成
func (s *Service) ChatCompletions(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
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
func (s *Service) ChatCompletionsStream(ctx context.Context, req *ChatRequest) (<-chan *ChatStreamResponse, <-chan error) {
	respChan := make(chan *ChatStreamResponse, 100)
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

func isNativeNonCompatibleURL(baseURL string) bool {
	u := strings.ToLower(baseURL)
	return strings.Contains(u, "api.anthropic.com") ||
		strings.Contains(u, "generativelanguage.googleapis.com")
}

// GetAPIConfig 根据模型名解析上游配置
func (s *Service) GetAPIConfig(modelName string) (*APIConfig, error) {
	llm := s.getConfig()
	if llm == nil {
		return nil, ErrProviderNotConfigured
	}
	if isNativeNonCompatibleURL(llm.BaseURL) {
		return nil, fmt.Errorf(
			"%w: LLM_BASE_URL %q is not OpenAI-compatible; use a compatible gateway",
			ErrProviderNotConfigured, llm.BaseURL,
		)
	}

	reqModel := modelName
	if reqModel == "" || reqModel == "auto-match" {
		reqModel = llm.DefaultModel
		if reqModel == "" && len(llm.Models) > 0 {
			reqModel = llm.Models[0]
		}
		if reqModel == "" {
			return nil, fmt.Errorf("%w: set LLM_DEFAULT_MODEL or pass model in request", ErrUnsupportedModel)
		}
	}

	return &APIConfig{
		BaseURL:  llm.BaseURL,
		APIKey:   llm.APIKey,
		Provider: llm.Provider,
		ReqModel: reqModel,
	}, nil
}

// GetModels 返回对外暴露的模型列表
func (s *Service) GetModels() []LLModelInfo {
	llm := s.getConfig()
	if llm == nil {
		return nil
	}
	out := make([]LLModelInfo, 0, len(llm.Models))
	for _, id := range llm.Models {
		out = append(out, LLModelInfo{
			ID: id, Object: "model", Provider: llm.Provider, Name: id,
		})
	}
	return out
}

// GetProvider 返回当前上游标签（单一配置，与 model 无关）
func (s *Service) GetProvider(modelName string) string {
	_ = modelName
	if llm := s.getConfig(); llm != nil {
		return llm.Provider
	}
	return "unknown"
}

func (s *Service) callWithClient(ctx context.Context, req *ChatRequest, apiConfig *APIConfig) (*ChatResponse, error) {
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

	var choices []ChatChoice
	for i, choice := range resp.Choices {
		choices = append(choices, ChatChoice{
			Index: i,
			Message: ChatMessage{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
			},
			FinishReason: string(choice.FinishReason),
		})
	}

	return &ChatResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Choices: choices,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (s *Service) streamWithClient(ctx context.Context, req *ChatRequest, apiConfig *APIConfig, responseChan chan<- *ChatStreamResponse, errorChan chan<- error) {
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

		streamResp := &ChatStreamResponse{
			ID:      response.ID,
			Object:  response.Object,
			Created: response.Created,
			Model:   response.Model,
		}

		if response.Usage != nil {
			streamResp.Usage = &Usage{
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
			streamResp.Choices = append(streamResp.Choices, ChatStreamChoice{
				Index: i, FinishReason: finishReason,
				Delta: ChatStreamDelta{
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
