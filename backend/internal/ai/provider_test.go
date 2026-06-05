package ai

import (
	"context"
	"testing"

	"github.com/qingketsing/novel2script/backend/internal/parser"
)

func TestMockProviderGeneratesStableScreenplay(t *testing.T) {
	novel, err := parser.ParseNovel("雨夜来信", sampleNovel)
	if err != nil {
		t.Fatalf("ParseNovel returned error: %v", err)
	}

	provider := NewMockProvider()
	output, err := provider.GenerateScreenplay(context.Background(), GenerateInput{
		Novel: novel,
	})
	if err != nil {
		t.Fatalf("GenerateScreenplay returned error: %v", err)
	}

	if output.Screenplay.Provider != "deepseek-v4" {
		t.Fatalf("expected deepseek-v4 provider, got %q", output.Screenplay.Provider)
	}
	if output.Screenplay.Mode != "mock" {
		t.Fatalf("expected mock mode, got %q", output.Screenplay.Mode)
	}
	if len(output.Screenplay.Acts) != 1 || len(output.Screenplay.Acts[0].Scenes) != 3 {
		t.Fatalf("expected one act with 3 scenes, got %+v", output.Screenplay.Acts)
	}
}

const sampleNovel = `第一章 雨夜来信
林舟在雨夜收到一封没有署名的信。

第二章 旧书店
林舟来到旧书店，寻找姐姐留下的线索。

第三章 街灯
街灯忽明忽暗，线索指向城市另一端。`
