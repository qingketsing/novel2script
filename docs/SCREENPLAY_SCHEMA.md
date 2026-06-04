# 剧本 YAML Schema 设计文档

## 设计目标

本 Schema 用于承接“AI 小说转剧本工具”的核心输出：把 3 个章节以上的小说文本转换为可编辑、可预览、可导出的结构化剧本 YAML。

它不是最终拍摄剧本格式，而是面向作者和编剧助理的剧本初稿结构。设计重点是：

- 保留小说章节来源，方便作者回溯原文。
- 把剧本拆成场景、角色、动作、对白、旁白和转场，便于后续编辑。
- 输出稳定的 YAML，方便前端预览、后端导出和后续多版本生成。
- 允许 mock mode 和真实 AI provider 共享同一份输出契约。
- 为后续分镜、局部重生成、版本对比预留扩展字段。

## 顶层结构

```yaml
schema_version: "1.0"
metadata:
  title: "示例剧本"
  source_type: "novel"
  source_chapter_count: 3
  language: "zh-CN"
  generated_by:
    provider: "deepseek-v4"
    mode: "mock"
  created_at: "2026-06-05T00:00:00+08:00"
characters:
  - id: "char_001"
    name: "林舟"
    role: "protagonist"
    description: "年轻小说作者，正在寻找失踪的姐姐。"
source_chapters:
  - id: "chapter_001"
    title: "第一章 雨夜来信"
    order: 1
    summary: "林舟在雨夜收到一封没有署名的信。"
screenplay:
  acts:
    - id: "act_001"
      title: "开端"
      order: 1
      scenes:
        - id: "scene_001"
          source_chapter_ids:
            - "chapter_001"
          heading:
            location: "旧书店"
            time: "夜"
            interior: true
          summary: "林舟在旧书店读到姐姐留下的线索。"
          characters:
            - "char_001"
          beats:
            - type: "action"
              text: "雨水敲打玻璃，林舟推门进入旧书店。"
            - type: "dialogue"
              character_id: "char_001"
              character_name: "林舟"
              text: "这封信，是谁放在这里的？"
            - type: "transition"
              text: "切至窗外的街灯。"
export:
  format: "yaml"
  filename: "sample-screenplay.yaml"
```

## 字段定义

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `schema_version` | string | 是 | Schema 版本号。MVP 固定为 `"1.0"`，后续兼容演进。 |
| `metadata` | object | 是 | 剧本生成元信息，包括标题、来源、语言、AI provider 和生成时间。 |
| `characters` | array | 是 | 角色列表。用于统一角色 ID，避免同一角色在多个场景中名称漂移。 |
| `source_chapters` | array | 是 | 原小说章节摘要列表。用于证明输入满足 3 个章节以上，并支持回溯来源。 |
| `screenplay` | object | 是 | 剧本主体。MVP 以幕和场景组织内容。 |
| `export` | object | 否 | 导出相关信息。前端下载和后端导出接口可使用。 |

## metadata

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `title` | string | 是 | 剧本标题。可来自小说标题，也可由系统生成。 |
| `source_type` | string | 是 | 固定为 `"novel"`，表示来源是小说文本。 |
| `source_chapter_count` | number | 是 | 输入章节数量，必须大于或等于 3。 |
| `language` | string | 是 | 输出语言，例如 `"zh-CN"`。 |
| `generated_by` | object | 是 | 生成来源信息。 |
| `created_at` | string | 否 | ISO 8601 时间字符串。 |

### generated_by

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `provider` | string | 是 | 设计上使用 `"deepseek-v4"`，MVP 不进行真实 API 调用。 |
| `mode` | string | 是 | `"mock"` 或 `"api"`。MVP 默认 `"mock"`。 |

## characters

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | 角色唯一 ID，例如 `char_001`。 |
| `name` | string | 是 | 角色名。 |
| `role` | string | 否 | 角色功能，例如 `protagonist`、`supporting`、`antagonist`。 |
| `description` | string | 否 | 角色简介、人物关系或性格提示。 |

## source_chapters

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | 章节唯一 ID，例如 `chapter_001`。 |
| `title` | string | 是 | 章节标题。没有标题时可由解析器生成。 |
| `order` | number | 是 | 章节顺序，从 1 开始。 |
| `summary` | string | 否 | 章节摘要，用于预览和剧本生成解释。 |

## screenplay

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `acts` | array | 是 | 剧本幕列表。MVP 可以只生成一幕，但仍保留幕结构。 |

### acts

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | 幕唯一 ID，例如 `act_001`。 |
| `title` | string | 否 | 幕标题，例如“开端”“发展”“高潮”。 |
| `order` | number | 是 | 幕顺序。 |
| `scenes` | array | 是 | 场景列表。 |

### scenes

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | string | 是 | 场景唯一 ID，例如 `scene_001`。 |
| `source_chapter_ids` | array | 是 | 场景对应的原小说章节 ID，可包含多个章节。 |
| `heading` | object | 是 | 场景头，描述地点、时间和内外景。 |
| `summary` | string | 否 | 场景摘要。 |
| `characters` | array | 否 | 本场出现的角色 ID。 |
| `beats` | array | 是 | 场景中的动作、对白、旁白和转场。 |

### heading

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `location` | string | 是 | 场景地点。 |
| `time` | string | 是 | 场景时间，例如“晨”“夜”“黄昏”。 |
| `interior` | boolean | 否 | `true` 表示内景，`false` 表示外景；未知时可省略。 |

### beats

`beats` 是剧本最小可编辑单元。MVP 支持以下类型：

| type | 必填字段 | 说明 |
| --- | --- | --- |
| `action` | `text` | 动作描写、画面描述、角色行为。 |
| `dialogue` | `character_id`, `character_name`, `text` | 角色对白。 |
| `narration` | `text` | 旁白、画外音或保留的叙述信息。 |
| `transition` | `text` | 转场，例如“切至”“淡出”。 |

通用字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `type` | string | 是 | beat 类型。 |
| `text` | string | 是 | 具体内容。 |
| `character_id` | string | 条件必填 | `type` 为 `dialogue` 时必填。 |
| `character_name` | string | 条件必填 | `type` 为 `dialogue` 时必填，方便前端直接展示。 |

## 校验规则

- `metadata.source_chapter_count` 必须大于或等于 3。
- `source_chapters` 数量必须大于或等于 3。
- `source_chapters[*].order` 必须连续递增。
- `characters[*].id` 必须唯一。
- `screenplay.acts[*].scenes[*].id` 必须唯一。
- `scenes[*].source_chapter_ids` 必须引用已存在的 `source_chapters[*].id`。
- `scenes[*].characters` 必须引用已存在的 `characters[*].id`。
- `dialogue` 类型 beat 必须包含 `character_id`、`character_name` 和 `text`。
- 所有 `text` 字段不能为空字符串。

## 设计原因

### 为什么保留 source_chapters

小说改剧本不是简单改写，而是从叙事文本中提取可表演、可拍摄的内容。保留 `source_chapters` 可以让作者知道每个场景来自哪些章节，也方便后续做局部重生成和差异对比。

### 为什么用 characters 统一角色

AI 生成时容易出现角色名称不一致、别名混用、配角重复的问题。统一角色表可以让前端预览更稳定，也能让后续分镜、对白统计和角色筛选有明确引用。

### 为什么用 acts 和 scenes

剧本天然以场景组织，但小说章节和剧本场景并不总是一一对应。`acts` 提供更高层的结构，`scenes` 承载具体戏剧单元。MVP 可以只生成一幕，后续可以扩展为三幕式或更多结构。

### 为什么用 beats

`beats` 把动作、对白、旁白、转场拆成最小编辑单元。这样前端不需要解析整段文本，就能直接做结构化预览、局部编辑和导出。它也让 mock mode 更容易生成稳定示例。

### 为什么 YAML 而不是纯文本

YAML 对作者更可读，也比纯文本更容易被程序解析。它适合在 MVP 中同时满足“人可编辑”和“系统可校验”的要求。后续如果需要 API 交换，也可以从同一结构转换为 JSON。

### 为什么不在 Schema 中绑定真实 AI 响应

MVP 的重点是验证主链路和输出结构，而不是验证模型能力。Schema 只记录 `provider` 和 `mode`，不依赖真实 `deepseek-v4` 响应格式。这样 mock mode、测试环境和未来真实 API 接入可以复用同一份产品契约。

## 扩展方向

- `versions`：支持同一输入生成多个剧本版本。
- `storyboards`：支持分镜号、景别、镜头运动和画面说明。
- `revision_notes`：记录用户编辑和 AI 再生成原因。
- `confidence`：记录角色识别、场景抽取或章节摘要的置信度。
- `source_spans`：记录输出片段对应的原文位置，支持更精细的回溯。
