# AI 提示词模板设计文档

## 文档目标

本文档定义 Novel2Script 后续接入 AI 生成能力时需要使用的提示词模板。当前 MVP 不接入真实大模型 API，默认使用 mock mode；但提示词模板需要提前明确 AI 的任务边界、输入格式、输出格式和稳定性要求。

设计目标：

- 让 mock mode 和未来真实 `deepseek-v4` 接入共享同一套任务定义。
- 保证输出围绕剧本 YAML Schema，不生成不可控的自由文本。
- 降低模型幻觉，要求无法判断的信息显式留空或标记为推断。
- 支持前端预览、后端导出、错误修复和后续测试。

## 通用约束

所有提示词都应遵守以下原则：

- 输出必须是结构化结果，避免解释性长文。
- 不得编造不存在的章节、角色和核心事件。
- 可以压缩、改写、剧本化表达，但不能改变故事主线。
- 角色名应保持一致；如存在别名，需要归并到同一角色。
- 对不确定信息使用空字符串、空数组或 `unknown`，不要强行补全。
- 默认语言为中文。
- MVP 阶段只使用 mock mode，不真实调用 `deepseek-v4`。

## 任务一：章节解析

### 用途

从用户上传或粘贴的 `.txt` / `.md` 小说文本中识别章节结构，验证是否满足 3 个章节以上。

### 输入

```text
你将收到一段小说文本，来源可能是 .txt 或 .md。
请识别章节标题、章节顺序和章节正文。

小说文本：
{{NOVEL_TEXT}}
```

### 输出格式

```yaml
chapters:
  - id: "chapter_001"
    title: "第一章 示例标题"
    order: 1
    text: "章节正文"
    word_count: 1200
validation:
  chapter_count: 3
  has_minimum_chapters: true
  errors: []
```

### 规则

- 章节 ID 从 `chapter_001` 开始递增。
- 如果原文没有明确标题，可生成 `第 1 章`、`第 2 章` 作为标题。
- `has_minimum_chapters` 必须基于实际识别结果判断。
- 章节不足 3 个时，`errors` 中需要说明原因。

## 任务二：角色提取

### 用途

从已解析章节中提取主要角色、配角、别名和简要人物关系。

### 输入

```text
请从以下小说章节中提取角色信息。
不要编造未出现的角色。
如果同一角色存在不同称呼，请归并为同一个角色。

章节列表：
{{CHAPTERS_YAML}}
```

### 输出格式

```yaml
characters:
  - id: "char_001"
    name: "林舟"
    aliases:
      - "小舟"
    role: "protagonist"
    description: "年轻小说作者，正在寻找失踪的姐姐。"
    relationships:
      - target_character_name: "林夏"
        relation: "姐姐"
```

### 规则

- 角色 ID 从 `char_001` 开始递增。
- 主角优先排在前面。
- `role` 可使用 `protagonist`、`supporting`、`antagonist`、`unknown`。
- 无法确定关系时，`relationships` 使用空数组。

## 任务三：场景提取

### 用途

把小说章节拆分为适合剧本化的场景单元，为后续剧本生成提供结构。

### 输入

```text
请根据章节内容和角色列表，提取可剧本化的场景。
每个场景应包含来源章节、地点、时间、出现角色和剧情摘要。

章节列表：
{{CHAPTERS_YAML}}

角色列表：
{{CHARACTERS_YAML}}
```

### 输出格式

```yaml
scenes:
  - id: "scene_001"
    source_chapter_ids:
      - "chapter_001"
    location: "旧书店"
    time: "夜"
    interior: true
    character_ids:
      - "char_001"
    summary: "林舟在旧书店收到一封没有署名的信。"
    dramatic_purpose: "引出姐姐失踪的线索。"
```

### 规则

- 场景 ID 从 `scene_001` 开始递增。
- 一个场景可以引用多个章节，但必须引用真实存在的章节 ID。
- 如果内外景无法判断，`interior` 可以省略。
- `dramatic_purpose` 应简短说明该场景在剧情中的作用。

## 任务四：剧本 YAML 生成

### 用途

根据章节、角色和场景信息，生成符合 `docs/SCREENPLAY_SCHEMA.md` 的剧本 YAML 初稿。

### 输入

```text
请将以下小说结构转换为剧本 YAML。
输出必须符合剧本 YAML Schema。
不要输出 Markdown 代码块，不要输出解释文字，只输出 YAML。

metadata:
  title: "{{TITLE}}"
  source_type: "novel"
  source_chapter_count: {{SOURCE_CHAPTER_COUNT}}
  language: "zh-CN"
  generated_by:
    provider: "deepseek-v4"
    mode: "{{GENERATION_MODE}}"

章节列表：
{{CHAPTERS_YAML}}

角色列表：
{{CHARACTERS_YAML}}

场景列表：
{{SCENES_YAML}}
```

### 输出格式

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
export:
  format: "yaml"
  filename: "screenplay.yaml"
```

### 规则

- `source_chapter_count` 必须等于输入章节数量。
- `source_chapters` 必须保留所有章节的 ID、标题、顺序和摘要。
- `characters` 必须使用角色提取阶段生成的角色 ID。
- `scenes[*].characters` 只能引用已存在的角色 ID。
- `beats` 必须至少包含 `action` 或 `dialogue`。
- 对白必须同时包含 `character_id`、`character_name` 和 `text`。
- 不要输出 Markdown 代码块。

## 任务五：Schema 校验与修复

### 用途

当生成结果不符合 YAML Schema 时，要求 AI 只修复结构问题，不改变故事内容。

### 输入

```text
以下 YAML 剧本未通过校验。
请根据错误信息修复 YAML。
不要改变故事主线、角色关系和场景顺序。
不要输出解释文字，只输出修复后的 YAML。

校验错误：
{{VALIDATION_ERRORS}}

原始 YAML：
{{SCREENPLAY_YAML}}
```

### 输出格式

输出修复后的完整 YAML。

### 规则

- 只修复字段缺失、类型错误、引用错误、空文本和格式问题。
- 不新增与原文无关的角色或场景。
- 如果缺少必要字段但无法确定内容，使用 `unknown` 或空数组。
- 修复后必须仍符合 `schema_version: "1.0"`。

## mock mode 约定

MVP 阶段不真实调用 `deepseek-v4`，但 mock mode 应模拟同一任务边界：

- 输入和输出结构与真实 provider 保持一致。
- 同一输入应得到稳定可复现的结果。
- mock 数据不能依赖外部网络。
- mock 输出必须符合 YAML Schema。
- mock 输出应覆盖章节、角色、场景、动作、对白和导出字段。

mock mode 的价值是保证后端、前端、导出、错误处理和 demo 可以稳定开发。真实 AI 接入只应替换 provider 实现，不应改变主链路和 YAML 契约。

## 提示词设计取舍

### 分阶段而不是一步生成

小说到剧本涉及章节解析、角色归并、场景拆分和剧本化表达。分阶段提示词更容易校验中间结果，也方便定位错误来源。MVP 可以在 mock mode 中简化执行，但任务定义仍保持分层。

### 强制结构化输出

自由文本很难被前端预览和后端导出复用。所有模板都要求 YAML 或 YAML-like 结构，方便校验、测试和后续转换。

### 限制创造性

本工具服务于作者改编自己的作品，不应随意改写主线。提示词允许将小说表达改写成剧本表达，但不鼓励新增大段剧情。

### 保留修复模板

真实模型输出可能出现字段缺失或格式错误。单独的修复模板可以把“生成”和“校验修复”分开，降低主生成提示词复杂度，也让错误处理更清楚。
