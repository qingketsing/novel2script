package validation

import (
	"strings"

	"github.com/qingketsing/novel2script/backend/internal/domain"
)

const (
	CodeEmptyText            = "EMPTY_TEXT"
	CodeInsufficientChapters = "INSUFFICIENT_CHAPTERS"
)

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Code + ": " + e.Message
}

func ValidateInput(content string) *Error {
	if strings.TrimSpace(content) == "" {
		return &Error{
			Code:    CodeEmptyText,
			Message: "请输入至少 3 个章节的小说文本。",
		}
	}
	return nil
}

func ValidateNovel(novel domain.Novel) *Error {
	if len(novel.Chapters) < 3 {
		return &Error{
			Code:    CodeInsufficientChapters,
			Message: "至少需要 3 个章节才能生成剧本初稿。",
		}
	}
	return nil
}
