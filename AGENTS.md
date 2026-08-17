# Sub2API 定制开发规范

本文件是本仓库的长期开发约束。开始任何开发、代码审查或上游同步前，必须先阅读本文件。除非用户明确要求，本项目的说明文档、测试说明和交付总结使用中文。

## 1. 目标

- 保持上游 Sub2API 功能可以持续同步。
- 将定制功能集中在独立 Extension 模块中，避免把业务逻辑散落到上游文件。
- Git 只允许本地提交；本项目当前不向任何远程仓库推送代码或标签。

## 2. Extension 架构

### 2.1 目录约定

后端定制模块统一放在：

```text
backend/internal/extensions/<extension-name>/
  handler/       # HTTP 输入输出和参数校验
  service/       # 定制业务逻辑
  repository/    # 定制数据访问
  routes.go      # 路由注册
  wire.go        # Wire ProviderSet（需要依赖注入时）
  contract.go    # 与核心模块交互的接口（按需）
```

前端定制代码统一放在：

```text
frontend/src/features/<extension-name>/
  api/            # 定制 API 客户端
  components/     # 定制组件
  views/          # 定制页面
  types/          # 定制类型
```

目录名使用小写短横线或下划线，并使用稳定、唯一的 Extension 名称作为命名空间。

### 2.2 依赖方向

- Extension 可以依赖核心模块公开且稳定的接口。
- 核心业务代码不得反向依赖 Extension 的具体实现。
- 不得在核心 Handler、Service 或 Repository 中直接嵌入定制业务判断。
- 需要读取核心数据时，优先通过接口或适配器；不能让定制代码到处调用 Ent 细节。
- 需要修改核心事务（认证、调度、计费、额度扣减）时，先定义最小接口和明确的适配层。
- 只有一个 Extension 时，使用显式注册函数；确实有多个模块后再抽象通用 Extension Registry，不提前建设复杂插件系统。

### 2.3 允许的集成点

核心仓库只保留尽可能少的集成改动：

1. `backend/internal/server/router.go` 中增加一次 Extension 路由注册。
2. `backend/internal/handler/wire.go`、`backend/internal/service/wire.go` 或 `backend/internal/repository/wire.go` 中增加必要的 Provider。
3. 增加独立的配置项和 Feature Flag，默认关闭或保持原有行为。
4. 只有在现有自定义菜单不能满足需求时，才修改前端集中路由或侧边栏。

集成文件中只做组装和转发，不放定制业务逻辑。所有定制逻辑必须留在 `backend/internal/extensions/<extension-name>/` 或 `frontend/src/features/<extension-name>/`。

### 2.4 生成代码

- 不得手工编辑 `backend/cmd/server/wire_gen.go`。
- 不得手工编辑 `backend/ent` 下的生成文件。
- 修改 Wire Provider 或 Ent Schema 后，使用项目既有的生成命令重新生成，并将生成结果作为同一提交的一部分检查。
- 不得为了减少冲突而把生成文件和源文件拆成互相不一致的提交。

### 2.5 数据库和迁移

- Extension 优先使用自己的数据库或独立 PostgreSQL schema。
- 共享数据库时，表、索引和设置键必须使用 Extension 命名空间，避免与上游重名。
- 已执行的上游迁移和 Extension 迁移一律不可修改；所有变更通过新增迁移文件完成。
- 新迁移必须幂等，并明确事务/非事务执行要求。
- 不直接依赖上游表的未文档化字段；需要核心数据时增加稳定的接口或只读适配器。

### 2.6 前端和外部页面

- 原生 Vue 页面放在 `frontend/src/features/<extension-name>/`，避免散落到核心 `views` 和公共组件。
- 独立页面、报表和运营工具优先使用项目已有的 `custom_menu_items` 或反向代理接入，减少前端路由冲突。
- 不得把长期有效的 JWT 放进外部页面 URL。需要单点登录时，使用同源代理或一次性、短时效票据交换。
- 外部页面必须服务端校验身份和权限，不能只相信 URL 中的 `user_id`、角色或来源参数。
- 所有跨窗口通信必须校验明确的 `Origin`，并配置最小化的 `frame-ancestors` 白名单。

### 2.7 测试和质量

- Extension 的单元测试与实现放在同一命名空间，测试名称包含 Extension 名称。
- API、权限、幂等、错误处理和迁移至少有契约测试。
- 不得对整个上游目录执行会产生大范围改写的自动格式化或 lint 修复。
- 上游同步后至少运行后端单元测试；涉及数据库、路由或前端时，同时运行对应集成测试和类型检查。

## 3. Git 管理规范

### 3.1 分支职责

当前仓库约定如下：

```text
main                 # 只跟踪远程上游，不放定制代码
custom/main          # 本地定制发布分支
custom/feature/*     # 本地定制功能分支
```

`main` 只用于保存上游基线。定制代码必须从 `custom/main` 或其功能分支开发，不得直接提交到 `main`。

首次建立本地定制分支：

```bash
git switch main
git switch -c custom/main
```

### 3.2 远程仓库和推送边界

- 当前 `origin` 仅作为上游代码的只读来源使用。
- 允许：`git fetch`、`git pull`、查看远程分支和标签。
- 禁止：`git push`、`git push --tags`、强制推送、创建远程分支、创建 Pull Request。
- 用户没有明确授权时，不修改远程仓库 URL、远程分支或远程标签。
- 所有开发成果只提交到本地 Git 历史；交付时说明本地提交号。

### 3.3 上游同步流程

在工作区干净、没有未提交改动时执行：

```bash
git fetch origin --prune --tags

git switch main
git pull --ff-only origin main

git switch custom/main
git merge --no-ff main -m "chore: sync upstream"
```

合并时优先保留上游核心行为，把定制差异限制在 Extension 目录和少数集成点。解决冲突后重新生成 Wire/Ent 代码并运行测试。未经明确授权，不执行任何 push 命令。

建议启用本地冲突记忆：

```bash
git config rerere.enabled true
git config merge.renormalize true
```

### 3.4 提交规范

- 一个提交只解决一个主题。
- Extension 提交使用 `feat(ext-<name>): ...`、`fix(ext-<name>): ...` 或 `test(ext-<name>): ...` 前缀。
- 上游同步使用 `chore: sync upstream`，不得与定制功能混在同一提交。
- 不提交环境配置、密钥、数据库数据、构建产物或无关格式化变更。
- 提交前检查 `git status`、`git diff --check` 和待提交文件清单。

## 4. 上游同步后的验收

每次同步后至少确认：

1. `main` 仍能快进到 `origin/main`，且没有定制提交。
2. `custom/main` 只包含清晰可识别的 Extension 提交和同步提交。
3. 后端单元测试通过；涉及数据库时运行集成测试。
4. 涉及前端时运行 `pnpm` 类型检查和相关测试。
5. 迁移 checksum、Wire 生成结果和 API 契约没有异常。

## 5. 当前仓库注意事项

- `DEV_GUIDE.md` 中记录的 Fork 地址可能与当前 Git 远端配置不一致；以实际 `git remote -v` 为准，未经授权不要擅自改远端。
- 当前项目使用 Go、Gin、Ent、Wire、Vue 和 pnpm；新增方案应优先遵循现有工具链。
- 许可证、CLA、上游 AI 服务条款及项目对商业化的声明需要单独审阅，不能仅因项目公开托管就默认可以任意运营或分发。

## 6. 本地工具与 Helper 扩展规范

- API Key 页面上的本地工具统一由 `frontend/src/features/local-tools/` 提供下拉入口，核心页面只挂载一次通用入口并传入 Key 与公开设置；工具的描述、Feature Flag、禁用规则、动作组件和参数统一在 `localToolRegistry.ts` 注册，新增工具不得修改 `KeysView.vue`。
- CC Switch 继续使用其官方 `ccswitch://` 协议，不经过 Sub2API Helper。
- 需要本机安装、配置或启动的工具统一复用 Sub2API Helper 任务协议。任务必须包含 `protocol_version`、`tool_id`、`tool_version`、`minimum_helper_version` 和由 Adapter 独立解码的受限 `payload`，并保留已有协议所需的兼容字段；新增后端模块通过受限 `extension_id` 使用自己的同源兑换与状态接口。
- Helper 只能按编译期 Registry 中显式绑定的 `extension_id -> tool_id` 调用白名单 Adapter；站点信任也必须绑定 origin 与 `extension_id`，不得执行后端下发的任意 Shell、PowerShell、可执行文件路径或参数数组。
- 首次新增 Hermes、OpenClaw 等工具时必须注册 Extension 与 Adapter 的绑定并发布新版 Helper；后续仅调整已兼容的后端任务定义时可以复用已发布 Helper。新增底层安装能力或安全修复时必须提高 `minimum_helper_version`。
- 新工具的 UI、后端任务定义和 Helper adapter 分别保持独立命名空间，禁止把工具专用安装逻辑重新写回 API Key 核心页面或通用任务分派器。
- 本地工具后端路由统一通过 `backend/internal/extensions/local-tools/` 接入，核心 `server/router.go` 只保留一次聚合注册；新增工具不得直接向核心 Router 增加专用依赖。
