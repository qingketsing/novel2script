package ai

import (
	"fmt"
	"strings"

	"github.com/qingketsing/novel2script/backend/internal/domain"
)

func BuildScreenplayPrompt(novel domain.Novel) string {
	var b strings.Builder

	writeLine(&b, "你是小说改编剧本助手。请把输入小说改写为结构化剧本 YAML。")
	writeLine(&b, "")
	writeLine(&b, "输出要求：")
	writeLine(&b, "- 只输出 YAML，不要输出解释、前言或总结。")
	writeLine(&b, "- 不要输出 Markdown 代码块，不要使用 ``` 包裹结果。")
	writeLine(&b, "- YAML 必须符合下方核心字段结构。")
	writeLine(&b, "- 保留 source chapter references：每个 scene 必须包含 source_chapter_ids，并引用输入中的 chapter id。")
	writeLine(&b, "- 动作和对白分离：action beat 只写画面和动作，dialogue beat 必须包含 character_id、character_name、text。")
	writeLine(&b, "- 不要编造重大剧情，不要改变主要事件因果；只能基于原章节内容提炼、压缩和剧本化。")
	writeLine(&b, "- 所有 text 字段不能为空。")
	writeLine(&b, "- metadata.title 必须使用输入小说标题，不要输出“未命名小说”。")
	writeLine(&b, "- scene.characters 必须引用 characters 中已经定义的 id。")
	writeLine(&b, "- dialogue.character_id 必须引用 characters 中已经定义的 id。")
	writeLine(&b, "- beat.type 只能使用 action、dialogue 或 note。")
	writeLine(&b, "- 每个 scene 优先包含至少一个 dialogue beat；如果原文完全不适合对白，才可以只输出 action beat。")
	writeLine(&b, "")
	writeLine(&b, "YAML 核心字段要求：")
	writeLine(&b, "schema_version: \"1.0\"")
	writeLine(&b, "metadata:")
	writeLine(&b, "  title: string")
	writeLine(&b, "  source_type: \"novel\"")
	writeLine(&b, "  source_chapter_count: number")
	writeLine(&b, "  language: \"zh-CN\"")
	writeLine(&b, "  generated_by:")
	writeLine(&b, "    provider: \"deepseek-v4\"")
	writeLine(&b, "    mode: \"api\"")
	writeLine(&b, "characters:")
	writeLine(&b, "  - id: \"char_001\"")
	writeLine(&b, "    name: string")
	writeLine(&b, "    role: protagonist | supporting | antagonist")
	writeLine(&b, "    description: string")
	writeLine(&b, "source_chapters:")
	writeLine(&b, "  - id: \"chapter_001\"")
	writeLine(&b, "    title: string")
	writeLine(&b, "    order: 1")
	writeLine(&b, "    summary: string")
	writeLine(&b, "screenplay:")
	writeLine(&b, "  acts:")
	writeLine(&b, "    - id: \"act_001\"")
	writeLine(&b, "      title: string")
	writeLine(&b, "      order: 1")
	writeLine(&b, "      scenes:")
	writeLine(&b, "        - id: \"scene_001\"")
	writeLine(&b, "          source_chapter_ids:")
	writeLine(&b, "            - \"chapter_001\"")
	writeLine(&b, "          heading:")
	writeLine(&b, "            location: string")
	writeLine(&b, "            time: string")
	writeLine(&b, "            interior: true")
	writeLine(&b, "          summary: string")
	writeLine(&b, "          characters:")
	writeLine(&b, "            - \"char_001\"")
	writeLine(&b, "          beats:")
	writeLine(&b, "            - type: \"action\"")
	writeLine(&b, "              text: string")
	writeLine(&b, "            - type: \"dialogue\"")
	writeLine(&b, "              character_id: \"char_001\"")
	writeLine(&b, "              character_name: string")
	writeLine(&b, "              text: string")
	writeLine(&b, "")
	writeLine(&b, "固定字段：")
	writeLine(&b, "- metadata.generated_by.provider 必须是 deepseek-v4。")
	writeLine(&b, "- metadata.generated_by.mode 必须是 api。")
	writeLine(&b, "")
	writeLine(&b, "输入小说：")
	writeLine(&b, fmt.Sprintf("标题：%s", fallbackTitle(novel.Title)))
	writeLine(&b, fmt.Sprintf("章节数量：%d", len(novel.Chapters)))
	for _, chapter := range novel.Chapters {
		writeLine(&b, "")
		writeLine(&b, fmt.Sprintf("## %s", chapter.ID))
		writeLine(&b, fmt.Sprintf("title: %s", fallbackTitle(chapter.Title)))
		writeLine(&b, fmt.Sprintf("order: %d", chapter.Order))
		writeLine(&b, "content:")
		writeBlock(&b, chapter.Content)
	}

	return b.String()
}

func writeLine(b *strings.Builder, value string) {
	b.WriteString(value)
	b.WriteByte('\n')
}

func writeBlock(b *strings.Builder, value string) {
	content := strings.TrimSpace(value)
	if content == "" {
		writeLine(b, "  本章暂无正文。")
		return
	}
	for _, line := range strings.Split(content, "\n") {
		writeLine(b, "  "+strings.TrimRight(line, " \t"))
	}
}

func fallbackTitle(value string) string {
	title := strings.TrimSpace(value)
	if title == "" {
		return "未命名"
	}
	return title
}
