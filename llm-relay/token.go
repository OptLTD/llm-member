package relay

import (
	"sync"

	"github.com/tiktoken-go/tokenizer"
	"github.com/tiktoken-go/tokenizer/codec"
)

// TokenCounter 使用 tiktoken 统计消息 token 数量。
type TokenCounter struct {
	tokenEncoderMap map[string]tokenizer.Codec

	tokenEncoderMutex   sync.RWMutex
	defaultTokenEncoder tokenizer.Codec
}

// NewTokenCounter 创建 TokenCounter
func NewTokenCounter() *TokenCounter {
	return &TokenCounter{
		defaultTokenEncoder: codec.NewCl100kBase(),
		tokenEncoderMap:     make(map[string]tokenizer.Codec),
	}
}

func (ts *TokenCounter) getTokenEncoder(model string) tokenizer.Codec {
	ts.tokenEncoderMutex.RLock()
	if encoder, exists := ts.tokenEncoderMap[model]; exists {
		ts.tokenEncoderMutex.RUnlock()
		return encoder
	}
	ts.tokenEncoderMutex.RUnlock()

	ts.tokenEncoderMutex.Lock()
	defer ts.tokenEncoderMutex.Unlock()

	if encoder, exists := ts.tokenEncoderMap[model]; exists {
		return encoder
	}

	modelCodec, err := tokenizer.ForModel(tokenizer.Model(model))
	if err != nil {
		ts.tokenEncoderMap[model] = ts.defaultTokenEncoder
		return ts.defaultTokenEncoder
	}

	ts.tokenEncoderMap[model] = modelCodec
	return modelCodec
}

func (ts *TokenCounter) getTokenNum(tokenEncoder tokenizer.Codec, text string) int {
	if text == "" || tokenEncoder == nil {
		return 0
	}
	tkm, _ := tokenEncoder.Count(text)
	return tkm
}

// CountMsgsToken 统计消息的 token 数量
func (ts *TokenCounter) CountMsgsToken(msgs []ChatMessage, model string, stream bool) (int, error) {
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

// CountTextToken 统计文本的 token 数量
func (ts *TokenCounter) CountTextToken(text string, model string) int {
	if text == "" {
		return 0
	}
	tokenEncoder := ts.getTokenEncoder(model)
	return ts.getTokenNum(tokenEncoder, text)
}
