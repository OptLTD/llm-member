package service

import (
	"fmt"
	"llm-member/internal/model"
	"sync"

	"github.com/tiktoken-go/tokenizer"
	"github.com/tiktoken-go/tokenizer/codec"
)

type TokenService struct {
	tokenEncoderMap map[string]tokenizer.Codec

	tokenEncoderMutex   sync.RWMutex
	defaultTokenEncoder tokenizer.Codec
}

func NewTokenService() *TokenService {
	return &TokenService{
		defaultTokenEncoder: codec.NewCl100kBase(),
		tokenEncoderMap:     make(map[string]tokenizer.Codec),
	}
}

// CheckUsage 检查用户使用限制
func (ts *TokenService) CheckUsage(user *model.UserModel) error {
	if user.ApiUsage == nil || user.ApiLimit == nil {
		return nil
	}

	usage, limit := user.ApiUsage, user.ApiLimit
	switch limit.LimitMethod {
	case "tokens":
		if usage.TodayTokens >= limit.DailyTokens {
			return fmt.Errorf("已达到每日 Token 限制 (%d/%d)", usage.TodayTokens, limit.DailyTokens)
		}
		if usage.TotalTokens >= limit.MonthlyTokens {
			return fmt.Errorf("已达到每月 Token 限制 (%d/%d)", usage.TotalTokens, limit.MonthlyTokens)
		}
	case "requests":
		if usage.TodayRequests >= limit.DailyRequests {
			return fmt.Errorf("已达到每日请求限制 (%d/%d)", usage.TodayRequests, limit.DailyRequests)
		}
		if usage.TotalRequests >= limit.MonthlyRequests {
			return fmt.Errorf("已达到每月请求限制 (%d/%d)", usage.TotalRequests, limit.MonthlyRequests)
		}
	case "projects":
		if usage.TodayProjects >= limit.DailyProjects {
			return fmt.Errorf("已达到每日项目限制 (%d/%d)", usage.TodayProjects, limit.DailyProjects)
		}
		if usage.TotalProjects >= limit.MonthlyProjects {
			return fmt.Errorf("已达到每月项目限制 (%d/%d)", usage.TotalProjects, limit.MonthlyProjects)
		}
	}
	return nil
}

func (ts *TokenService) getTokenEncoder(model string) tokenizer.Codec {
	// First, try to get the encoder from cache with read lock
	ts.tokenEncoderMutex.RLock()
	if encoder, exists := ts.tokenEncoderMap[model]; exists {
		ts.tokenEncoderMutex.RUnlock()
		return encoder
	}
	ts.tokenEncoderMutex.RUnlock()

	// If not in cache, create new encoder with write lock
	ts.tokenEncoderMutex.Lock()
	defer ts.tokenEncoderMutex.Unlock()

	// Double-check if another goroutine already created the encoder
	if encoder, exists := ts.tokenEncoderMap[model]; exists {
		return encoder
	}

	// Create new encoder
	modelCodec, err := tokenizer.ForModel(tokenizer.Model(model))
	if err != nil {
		// Cache the default encoder for this model to avoid repeated failures
		ts.tokenEncoderMap[model] = ts.defaultTokenEncoder
		return ts.defaultTokenEncoder
	}

	// Cache the new encoder
	ts.tokenEncoderMap[model] = modelCodec
	return modelCodec
}

func (ts *TokenService) getTokenNum(tokenEncoder tokenizer.Codec, text string) int {
	if text == "" {
		return 0
	}
	if tokenEncoder == nil {
		return 0
	}
	tkm, _ := tokenEncoder.Count(text)
	return tkm
}

// CountMsgsToken 统计消息的token数量
func (ts *TokenService) CountMsgsToken(msgs []model.ChatMessage, model string, stream bool) (int, error) {
	tokenEncoder := ts.getTokenEncoder(model)
	// Reference:
	// https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
	// https://github.com/pkoukk/tiktoken-go/issues/6
	//
	// Every message follows <|start|>{role/name}\n{content}<|end|>\n
	var tokensPerMessage int
	if model == "gpt-3.5-turbo-0301" {
		tokensPerMessage = 4
	} else {
		tokensPerMessage = 3
	}
	tokenNum := 0
	for _, message := range msgs {
		tokenNum += tokensPerMessage
		tokenNum += ts.getTokenNum(tokenEncoder, message.Role)
		tokenNum += ts.getTokenNum(tokenEncoder, message.Content)
	}
	tokenNum += 3 // Every reply is primed with <|start|>assistant<|message|>
	return tokenNum, nil
}

// CountTextToken 统计文本的token数量
func (ts *TokenService) CountTextToken(text string, model string) int {
	if text == "" {
		return 0
	}
	tokenEncoder := ts.getTokenEncoder(model)
	return ts.getTokenNum(tokenEncoder, text)
}
