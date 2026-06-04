# AI 小说转剧本工具任务板

## 项目边界

- 目标用户：希望把 3 个章节以上小说文本快速改编为剧本初稿的小说作者、编剧助理和内容团队。
- 核心场景：上传或粘贴 `.txt` / `.md` 小说文本，系统解析章节、角色、场景和情节节点，经 AI mock 处理后生成可编辑的结构化剧本 YAML。
- 技术约束：后端使用 Golang；AI API 来源按 `deepseek-v4` 设计，但 MVP 不接入真实大模型，必须提供稳定 mock mode。
- 输出要求：结构化 YAML、可导出、带示例、带 YAML Schema 设计文档和设计原因说明。
- 质量要求：主链路稳定、UI 干净、错误处理明确、README 专业、demo 讲得清楚。

## P0 必做任务

| 任务 | 建议分支名 | PR 标题 | 允许修改的文件 | 验收标准 | 风险等级 |
| --- | --- | --- | --- | --- | --- |
| 搭建 Golang 后端项目骨架 | `backend/bootstrap-go-service` | `feat: bootstrap golang backend service` | `backend/**`, `go.mod`, `go.sum` | 后端服务可本地启动；提供健康检查接口；目录结构能承载解析、生成、导出模块；不接入真实 AI | 中 |
| 实现小说文本上传与解析 | `backend/parse-novel-input` | `feat: parse txt and markdown novel input` | `backend/**` | 支持 `.txt` / `.md` 上传或文本提交；校验至少 3 个章节；能返回章节列表、标题、正文长度和基础错误信息 | 高 |
| 设计 AI mock 生成流程 | `backend/mock-script-generation` | `feat: add mock script generation pipeline` | `backend/**`, `docs/PROMPT_TEMPLATES.md` | 使用 `deepseek-v4` 命名抽象 AI provider；默认 mock mode；根据章节、角色、场景生成稳定可复现的剧本 YAML 草稿；无真实 API 调用 | 高 |
| 定义结构化剧本 YAML Schema 文档 | `docs/yaml-schema-design` | `docs: add screenplay yaml schema design` | `docs/PROJECT_CONTEXT.md`, `docs/PROMPT_TEMPLATES.md`, `docs/SCREENPLAY_SCHEMA.md` | 文档说明 YAML 字段、层级、示例、必填项、可扩展字段；解释为何围绕章节、场景、角色、对白、动作、转场设计 | 中 |
| 实现 YAML 导出接口 | `backend/export-yaml` | `feat: add screenplay yaml export endpoint` | `backend/**` | 前端可调用导出接口；返回标准 YAML 文件内容；包含文件名、内容类型和错误响应；输出符合 Schema 文档 | 高 |
| 搭建前端主链路页面 | `frontend/bootstrap-main-flow` | `feat: bootstrap frontend conversion flow` | `frontend/**` | 用户可以上传或粘贴文本、触发生成、看到生成状态、下载 YAML；UI 干净；加载、空状态和失败状态可见 | 高 |
| 补充专业 README | `docs/readme-mvp-guide` | `docs: add mvp usage and demo guide` | `Readme.md` | README 说明定位、MVP 边界、运行方式、mock mode、示例输入输出、demo 讲解路径和设计取舍 | 中 |

## P1 核心功能任务

| 任务 | 建议分支名 | PR 标题 | 允许修改的文件 | 验收标准 | 风险等级 |
| --- | --- | --- | --- | --- | --- |
| 构建结构化剧本预览界面 | `frontend/structured-preview` | `feat: add structured screenplay preview` | `frontend/**` | 以章节、场景、角色、对白、动作分组展示 YAML 结果；支持一键加载示例；长文本可读且不挤压布局 | 中 |
| 增加示例小说与示例 YAML | `docs/sample-novel-and-output` | `docs: add sample novel and screenplay yaml` | `docs/**` | 提供 3 个章节以上示例小说片段和对应 YAML 输出；示例能用于 demo；不包含真实版权文本 | 低 |
| 增强后端错误处理 | `backend/error-handling` | `fix: improve backend validation and error responses` | `backend/**` | 上传格式错误、章节不足、生成失败、导出失败均返回结构化错误；前端可据此展示用户友好提示 | 高 |
| 增加端到端主链路验证 | `test/main-flow-validation` | `test: cover novel to yaml main flow` | `backend/**`, `frontend/**` | 覆盖上传解析、mock 生成、YAML 导出主链路；测试不依赖真实 AI 或外部网络 | 中 |
| 梳理项目上下文文档 | `docs/project-context` | `docs: add project context and design tradeoffs` | `docs/PROJECT_CONTEXT.md` | 说明用户定位、使用场景、MVP 边界、技术取舍、为什么先 mock AI、后续真实 API 接入策略 | 低 |

## P2 加分功能任务

| 任务 | 建议分支名 | PR 标题 | 允许修改的文件 | 验收标准 | 风险等级 |
| --- | --- | --- | --- | --- | --- |
| 多版本剧本生成 | `feature/multi-version-generation` | `feat: support multiple screenplay versions` | `backend/**`, `frontend/**` | 同一小说输入可生成多个不同风格或结构版本；版本可切换、比较和导出；mock mode 下结果稳定 | 高 |
| 分镜生成 | `feature/storyboard-generation` | `feat: add storyboard generation draft` | `backend/**`, `frontend/**`, `docs/**` | 基于剧本场景生成分镜草案；包含镜号、景别、画面描述、角色、对白引用和备注；输出仍可结构化导出 | 高 |
| 剧本局部编辑与再生成 | `feature/partial-regeneration` | `feat: support partial screenplay regeneration` | `backend/**`, `frontend/**` | 用户可选择单个场景或角色段落重新生成；不会覆盖未选中的内容；提供确认和错误提示 | 高 |
| 导出格式扩展 | `feature/export-formats` | `feat: add additional export formats` | `backend/**`, `frontend/**`, `docs/**` | 在 YAML 之外支持 Markdown 或 JSON 导出；格式说明清晰；不影响默认 YAML 主链路 | 中 |
| Demo 演示脚本优化 | `docs/demo-script` | `docs: add demo walkthrough script` | `docs/**`, `Readme.md` | 提供 3 分钟 demo 讲解顺序，覆盖真实痛点、输入、生成、预览、导出、mock mode 和设计取舍 | 低 |

## 执行原则

- P0 优先保障从输入到 YAML 导出的闭环，避免过早投入复杂编辑器和真实大模型接入。
- 每个 PR 应保持文件边界清晰，避免把文档、前端、后端和测试混在同一个大改动里。
- 所有 AI 能力先通过 mock mode 验证产品结构和主链路，待 Schema、错误处理和 demo 稳定后再评估真实 `deepseek-v4` 接入。
- YAML Schema 是产品契约，后端生成、前端预览、导出文件和示例文档都应围绕同一结构演进。
