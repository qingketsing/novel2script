package generator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/qingketsing/novel2script/backend/internal/domain"
)

var chineseNamePattern = regexp.MustCompile(`[\p{Han}]{2,3}`)

func GenerateMockScreenplay(novel domain.Novel) domain.Screenplay {
	characters := extractMockCharacters(novel)
	scenes := make([]domain.Scene, 0, len(novel.Chapters))
	mainCharacter := characters[0]

	for i, chapter := range novel.Chapters {
		sceneID := fmt.Sprintf("scene_%03d", i+1)
		scenes = append(scenes, domain.Scene{
			ID:               sceneID,
			SourceChapterIDs: []string{chapter.ID},
			Heading: domain.Heading{
				Location: "主要场景",
				Time:     "日",
				Interior: true,
			},
			Summary:    chapter.Summary,
			Characters: []string{mainCharacter.ID},
			Beats: []domain.Beat{
				{
					Type: "action",
					Text: fmt.Sprintf("%s的关键情节展开，%s进入新的行动节点。", chapter.Title, mainCharacter.Name),
				},
				{
					Type:          "dialogue",
					CharacterID:   mainCharacter.ID,
					CharacterName: mainCharacter.Name,
					Text:          "这件事必须继续查下去。",
				},
			},
		})
	}

	return domain.Screenplay{
		SchemaVersion:  "1.0",
		Title:          novel.Title,
		SourceType:     "novel",
		Language:       "zh-CN",
		Provider:       "deepseek-v4",
		Mode:           "mock",
		CreatedAt:      "2026-06-05T00:00:00+08:00",
		Characters:     characters,
		SourceChapters: novel.Chapters,
		Acts: []domain.Act{
			{
				ID:     "act_001",
				Title:  "开端",
				Order:  1,
				Scenes: scenes,
			},
		},
	}
}

func extractMockCharacters(novel domain.Novel) []domain.Character {
	name := firstLikelyName(novel)
	if name == "" {
		name = "Unknown"
	}

	return []domain.Character{
		{
			ID:          "char_001",
			Name:        name,
			Role:        "protagonist",
			Description: "根据输入文本生成的 mock 主角，用于稳定演示剧本结构。",
		},
	}
}

func firstLikelyName(novel domain.Novel) string {
	for _, chapter := range novel.Chapters {
		for _, candidate := range chineseNamePattern.FindAllString(chapter.Content, -1) {
			if isLikelyChapterWord(candidate) {
				continue
			}
			return candidate
		}
	}
	return ""
}

func isLikelyChapterWord(value string) bool {
	blocked := []string{"第一", "第二", "第三", "章节", "雨夜", "旧书", "街灯", "没有", "署名", "城市", "另一"}
	for _, item := range blocked {
		if strings.Contains(value, item) {
			return true
		}
	}
	return false
}
