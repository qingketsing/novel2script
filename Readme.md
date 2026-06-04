# Novel2Script

AI 小说转剧本工具，面向希望把小说作品快速改编为剧本初稿的作者、编剧助理和内容团队。

项目目标是降低小说改编剧本的第一步门槛：用户输入 3 个章节以上的小说文本，系统解析章节、角色、场景和情节节点，并生成可编辑、可继续打磨的结构化剧本 YAML。

## 核心定位

小说作者通常已经有完整故事、人物和叙事文本，但把小说改成剧本需要重新组织场景、对白、动作、转场和人物出场关系。Novel2Script 的 MVP 不追求直接产出最终拍摄稿，而是优先提供一个稳定、结构化、可编辑的剧本初稿。

## MVP 范围

P0 主链路：

- 支持上传或粘贴 `.txt` / `.md` 小说文本。
- 校验输入至少包含 3 个章节。
- 解析章节、角色、场景和基础情节信息。
- 使用 AI mock mode 生成结构化剧本。
- 导出 YAML 格式剧本。
- 提供干净的前端主链路和明确错误提示。

暂不做：

- 不接入真实大模型 API。
- 不做复杂在线协同编辑。
- 不做完整分镜生产系统。
- 不把输出包装成最终可拍摄剧本。

## 技术方向

- 后端语言：Golang。
- AI provider 设计来源：`deepseek-v4`。
- MVP 默认模式：mock mode。
- 输出格式：YAML。
- 前端目标：输入、生成、预览、导出四步主链路清晰稳定。

mock mode 的目的不是模拟模型能力上限，而是先验证产品结构、输出契约、错误处理和 demo 链路。等主链路稳定后，再评估真实 API 接入。

## 结构化输出示例

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
source_chapters:
  - id: "chapter_001"
    title: "第一章 雨夜来信"
    order: 1
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
          characters:
            - "char_001"
          beats:
            - type: "action"
              text: "雨水敲打玻璃，林舟推门进入旧书店。"
            - type: "dialogue"
              character_id: "char_001"
              character_name: "林舟"
              text: "这封信，是谁放在这里的？"
```

## Demo 主链路

1. 准备一段包含 3 个章节以上的小说文本。
2. 在前端上传 `.txt` / `.md`，或直接粘贴文本。
3. 系统解析章节并展示基础信息。
4. 触发 mock AI 生成剧本 YAML。
5. 在前端预览结构化结果。
6. 一键导出 YAML 文件，供作者继续编辑。

## 项目文档

- [任务板](docs/TASK_BOARD.md)：拆分 P0、P1、P2 任务、建议分支、PR 标题、允许修改文件和验收标准。
- [YAML Schema 设计文档](docs/SCREENPLAY_SCHEMA.md)：定义剧本 YAML 输出契约，并说明字段设计原因。
- 项目上下文文档：说明用户定位、MVP 边界、设计取舍和后续演进。
- 提示词模板文档：后续维护 AI 解析和生成提示词，不在当前阶段优先展开。

## 当前状态

项目处于文档和产品结构定稿阶段。当前优先事项是先明确任务拆分、YAML Schema、MVP 边界和 demo 叙事，再进入 Golang 后端与前端主链路实现。
