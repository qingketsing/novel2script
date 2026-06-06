package ai

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrDeepSeekAPIKeyRequired  = errors.New("deepseek api key is required")
	ErrDeepSeekBaseURLRequired = errors.New("deepseek base url is required")
	ErrDeepSeekModelRequired   = errors.New("deepseek model is required")
)

type DeepSeekConfig struct {
	APIKey  string
	BaseURL string
	Model   string
}

type DeepSeekProvider struct {
	cfg           DeepSeekConfig
	yamlGenerator deepSeekYAMLGenerator
}

func NewDeepSeekProvider(cfg DeepSeekConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, ErrDeepSeekAPIKeyRequired
	}
	if cfg.BaseURL == "" {
		return nil, ErrDeepSeekBaseURLRequired
	}
	if cfg.Model == "" {
		return nil, ErrDeepSeekModelRequired
	}

	client, err := NewDeepSeekClient(DeepSeekClientConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.BaseURL,
		Model:   cfg.Model,
	}, nil)
	if err != nil {
		return nil, err
	}

	return DeepSeekProvider{
		cfg:           cfg,
		yamlGenerator: client,
	}, nil
}

func (p DeepSeekProvider) GenerateScreenplay(ctx context.Context, input GenerateInput) (GenerateOutput, error) {
	prompt := BuildScreenplayPrompt(input.Novel)
	rawYAML, err := p.yamlGenerator.GenerateYAML(ctx, prompt)
	if err != nil {
		return GenerateOutput{}, err
	}
	if err := ValidateScreenplayYAML(rawYAML); err != nil {
		repairedYAML, repairErr := p.repairYAML(ctx, rawYAML, err)
		if repairErr != nil {
			return GenerateOutput{}, repairErr
		}
		return GenerateOutput{RawYAML: repairedYAML}, nil
	}

	return GenerateOutput{RawYAML: rawYAML}, nil
}

func (p DeepSeekProvider) repairYAML(ctx context.Context, rawYAML string, validationErr error) (string, error) {
	repairPrompt := buildYAMLRepairPrompt(rawYAML, validationErr)
	repairedYAML, err := p.yamlGenerator.GenerateYAML(ctx, repairPrompt)
	if err != nil {
		return "", err
	}
	if err := ValidateScreenplayYAML(repairedYAML); err != nil {
		return "", err
	}
	return repairedYAML, nil
}

func buildYAMLRepairPrompt(rawYAML string, validationErr error) string {
	var detail string
	var yamlErr YAMLValidationError
	if AsYAMLValidationError(validationErr, &yamlErr) {
		detail = fmt.Sprintf("校验失败路径：%s\n校验失败原因：%s", yamlErr.Path, yamlErr.Message)
	} else {
		detail = fmt.Sprintf("校验失败原因：%s", validationErr.Error())
	}

	return fmt.Sprintf(`你是剧本 YAML 修复助手。请修复下面这份 YAML，使其满足剧本 YAML 结构要求。

输出要求：
- 只输出 YAML，不要输出解释、前言或总结。
- 不要输出 Markdown 代码块，不要使用 `+"```"+` 包裹结果。
- 保留原始剧情、角色、章节引用和已有字段含义。
- 只修复结构、缺失字段、引用关系和格式问题。
- 不要编造重大剧情。

%s

原始 YAML：
%s
`, detail, rawYAML)
}

type deepSeekYAMLGenerator interface {
	GenerateYAML(ctx context.Context, prompt string) (string, error)
}
