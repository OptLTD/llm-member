package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"llm-member/internal/config"
	"llm-member/internal/model"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

// APIConfig 通用 API 配置
type APIConfig struct {
	BaseURL string
	APIKey  string

	Provider string
	ReqModel string

	Compatible bool
}

type RelayService struct {
}

func NewRelayService() *RelayService {
	return &RelayService{}
}

func (s *RelayService) ChatCompletions(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	apiConfig, err := s.GetAPIConfig(req.Model)
	if err != nil {
		return nil, err
	}
	if req.Model == "" || req.Model == "auto-match" {
		req.Model = apiConfig.ReqModel
	}
	if apiConfig.Compatible {
		return s.callWithClient(ctx, req, apiConfig)
	} else {
		return s.callWithHTTP(ctx, req, apiConfig)
	}
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
		if apiConfig.Compatible {
			s.streamWithClient(ctx, req, apiConfig, respChan, errorChan)
		} else {
			s.streamWithHTTP(ctx, req, apiConfig, respChan, errorChan)
		}
	}()

	return respChan, errorChan
}

func (s *RelayService) GetAPIConfig(model string) (*APIConfig, error) {
	apiConfig := &APIConfig{Compatible: true}
	apiConfig.Provider = s.GetProvider(model)
	if model == "" || model == "auto-match" {
		if models := s.GetModels(); len(models) > 0 {
			provider := models[0].Provider
			apiConfig.Provider = provider
			apiConfig.ReqModel = models[0].ID
		}
	}
	if apiConfig.Provider == "unknown" {
		return nil, fmt.Errorf("unsupported model: %s", model)
	}

	if provider := config.GetProvider(apiConfig.Provider); provider == nil {
		fmt.Printf("[LLM] Provider %s not found in config\n", provider)
		return nil, fmt.Errorf("provider %s not configured", provider)
	} else {
		apiConfig.APIKey = provider.APIKey
		apiConfig.BaseURL = provider.BaseURL
	}

	// 只有少数提供商需要特殊处理
	switch apiConfig.Provider {
	case "claude", "gemini":
		apiConfig.Compatible = false
	}

	return apiConfig, nil
}

func (s *RelayService) GetModels() []model.LLModelInfo {
	var modelList []model.LLModelInfo

	// OpenAI 模型
	if config.HasProvider("openai") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "gpt-4", Object: "model", Provider: "openai", Name: "GPT-4"},
			{ID: "gpt-4-turbo", Object: "model", Provider: "openai", Name: "GPT-4 Turbo"},
			{ID: "gpt-3.5-turbo", Object: "model", Provider: "openai", Name: "GPT-3.5 Turbo"},
			{ID: "gpt-4o", Object: "model", Provider: "openai", Name: "GPT-4o"},
			{ID: "gpt-4o-mini", Object: "model", Provider: "openai", Name: "GPT-4o Mini"},
		}...)
	}

	// Claude 模型
	if config.HasProvider("claude") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "claude-3-5-sonnet-20241022", Object: "model", Provider: "claude", Name: "Claude 3.5 Sonnet"},
			{ID: "claude-3-5-haiku-20241022", Object: "model", Provider: "claude", Name: "Claude 3.5 Haiku"},
			{ID: "claude-3-opus-20240229", Object: "model", Provider: "claude", Name: "Claude 3 Opus"},
		}...)
	}

	// 通义千问模型
	if config.HasProvider("qwen") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "qwen-turbo", Object: "model", Provider: "qwen", Name: "通义千问 Turbo"},
			{ID: "qwen-plus", Object: "model", Provider: "qwen", Name: "通义千问 Plus"},
			{ID: "qwen-max", Object: "model", Provider: "qwen", Name: "通义千问 Max"},
			{ID: "qwen2.5-72b-instruct", Object: "model", Provider: "qwen", Name: "通义千问 2.5 72B"},
		}...)
	}

	// 豆包模型
	if config.HasProvider("doubao") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "doubao-seed-1-6-250615", Object: "model", Provider: "doubao", Name: "豆包 Seed 1.6"},
		}...)
	}

	// 智谱清言模型
	if config.HasProvider("bigmodel") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "glm-4", Object: "model", Provider: "bigmodel", Name: "GLM-4"},
			{ID: "glm-4-plus", Object: "model", Provider: "bigmodel", Name: "GLM-4 Plus"},
			{ID: "glm-4-air", Object: "model", Provider: "bigmodel", Name: "GLM-4 Air"},
			{ID: "glm-4-flash", Object: "model", Provider: "bigmodel", Name: "GLM-4 Flash"},
		}...)
	}

	// Grok 模型
	if config.HasProvider("grok") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "grok-beta", Object: "model", Provider: "grok", Name: "Grok Beta"},
			{ID: "grok-vision-beta", Object: "model", Provider: "grok", Name: "Grok Vision Beta"},
		}...)
	}

	// Gemini 模型
	if config.HasProvider("gemini") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "gemini-1.5-pro", Object: "model", Provider: "gemini", Name: "Gemini 1.5 Pro"},
			{ID: "gemini-1.5-flash", Object: "model", Provider: "gemini", Name: "Gemini 1.5 Flash"},
			{ID: "gemini-pro", Object: "model", Provider: "gemini", Name: "Gemini Pro"},
		}...)
	}

	// OpenRouter 模型
	if config.HasProvider("openrouter") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "openai/gpt-4o", Object: "model", Provider: "openrouter", Name: "GPT-4o (OpenRouter)"},
			{ID: "anthropic/claude-3.5-sonnet", Object: "model", Provider: "openrouter", Name: "Claude 3.5 Sonnet (OpenRouter)"},
			{ID: "google/gemini-pro-1.5", Object: "model", Provider: "openrouter", Name: "Gemini Pro 1.5 (OpenRouter)"},
		}...)
	}

	// SiliconFlow 模型
	if config.HasProvider("siliconflow") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "qwen/qwen2.5-72b-instruct", Object: "model", Provider: "siliconflow", Name: "Qwen2.5 72B"},
			{ID: "meta-llama/llama-3.1-405b-instruct", Object: "model", Provider: "siliconflow", Name: "Llama 3.1 405B"},
		}...)
	}

	// DeepSeek 模型
	if config.HasProvider("deepseek") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "deepseek-chat", Object: "model", Provider: "deepseek", Name: "DeepSeek Chat"},
			{ID: "deepseek-coder", Object: "model", Provider: "deepseek", Name: "DeepSeek Coder"},
			{ID: "deepseek-reasoner", Object: "model", Provider: "deepseek", Name: "DeepSeek Reasoner"},
		}...)
	}

	// OpenAI-Like 模型
	if config.HasProvider("openai-like") {
		modelList = append(modelList, []model.LLModelInfo{
			{ID: "custom-model", Object: "model", Provider: "openai-like", Name: "Custom Model"},
		}...)
	}

	return modelList
}

func (s *RelayService) GetProvider(model string) string {
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

	// DeepSeek 模型
	if strings.HasPrefix(model, "deepseek-") {
		return "deepseek"
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
	if strings.Contains(model, "qwen/") {
		return "siliconflow"
	}

	// OpenAI-Like 模型
	if model == "custom-model" {
		return "openai-like"
	}

	return "unknown"
}

// callWithClient 使用 OpenAI 客户端调用
func (s *RelayService) callWithClient(ctx context.Context, req *model.ChatRequest, apiConfig *APIConfig) (*model.ChatResponse, error) {
	fmt.Printf("[LLM] curr model: %s, BaseURL: %s\n", req.Model, apiConfig.BaseURL)

	config := openai.DefaultConfig(apiConfig.APIKey)
	config.BaseURL = apiConfig.BaseURL
	client := openai.NewClientWithConfig(config)

	// 转换消息格式
	var messages []openai.ChatCompletionMessage
	for _, msg := range req.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role: msg.Role, Content: msg.Content,
		})
	}

	// 构建请求
	chatReq := openai.ChatCompletionRequest{
		Model: req.Model, Messages: messages,
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
		fmt.Printf("[LLM] OpenAI client error for model %s: %v\n", req.Model, err)
		return nil, err
	}

	// 转换响应格式
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

// callWithHTTP 使用 HTTP 调用
func (s *RelayService) callWithHTTP(ctx context.Context, req *model.ChatRequest, apiConfig *APIConfig) (*model.ChatResponse, error) {
	fmt.Printf("[LLM] Using HTTP client for model: %s, BaseURL: %s\n", req.Model, apiConfig.BaseURL)

	// 构建请求体
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiConfig.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiConfig.APIKey)

	// 发送请求
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

	// 解析响应
	var chatResp model.ChatResponse
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

// streamWithClient 使用 OpenAI 客户端进行流式调用
func (s *RelayService) streamWithClient(ctx context.Context, req *model.ChatRequest, apiConfig *APIConfig, responseChan chan<- *model.ChatStreamResponse, errorChan chan<- error) {
	fmt.Printf("[LLM] curr model: %s, BaseURL: %s\n", req.Model, apiConfig.BaseURL)

	config := openai.DefaultConfig(apiConfig.APIKey)
	config.BaseURL = apiConfig.BaseURL
	client := openai.NewClientWithConfig(config)

	// 转换消息格式
	var messages []openai.ChatCompletionMessage
	for _, msg := range req.Messages {
		messages = append(messages, openai.ChatCompletionMessage{
			Role: msg.Role, Content: msg.Content,
		})
	}

	// 构建请求
	chatReq := openai.ChatCompletionRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   true,
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

	// 创建流式请求
	stream, err := client.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		fmt.Printf("[LLM] OpenAI client stream error for model %s: %v\n", req.Model, err)
		errorChan <- err
		return
	}
	defer stream.Close()

	// 读取流式响应
	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("[LLM] Stream receive error: %v\n", err)
			errorChan <- err
			return
		}

		// 转换为我们的格式
		streamResp := &model.ChatStreamResponse{
			ID:      response.ID,
			Object:  response.Object,
			Created: response.Created,
			Model:   response.Model,
		}

		for i, choice := range response.Choices {
			finishReason := (*string)(nil)
			if choice.FinishReason != "" {
				reason := string(choice.FinishReason)
				finishReason = &reason
			}

			streamChoice := model.ChatStreamChoice{
				Index: i, FinishReason: finishReason,
				Delta: model.ChatStreamDelta{
					Role:    choice.Delta.Role,
					Content: choice.Delta.Content,
				},
			}
			streamResp.Choices = append(streamResp.Choices, streamChoice)
		}

		select {
		case responseChan <- streamResp:
		case <-ctx.Done():
			return
		}
	}
}

// streamWithHTTP 使用 HTTP 进行流式调用
func (s *RelayService) streamWithHTTP(ctx context.Context, req *model.ChatRequest, apiConfig *APIConfig, responseChan chan<- *model.ChatStreamResponse, errorChan chan<- error) {
	fmt.Printf("[LLM] Using HTTP client stream for model: %s, BaseURL: %s\n", req.Model, apiConfig.BaseURL)

	// 设置流式请求
	streamReq := *req
	streamReq.Stream = true

	// 构建请求体
	reqBody, err := json.Marshal(streamReq)
	if err != nil {
		errorChan <- err
		return
	}

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiConfig.BaseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		errorChan <- err
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiConfig.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	// 发送请求
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		errorChan <- err
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		errorChan <- fmt.Errorf("API error: %s", string(body))
		return
	}

	// 读取流式响应
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		// 处理 SSE 格式
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// 检查是否是结束标记
			if data == "[DONE]" {
				break
			}

			// 解析 JSON 数据
			var streamResp model.ChatStreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				fmt.Printf("[LLM] Failed to parse stream data: %v\n", err)
				continue
			}

			select {
			case responseChan <- &streamResp:
			case <-ctx.Done():
				return
			}
		}
	}

	if err := scanner.Err(); err != nil {
		errorChan <- err
	}
}
