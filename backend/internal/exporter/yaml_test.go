package exporter

import (
	"strings"
	"testing"

	"github.com/qingketsing/novel2script/backend/internal/generator"
	"github.com/qingketsing/novel2script/backend/internal/parser"
)

func TestExportYAMLContainsRequiredFields(t *testing.T) {
	novel, err := parser.ParseNovel("雨夜来信", sampleNovel)
	if err != nil {
		t.Fatalf("ParseNovel returned error: %v", err)
	}

	out := ExportYAML(generator.GenerateMockScreenplay(novel))

	required := []string{
		`schema_version: "1.0"`,
		`metadata:`,
		`provider: "deepseek-v4"`,
		`mode: "mock"`,
		`characters:`,
		`source_chapters:`,
		`screenplay:`,
		`acts:`,
		`scenes:`,
		`beats:`,
		`type: "action"`,
	}
	for _, want := range required {
		if !strings.Contains(out, want) {
			t.Fatalf("expected YAML to contain %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "```") {
		t.Fatalf("YAML must not contain Markdown code fences:\n%s", out)
	}
}

const sampleNovel = `第一章 雨夜来信
林舟在雨夜收到一封没有署名的信。

第二章 旧书店
林舟来到旧书店，寻找姐姐留下的线索。

第三章 街灯
街灯忽明忽暗，线索指向城市另一端。`
