# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目现状

IsleTalk 目前**只有治理与流程基础设施，没有应用代码**。仓库里是 Spec Kit 工具链（`.specify/`）、Claude Code 侧的技能入口（`.claude/skills/`）、项目宪法，以及一套仓库自身的治理资产（CI、模板、校验器）。

因此本仓库**没有 build / lint / test 命令**——实现语言尚未确定（多语言栈，按服务逐个定）。不要臆造 `npm test`、`go build` 之类的命令；先读宪法 A 组「语言」小节与对应 feature 的 `plan.md`。

`README.md` 已有项目概览、原则摘要与技术架构决策表，与本文件互补，不重复。

## 常用命令

**治理校验** —— 本仓库唯一的"测试"，也是 CI 的第一个作业：

```bash
python .github/scripts/check-governance.py
```

Windows 上必须用 `python`（本机 3.13.7），**不能用 `python3`**——那是指向 Microsoft Store 的应用别名，执行脚本会静默退出 49 且无输出。CI 跑在 ubuntu 上用 `python3` 没问题。

Windows 控制台默认 GBK，中文输出会变乱码，需要时加 `PYTHONIOENCODING=utf-8`。

校验内容：宪法无 Sync Impact Report 残留、版本行格式合规、五条原则与 A–G 组齐全、无未替换的 `[占位符]`、仓库必备文件存在、`specs/` 下 spec/plan/tasks 的依赖关系成立。退出码 0 = 通过，1 = 逐条列出问题。

**CI**（`.github/workflows/ci.yml`）两个作业：`治理校验`、`密钥扫描`（gitleaks）。触发于 main 与 develop 的推送及 PR。语言相关门禁（覆盖率 ≥80%、依赖方向检查、SAST/SCA、契约破坏检测）在工作流里留了占位，待首个服务定语言后补入。

**Spec Kit 底层脚本**（技能内部调用，需手工执行时）：

```bash
powershell -NoProfile -ExecutionPolicy Bypass -File .specify/scripts/powershell/check-prerequisites.ps1 -Json
powershell -NoProfile -ExecutionPolicy Bypass -File .specify/scripts/powershell/create-new-feature.ps1 -Json -ShortName "<slug>" -Number <n> "<描述>"
powershell -NoProfile -ExecutionPolicy Bypass -File .specify/scripts/powershell/setup-plan.ps1 -Json
powershell -NoProfile -ExecutionPolicy Bypass -File .specify/scripts/powershell/setup-tasks.ps1 -Json
powershell -NoProfile -ExecutionPolicy Bypass -File .specify/scripts/powershell/resolve-template.ps1 constitution-template -Json
```

`check-prerequisites.ps1` 的 `-RequireSpec` / `-RequireTasks` 决定校验哪份产物；`-Json` 输出 `FEATURE_DIR` 与 `AVAILABLE_DOCS`（绝对路径）。

**发布与仓库操作**：`gh` CLI 已认证。常用：`gh pr create --base develop --head <分支>`、`gh pr merge <n> --squash --subject "..." --body "..."`、`gh pr list --author "app/dependabot"`、`gh run list`。

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

辅助命令：`/speckit-checklist`、`/speckit-converge`、`/speckit-taskstoissues`、`/speckit-constitution`。这些是 `.claude/skills/speckit-*/SKILL.md` 定义的技能，不是可执行文件。

### 产物位置

feature 产物统一在 `specs/{编号}-{slug}/`（如 `specs/001-repo-governance-baseline/`）：

| 文件 | 阶段 | 必需性 |
|---|---|---|
| `spec.md` | specify | 必需，含 Given/When/Then 验收标准与范围边界 |
| `plan.md` | plan | 必需，含 Constitution Check 章节 |
| `tasks.md` | tasks | implement 必需 |
| `research.md` / `data-model.md` / `quickstart.md` / `contracts/` | plan | 可选，存在即被识别 |

spec 目录名与 git 分支名**相互独立**，由 `.specify/feature.json`（本机状态，不入库）关联，两者的编号必须一致。分支用 `feature/{编号}-{slug}`，spec 目录用 `{编号}-{slug}`。

## 宪法：最高规则来源

`.specify/memory/constitution.md`（当前 v2.6.0）与任何其他文档冲突时以宪法为准。**动手前先读它**——尤其对应 feature 涉及的组。

| 原则 | 一句话 |
|---|---|
| I. 代码质量优先 | 函数 ≤50 行、文件 ≤800 行；错误显式处理；输入边界校验；密钥走环境变量 |
| II. 测试覆盖率 ≥ 80% | CI 硬门禁；关键路径接近 100% 分支覆盖；禁止删测试刷数字 |
| III. API 返回格式统一 | 统一信封 `{success, data, message, meta}`；错误码唯一注册表；列表必须分页 |
| IV. 规格驱动开发 | 流程不可协商，禁止跳步 |
| V. 标准分层业务架构 | 微服务 + 服务内 DDD 四层；gRPC/proto 契约；服务独占数据 |

附加约束分 A–G 七组：A 技术栈与依赖选型、B API 与数据契约治理、C 质量测试与性能、D 可靠性与可观测性、E 安全隐私与供应链、F 交付分支与协作、G 用户体验与运营支持，各组内含具体阈值。

### 修订宪法的规矩

只有 `/speckit-constitution` 能改宪法，版本按语义化：

- **MAJOR** —— 删除或重定义原则、破坏性治理变更（含服务形态、服务间协议、数据边界的改动）
- **MINOR** —— 新增原则/章节、对既有约束的实质性扩展
- **PATCH** —— 措辞澄清、非语义重构

修订后文件顶部会有一段 **Sync Impact Report** HTML 注释，那是临时审阅材料，**提交前必须删除**（治理校验器会检查）。

宪法里的 `TODO(TECH_STACK)` 是有意保留的待办清单：每确定一门语言或一项关键组件，须登记其工具链细则并以 MINOR 版本修订宪法。

## 已定的架构决策

不要重新发明，也不要在 `plan.md` 里违背（要偏离就得走豁免流程）：

- **微服务**，每个服务内部 DDD 四层 `api → application → domain → infrastructure`，依赖单向向内，`domain` 层零框架依赖
- **服务间 gRPC** 同步为主，契约以 `.proto` 定义、版本化、带契约测试；异步事件走专用消息队列
- **数据服务独占**（database per service），禁止跨服务直连他方库表；集合名带服务前缀
- **MongoDB** 默认持久化（例外须在 `plan.md` 论证）；**Redis** 做 cache-aside 缓存、分布式锁、限流、会话，不得作唯一事实来源
- **多语言栈**：每个服务可选最适配语言，服务内部统一；选型须在 `plan.md` 论证并经评审
- **本地依赖**用 Docker Compose 编排，镜像版本固定，不得用 `latest`

## 分支与发布

Git Flow + squash merge（宪法 F 组有完整细则）：

- `main` 生产分支（只收 release 与 hotfix）、`develop` 集成分支、`feature/` `release/{版本}` `hotfix/`
- 特性分支从最新 `develop` 切；hotfix 从 `main` 的发布 tag 切，**合入 main 后必须同步回 develop**
- 分支名 `{type}/{编号}-{kebab-case-slug}`；编号须能对应 `specs/` 下的 feature 目录
- 合入 develop 与 main 一律 squash；main 与 develop 必须开分支保护
- 已合并分支立即删除（仓库开了 `delete_branch_on_merge`）；特性分支 ≤3 天

**发布流程**（当前 `v0.1.0` 已发布，main 精确等于该标签）：

```
release/{版本}  ← 从 develop 切
   ├─ 补 CHANGELOG.md 新版本段（Keep a Changelog + SemVer）
   ├─ PR → main，squash 合并，提交信息用 Conventional Commits（如 chore(release): v0.2.0）
   ├─ 打 annotated tag v{MAJOR}.{MINOR}.{PATCH}，指向 main 的 squash 提交
   └─ PR → develop 回合内容，然后删除 release 分支
```

`main` MUST 始终精确等于最新发布标签。**依赖升级 MUST NOT 直连 main**（`.github/dependabot.yml` 已设 `target-branch: develop`），否则会绕过 release 流程。

## 平台保护现状

仓库为**公开**（github.com/qkja/IsleTalk）。已启用：

- `main` / `develop` 分支保护：CI 必须全绿、必须同步、禁直接推送、禁 force push、删除保护、线性历史、管理员同样受限
- 仓库规则集：`v*` 标签禁删除、禁改写
- GitHub 内置 secret scanning 与 push protection、Dependabot 漏洞告警与安全更新
- 合并设置：仅 squash、PR 标题作默认提交信息、合并后自动删分支

> **评审人数的限时例外**：GitHub 禁止 PR 作者批准自己的 PR，单一维护者仓库无法满足"≥1 名非作者评审"。
> 期间 `required_approving_review_count` 暂设为 0，其余保护全部生效。
> 引入第二名评审者后必须恢复为 ≥1 名并修订宪法（见宪法 F 组「协作与知识」的例外条款）。

## 环境注意事项

- **Windows + PowerShell**：Spec Kit 脚本全是 `.ps1`，调用时加 `-NoProfile -ExecutionPolicy Bypass`
- **行尾**：`.gitattributes` 规定 `* text=auto eol=lf`，但 `*.ps1` / `*.bat` 用 CRLF
- **Python**：本机用 `python` 而非 `python3`（见上）；CI 上用 `python3`
- **GitHub**：`gh` CLI 已认证（账号 qkja），仓库为公开；`/speckit-taskstoissues` 依赖它
- **本地 Docker**：MongoDB、Redis、消息中间件的镜像本机已有，起本地依赖用 Docker Compose

## 不要做的事

- 不要跳过 Spec Kit 流程直接实现功能——这是宪法原则 IV 的硬性要求
- 不要在 `main` / `develop` 上直接提交，不要 force push 共享分支
- 不要让依赖更新直连 `main`（必须经 `develop` → `release`）
- 不要硬编码密钥、凭据、生产连接串
- 不要在没有 `spec.md` / `plan.md` / `tasks.md` 的情况下声称某个功能"已完成"
