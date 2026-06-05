package exporter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/qingketsing/novel2script/backend/internal/domain"
)

func ExportYAML(screenplay domain.Screenplay) string {
	var b strings.Builder

	line(&b, 0, "schema_version: %s", quote(screenplay.SchemaVersion))
	line(&b, 0, "metadata:")
	line(&b, 1, "title: %s", quote(screenplay.Title))
	line(&b, 1, "source_type: %s", quote(screenplay.SourceType))
	line(&b, 1, "source_chapter_count: %d", len(screenplay.SourceChapters))
	line(&b, 1, "language: %s", quote(screenplay.Language))
	line(&b, 1, "generated_by:")
	line(&b, 2, "provider: %s", quote(screenplay.Provider))
	line(&b, 2, "mode: %s", quote(screenplay.Mode))
	line(&b, 1, "created_at: %s", quote(screenplay.CreatedAt))
	line(&b, 0, "characters:")
	for _, character := range screenplay.Characters {
		line(&b, 1, "- id: %s", quote(character.ID))
		line(&b, 2, "name: %s", quote(character.Name))
		line(&b, 2, "role: %s", quote(character.Role))
		line(&b, 2, "description: %s", quote(character.Description))
	}
	line(&b, 0, "source_chapters:")
	for _, chapter := range screenplay.SourceChapters {
		line(&b, 1, "- id: %s", quote(chapter.ID))
		line(&b, 2, "title: %s", quote(chapter.Title))
		line(&b, 2, "order: %d", chapter.Order)
		line(&b, 2, "summary: %s", quote(chapter.Summary))
	}
	line(&b, 0, "screenplay:")
	line(&b, 1, "acts:")
	for _, act := range screenplay.Acts {
		line(&b, 2, "- id: %s", quote(act.ID))
		line(&b, 3, "title: %s", quote(act.Title))
		line(&b, 3, "order: %d", act.Order)
		line(&b, 3, "scenes:")
		for _, scene := range act.Scenes {
			line(&b, 4, "- id: %s", quote(scene.ID))
			line(&b, 5, "source_chapter_ids:")
			for _, chapterID := range scene.SourceChapterIDs {
				line(&b, 6, "- %s", quote(chapterID))
			}
			line(&b, 5, "heading:")
			line(&b, 6, "location: %s", quote(scene.Heading.Location))
			line(&b, 6, "time: %s", quote(scene.Heading.Time))
			line(&b, 6, "interior: %t", scene.Heading.Interior)
			line(&b, 5, "summary: %s", quote(scene.Summary))
			line(&b, 5, "characters:")
			for _, characterID := range scene.Characters {
				line(&b, 6, "- %s", quote(characterID))
			}
			line(&b, 5, "beats:")
			for _, beat := range scene.Beats {
				line(&b, 6, "- type: %s", quote(beat.Type))
				if beat.CharacterID != "" {
					line(&b, 7, "character_id: %s", quote(beat.CharacterID))
				}
				if beat.CharacterName != "" {
					line(&b, 7, "character_name: %s", quote(beat.CharacterName))
				}
				line(&b, 7, "text: %s", quote(beat.Text))
			}
		}
	}
	line(&b, 0, "export:")
	line(&b, 1, "format: %s", quote("yaml"))
	line(&b, 1, "filename: %s", quote("mock-screenplay.yaml"))

	return b.String()
}

func line(b *strings.Builder, indent int, format string, args ...any) {
	b.WriteString(strings.Repeat("  ", indent))
	b.WriteString(fmt.Sprintf(format, args...))
	b.WriteByte('\n')
}

func quote(value string) string {
	return strconv.Quote(value)
}
