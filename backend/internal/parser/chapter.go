package parser

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/qingketsing/novel2script/backend/internal/domain"
)

var chapterTitlePattern = regexp.MustCompile(`^\s*(?:#{1,6}\s*)?(第[零一二三四五六七八九十百千万0-9]+[章节回部卷][^\n]*|Chapter\s+[0-9]+[^\n]*)\s*$`)

func ParseNovel(title, content string) (domain.Novel, error) {
	chapters := parseChapters(content)
	if strings.TrimSpace(title) == "" {
		title = "未命名小说"
	}

	return domain.Novel{
		Title:    strings.TrimSpace(title),
		Content:  content,
		Chapters: chapters,
	}, nil
}

func parseChapters(content string) []domain.Chapter {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	var chapters []domain.Chapter
	var current *domain.Chapter
	var body []string

	flush := func() {
		if current == nil {
			return
		}
		current.Content = strings.TrimSpace(strings.Join(body, "\n"))
		current.Summary = summarize(current.Content)
		chapters = append(chapters, *current)
		body = nil
	}

	for _, line := range lines {
		if title, ok := chapterTitle(line); ok {
			flush()
			order := len(chapters) + 1
			current = &domain.Chapter{
				ID:    fmt.Sprintf("chapter_%03d", order),
				Title: title,
				Order: order,
			}
			continue
		}
		if current != nil {
			body = append(body, line)
		}
	}
	flush()

	if len(chapters) == 0 && strings.TrimSpace(content) != "" {
		chapters = append(chapters, domain.Chapter{
			ID:      "chapter_001",
			Title:   "第一章",
			Order:   1,
			Content: strings.TrimSpace(content),
			Summary: summarize(content),
		})
	}

	return chapters
}

func chapterTitle(line string) (string, bool) {
	match := chapterTitlePattern.FindStringSubmatch(line)
	if match == nil {
		return "", false
	}
	return strings.TrimSpace(match[1]), true
}

func summarize(content string) string {
	text := strings.Join(strings.Fields(content), "")
	if text == "" {
		return "本章暂无摘要。"
	}
	runes := []rune(text)
	if len(runes) > 48 {
		return string(runes[:48]) + "..."
	}
	return text
}
