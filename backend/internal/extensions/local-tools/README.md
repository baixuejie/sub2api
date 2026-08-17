# 本地工具后端注册入口

该目录是核心 Router 与各本地工具 Extension 之间唯一的组装入口。`server/router.go` 只调用一次 `localtools.RegisterRoutes`，不直接依赖 DeepSeek Harness、Hermes 或 OpenClaw。

新增工具时：

1. 在 `backend/internal/extensions/<tool-name>/` 实现独立路由、会话、一次性票据、兑换和状态上报。
2. 仅在本目录的 `RegisterRoutes` 中追加该 Extension 的注册调用。
3. 不把工具业务逻辑、任务 payload 或安装命令写入本目录和核心 Router。
