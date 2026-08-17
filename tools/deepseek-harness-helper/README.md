# DeepSeek Harness Helper

独立的跨平台 Go Helper，用于处理 Sub2API 发起的 `sub2api-harness://bootstrap` 安装请求。Helper 不安装 Node.js；本机必须预先安装 Node.js `>= 22.19.0` 和 npm。

## 行为

1. 严格解析：

   ```text
   sub2api-harness://bootstrap?server=<origin>&ticket=<opaque>&operation_id=<opaque>
   ```

   `server` 只接受 HTTPS，或 `localhost` / loopback IP 的 HTTP origin。禁止 userinfo、额外路径、query 和 fragment。

2. 在访问服务器前显示系统原生确认框，明确展示规范化后的 origin。用户批准后写入私有 `trusted-sites.json`；同一 origin 后续不重复询问，其他站点必须单独批准。
3. `POST <server>/api/v1/deepseek-harness/exchange`，body 为 `{"ticket":"..."}`，读取标准 `{"data": ...}` envelope。
4. 使用 exchange 返回的 `status_url` 和 `Bearer <event_token>` 依次上报：
   `checking_environment`、`installing`、`configuring`、`starting`、`completed`；失败上报 `failed`。
5. 在 Helper 私有用户数据目录执行固定版本安装：

   ```text
   npm install --prefix <runtime> --no-audit --no-fund --save-exact @deepseek-ai/dsh@<dsh_version>
   ```

   不通过 shell 执行，`dsh_version` 只接受固定版本 `0.1.0-rc.6`。
6. 在私有 `DSH_HOME` 串行更新：
   - `settings.yaml`: 仅清理 Helper 明确定义的六个受管 route 及 legacy `sub2api`，保留其他同前缀用户 provider，再写入当前 `llm-pi-ai.providers[route]` 和 `agent-default-model`
   - `.credentials.yaml`: 固定写入 `SUB2API_API_KEY: api_key`
   - 非 Sub2API provider 已占用该凭据名时拒绝覆盖
   - 两个文件使用与 DSH rc.6 相同的 `<文件>.lock` 写锁和原子替换，覆盖前更新 `.bak`；配置成功后不做整文件反向覆盖，后续状态上报失败只记录警告，避免覆盖用户或 DSH 的新修改
7. 仅启动或复用 Helper 自己记录的 DSH 进程：

   ```text
   node <private-runtime>/node_modules/@deepseek-ai/dsh/lib/bin.js \
     --profile web --host 127.0.0.1 --port 0
   ```

   DSH 的 stdout/stderr 直接追加到私有日志文件，Helper 只从本次启动后的新增日志中解析完整匹配 `dsh web: http://127.0.0.1:<port>` 的行，避免 Helper 退出后子进程失去 pipe 读端。因为凭据引用固定为 `SUB2API_API_KEY`，写入新 endpoint 和新 key 前会先停止状态文件中可验证的 Helper-managed DSH，避免热加载产生“新 endpoint + 旧 key”的组合；状态文件损坏、runtime 不匹配，或 PID 的进程创建时间/可执行文件身份不匹配时直接中止，避免 PID 复用误杀其他进程。两个文件完成后再启动，并同时验证进程身份与包含 Harness 标识的 loopback HTTP。终态事件可严格幂等重试；本地 Harness 已成功启动后，无论状态响应明确失败还是网络/5xx 重试后仍无法确认，Helper 都保留配置与进程并输出警告。

## 数据目录

由 `os.UserConfigDir()` 决定：

- Windows: `%AppData%\\Sub2API\\DeepSeekHarnessHelper`
- Linux: `${XDG_CONFIG_HOME:-$HOME/.config}/sub2api/deepseek-harness-helper`
- macOS: `$HOME/Library/Application Support/sub2api/deepseek-harness-helper`

其中包含私有 npm runtime、`dsh-home`、`trusted-sites.json`、`process.json` 和 `dsh.log`。Unix 目录权限收紧为 `0700`、配置文件为 `0600`；Windows 使用 `icacls.exe` 移除继承，并仅授予当前用户和 SYSTEM 完全控制。

## 构建与测试

需要 Go 1.22+：

```bash
go mod download
go test ./...
go vet ./...
```

PowerShell 构建所有目标：

```powershell
./scripts/build.ps1 -Version 0.1.0
```

POSIX shell 构建：

```bash
VERSION=0.1.0 ./scripts/build.sh
```

输出在 `dist/`：Windows amd64/arm64 自安装 `.exe`、Linux amd64/arm64 `.tar.gz`、macOS amd64/arm64 `.app.tar.gz`，以及 `SHA256SUMS`。仓库工作流 `.github/workflows/deepseek-harness-helper-build.yml` 只生成并保留未签名 CI artifact，不会直接发布；受保护的发布流水线完成 Windows 签名以及 macOS codesign/notarize 后，才能上传到 `dsh-helper-v*` Release。

## 协议注册

### Windows

```powershell
.\deepseek-harness-helper-windows-amd64.exe
```

无参数运行会先复制到当前用户配置目录的稳定 `bin` 路径，再写入 `HKCU\\Software\\Classes\\sub2api-harness`。command 使用完整且带引号的 executable 路径和单个 `%1` URI 参数，不需要管理员权限；`register-protocol` 提供相同行为。

### Linux

```bash
./deepseek-harness-helper
```

无参数运行会复制到当前用户配置目录的稳定 `bin` 路径，写入 `${XDG_DATA_HOME:-$HOME/.local/share}/applications/deepseek-harness-helper.desktop`，然后执行：

```text
xdg-mime default deepseek-harness-helper.desktop x-scheme-handler/sub2api-harness
```

因此桌面环境必须提供 `xdg-mime`。

### macOS

可分发协议处理器必须是 `.app`。每个 macOS `.tar.gz` 中包含：

```text
DeepSeek Harness Helper.app/
  Contents/Info.plist
  Contents/MacOS/deepseek-harness-helper
```

`Info.plist` 已声明 `CFBundleURLTypes`。发布方必须 codesign、notarize，并让用户启动一次 app 以由 LaunchServices 注册。仓库不包含签名身份或 notarization 凭据。

## CLI

```text
deepseek-harness-helper
deepseek-harness-helper <sub2api-harness://bootstrap?...>
deepseek-harness-helper register-protocol
deepseek-harness-helper --version
```

成功完成后打印并尝试打开 loopback Harness URL。浏览器打开命令也使用参数数组，不经过 shell。

## 安全和限制

- Node 缺失、npm 缺失或 Node 低于 `22.19.0` 时立即失败并上报，不自动安装或升级 Node。
- 未经本机用户确认的 origin 不会收到任何网络请求。确认框显示精确 origin；只应批准自己信任的 Sub2API 站点。
- Exchange 不携带长期认证；ticket 应由服务端保持短时、单次使用。API key 只写入私有 credentials 文件，不进入命令参数或 Helper 日志。
- 固定凭据槽意味着当前 DSH_HOME 只保留一个 Sub2API-managed provider；再次从其他 Key/分组安装会替换上一次的 `sub2api-*` provider 和默认模型。
- `status_url` 必须与 exchange server 同 origin，且 path 必须精确对应 `operation_id`，防止 event token 被转发到其他主机。
- Provider `base_url` 要求 HTTPS；仅本机开发允许 loopback HTTP。
- npm registry 与 TLS 仍属于本机 npm 配置的信任边界；生产发布应同时固定 lockfile、校验发布 artifact，并考虑专用 registry policy。
- YAML 使用 AST 合并并保留未知 namespace 和未替换节点；被替换的目标 provider 和 `agent-default-model` 节点的原注释/格式不会保留。
- Unix 通过 process group 终止启动失败的进程；Windows 使用独立 process group，但超时清理最终调用 `Kill`，不会接管或终止状态文件之外的 PID。
- 除站点信任确认框外当前没有独立进度 GUI；协议处理完成后打开浏览器。并发 bootstrap 会通过全局 bootstrap 锁、配置文件锁和启动锁串行处理。
- 仓库不包含 Windows 代码签名证书或 macOS Developer ID/notarization 凭据；正式发布前必须在发布流水线接入签名，签名前应保持站点 Feature Flag 关闭。
