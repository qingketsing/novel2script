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
- 支持 mock mode 与 DeepSeek API mode 生成结构化剧本。
- 导出 YAML 格式剧本。
- 提供干净的前端主链路和明确错误提示。

暂不做：

- 不做复杂在线协同编辑。
- 不做完整分镜生产系统。
- 不把输出包装成最终可拍摄剧本。

## 技术方向

- 后端语言：Golang。
- AI provider 设计来源：`deepseek-v4`。
- MVP 默认模式：mock mode，可通过环境变量切换 DeepSeek API mode。
- 输出格式：YAML。
- 前端目标：输入、生成、预览、导出四步主链路清晰稳定。

mock mode 的目的不是模拟模型能力上限，而是保证本地开发、自动化测试和比赛 demo 具备稳定 fallback。DeepSeek API mode 用于真实 AI 生成，模型输出会经过 YAML 校验后再返回前端。

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
4. 触发 mock 或 DeepSeek API 生成剧本 YAML。
5. 在前端预览结构化结果。
6. 一键导出 YAML 文件，供作者继续编辑。

## 本地运行

启动后端：

```bash
go run ./backend/cmd/server
```

启动前端：

```bash
cd frontend
npm ci
npm run dev
```

默认访问地址：

```text
前端：http://127.0.0.1:5173/
后端：http://localhost:8080
健康检查：http://localhost:8080/health
```

## AI 运行配置

默认使用 mock mode，不需要 API key：

```bash
AI_MODE=mock
```

启用 DeepSeek API mode 时，需要配置：

```bash
AI_MODE=deepseek
DEEPSEEK_API_KEY=your_api_key_here
DEEPSEEK_BASE_URL=https://api.deepseek.com
DEEPSEEK_MODEL=deepseek-v4-flash
DEEPSEEK_TIMEOUT_SECONDS=30
```

说明：

- `AI_MODE=mock`：稳定 demo fallback，不访问外部模型。
- `AI_MODE=deepseek`：调用 DeepSeek-compatible chat completions API。
- `DEEPSEEK_API_KEY`：真实 API key，只能放在本地环境变量中，不能提交到仓库。
- `DEEPSEEK_BASE_URL`：默认可使用 `https://api.deepseek.com`。
- `DEEPSEEK_MODEL`：建议本地测试使用 `deepseek-v4-flash`，需要更强生成质量时可切换模型。
- `DEEPSEEK_TIMEOUT_SECONDS`：DeepSeek 请求超时时间，默认 30 秒。

PowerShell 示例：

```powershell
$env:AI_MODE="deepseek"
$env:DEEPSEEK_API_KEY="your_api_key_here"
$env:DEEPSEEK_BASE_URL="https://api.deepseek.com"
$env:DEEPSEEK_MODEL="deepseek-v4-flash"
$env:DEEPSEEK_TIMEOUT_SECONDS="30"

go run ./backend/cmd/server
```

如果使用 `.env` 文件，请只保存在本地，并确认不会提交。当前后端读取的是进程环境变量，不会自动加载 `.env` 文件。

## API Smoke Test

后端启动后，可以用 3 章样例测试生成接口：

```bash
curl -X POST http://localhost:8080/api/convert \
  -H "Content-Type: application/json" \
  -d '{
    "title": "雨夜来信",
    "input_type": "md",
    "content": "# 第一章 雨夜来信\n林舟在雨夜收到一封没有署名的信。\n\n# 第二章 旧书店\n林舟来到旧书店，寻找姐姐留下的线索。\n\n# 第三章 街灯\n街灯忽明忽暗，线索指向城市另一端。"
  }'
```

成功响应会包含：

```json
{
  "screenplay_yaml": "schema_version: \"1.0\"\n...",
  "chapter_count": 3,
  "mode": "api"
}
```

在 mock mode 下，`mode` 会返回 `mock`。

## 项目文档

- [任务板](docs/TASK_BOARD.md)：拆分 P0、P1、P2 任务、建议分支、PR 标题、允许修改文件和验收标准。
- [YAML Schema 设计文档](docs/SCREENPLAY_SCHEMA.md)：定义剧本 YAML 输出契约，并说明字段设计原因。
- 项目上下文文档：说明用户定位、MVP 边界、设计取舍和后续演进。
- 提示词模板文档：后续维护 AI 解析和生成提示词，不在当前阶段优先展开。

## 当前状态

项目已完成 MVP 主链路：前端输入或上传小说文本，后端解析 3 个以上章节，并通过 mock mode 或 DeepSeek API mode 生成结构化剧本 YAML。当前优先事项是继续增强 AI 输出兜底、完善 demo 说明，并打磨前端预览体验。
