package ai

import (
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type YAMLValidationError struct {
	Path    string
	Message string
}

// Error 返回包含失败 YAML 路径的可读校验错误。
func (e YAMLValidationError) Error() string {
	if e.Path == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

// AsYAMLValidationError 将错误解包为 YAMLValidationError，方便调用方读取 path 和 message。
func AsYAMLValidationError(err error, target *YAMLValidationError) bool {
	return errors.As(err, target)
}

// ValidateScreenplayYAML 解析 AI 输出，并校验剧本 YAML 的最小结构契约。
func ValidateScreenplayYAML(yamlText string) error {
	if strings.TrimSpace(yamlText) == "" {
		return validationError("$", "YAML 不能为空")
	}

	var doc screenplayYAML
	if err := yaml.Unmarshal([]byte(yamlText), &doc); err != nil {
		return validationError("$", "YAML 解析失败")
	}

	return validateScreenplayDocument(doc)
}

func validateScreenplayDocument(doc screenplayYAML) error {
	if isBlank(doc.SchemaVersion) {
		return validationError("schema_version", "schema_version 不能为空")
	}
	if isBlank(doc.Metadata.Title) {
		return validationError("metadata.title", "metadata.title 不能为空")
	}
	if doc.Metadata.SourceChapterCount == 0 {
		return validationError("metadata.source_chapter_count", "metadata.source_chapter_count 不能为空")
	}
	if isBlank(doc.Metadata.GeneratedBy.Provider) {
		return validationError("metadata.generated_by.provider", "metadata.generated_by.provider 不能为空")
	}
	if isBlank(doc.Metadata.GeneratedBy.Mode) {
		return validationError("metadata.generated_by.mode", "metadata.generated_by.mode 不能为空")
	}
	if len(doc.Characters) == 0 {
		return validationError("characters", "characters 必须至少包含一个角色")
	}
	if len(doc.SourceChapters) == 0 {
		return validationError("source_chapters", "source_chapters 必须至少包含一个章节")
	}
	if len(doc.Screenplay.Acts) == 0 {
		return validationError("screenplay.acts", "screenplay.acts 必须至少包含一幕")
	}

	for actIndex, act := range doc.Screenplay.Acts {
		if len(act.Scenes) == 0 {
			return validationError(fmt.Sprintf("screenplay.acts[%d].scenes", actIndex), "act.scenes 必须至少包含一个场景")
		}
		for sceneIndex, scene := range act.Scenes {
			if err := validateScene(scene, actIndex, sceneIndex); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateScene(scene yamlScene, actIndex, sceneIndex int) error {
	path := fmt.Sprintf("screenplay.acts[%d].scenes[%d]", actIndex, sceneIndex)
	if isBlank(scene.ID) {
		return validationError(path+".id", "scene.id 不能为空")
	}
	if len(scene.SourceChapterIDs) == 0 {
		return validationError(path+".source_chapter_ids", "source_chapter_ids 必须至少包含一个章节 ID")
	}
	if isBlank(scene.Heading.Location) {
		return validationError(path+".heading.location", "heading.location 不能为空")
	}
	if isBlank(scene.Heading.Time) {
		return validationError(path+".heading.time", "heading.time 不能为空")
	}
	if isBlank(scene.Summary) {
		return validationError(path+".summary", "scene.summary 不能为空")
	}
	if scene.Characters == nil {
		return validationError(path+".characters", "scene.characters 必须是数组")
	}
	if len(scene.Beats) == 0 {
		return validationError(path+".beats", "beats 必须至少包含一个 beat")
	}

	for beatIndex, beat := range scene.Beats {
		if err := validateBeat(beat, fmt.Sprintf("%s.beats[%d]", path, beatIndex)); err != nil {
			return err
		}
	}

	return nil
}

func validateBeat(beat yamlBeat, path string) error {
	if isBlank(beat.Type) {
		return validationError(path+".type", "beat.type 不能为空")
	}

	switch beat.Type {
	case "action":
		if isBlank(beat.Text) {
			return validationError(path+".text", "action beat 必须包含 text")
		}
	case "dialogue":
		if isBlank(beat.CharacterID) {
			return validationError(path+".character_id", "dialogue beat 必须包含 character_id")
		}
		if isBlank(beat.CharacterName) {
			return validationError(path+".character_name", "dialogue beat 必须包含 character_name")
		}
		if isBlank(beat.Text) {
			return validationError(path+".text", "dialogue beat 必须包含 text")
		}
	default:
		if isBlank(beat.Text) {
			return validationError(path+".text", "beat.text 不能为空")
		}
	}

	return nil
}

func validationError(path, message string) error {
	return YAMLValidationError{Path: path, Message: message}
}

func isBlank(value string) bool {
	return strings.TrimSpace(value) == ""
}

// screenplayYAML 只映射校验 AI YAML 输出所需的字段，不作为领域模型使用。
type screenplayYAML struct {
	SchemaVersion string `yaml:"schema_version"`
	Metadata      struct {
		Title              string `yaml:"title"`
		SourceChapterCount int    `yaml:"source_chapter_count"`
		GeneratedBy        struct {
			Provider string `yaml:"provider"`
			Mode     string `yaml:"mode"`
		} `yaml:"generated_by"`
	} `yaml:"metadata"`
	Characters     []struct{}     `yaml:"characters"`
	SourceChapters []struct{}     `yaml:"source_chapters"`
	Screenplay     yamlScreenplay `yaml:"screenplay"`
}

type yamlScreenplay struct {
	Acts []yamlAct `yaml:"acts"`
}

type yamlAct struct {
	Scenes []yamlScene `yaml:"scenes"`
}

type yamlScene struct {
	ID               string   `yaml:"id"`
	SourceChapterIDs []string `yaml:"source_chapter_ids"`
	Heading          struct {
		Location string `yaml:"location"`
		Time     string `yaml:"time"`
	} `yaml:"heading"`
	Summary    string     `yaml:"summary"`
	Characters []string   `yaml:"characters"`
	Beats      []yamlBeat `yaml:"beats"`
}

type yamlBeat struct {
	Type          string `yaml:"type"`
	Text          string `yaml:"text"`
	CharacterID   string `yaml:"character_id"`
	CharacterName string `yaml:"character_name"`
}
