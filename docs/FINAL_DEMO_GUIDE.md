# 最终 Demo 指南

本文档用于说明 Novel2Script MVP 如何运行、验证、演示和排障。

项目不是“AI 直接写出最终可拍摄剧本”，而是：

```text
3 章以上小说输入
-> 章节解析
-> AI 辅助生成剧本
-> YAML 结构校验与一致性校验
-> 输出可编辑的结构化剧本 YAML
-> 前端预览与导出
```

## 项目定位

小说作者通常已经有完整故事、人物和叙事文本，但把小说改编成剧本时，需要重新整理场景、动作、对白、人物出场和原文来源关系。

Novel2Script 的 MVP 聚焦“剧本初稿”阶段：

- 降低小说改编剧本的第一步门槛
- 输出结构化 YAML，而不是松散文本
- 保留人工继续编辑和打磨的空间
- 保留剧本内容和原小说章节之间的追溯关系
- 提供稳定的 mock mode，保证本地开发和现场 demo 不被外部 API 影响

## 推荐演示路线

1. 打开前端：

```text
http://localhost:5173
```

2. 加载或粘贴 3 章以上小说文本。

推荐样例：

```text
docs/examples/novel-example.md
docs/examples/api-smoke/novel-3chapters-short.md
docs/examples/api-smoke/novel-5chapters-medium.md
docs/examples/api-smoke/novel-6chapters-long.md
```

3. 点击生成剧本 YAML。

4. 讲解 YAML 的核心结构：

```text
schema_version
metadata
characters
source_chapters
screenplay.acts.scenes
beats
```

5. 说明结构化输出的价值：

- `source_chapters` 保留剧本和原小说章节的追溯关系。
- `characters` 帮助作者维护人物一致性。
- `scenes` 是剧本编辑的核心单位。
- `beats` 将动作和对白拆成可编辑节点。
- YAML 可读、可编辑、可校验，适合作为剧本初稿格式。

6. 展示复制或下载 YAML。

7. 说明 mock mode 保证演示稳定，API mode 展示真实 DeepSeek 生成能力。

## 启动 Mock Mode

mock mode 不需要 API key，是最稳定的现场演示方式。

```bash
docker compose up --build
```

打开：

```text
前端：http://localhost:5173
后端：http://localhost:8080
健康检查：http://localhost:8080/health
```

预期结果：

```text
response mode: mock
chapter_count: 输入章节数
screenplay_yaml: 非空 YAML
```

适合使用 mock mode 的场景：

- 没有 API key
- 网络不稳定
- 希望现场演示主链路稳定可控

## 启动 DeepSeek API Mode

API mode 需要本地提供 `DEEPSEEK_API_KEY`。

PowerShell：

```powershell
$env:DEEPSEEK_API_KEY="your_api_key_here"
$env:DEEPSEEK_BASE_URL="https://api.deepseek.com"
$env:DEEPSEEK_MODEL="deepseek-v4-flash"

docker compose -f docker-compose.yml -f docker-compose.api.yml up --build
```

macOS / Linux：

```bash
DEEPSEEK_API_KEY=your_api_key_here \
DEEPSEEK_BASE_URL=https://api.deepseek.com \
DEEPSEEK_MODEL=deepseek-v4-flash \
docker compose -f docker-compose.yml -f docker-compose.api.yml up --build
```

预期结果：

```text
response mode: api
metadata.generated_by.mode: api
```

真实 API key 只能放在本地环境变量或本地 `.env` 文件里，不能提交到仓库。

## 运行 API Quality Smoke

API mode 启动后，运行：

```powershell
powershell -ExecutionPolicy Bypass -File scripts\api-mode-quality-smoke.ps1
```

脚本会检查：

- 后端 `/health`
- 3 / 5 / 6 章固定小说样例
- `response.mode == api`
- `chapter_count` 是否正确
- YAML 是否包含核心结构字段
- 是否包含 source chapter 引用
- 是否没有 Markdown code fence

预期输出：

```text
3chapters-short   PASS   mode=api
5chapters-medium  PASS   mode=api
6chapters-long    PASS   mode=api
```

如果本地验证时允许 fallback：

```powershell
powershell -ExecutionPolicy Bypass -File scripts\api-mode-quality-smoke.ps1 -AllowMockFallback
```

## 日志与排障

后端日志是排查问题的主要入口：

```bash
docker compose logs backend
```

关键日志：

```text
http request completed
convert request completed
convert pipeline started
novel parsed
screenplay generation started
deepseek generation started
deepseek generation returned
deepseek yaml validation succeeded
deepseek yaml validation failed
deepseek yaml repair succeeded
deepseek yaml repair failed
convert fallback activated
convert fallback completed
```

重点字段：

```text
request_id
method
path
status
duration_ms
chapter_count
mode
yaml_length
prompt_length
timeout_ms
error_code
```

日志不会记录：

- API key
- 完整小说正文
- 完整 prompt
- 完整 YAML 输出

## 常见问题

### 前端启动时有很多 Nginx 日志

这是正常现象。前端 Docker 镜像使用 Nginx 托管静态文件，`start worker process` 这类日志是 Nginx 启动 notice，不是前端业务错误。

排查后端生成问题时，优先看：

```bash
docker compose logs backend
```

### API mode 返回了 mock

通常说明触发了 fallback。

查看后端日志：

```text
convert fallback activated
```

常见原因：

- API key 缺失或无效
- DeepSeek 请求失败
- DeepSeek 请求超时
- AI 返回 YAML 未通过校验且修复失败

### YAML 校验失败

查看后端日志：

```text
deepseek yaml validation failed
deepseek yaml repair succeeded
deepseek yaml repair failed
```

后端会在返回前校验 YAML 结构和业务一致性。如果修复失败且开启 fallback，系统会回退到 mock，保证 demo 主链路仍可完成。

### API mode 生成较慢

长文本生成需要更长时间。后端会根据输入规模估算 DeepSeek 动态超时。

查看日志：

```text
deepseek generation started timeout_ms=...
deepseek generation returned duration_ms=...
```

### 端口冲突

默认端口：

```text
前端：5173
后端：8080
```

可通过本地环境变量覆盖：

```text
FRONTEND_PORT
BACKEND_PORT
```

## 最终验收清单

演示前建议运行：

```bash
go test ./...
```

然后确认：

- Docker mock mode 可以启动。
- 前端可以打开 `http://localhost:5173`。
- 后端健康检查返回 `{"status":"ok"}`。
- mock `/api/convert` 返回 `mode=mock`。
- Docker API mode 可以用本地 `DEEPSEEK_API_KEY` 启动。
- API `/api/convert` 返回 `mode=api`。
- `scripts/api-mode-quality-smoke.ps1` 通过。
- YAML Schema 文档存在：`docs/SCREENPLAY_SCHEMA.md`。
- 示例小说和示例 YAML 存在：`docs/examples/`。

## 演示讲解重点

- MVP 边界是生成“可编辑剧本初稿”，不是直接生成最终拍摄稿。
- YAML 被选中是因为它可读、结构化、可校验。
- scene 是剧本编辑的核心单位。
- source chapter 引用保留了与原小说的追溯关系。
- mock mode 保证 demo 稳定。
- API mode 展示真实 AI 生成能力。
- 校验、修复、fallback、动态超时和请求日志让 AI 链路更可控、更容易排查。
