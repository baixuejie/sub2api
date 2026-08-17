# 本地工具扩展

该目录统一承载 API Key 页面上的“本地工具”入口。核心页面只挂载 `LocalToolsAction` 并传入当前 API Key、公开站点设置和 Feature Flag。

## 当前工具

- CC Switch：使用官方 `ccswitch://` 协议导入配置，不经过 Sub2API Helper。
- DeepSeek Harness：通过后端短时票据和 Sub2API Local Tools Helper 完成本机安装、配置与启动。

## 新增工具

以 Hermes 或 OpenClaw 为例：

1. 在 `frontend/src/features/<tool-name>/` 增加工具自己的交互组件，并在 `localToolRegistry.ts` 注册描述、Feature Flag、禁用规则、组件和组件参数；不要修改 `KeysView.vue`，也不要把安装流程写入通用菜单组件。
2. 在 `backend/internal/extensions/<tool-name>/` 增加 profile、session、exchange 和 event 接口，并只通过 `backend/internal/extensions/local-tools/` 聚合注册。启动 URI 使用受限的 `extension_id=<tool-name>`，Helper 会访问同源 `/api/v1/<tool-name>/exchange` 和对应状态接口。
3. 后端任务返回 `protocol_version`、`tool_id`、`tool_version`、`minimum_helper_version` 和由 Adapter 独立解码的 `payload`，长期凭据只能在一次性票据兑换后返回给 Helper。
4. 首次接入新工具时，在 Helper 的 Adapter Registry 中显式注册 `extension_id -> tool_id` 绑定并发布新版 Helper；后续协议兼容的后端调整无需用户更新。新增底层能力时提高 `minimum_helper_version`。
5. Adapter 只能调用编译进 Helper 的固定 Go 逻辑。不得允许后端下发 Shell、PowerShell、任意可执行文件路径或参数数组。

每个新工具至少覆盖：Key 所有权与状态校验、票据单次消费、origin 与 Extension 级信任、`extension_id -> tool_id` 绑定、同源 URL 校验、未知工具拒绝、最低 Helper 版本提示、失败状态回报和前端菜单权限。
