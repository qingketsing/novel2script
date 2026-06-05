package generator

import (
	"reflect"
	"testing"

	"github.com/qingketsing/novel2script/backend/internal/parser"
)

func TestMockGeneratorIsStable(t *testing.T) {
	novel, err := parser.ParseNovel("雨夜来信", sampleNovel)
	if err != nil {
		t.Fatalf("ParseNovel returned error: %v", err)
	}

	first := GenerateMockScreenplay(novel)
	second := GenerateMockScreenplay(novel)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("expected stable mock output\nfirst: %+v\nsecond: %+v", first, second)
	}
	if len(first.Characters) == 0 {
		t.Fatal("expected at least one character")
	}
	if first.Characters[0].ID != "char_001" {
		t.Fatalf("unexpected first character id: %q", first.Characters[0].ID)
	}
	if len(first.Acts) != 1 || len(first.Acts[0].Scenes) != 3 {
		t.Fatalf("expected one act with 3 scenes, got %+v", first.Acts)
	}
}

func TestMockScenesReferenceRealChaptersAndCharacters(t *testing.T) {
	novel, err := parser.ParseNovel("雨夜来信", sampleNovel)
	if err != nil {
		t.Fatalf("ParseNovel returned error: %v", err)
	}

	screenplay := GenerateMockScreenplay(novel)

	for i, scene := range screenplay.Acts[0].Scenes {
		if scene.ID == "" {
			t.Fatalf("scene %d missing id", i)
		}
		if len(scene.SourceChapterIDs) != 1 || scene.SourceChapterIDs[0] != novel.Chapters[i].ID {
			t.Fatalf("scene %d references wrong chapter: %+v", i, scene.SourceChapterIDs)
		}
		if len(scene.Characters) == 0 || scene.Characters[0] != screenplay.Characters[0].ID {
			t.Fatalf("scene %d references wrong character: %+v", i, scene.Characters)
		}
		if len(scene.Beats) == 0 || scene.Beats[0].Type != "action" {
			t.Fatalf("scene %d missing action beat: %+v", i, scene.Beats)
		}
	}
}

const sampleNovel = `第一章 雨夜来信
林舟在雨夜收到一封没有署名的信。

第二章 旧书店
林舟来到旧书店，寻找姐姐留下的线索。

第三章 街灯
街灯忽明忽暗，线索指向城市另一端。`
