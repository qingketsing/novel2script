package ai

import (
	"strings"
	"testing"

	"github.com/qingketsing/novel2script/backend/internal/domain"
)

func TestBuildScreenplayPromptIncludesNovelChapters(t *testing.T) {
	prompt := BuildScreenplayPrompt(samplePromptNovel())

	required := []string{
		"雨夜来信",
		"chapter_001",
		"第一章 雨夜来信",
		"林舟在雨夜收到一封没有署名的信。",
		"chapter_002",
		"第二章 旧书店",
		"林舟来到旧书店，寻找姐姐留下的线索。",
		"chapter_003",
		"第三章 街灯",
		"街灯忽明忽暗，线索指向城市另一端。",
	}
	for _, want := range required {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q\n%s", want, prompt)
		}
	}
}

func TestBuildScreenplayPromptDefinesYAMLContract(t *testing.T) {
	prompt := BuildScreenplayPrompt(samplePromptNovel())

	required := []string{
		"只输出 YAML",
		"不要输出 Markdown 代码块",
		"schema_version",
		"metadata",
		"generated_by",
		"characters",
		"source_chapters",
		"screenplay",
		"acts",
		"scenes",
		"beats",
		"source_chapter_ids",
		"action",
		"dialogue",
	}
	for _, want := range required {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q\n%s", want, prompt)
		}
	}
}

func TestBuildScreenplayPromptSetsGenerationRules(t *testing.T) {
	prompt := BuildScreenplayPrompt(samplePromptNovel())

	required := []string{
		"保留 source chapter references",
		"动作和对白分离",
		"不要编造重大剧情",
		"不要改变主要事件因果",
		"metadata.title 必须使用输入小说标题",
		"scene.characters 必须引用 characters 中已经定义的 id",
		"dialogue.character_id 必须引用 characters 中已经定义的 id",
		"beat.type 只能使用 action、dialogue 或 note",
		"每个 scene 优先包含至少一个 dialogue beat",
		"metadata.generated_by.provider",
		"deepseek-v4",
		"metadata.generated_by.mode",
	}
	for _, want := range required {
		if !strings.Contains(prompt, want) {
			t.Fatalf("expected prompt to contain %q\n%s", want, prompt)
		}
	}
}

func TestBuildScreenplayPromptIsStable(t *testing.T) {
	novel := samplePromptNovel()

	first := BuildScreenplayPrompt(novel)
	second := BuildScreenplayPrompt(novel)

	if first != second {
		t.Fatalf("expected stable prompt\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

func samplePromptNovel() domain.Novel {
	return domain.Novel{
		Title: "雨夜来信",
		Chapters: []domain.Chapter{
			{
				ID:      "chapter_001",
				Title:   "第一章 雨夜来信",
				Order:   1,
				Content: "林舟在雨夜收到一封没有署名的信。",
			},
			{
				ID:      "chapter_002",
				Title:   "第二章 旧书店",
				Order:   2,
				Content: "林舟来到旧书店，寻找姐姐留下的线索。",
			},
			{
				ID:      "chapter_003",
				Title:   "第三章 街灯",
				Order:   3,
				Content: "街灯忽明忽暗，线索指向城市另一端。",
			},
		},
	}
}
