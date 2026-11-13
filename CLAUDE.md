# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目现状

IsleTalk 目前**只有治理与流程基础设施，没有应用代码**。仓库里是 Spec Kit 工具链（`.specify/`）、Claude Code 侧的技能入口（`.claude/skills/`）与项目宪法。

因此本仓库**没有 build / lint / test 命令**——实现语言尚未确定（多语言栈，按服务逐个定）。不要臆造 `npm test`、`go build` 之类的命令；先读宪法 A 组「语言」小节与对应 feature 的 `plan.md`。

`README.md` 已有项目概览与原则摘要，与本文件互补，不重复。

## 唯一入口：Spec Kit 流程

宪法原则 IV 规定所有需求 MUST 严格走以下流程，禁止跳步。**这是本仓库与其他项目最大的不同——不要绕过它直接写代码。**

```
/speckit-specify    → spec.md    功能规格与验收标准
/speckit-clarify    → 回写 spec   消除歧义（规格有歧义时 MUST 执行）
/speckit-plan       → plan.md    技术方案、契约、测试策略
/speckit-tasks      → tasks.md   依赖有序、可验证的任务清单
/speckit-implement  → 实现        按 tasks 顺序，业务逻辑先写失败测试
/speckit-analyze    → 复核        spec/plan/tasks 一致性，合并前必跑
```

辅助命令：`/speckit-checklist`（生成验收清单）、`/speckit-converge`（补齐未实现任务）、`/speckit-taskstoissues`（任务转 GitHub Issue）、`/speckit-constitution`（修订宪法）。

这些命令是 `.claude/skills/speckit-*/SKILL.md` 定义的技能，不是可执行文件。

### 产物位置

feature 产物统一在 `specs/{编号}-{slug}/`（如 `specs/001-repo-governance-baseline/`）：

| 文件 | 阶段 | 来源 |
|---|---|---|
| `spec.md` | specify | 必需，含 Given/When/Then 验收标准与范围边界 |
| `plan.md` | plan | 必需，含 Constitution Check 章节 |
| `tasks.md` | tasks | implement 必需 |
| `research.md` / `data-model.md` / `quickstart.md` / `contracts/` | plan | 可选，存在即被识别 |

spec 目录名与 git 分支名**相互独立**，由 `.specify/feature.json`（本机状态，不入库）关联，两者的编号必须一致。分支用 `feature/{编号}-{slug}`，spec 目录用 `{编号}-{slug}`。

### 底层脚本

技能内部调用 `.specify/scripts/powershell/` 下的 PowerShell 脚本。需要手工调用时用：

```bash
powershell -NoProfile -ExecutionPolicy Bypass -File .specify/scripts/powershell/check-prerequisites.ps1 -Json
powershell -NoProfile -ExecutionPolicy Bypass -File .specify/scripts/powershell/create-new-feature.ps1 -Json "<feature 描述>"
powershell -NoProfile -ExecutionPolicy Bypass -File .specify/scripts/powershell/setup-plan.ps1 -Json
powershell -NoProfile -ExecutionPolicy Bypass -File .specify/scripts/powershell/setup-tasks.ps1 -Json
powershell -NoProfile -ExecutionPolicy Bypass -File .specify/scripts/powershell/resolve-template.ps1 constitution-template -Json
```

`check-prerequisites.ps1` 的 `-RequireSpec` / `-RequireTasks` 开关决定校验哪份产物；`-Json` 输出 `FEATURE_DIR` 与 `AVAILABLE_DOCS`，路径是绝对路径。

## 宪法：最高规则来源

`.specify/memory/constitution.md`（当前 v2.5.0）与任何其他文档冲突时以宪法为准。**动手前先读它**——尤其对应 feature 涉及的组。

五条核心原则：

| 原则 | 一句话 |
|---|---|
| I. 代码质量优先 | 函数 ≤50 行、文件 ≤800 行；错误显式处理；输入边界校验；密钥走环境变量 |
| II. 测试覆盖率 ≥ 80% | CI 硬门禁；关键路径接近 100% 分支覆盖；禁止删测试刷数字 |
| III. API 返回格式统一 | 统一信封 `{success, data, message, meta}`；错误码唯一注册表；列表必须分页 |
| IV. 规格驱动开发 | 流程不可协商，禁止跳步 |
| V. 标准分层业务架构 | 微服务 + 服务内 DDD 四层；gRPC/proto 契约；服务独占数据 |

附加约束分 A–G 七组：A 技术栈与依赖选型、B API 与数据契约治理、C 质量测试与性能、D 可靠性与可观测性、E 安全隐私与供应链、F 交付分支与协作、G 用户体验与运营支持。另有 A–G 组各自的具体阈值。

### 修订宪法的规矩

只有 `/speckit-constitution` 能改宪法，版本按语义化：

- **MAJOR** —— 删除或重定义原则、破坏性治理变更（含服务形态、服务间协议、数据边界的改动）
- **MINOR** —— 新增原则/章节、对既有约束的实质性扩展
- **PATCH** —— 措辞澄清、非语义重构

修订后文件顶部会有一段 **Sync Impact Report** HTML 注释，那是临时审阅材料，**提交前必须删除**。

宪法里的 `TODO(TECH_STACK)` 是有意保留的待办清单：每确定一门语言或一项关键组件，须登记其工具链细则并以 MINOR 版本修订宪法。

## 已定的架构决策

不要重新发明，也不要在 `plan.md` 里违背（要偏离就得走豁免流程）：

- **微服务**，每个服务内部 DDD 四层 `api → application → domain → infrastructure`，依赖单向向内，`domain` 层零框架依赖
- **服务间 gRPC** 同步为主，契约以 `.proto` 定义、版本化、带契约测试；异步事件走专用消息队列
- **数据服务独占**（database per service），禁止跨服务直连他方库表；集合名带服务前缀
- **MongoDB** 默认持久化（例外须在 `plan.md` 论证）；**Redis** 做 cache-aside 缓存、分布式锁、限流、会话，不得作唯一事实来源
- **多语言栈**：每个服务可选最适配语言，服务内部统一；选型须在 `plan.md` 论证并经评审
- **本地依赖**用 Docker Compose 编排，镜像版本固定，不得用 `latest`

## 分支模型

Git Flow + squash merge（宪法 F 组有完整细则）：

- `main` 生产分支（只收 release 与 hotfix）、`develop` 集成分支、`feature/` `release/` `hotfix/`
- 特性分支从最新 `develop` 切；hotfix 从 `main` 的发布 tag 切，**合入 main 后必须同步回 develop**
- 分支名 `{type}/{编号}-{kebab-case-slug}`；编号须能对应 `specs/` 下的 feature 目录
- 合入 develop 与 main 一律 squash；main 与 develop 必须开分支保护（CI 全绿 + 禁直接推送 + 禁 force push + 删除保护 + 必须同步 + 线性历史）
- 已合并分支立即删除；特性分支 ≤3 天

> **评审人数的限时例外**：GitHub 禁止作者批准自己的 PR，单一维护者仓库无法满足"≥1 名非作者评审"。
> 期间分支保护的 `required_approving_review_count` 暂设为 0，其余保护全部生效。
> 引入第二名评审者后必须恢复为 ≥1 名，并修订宪法（见宪法 F 组「协作与知识」的例外条款）。

## 环境注意事项

- **Windows + PowerShell**：所有 Spec Kit 脚本是 `.ps1`（需 pwsh/PowerShell），调用时加 `-NoProfile -ExecutionPolicy Bypass`
- **行尾**：`.gitattributes` 规定 `* text=auto eol=lf`，但 `*.ps1` / `*.bat` 用 CRLF。新增文件遵循该规则，别手动改 CRLF
- **GitHub**：`gh` CLI 可用；`/speckit-taskstoissues` 依赖它。仓库为私有
- **本地 Docker**：MongoDB、Redis、消息中间件的镜像本机已有，起本地依赖用 Docker Compose

## 不要做的事

- 不要跳过 Spec Kit 流程直接实现功能——这是宪法原则 IV 的硬性要求
- 不要在 `main` / `develop` 上直接提交，不要 force push 共享分支
- 不要硬编码密钥、凭据、生产连接串
- 不要把 `md_db.md` / `md_gen.md` / `md_rpc.md` 提交进去——它们是工具抓取时残留的 404 页面垃圾文件，应删除
- 不要在没有 `spec.md` / `plan.md` / `tasks.md` 的情况下声称某个功能"已完成"
