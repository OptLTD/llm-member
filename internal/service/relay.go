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

	"github.com/OptLTD/member/internal/config"
	"github.com/OptLTD/member/internal/consts"
	"github.com/OptLTD/member/internal/model"

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

func isNativeNonCompatibleURL(baseURL string) bool {
	u := strings.ToLower(baseURL)
	return strings.Contains(u, "api.anthropic.com") ||
		strings.Contains(u, "generativelanguage.googleapis.com")
}

func (s *RelayService) GetAPIConfig(modelName string) (*APIConfig, error) {
	llm := config.GetLLM()
	if llm == nil {
		return nil, consts.ErrProviderNotConfigured
	}
	if isNativeNonCompatibleURL(llm.BaseURL) {
		return nil, fmt.Errorf(
			"%w: LLM_BASE_URL %q is not OpenAI-compatible; use a compatible gateway",
			consts.ErrProviderNotConfigured, llm.BaseURL,
		)
	}

	reqModel := modelName
	if reqModel == "" || reqModel == "auto-match" {
		reqModel = llm.DefaultModel
		if reqModel == "" && len(llm.Models) > 0 {
			reqModel = llm.Models[0]
		}
		if reqModel == "" {
			return nil, fmt.Errorf("%w: set LLM_DEFAULT_MODEL or pass model in request", consts.ErrUnsupportedModel)
		}
	}

	return &APIConfig{
		BaseURL:  llm.BaseURL,
		APIKey:   llm.APIKey,
		Provider: llm.Provider,
		ReqModel: reqModel,
	}, nil
}

func (s *RelayService) GetModels() []model.LLModelInfo {
	llm := config.GetLLM()
	if llm == nil {
		return nil
	}
	out := make([]model.LLModelInfo, 0, len(llm.Models))
	for _, id := range llm.Models {
		out = append(out, model.LLModelInfo{
			ID: id, Object: "model", Provider: llm.Provider, Name: id,
		})
	}
	return out
}

// GetProvider 返回当前上游标签（单一配置，与 model 无关）
func (s *RelayService) GetProvider(modelName string) string {
	_ = modelName
	if llm := config.GetLLM(); llm != nil {
		return llm.Provider
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
