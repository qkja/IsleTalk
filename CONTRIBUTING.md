# Contributing to IsleTalk

本仓库的一切变更受 [项目宪法](../.specify/memory/constitution.md) 约束。**宪法与本文冲突时以宪法为准**；本文只补充操作性说明。

与普通项目最大的不同：IsleTalk 采用**规格驱动开发（SDD）**。所有功能开发 MUST 走 Spec Kit 流程，禁止直接写代码。

## 开始之前

1. 读一遍宪法，至少读与你改动相关的那几条原则和 A–G 组。
2. 确认你的改动落在哪个 feature。若还没有，用 `/speckit-specify` 建一个新的。
3. 本仓库目前**没有应用代码**，也没有 build / lint / test 命令——实现语言按服务逐个确定。

## 标准流程

```
/speckit-specify    → spec.md    功能规格与验收标准
/speckit-clarify    → 回写 spec   消除歧义（规格有歧义时 MUST 执行）
/speckit-plan       → plan.md    技术方案、契约、测试策略
/speckit-tasks      → tasks.md   依赖有序、可验证的任务清单
/speckit-implement  → 实现        按 tasks 顺序，业务逻辑先写失败测试
/speckit-analyze    → 复核        合并前必跑
```

产物在 `specs/{编号}-{slug}/`，例如 `specs/001-repo-governance-baseline/`。

> spec 目录名与 git 分支名**相互独立**，由 `.specify/feature.json`（本机状态，不入库）关联；两者的编号必须一致。

辅助命令：`/speckit-checklist`、`/speckit-converge`、`/speckit-taskstoissues`、`/speckit-constitution`。

## 分支模型

Git Flow + squash merge。

| 分支 | 用途 | 切出点 | 回合点 |
|---|---|---|---|
| `main` | 生产分支 | — | 只接受 `release/` 与 `hotfix/` |
| `develop` | 集成分支 | — | 接受所有 `feature/` |
| `feature/{编号}-{slug}` | 特性分支 | 最新 `develop` | `develop` |
| `release/{版本}` | 发布分支 | `develop` | `main` + 回合 `develop` |
| `hotfix/{编号}-{slug}` | 线上紧急修复 | `main` 的发布 tag | `main` **且必须同步回合 `develop`** |

规则：

- 分支名必须是 `{type}/{编号}-{kebab-case-slug}`，编号对应 `specs/` 下的 feature 目录。
- 特性分支短命：≤3 天或一个迭代；落后 `develop` 超过 3 天 MUST 先同步。
- 已合并分支 MUST 立即删除；超过 14 天无进展的分支 MUST 关闭或重新评估。
- 一次 PR 只做一个任务；一个提交只做一件事。

## 提交与 PR

- 提交信息遵循 **Conventional Commits**。因为合并用 squash，**PR 标题就是主干上的提交信息**，请直接按规范写标题。
- PR 必须引用任务或 issue 号。
- PR 描述必须说明：变更意图、影响面、验证方式。
- 合并方式一律 **squash**（仓库已关闭 merge commit 与 rebase merge）。
- 每个 PR 至少 1 名非作者评审通过；新增服务、数据库迁移、破坏性契约变更需 ≥2 名。

### 合并前自查

- [ ] `spec.md` / `plan.md` / `tasks.md` 齐备且与实现一致
- [ ] `/speckit-analyze` 已运行，发现项已处置
- [ ] 格式化、静态检查、类型检查、构建全绿
- [ ] 覆盖率 ≥ 80%
- [ ] 无 CRITICAL / HIGH 级评审问题
- [ ] 无硬编码密钥、无调试残留
- [ ] 新增对外接口有契约测试与文档
- [ ] 新增告警有 runbook，新增服务有 README
- [ ] 改动用户界面时通过可访问性检查（WCAG 2.2 AA）

## 本地开发

依赖（MongoDB / Redis / 消息中间件）用 Docker Compose 起，镜像版本固定，不用 `latest`。

Spec Kit 脚本是 PowerShell：

```bash
powershell -NoProfile -ExecutionPolicy Bypass -File .specify/scripts/powershell/check-prerequisites.ps1 -Json
```

行尾：仓库统一 LF，仅 `*.ps1` / `*.bat` 用 CRLF（见 `.gitattributes`）。

## 修订宪法

只有 `/speckit-constitution` 能改宪法。版本按语义化：破坏性治理变更 = MAJOR，新增约束 = MINOR，措辞澄清 = PATCH。

改完后文件顶部会有一段 **Sync Impact Report** HTML 注释——那是临时审阅材料，**提交前必须删除**，CI 会检查这一点。
