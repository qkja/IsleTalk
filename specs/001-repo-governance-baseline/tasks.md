# Tasks: 仓库治理基线

**Input**: Design documents from `/specs/001-repo-governance-baseline/`
**Prerequisites**: plan.md, spec.md
**Tests**: 本 feature 无业务代码；治理校验脚本的正确性以反向测试覆盖（T009）

**Format**: `[ID] [P?] [Story] Description`

## Phase 1: Setup

**Purpose**: 建立仓库治理资产与 Spec Kit 产物骨架

- [x] T001 建 feature 目录与 spec.md（`create-new-feature.ps1` 生成后填充）→ `specs/001-repo-governance-baseline/spec.md`
- [x] T002 写实现计划 → `specs/001-repo-governance-baseline/plan.md`
- [x] T003 写本任务清单 → `specs/001-repo-governance-baseline/tasks.md`
- [x] T004 [US1] 取 Apache-2.0 许可证全文 → `LICENSE`
- [x] T005 [US1] 写贡献指南（流程、分支模型、提交与 PR 规范、合并前自查）→ `CONTRIBUTING.md`
- [x] T006 [US1] 写归属规则（默认归属 + 按服务细化示例）→ `CODEOWNERS`
- [x] T007 [US1] 写漏洞报告策略（渠道、分级时限、响应承诺、处置流程）→ `SECURITY.md`

## Phase 2: 自动化门禁

**Purpose**: 让治理要求可被机器检查，而非依赖人工记忆

- [x] T008 [US2] 写治理校验器：宪法无临时块、版本行合规、章节齐全、无占位符、必备文件存在、specs 产物依赖关系成立 → `.github/scripts/check-governance.py`
- [x] T009 [US2] 反向测试校验器：注入 Sync Impact Report 应失败（退出码 1），移除后应通过（退出码 0）
- [x] T010 [US2] 写 CI 工作流：治理校验作业 + gitleaks 密钥扫描作业 + 语言门禁占位 → `.github/workflows/ci.yml`
- [x] T011 [US2] 清除宪法顶部的 Sync Impact Report 临时块（治理条款要求提交前删除）

## Phase 3: 协作入口

**Purpose**: 让提交与评审有统一格式

- [x] T012 [US3] 写 PR 模板（变更意图、影响面、验证方式、合并前自查清单、破坏性变更）→ `.github/pull_request_template.md`
- [x] T013 [US3] 写 Issue 模板（缺陷、需求、禁止空白 issue 并链接安全渠道）→ `.github/ISSUE_TEMPLATE/bug_report.md`、`.github/ISSUE_TEMPLATE/feature_request.md`、`.github/ISSUE_TEMPLATE/config.yml`

## Phase 4: 仓库设置与清理

**Purpose**: 让仓库行为与宪法 F 组一致，并清除残留

- [x] T014 [US4] 关闭 merge commit 与 rebase merge，开启 PR 标题作默认提交信息，开启合并后自动删分支
- [x] T015 清除根目录工具抓取残留文件 `md_db.md`、`md_gen.md`、`md_rpc.md`
- [x] T016 [US4] 写 Claude Code 工作指引 → `CLAUDE.md`
- [x] T017 宪法补充说明：spec 目录名与 git 分支名相互独立，由 `.specify/feature.json` 关联（版本 2.5.1）

## Phase 5: 合入

**Purpose**: 走完标准流程并合入集成分支

- [x] T018 运行 `/speckit-analyze` 复核 spec / plan / tasks 一致性
- [x] T019 推送分支并开 PR 到 `develop`，squash 合并
