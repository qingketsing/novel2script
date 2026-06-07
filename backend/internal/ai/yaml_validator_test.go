package ai

import (
	"strings"
	"testing"
)

func TestValidateScreenplayYAMLRejectsEmptyYAML(t *testing.T) {
	err := ValidateScreenplayYAML(" \n\t ")

	assertValidationError(t, err, "$", "YAML 不能为空")
}

func TestValidateScreenplayYAMLRejectsInvalidYAML(t *testing.T) {
	err := ValidateScreenplayYAML("metadata:\n  title: [")

	assertValidationError(t, err, "$", "YAML 解析失败")
}

func TestValidateScreenplayYAMLRejectsMissingSchemaVersion(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `schema_version: "1.0"`+"\n", "", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "schema_version", "schema_version 不能为空")
}

func TestValidateScreenplayYAMLRejectsMissingMetadataTitle(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `  title: "雨夜来信"`+"\n", "", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "metadata.title", "metadata.title 不能为空")
}

func TestValidateScreenplayYAMLRejectsEmptyCharacters(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, charactersBlock, "characters: []\n", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "characters", "characters 必须至少包含一个角色")
}

func TestValidateScreenplayYAMLRejectsEmptySourceChapters(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, sourceChaptersBlock, "source_chapters: []\n", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "source_chapters", "source_chapters 必须至少包含一个章节")
}

func TestValidateScreenplayYAMLRejectsSourceChapterCountMismatch(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, "  source_chapter_count: 3", "  source_chapter_count: 2", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "metadata.source_chapter_count", "metadata.source_chapter_count 必须与 source_chapters 数量一致")
}

func TestValidateScreenplayYAMLRejectsInputChapterCountMismatch(t *testing.T) {
	err := ValidateScreenplayYAMLForChapterCount(validScreenplayYAML, 4)

	assertValidationError(t, err, "metadata.source_chapter_count", "metadata.source_chapter_count 必须与输入章节数量一致")
}

func TestValidateScreenplayYAMLRejectsUnreferencedSourceChapter(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, "            - \"chapter_003\"\n", "", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "source_chapters[2].id", "每个 source chapter 必须至少被一个 scene 引用")
}

func TestValidateScreenplayYAMLRejectsEmptyActs(t *testing.T) {
	yamlText := strings.Split(validScreenplayYAML, "screenplay:\n")[0] + "screenplay:\n  acts: []\n"

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "screenplay.acts", "screenplay.acts 必须至少包含一幕")
}

func TestValidateScreenplayYAMLRejectsSceneMissingID(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `        - id: "scene_001"`+"\n", `        - ignored: "field"`+"\n", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "screenplay.acts[0].scenes[0].id", "scene.id 不能为空")
}

func TestValidateScreenplayYAMLRejectsSceneMissingSourceChapterIDs(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `          source_chapter_ids:
            - "chapter_001"
            - "chapter_002"
            - "chapter_003"
`, "", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "screenplay.acts[0].scenes[0].source_chapter_ids", "source_chapter_ids 必须至少包含一个章节 ID")
}

func TestValidateScreenplayYAMLRejectsSceneMissingHeadingLocation(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `            location: "旧书店"`+"\n", "", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "screenplay.acts[0].scenes[0].heading.location", "heading.location 不能为空")
}

func TestValidateScreenplayYAMLRejectsEmptyBeats(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, beatsBlock, "          beats: []\n", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "screenplay.acts[0].scenes[0].beats", "beats 必须至少包含一个 beat")
}

func TestValidateScreenplayYAMLRejectsDialogueMissingCharacterID(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `              character_id: "char_001"`+"\n", "", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "screenplay.acts[0].scenes[0].beats[1].character_id", "dialogue beat 必须包含 character_id")
}

func TestValidateScreenplayYAMLRejectsDialogueMissingCharacterName(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `              character_name: "林舟"`+"\n", "", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "screenplay.acts[0].scenes[0].beats[1].character_name", "dialogue beat 必须包含 character_name")
}

func TestValidateScreenplayYAMLRejectsDialogueMissingText(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `              text: "这封信是谁放在这里的？"`+"\n", "", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "screenplay.acts[0].scenes[0].beats[1].text", "dialogue beat 必须包含 text")
}

func TestValidateScreenplayYAMLRejectsWrongGeneratedProvider(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `    provider: "deepseek-v4"`, `    provider: "other"`, 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "metadata.generated_by.provider", "metadata.generated_by.provider 必须是 deepseek-v4")
}

func TestValidateScreenplayYAMLRejectsWrongGeneratedMode(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `    mode: "api"`, `    mode: "mock"`, 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "metadata.generated_by.mode", "metadata.generated_by.mode 必须是 api")
}

func TestValidateScreenplayYAMLRejectsDuplicateCharacterID(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, charactersBlock, `characters:
  - id: "char_001"
    name: "林舟"
    role: "protagonist"
    description: "年轻作者"
  - id: "char_001"
    name: "林舟复制"
    role: "supporting"
    description: "重复角色"
`, 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "characters[1].id", "character.id 不能重复")
}

func TestValidateScreenplayYAMLRejectsMissingCharacterName(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `    name: "林舟"`+"\n", "", 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "characters[0].name", "character.name 不能为空")
}

func TestValidateScreenplayYAMLRejectsDuplicateSourceChapterID(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, sourceChaptersBlock, `source_chapters:
  - id: "chapter_001"
    title: "第一章 雨夜来信"
    order: 1
    summary: "林舟收到信。"
  - id: "chapter_001"
    title: "第一章 重复"
    order: 2
    summary: "重复章节。"
`, 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "source_chapters[1].id", "source_chapters.id 不能重复")
}

func TestValidateScreenplayYAMLRejectsUnknownSceneCharacter(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `            - "char_001"`, `            - "char_missing"`, 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "screenplay.acts[0].scenes[0].characters[0]", "scene.characters 必须引用已定义角色")
}

func TestValidateScreenplayYAMLRejectsUnknownDialogueCharacter(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `              character_id: "char_001"`, `              character_id: "char_missing"`, 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "screenplay.acts[0].scenes[0].beats[1].character_id", "dialogue.character_id 必须引用已定义角色")
}

func TestValidateScreenplayYAMLRejectsDialogueCharacterNameMismatch(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `              character_name: "林舟"`, `              character_name: "许岚"`, 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "screenplay.acts[0].scenes[0].beats[1].character_name", "dialogue.character_name 必须与 characters 中对应角色名称一致")
}

func TestValidateScreenplayYAMLRejectsUnknownSourceChapter(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `            - "chapter_001"`, `            - "chapter_missing"`, 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "screenplay.acts[0].scenes[0].source_chapter_ids[0]", "source_chapter_ids 必须引用已定义章节")
}

func TestValidateScreenplayYAMLRejectsUnsupportedBeatType(t *testing.T) {
	yamlText := strings.Replace(validScreenplayYAML, `            - type: "action"`, `            - type: "camera"`, 1)

	err := ValidateScreenplayYAML(yamlText)

	assertValidationError(t, err, "screenplay.acts[0].scenes[0].beats[0].type", "beat.type 只能是 action、dialogue 或 note")
}

func TestValidateScreenplayYAMLAcceptsValidYAML(t *testing.T) {
	if err := ValidateScreenplayYAML(validScreenplayYAML); err != nil {
		t.Fatalf("ValidateScreenplayYAML returned error: %v", err)
	}
}

func assertValidationError(t *testing.T, err error, path, message string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected validation error")
	}
	var validationErr YAMLValidationError
	if !AsYAMLValidationError(err, &validationErr) {
		t.Fatalf("expected YAMLValidationError, got %T: %v", err, err)
	}
	if validationErr.Path != path {
		t.Fatalf("expected path %q, got %q", path, validationErr.Path)
	}
	if validationErr.Message != message {
		t.Fatalf("expected message %q, got %q", message, validationErr.Message)
	}
}

const charactersBlock = `characters:
  - id: "char_001"
    name: "林舟"
    role: "protagonist"
    description: "年轻作者"
`

const sourceChaptersBlock = `source_chapters:
  - id: "chapter_001"
    title: "第一章 雨夜来信"
    order: 1
    summary: "林舟收到信。"
  - id: "chapter_002"
    title: "第二章 旧书店"
    order: 2
    summary: "林舟寻找线索。"
  - id: "chapter_003"
    title: "第三章 街灯"
    order: 3
    summary: "林舟确认下一步行动。"
`

const beatsBlock = `          beats:
            - type: "action"
              text: "雨水敲打玻璃，林舟推门进入旧书店。"
            - type: "dialogue"
              character_id: "char_001"
              character_name: "林舟"
              text: "这封信是谁放在这里的？"
`

const validScreenplayYAML = `schema_version: "1.0"
metadata:
  title: "雨夜来信"
  source_type: "novel"
  source_chapter_count: 3
  language: "zh-CN"
  generated_by:
    provider: "deepseek-v4"
    mode: "api"
characters:
  - id: "char_001"
    name: "林舟"
    role: "protagonist"
    description: "年轻作者"
source_chapters:
  - id: "chapter_001"
    title: "第一章 雨夜来信"
    order: 1
    summary: "林舟收到信。"
  - id: "chapter_002"
    title: "第二章 旧书店"
    order: 2
    summary: "林舟寻找线索。"
  - id: "chapter_003"
    title: "第三章 街灯"
    order: 3
    summary: "林舟确认下一步行动。"
screenplay:
  acts:
    - id: "act_001"
      title: "开端"
      order: 1
      scenes:
        - id: "scene_001"
          source_chapter_ids:
            - "chapter_001"
            - "chapter_002"
            - "chapter_003"
          heading:
            location: "旧书店"
            time: "夜"
            interior: true
          summary: "林舟在旧书店读到线索。"
          characters:
            - "char_001"
          beats:
            - type: "action"
              text: "雨水敲打玻璃，林舟推门进入旧书店。"
            - type: "dialogue"
              character_id: "char_001"
              character_name: "林舟"
              text: "这封信是谁放在这里的？"
`
