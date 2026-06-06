# API Mode Runtime Report

本文档用于记录 DeepSeek API mode 下的生成耗时和输出规模。它不是自动化评测脚本，只作为人工 smoke test 的结果记录，帮助判断长文本生成是否稳定、是否触发 fallback，以及输出 YAML 是否达到预期规模。

## 运行前提

使用 DeepSeek API mode 启动服务：

```bash
DEEPSEEK_API_KEY=your_api_key_here docker compose -f docker-compose.yml -f docker-compose.api.yml up --build
```

确认前端和后端可访问：

```text
frontend: http://localhost:5173
backend:  http://localhost:8080
```

为了验证真实 API 稳定性，建议关闭 fallback 后再跑一次：

```bash
AI_FALLBACK_TO_MOCK=false DEEPSEEK_API_KEY=your_api_key_here docker compose -f docker-compose.yml -f docker-compose.api.yml up --build
```

如果只是演示链路稳定性，可以保留 fallback：

```bash
AI_FALLBACK_TO_MOCK=true
```

## 记录字段

每次 API mode 测试记录以下字段：

| 字段 | 说明 |
| --- | --- |
| `case` | 测试样例名称。 |
| `chapters` | 输入章节数 / 响应 `chapter_count`。 |
| `content_length` | 后端日志中的输入正文长度。 |
| `prompt_length` | 后端日志中的 prompt 长度。 |
| `timeout_ms` | 后端为该输入计算的动态超时时间。 |
| `duration_ms` | 后端完成 `/api/convert` 的总耗时。 |
| `mode` | 响应模式，真实 API 成功应为 `api`。 |
| `yaml_length` | 后端日志中的 YAML 长度。 |
| `fallback` | 是否出现 `convert fallback activated`。 |
| `repair` | 是否出现 `deepseek yaml repair succeeded`。 |
| `notes` | 其他观察，例如是否超过 60 秒、是否出现 context canceled。 |

## 日志采集方式

查看后端日志：

```bash
docker compose logs backend
```

关键日志：

```text
deepseek generation started
deepseek generation returned
deepseek yaml validation succeeded
deepseek yaml repair succeeded
convert fallback activated
screenplay generation completed
convert request completed
http request completed
```

判断规则：

- `mode=api`：真实 DeepSeek API 输出通过校验并返回。
- `mode=mock` 且出现 `convert fallback activated`：真实 API 失败后回退 mock。
- `duration_ms` 约为 `60000` 且错误为 `context canceled`：通常是上游代理或浏览器请求链路取消。
- `duration_ms` 接近 `timeout_ms` 且错误为 `context deadline exceeded`：通常是后端动态超时触发。

## 运行记录

| case | chapters | content_length | prompt_length | timeout_ms | duration_ms | mode | yaml_length | fallback | repair | notes |
| --- | ---: | ---: | ---: | ---: | ---: | --- | ---: | --- | --- | --- |
| N17 6 chapters long | 6 / 6 | 10822 | 14073 | 134000 | 106630 | api | 25519 | no | no | 长文本超过 60 秒后仍成功返回，验证 Nginx API 代理超时修复生效。 |

## 建议后续补充样例

后续可以继续补充以下固定输入：

```text
docs/examples/api-smoke/novel-3chapters-short.md
docs/examples/api-smoke/novel-5chapters-medium.md
docs/examples/api-smoke/novel-6chapters-long.md
```

推荐每个样例至少运行一次 fallback 关闭的 API mode 测试，并记录：

- `mode` 是否为 `api`
- `chapter_count` 是否正确
- `yaml_length` 是否明显异常
- 是否出现 YAML repair
- 是否出现 fallback
- 是否出现 60 秒附近的 `context canceled`
