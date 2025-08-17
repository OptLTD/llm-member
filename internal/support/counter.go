package support

// copy from :
// https://github.com/QuantumNous/new-api/blob/main/service/token_counter.go
import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/tiktoken-go/tokenizer"
	"github.com/tiktoken-go/tokenizer/codec"
)

// 常量定义
const (
	GetMediaToken          = true
	GetMediaTokenNotStream = false
)

// 渠道类型常量
const (
	ChannelTypeGemini    = 1
	ChannelTypeVertexAi  = 2
	ChannelTypeAnthropic = 3
)

// 内容类型常量
const (
	ContentTypeText       = "text"
	ContentTypeImageURL   = "image_url"
	ContentTypeInputAudio = "input_audio"
	ContentTypeFile       = "file"
	ContentTypeVideoUrl   = "video_url"
)

// RelayInfo 中继信息结构
type RelayInfo struct {
	ChannelType       int
	IsFirstRequest    bool
	InputAudioFormat  string
	OutputAudioFormat string
	RealtimeTools     []string
}

// MessageImageUrl 图片消息结构
type MessageImageUrl struct {
	Url      string `json:"url"`
	Detail   string `json:"detail,omitempty"`
	MimeType string `json:"mime_type,omitempty"`
}

// GeneralOpenAIRequest OpenAI请求结构
type GeneralOpenAIRequest struct {
	Messages []Message `json:"messages"`
	Model    string    `json:"model"`
	Stream   bool      `json:"stream,omitempty"`
	Tools    []Tool    `json:"tools,omitempty"`
}

// ClaudeRequest Claude请求结构
type ClaudeRequest struct {
	Messages []ClaudeMessage `json:"messages"`
	Model    string          `json:"model"`
	Stream   bool            `json:"stream,omitempty"`
	System   string          `json:"system,omitempty"`
	Tools    interface{}     `json:"tools,omitempty"`
}

// ClaudeMessage Claude消息结构
type ClaudeMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

// Message 消息结构
type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
	Name    *string     `json:"name,omitempty"`
}

// MessageContent 消息内容结构
type MessageContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// Tool 工具结构
type Tool struct {
	Name        string       `json:"name"`
	Description string       `json:"description"`
	InputSchema interface{}  `json:"input_schema"`
	Function    ToolFunction `json:"function"`
}

// ToolFunction 工具函数结构
type ToolFunction struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Parameters  interface{} `json:"parameters"`
}

// RealtimeEvent 实时事件结构
type RealtimeEvent struct {
	Type    string           `json:"type"`
	Session *RealtimeSession `json:"session,omitempty"`
	Delta   string           `json:"delta,omitempty"`
	Audio   string           `json:"audio,omitempty"`
	Item    *RealtimeItem    `json:"item,omitempty"`
}

// RealtimeSession 实时会话结构
type RealtimeSession struct {
	Instructions string `json:"instructions"`
}

// RealtimeItem 实时项目结构
type RealtimeItem struct {
	Type    string                `json:"type"`
	Content []RealtimeItemContent `json:"content,omitempty"`
}

// RealtimeItemContent 实时项目内容结构
type RealtimeItemContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// ChatCompletionsStreamResponseChoice 流式响应选择结构
type ChatCompletionsStreamResponseChoice struct {
	Delta StreamDelta `json:"delta"`
}

// StreamDelta 流式增量结构
type StreamDelta struct {
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall 工具调用结构
type ToolCall struct {
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction 工具调用函数结构
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// 实时事件类型常量
const (
	RealtimeEventTypeSessionUpdate                  = "session.update"
	RealtimeEventResponseAudioDelta                 = "response.audio.delta"
	RealtimeEventResponseAudioTranscriptionDelta    = "response.audio_transcription.delta"
	RealtimeEventResponseFunctionCallArgumentsDelta = "response.function_call_arguments.delta"
	RealtimeEventInputAudioBufferAppend             = "input_audio_buffer.append"
	RealtimeEventConversationItemCreated            = "conversation.item.created"
	RealtimeEventTypeResponseDone                   = "response.done"
)

// tokenEncoderMap won't grow after initialization
var defaultTokenEncoder tokenizer.Codec

// tokenEncoderMap is used to store token encoders for different models
var tokenEncoderMap = make(map[string]tokenizer.Codec)

// tokenEncoderMutex protects tokenEncoderMap for concurrent access
var tokenEncoderMutex sync.RWMutex

func InitTokenEncoders() {
	log.Println("initializing token encoders")
	defaultTokenEncoder = codec.NewCl100kBase()
	log.Println("token encoders initialized")
}

func getTokenEncoder(model string) tokenizer.Codec {
	// First, try to get the encoder from cache with read lock
	tokenEncoderMutex.RLock()
	if encoder, exists := tokenEncoderMap[model]; exists {
		tokenEncoderMutex.RUnlock()
		return encoder
	}
	tokenEncoderMutex.RUnlock()

	// If not in cache, create new encoder with write lock
	tokenEncoderMutex.Lock()
	defer tokenEncoderMutex.Unlock()

	// Double-check if another goroutine already created the encoder
	if encoder, exists := tokenEncoderMap[model]; exists {
		return encoder
	}

	// Create new encoder
	modelCodec, err := tokenizer.ForModel(tokenizer.Model(model))
	if err != nil {
		// Cache the default encoder for this model to avoid repeated failures
		tokenEncoderMap[model] = defaultTokenEncoder
		return defaultTokenEncoder
	}

	// Cache the new encoder
	tokenEncoderMap[model] = modelCodec
	return modelCodec
}

func getTokenNum(tokenEncoder tokenizer.Codec, text string) int {
	if text == "" {
		return 0
	}
	tkm, _ := tokenEncoder.Count(text)
	return tkm
}

// 简化的图片token计算函数
func getImageToken(info *RelayInfo, imageUrl *MessageImageUrl, model string, stream bool) (int, error) {
	if imageUrl == nil {
		return 0, fmt.Errorf("image_url_is_nil")
	}
	baseTokens := 85
	if model == "glm-4v" {
		return 1047, nil
	}
	if imageUrl.Detail == "low" {
		return baseTokens, nil
	}
	if !GetMediaTokenNotStream && !stream {
		return 3 * baseTokens, nil
	}

	// 简化的图片计费逻辑
	if imageUrl.Detail == "auto" || imageUrl.Detail == "" {
		imageUrl.Detail = "high"
	}

	// tileTokens := 170
	if strings.HasPrefix(model, "gpt-4o-mini") {
		// tileTokens = 5667
		baseTokens = 2833
	}
	// 是否统计图片token
	if !GetMediaToken {
		return 3 * baseTokens, nil
	}
	if info.ChannelType == ChannelTypeGemini || info.ChannelType == ChannelTypeVertexAi || info.ChannelType == ChannelTypeAnthropic {
		return 3 * baseTokens, nil
	}

	// 简化处理，返回固定值
	return 3 * baseTokens, nil
}

func CountTokenChatRequest(info *RelayInfo, request GeneralOpenAIRequest) (int, error) {
	tkm := 0
	msgTokens, err := CountTokenMessages(info, request.Messages, request.Model, request.Stream)
	if err != nil {
		return 0, err
	}
	tkm += msgTokens
	if request.Tools != nil {
		openaiTools := request.Tools
		countStr := ""
		for _, tool := range openaiTools {
			countStr = tool.Function.Name
			if tool.Function.Description != "" {
				countStr += tool.Function.Description
			}
			if tool.Function.Parameters != nil {
				countStr += fmt.Sprintf("%v", tool.Function.Parameters)
			}
		}
		toolTokens := CountTokenInput(countStr, request.Model)
		tkm += 8
		tkm += toolTokens
	}

	return tkm, nil
}

func CountTokenClaudeRequest(request ClaudeRequest, model string) (int, error) {
	tkm := 0

	// Count tokens in messages
	msgTokens, err := CountTokenClaudeMessages(request.Messages, model, request.Stream)
	if err != nil {
		return 0, err
	}
	tkm += msgTokens

	// Count tokens in system message
	if request.System != "" {
		systemTokens := CountTokenInput(request.System, model)
		tkm += systemTokens
	}

	if request.Tools != nil {
		// check is array
		if tools, ok := request.Tools.([]any); ok {
			if len(tools) > 0 {
				// 简化处理，直接估算
				tkm += len(tools) * 100 // 每个工具估算100个token
			}
		} else {
			return 0, errors.New("tools: Input should be a valid list")
		}
	}

	return tkm, nil
}

func CountTokenClaudeMessages(messages []ClaudeMessage, model string, stream bool) (int, error) {
	tokenEncoder := getTokenEncoder(model)
	tokenNum := 0

	for _, message := range messages {
		// Count tokens for role
		tokenNum += getTokenNum(tokenEncoder, message.Role)
		if content, ok := message.Content.(string); ok {
			tokenNum += getTokenNum(tokenEncoder, content)
		} else {
			// 复杂内容简化处理
			contentJSON, _ := json.Marshal(message.Content)
			tokenNum += getTokenNum(tokenEncoder, string(contentJSON))
		}
	}

	// Add a constant for message formatting
	tokenNum += len(messages) * 2

	return tokenNum, nil
}

func CountTokenClaudeTools(tools []Tool, model string) (int, error) {
	tokenEncoder := getTokenEncoder(model)
	tokenNum := 0

	for _, tool := range tools {
		tokenNum += getTokenNum(tokenEncoder, tool.Name)
		tokenNum += getTokenNum(tokenEncoder, tool.Description)

		schemaJSON, err := json.Marshal(tool.InputSchema)
		if err != nil {
			return 0, errors.New(fmt.Sprintf("marshal_tool_schema_fail: %s", err.Error()))
		}
		tokenNum += getTokenNum(tokenEncoder, string(schemaJSON))
	}

	// Add a constant for tool formatting
	tokenNum += len(tools) * 3

	return tokenNum, nil
}

func CountTokenRealtime(info *RelayInfo, request RealtimeEvent, model string) (int, int, error) {
	audioToken := 0
	textToken := 0
	switch request.Type {
	case RealtimeEventTypeSessionUpdate:
		if request.Session != nil {
			msgTokens := CountTextToken(request.Session.Instructions, model)
			textToken += msgTokens
		}
	case RealtimeEventResponseAudioDelta:
		// 简化音频token计算
		audioToken += len(request.Delta) / 100 // 简单估算
	case RealtimeEventResponseAudioTranscriptionDelta, RealtimeEventResponseFunctionCallArgumentsDelta:
		// count text token
		tkm := CountTextToken(request.Delta, model)
		textToken += tkm
	case RealtimeEventInputAudioBufferAppend:
		// 简化音频token计算
		audioToken += len(request.Audio) / 100 // 简单估算
	case RealtimeEventConversationItemCreated:
		if request.Item != nil {
			switch request.Item.Type {
			case "message":
				for _, content := range request.Item.Content {
					if content.Type == "input_text" {
						tokens := CountTextToken(content.Text, model)
						textToken += tokens
					}
				}
			}
		}
	case RealtimeEventTypeResponseDone:
		// count tools token
		if !info.IsFirstRequest {
			if info.RealtimeTools != nil && len(info.RealtimeTools) > 0 {
				for _, tool := range info.RealtimeTools {
					toolTokens := CountTokenInput(tool, model)
					textToken += 8
					textToken += toolTokens
				}
			}
		}
	}
	return textToken, audioToken, nil
}

func CountTokenMessages(info *RelayInfo, messages []Message, model string, stream bool) (int, error) {
	tokenEncoder := getTokenEncoder(model)
	// Reference:
	// https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb
	// https://github.com/pkoukk/tiktoken-go/issues/6
	//
	// Every message follows <|start|>{role/name}\n{content}<|end|>\n
	var tokensPerMessage int
	var tokensPerName int
	if model == "gpt-3.5-turbo-0301" {
		tokensPerMessage = 4
		tokensPerName = -1 // If there's a name, the role is omitted
	} else {
		tokensPerMessage = 3
		tokensPerName = 1
	}
	tokenNum := 0
	for _, message := range messages {
		tokenNum += tokensPerMessage
		tokenNum += getTokenNum(tokenEncoder, message.Role)
		if message.Content != nil {
			if message.Name != nil {
				tokenNum += tokensPerName
				tokenNum += getTokenNum(tokenEncoder, *message.Name)
			}
			// 简化内容处理
			if content, ok := message.Content.(string); ok {
				tokenNum += getTokenNum(tokenEncoder, content)
			} else {
				// 复杂内容简化处理
				contentJSON, _ := json.Marshal(message.Content)
				tokenNum += getTokenNum(tokenEncoder, string(contentJSON))
			}
		}
	}
	tokenNum += 3 // Every reply is primed with <|start|>assistant<|message|>
	return tokenNum, nil
}

func CountTokenInput(input any, model string) int {
	switch v := input.(type) {
	case string:
		return CountTextToken(v, model)
	case []string:
		text := ""
		for _, s := range v {
			text += s
		}
		return CountTextToken(text, model)
	case []interface{}:
		text := ""
		for _, item := range v {
			text += fmt.Sprintf("%v", item)
		}
		return CountTextToken(text, model)
	}
	return CountTokenInput(fmt.Sprintf("%v", input), model)
}

func CountTokenStreamChoices(messages []ChatCompletionsStreamResponseChoice, model string) int {
	tokens := 0
	for _, message := range messages {
		tkm := CountTokenInput(message.Delta.Content, model)
		tokens += tkm
		if message.Delta.ToolCalls != nil {
			for _, tool := range message.Delta.ToolCalls {
				tkm := CountTokenInput(tool.Function.Name, model)
				tokens += tkm
				tkm = CountTokenInput(tool.Function.Arguments, model)
				tokens += tkm
			}
		}
	}
	return tokens
}

func CountTTSToken(text string, model string) int {
	if strings.HasPrefix(model, "tts") {
		return utf8.RuneCountInString(text)
	} else {
		return CountTextToken(text, model)
	}
}

// 简化的音频token计算函数
func CountAudioTokenInput(audioBase64 string, audioFormat string) (int, error) {
	if audioBase64 == "" {
		return 0, nil
	}
	// 简化计算：基于base64长度估算
	duration := float64(len(audioBase64)) / 1000.0 // 简单估算
	return int(duration / 60 * 100 / 0.06), nil
}

func CountAudioTokenOutput(audioBase64 string, audioFormat string) (int, error) {
	if audioBase64 == "" {
		return 0, nil
	}
	// 简化计算：基于base64长度估算
	duration := float64(len(audioBase64)) / 1000.0 // 简单估算
	return int(duration / 60 * 200 / 0.24), nil
}

// CountTextToken 统计文本的token数量
func CountTextToken(text string, model string) int {
	if text == "" {
		return 0
	}
	tokenEncoder := getTokenEncoder(model)
	return getTokenNum(tokenEncoder, text)
}
