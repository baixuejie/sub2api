# DeepSeek Harness 一键安装 Extension

该 Extension 为 API Key 行提供 DeepSeek Harness 本地安装入口。功能默认关闭，管理员需在“系统设置 -> 站点设置 -> DeepSeek Harness 一键安装”中显式启用。

## 启用前提

1. 在站点设置中配置公开 `api_base_url`。生产环境必须使用 HTTPS；仅 loopback 本地开发允许 HTTP。Provider 端点会规范化为该地址加一个 `/v1`。
2. 发布并签名 `tools/deepseek-harness-helper/` 的六平台产物。默认下载版本为 `dsh-helper-v0.1.0`。
3. 确认 Redis 可用，然后再启用 Feature Flag `ext_deepseek_harness_enabled`。

可通过环境变量覆盖 Helper 发布位置：

```text
DEEPSEEK_HARNESS_HELPER_RELEASE_BASE_URL=https://downloads.example.com/dsh-helper-v0.1.0
DEEPSEEK_HARNESS_HELPER_RELEASES_PAGE=https://downloads.example.com/deepseek-harness-helper
```

构建工作流：`.github/workflows/deepseek-harness-helper-build.yml`。它只保留未签名 CI artifact，不会直接发布。受保护的发布流水线必须完成 Windows 代码签名、macOS codesign/notarize，再把六个平台文件和 `SHA256SUMS` 上传到 `dsh-helper-v*` Release。仓库不包含签名凭据，未签名产物不应在生产站点启用。

## 安全流程

1. 已登录用户仅提交 `api_key_id` 和可选模型。
2. 后端按 JWT 用户重新读取 Key，校验所有权、状态、过期、额度、分组和模型白名单。
3. Redis 保存 120 秒随机票据的 SHA-256 摘要，以及 1 小时安装会话；会话不保存明文 API Key。
4. 浏览器通过 `sub2api-harness://bootstrap` 把站点 origin 和一次性票据交给本地 Helper。
5. Helper 在任何网络请求前显示系统确认框；用户确认精确 origin 后才持久信任该站点，其他站点必须重新确认。
6. Helper 通过 HTTPS 兑换票据。后端再次校验 Key 后才在响应中返回当前 Key；票据使用 Redis `GETDEL` 原子消费。
7. Helper 使用独立事件令牌上报进度；后端只保存事件令牌哈希。状态通过 Redis `WATCH` 原子更新，终态不能被旧事件覆盖；内容完全一致的终态重试保持幂等，用于处理响应丢失。
8. 浏览器只轮询脱敏状态。完成 URL 必须是带端口的 loopback HTTP URL。

## API

```text
GET  /api/v1/deepseek-harness/profile?api_key_id=<id>   JWT
POST /api/v1/deepseek-harness/sessions                  JWT
GET  /api/v1/deepseek-harness/sessions/:id              JWT
POST /api/v1/deepseek-harness/exchange                  one-time ticket
POST /api/v1/deepseek-harness/sessions/:id/events       Bearer event token
```

所有接口都会在请求时读取 Feature Flag。关闭开关后，现有会话的查询、兑换和事件上报也会立即停止。

## 通用 Helper 契约

当前 DSH 接口继续保留原路径和 `dsh_version`，同时在 profile 和 exchange 任务中返回：

```json
{
  "protocol_version": "1",
  "tool_id": "deepseek-harness",
  "tool_version": "0.1.0-rc.6",
  "minimum_helper_version": "0.1.0"
}
```

Helper 通过 `tool_id` 分派二进制内注册的白名单 adapter。后续 Hermes、OpenClaw 等工具应新增独立 Extension 和 adapter，或复用已发布 Helper 的受控能力；服务端不能下发任意安装脚本、可执行文件或命令参数。只有现有 Helper 已具备所需能力时，新工具才可以免升级接入；新增底层能力时必须提高 `minimum_helper_version`。

## 模型策略

- OpenAI: `openai-responses`, 默认 `gpt-5.6-sol`
- Anthropic: `anthropic-messages`, 默认 `claude-opus-5`
- Grok: `openai-responses`, 默认 `grok-4.5`
- Gemini: `openai-completions`
- Antigravity: Claude 模型使用 `anthropic-messages`，其余使用 `openai-completions`
- Composite: `openai-completions`

启用分组 `ModelsListConfig` 时，以分组模型列表为候选；平台默认模型不在列表中时选择列表第一项。Helper 总是写入 `llm-pi-ai` provider 和 `agent-default-model`。由于凭据名按契约固定为 `SUB2API_API_KEY`，每次安装会清理上一次的 `sub2api-*` provider，只保留当前 Key 对应的一个 managed provider，避免旧 provider 静默读取新 Key。
