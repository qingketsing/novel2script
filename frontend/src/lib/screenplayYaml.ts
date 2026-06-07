import { parse } from "yaml";
import type {
  PreviewAct,
  PreviewBeat,
  PreviewCharacter,
  PreviewScene,
  PreviewSourceChapter,
  ScreenplayPreview,
} from "../types";

export type ParsePreviewResult =
  | { ok: true; preview: ScreenplayPreview }
  | { ok: false; message: string };

export function parseScreenplayPreview(yamlText: string): ParsePreviewResult {
  try {
    const doc = parse(yamlText) as unknown;
    if (!isRecord(doc)) {
      return { ok: false, message: "YAML 顶层结构不是对象。" };
    }

    const metadata = recordValue(doc, "metadata");
    const generatedBy = recordValue(metadata, "generated_by");
    const screenplay = recordValue(doc, "screenplay");

    return {
      ok: true,
      preview: {
        title: stringValue(metadata?.title, "未命名剧本"),
        sourceChapterCount: numberValue(metadata?.source_chapter_count),
        generatedMode: stringValue(generatedBy?.mode, "unknown"),
        characters: arrayValue(doc.characters).map(toCharacter),
        sourceChapters: arrayValue(doc.source_chapters).map(toSourceChapter),
        acts: arrayValue(screenplay?.acts).map(toAct),
      },
    };
  } catch {
    return { ok: false, message: "YAML 无法解析，请查看源码确认格式。" };
  }
}

function toCharacter(value: unknown): PreviewCharacter {
  const item = asRecord(value);
  return {
    id: stringValue(item.id, "unknown"),
    name: stringValue(item.name, "未命名角色"),
    role: stringValue(item.role, "未标注"),
    description: stringValue(item.description, ""),
    motivation: stringValue(item.motivation, ""),
  };
}

function toSourceChapter(value: unknown): PreviewSourceChapter {
  const item = asRecord(value);
  return {
    id: stringValue(item.id, "chapter_unknown"),
    title: stringValue(item.title, "未命名章节"),
    order: numberValue(item.order),
    summary: stringValue(item.summary, ""),
  };
}

function toAct(value: unknown): PreviewAct {
  const item = asRecord(value);
  return {
    id: stringValue(item.id, "act_unknown"),
    title: stringValue(item.title, "未命名幕"),
    order: numberValue(item.order),
    scenes: arrayValue(item.scenes).map(toScene),
  };
}

function toScene(value: unknown): PreviewScene {
  const item = asRecord(value);
  const heading = recordValue(item, "heading");
  return {
    id: stringValue(item.id, "scene_unknown"),
    sourceChapterIds: arrayValue(item.source_chapter_ids).map((id) => stringValue(id, "")),
    heading: {
      location: stringValue(heading?.location, "未标注地点"),
      time: stringValue(heading?.time, "未标注时间"),
      interior: typeof heading?.interior === "boolean" ? heading.interior : null,
    },
    summary: stringValue(item.summary, ""),
    characters: arrayValue(item.characters).map((id) => stringValue(id, "")),
    beats: arrayValue(item.beats).map(toBeat),
  };
}

function toBeat(value: unknown): PreviewBeat {
  const item = asRecord(value);
  return {
    type: stringValue(item.type, "note"),
    text: stringValue(item.text, ""),
    characterId: stringValue(item.character_id, ""),
    characterName: stringValue(item.character_name, ""),
  };
}

function recordValue(value: unknown, key: string): Record<string, unknown> | undefined {
  if (!isRecord(value)) {
    return undefined;
  }
  const nested = value[key];
  return isRecord(nested) ? nested : undefined;
}

function asRecord(value: unknown): Record<string, unknown> {
  return isRecord(value) ? value : {};
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function arrayValue(value: unknown): unknown[] {
  return Array.isArray(value) ? value : [];
}

function stringValue(value: unknown, fallback: string): string {
  return typeof value === "string" && value.trim() ? value.trim() : fallback;
}

function numberValue(value: unknown): number {
  return typeof value === "number" && Number.isFinite(value) ? value : 0;
}
