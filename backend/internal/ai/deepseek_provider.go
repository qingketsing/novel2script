package ai

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"
	"unicode/utf8"

	"github.com/qingketsing/novel2script/backend/internal/domain"
	"github.com/qingketsing/novel2script/backend/internal/observability"
)

const (
	deepSeekRequestTimeoutMin = 45 * time.Second
	deepSeekRequestTimeoutMax = 300 * time.Second

	deepSeekTimeoutBaseOverhead = 15 * time.Second
	deepSeekTimeoutSafetyBuffer = 20 * time.Second

	deepSeekInputTokensPerSecond  = 1500.0
	deepSeekOutputTokensPerSecond = 40.0

	deepSeekOutputBaseTokens       = 350.0
	deepSeekOutputTokensPerChapter = 450.0
	deepSeekOutputInputTokenRatio  = 0.08
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
	Timeout time.Duration
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
		Timeout: cfg.Timeout,
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
	logger := observability.Logger(ctx)
	requestID := observability.RequestID(ctx)
	timeout := p.timeoutForInput(input)
	requestCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	prompt := BuildScreenplayPrompt(input.Novel)
	logger.InfoContext(ctx, "deepseek generation started",
		"request_id", requestID,
		"chapter_count", len(input.Novel.Chapters),
		"timeout_ms", timeout.Milliseconds(),
		"input_tokens_estimate", estimateNovelTokens(input.Novel),
		"prompt_length", len(prompt),
	)

	generateStart := time.Now()
	rawYAML, err := p.yamlGenerator.GenerateYAML(requestCtx, prompt)
	if err != nil {
		logger.WarnContext(ctx, "deepseek generation failed",
			"request_id", requestID,
			"duration_ms", time.Since(generateStart).Milliseconds(),
			"error", err.Error(),
		)
		return GenerateOutput{}, err
	}
	logger.InfoContext(ctx, "deepseek generation returned",
		"request_id", requestID,
		"duration_ms", time.Since(generateStart).Milliseconds(),
		"yaml_length", len(rawYAML),
	)

	if err := ValidateScreenplayYAML(rawYAML); err != nil {
		logger.WarnContext(ctx, "deepseek yaml validation failed",
			"request_id", requestID,
			"error", err.Error(),
		)
		repairStart := time.Now()
		repairedYAML, repairErr := p.repairYAML(requestCtx, rawYAML, err)
		if repairErr != nil {
			logger.WarnContext(ctx, "deepseek yaml repair failed",
				"request_id", requestID,
				"duration_ms", time.Since(repairStart).Milliseconds(),
				"error", repairErr.Error(),
			)
			return GenerateOutput{}, repairErr
		}
		logger.InfoContext(ctx, "deepseek yaml repair succeeded",
			"request_id", requestID,
			"duration_ms", time.Since(repairStart).Milliseconds(),
			"yaml_length", len(repairedYAML),
		)
		return GenerateOutput{RawYAML: repairedYAML}, nil
	}
	logger.InfoContext(ctx, "deepseek yaml validation succeeded",
		"request_id", requestID,
		"yaml_length", len(rawYAML),
	)

	return GenerateOutput{RawYAML: rawYAML}, nil
}

func (p DeepSeekProvider) timeoutForInput(input GenerateInput) time.Duration {
	inputTokens := estimateNovelTokens(input.Novel)
	outputTokens := deepSeekOutputBaseTokens +
		float64(len(input.Novel.Chapters))*deepSeekOutputTokensPerChapter +
		float64(inputTokens)*deepSeekOutputInputTokenRatio

	inputBudget := secondsForTokens(inputTokens, deepSeekInputTokensPerSecond)
	outputBudget := secondsForTokens(int(math.Ceil(outputTokens)), deepSeekOutputTokensPerSecond)
	timeout := deepSeekTimeoutBaseOverhead + inputBudget + outputBudget + deepSeekTimeoutSafetyBuffer

	if p.cfg.Timeout > timeout {
		timeout = p.cfg.Timeout
	}
	if timeout < deepSeekRequestTimeoutMin {
		return deepSeekRequestTimeoutMin
	}
	if timeout > deepSeekRequestTimeoutMax {
		return deepSeekRequestTimeoutMax
	}
	return timeout
}

func estimateNovelTokens(novel domain.Novel) int {
	runes := utf8.RuneCountInString(novel.Title) + utf8.RuneCountInString(novel.Content)
	for _, chapter := range novel.Chapters {
		runes += utf8.RuneCountInString(chapter.Title)
		runes += utf8.RuneCountInString(chapter.Summary)
		runes += utf8.RuneCountInString(chapter.Content)
	}
	return runes
}

func secondsForTokens(tokens int, tokensPerSecond float64) time.Duration {
	if tokens <= 0 {
		return 0
	}
	seconds := math.Ceil(float64(tokens) / tokensPerSecond)
	return time.Duration(seconds) * time.Second
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
