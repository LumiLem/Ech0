package agent

import (
	"context"
	"errors"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino-ext/components/model/ollama"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/schema"
	"google.golang.org/genai"

	commonModel "github.com/lin-snow/ech0/internal/model/common"
	model "github.com/lin-snow/ech0/internal/model/setting"
)

const (
	GEN_RECENT = "gen_recent"
)

func Generate(ctx context.Context, setting model.AgentSetting, in []*schema.Message) (string, error) {
	if !setting.Enable {
		return "", errors.New(commonModel.AGENT_NOT_ENABLED)
	}
	if setting.Model == "" {
		return "", errors.New(commonModel.AGENT_MODEL_MISSING)
	}
	if setting.Provider == "" {
		return "", errors.New(commonModel.AGENT_PROVIDER_NOT_FOUND)
	}
	if setting.ApiKey == "" {
		return "", errors.New(commonModel.AGENT_API_KEY_MISSING)
	}

	baseURL := ""
	if setting.BaseURL != "" {
		baseURL = setting.BaseURL
	}

	apiKey := setting.ApiKey
	model := setting.Model
	prompt := setting.Prompt
	if prompt != "" {
		// 在对话开头添加系统提示
		in = append(in, &schema.Message{
			Role:    schema.User,
			Content: prompt,
		})
	}

	var resp *schema.Message
	var genErr error

	// 选择服务提供商
	switch setting.Provider {
	case string(commonModel.OpenAI):
		cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:  apiKey,
			Model:   model,
			BaseURL: baseURL,
		})
		if err != nil {
			return "", err
		}

		resp, genErr = cm.Generate(ctx, in)

	case string(commonModel.Anthropic):
		var baseURLPtr *string = nil
		if len(baseURL) > 0 {
			baseURLPtr = &baseURL
		}

		cm, err := claude.NewChatModel(ctx, &claude.Config{
			APIKey:  apiKey,
			Model:   model,
			BaseURL: baseURLPtr,
		})
		if err != nil {
			return "", err
		}

		resp, genErr = cm.Generate(ctx, in)

	case string(commonModel.Gemini):
		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey: apiKey,
		})
		if err != nil {
			return "", err
		}
		cm, err := gemini.NewChatModel(ctx, &gemini.Config{
			Client: client,
			Model:  model,
		})
		if err != nil {
			return "", err
		}

		resp, genErr = cm.Generate(ctx, in)

	case string(commonModel.Qwen):
		cm, err := qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
			APIKey:  apiKey,
			Model:   setting.Model,
			BaseURL: baseURL,
		})
		if err != nil {
			return "", err
		}

		resp, genErr = cm.Generate(ctx, in)

	case string(commonModel.DeepSeek):
		cm, err := deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
			APIKey:  apiKey,
			Model:   model,
			BaseURL: baseURL,
		})
		if err != nil {
			return "", err
		}

		resp, genErr = cm.Generate(ctx, in)

	case string(commonModel.Ollama):
		cm, err := ollama.NewChatModel(ctx, &ollama.ChatModelConfig{
			Model:   model,
			BaseURL: baseURL,
		})
		if err != nil {
			return "", err
		}

		resp, genErr = cm.Generate(ctx, in)

	case string(commonModel.Custom):
		cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:  apiKey,
			Model:   model,
			BaseURL: baseURL,
		})
		if err != nil {
			return "", err
		}

		resp, genErr = cm.Generate(ctx, in)

	default:
		return "", errors.New(commonModel.AGENT_PROVIDER_NOT_FOUND)
	}

	if genErr != nil {
		return "", genErr
	}

	return resp.Content, nil
}
