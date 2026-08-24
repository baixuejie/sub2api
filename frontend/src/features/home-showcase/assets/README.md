# 首页工具图标

首页只加载本目录内的固定资源，不在浏览器运行时请求 GitHub 或第三方 CDN。资源均来自工具官方仓库的公开品牌文件，下载后随前端构建发布。

- `openclaw-official.svg`：OpenClaw 官方仓库 `apps/linux/src-tauri/icons/icon.svg`，来源：`openclaw/openclaw`，MIT/项目许可证。
- `ccswitch-official.png`：CC Switch 官方仓库 `src-tauri/icons/128x128.png`，来源：`farion1231/cc-switch`，MIT。
- `deepseek-harness-official.svg`：DeepSeek Harness 官方仓库 `website/public/favicon.svg`，来源：`deepseek-ai/deepseek-harness`，MIT。
- `hermes-official.png`：Hermes Agent 官方仓库 `website/static/img/logo.png`，来源：`NousResearch/hermes-agent`，MIT。

Claude 使用项目现有 `ModelIcon.vue` 中的 Claude/Anthropic 官方品牌路径。Codex 官方仓库没有单独发布可复用的图标文件，因此使用 OpenAI 官方标志并在界面中明确显示 `Codex` 名称，不把 OpenAI 标志冒充成独立 Codex Logo。
