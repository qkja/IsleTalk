# Implementation Plan: 仓库治理基线

**Feature Branch**: `feature/001-repo-governance-baseline`
**Spec**: `specs/001-repo-governance-baseline/spec.md`
**Created**: 2026-11-13

## Summary

为 IsleTalk 仓库补齐宪法 F 组「仓库卫生与协作入口」要求的必备资产（LICENSE、CONTRIBUTING.md、CODEOWNERS、SECURITY.md、`.github/`），并建立唯一的治理合规自动化检查点（CI 治理校验 + 密钥扫描）。同时把仓库的合并设置调整为与宪法 F 组「合并策略」一致（仅 squash、PR 标题作默认提交信息、合并后自动删分支）。

本 feature 不产生应用代码。其"实现"全部是仓库治理资产，正确性由治理校验脚本自身的反向测试覆盖。

## Technical Context

**Language**: Python 3（仅 CI 校验脚本；脚本运行在 ubuntu runner，仓库本地 Windows 亦可用 `python` 执行）
**Infrastructure**: GitHub Actions（ubuntu-latest）
**External Dependencies**: GitHub REST API（仓库设置）、gitleaks（密钥扫描）
**Repository**: github.com/qkja/IsleTalk（私有）

## Constitution Check

| 宪法条款 | 本次如何满足 |
|---|---|
| F 组「仓库卫生与协作入口」 | 补齐 LICENSE / CONTRIBUTING.md / CODEOWNERS / SECURITY.md / .gitignore / .gitattributes / CI 配置 |
| F 组「保护规则」 | 仓库设置对齐（仅 squash、合并后自动删分支）；平台级分支保护受套餐限制，已在 spec 的 Assumptions 中显式记录为待解禁项 |
| F 组「合并策略」 | 关闭 merge commit 与 rebase merge，开启 PR 标题作默认提交信息 |
| F 组「标签与发布」 | 仓库设置不涉及 tag；标签保护待套餐解锁 |
| II 组「测试覆盖率 ≥ 80%」 | 本 feature 无业务代码；治理校验脚本的正确性以反向测试覆盖（注入违规 → 期望失败 → 移除 → 期望通过） |
| E 组「密钥扫描」 | GitHub secret scanning 受套餐限制，改用 gitleaks 在 CI 中自行实现 |
| 原则 IV「规格驱动开发」 | 本 feature 自身走 specify → plan → tasks → implement → analyze，产物在 `specs/001-repo-governance-baseline/` |

**门禁关联**：本 feature 完成后，「产物门禁」「流程门禁」「分支管理门禁」应全部为绿；「语言选型门禁」「数据与中间件门禁」「生产就绪门禁」因尚无服务而不适用。

## Project Structure

### Documentation (this feature)

```
specs/001-repo-governance-baseline/
├── spec.md          功能规格与验收标准
├── plan.md          本文件
└── tasks.md         依赖有序的任务清单
```

### Root Repository Governance Assets

```
LICENSE                              Apache-2.0
CONTRIBUTING.md                      贡献与流程说明
CODEOWNERS                           路径归属规则
SECURITY.md                          漏洞报告与响应流程
CLAUDE.md                            Claude Code 工作指引
.github/
├── pull_request_template.md         PR 模板与合并前自查清单
├── ISSUE_TEMPLATE/
│   ├── bug_report.md
│   ├── feature_request.md
│   └── config.yml
├── scripts/
│   └── check-governance.py          治理合规校验器
└── workflows/
    └── ci.yml                       CI：治理校验 + 密钥扫描 + 语言门禁占位
```

## Complexity Tracking

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|------------------------------------|
| 引入 Python 校验脚本而非纯 shell | 需解析宪法 Markdown（版本行、章节、占位符）并遍历 specs/ 目录做依赖关系判断 | 纯 shell 处理多行正则与目录遍历易错，且 Windows 本地开发者无法直接复用；Python 3 在 ubuntu runner 与 Windows 上均可用 |
| 用 gitleaks 自行扫描而非 GitHub secret scanning | 私有仓库 + Free 套餐下 GitHub secret scanning 不可用 | 等待套餐升级会让密钥扫描在可预见的时间内完全缺失 |
| 在 CI 中预留语言门禁占位而非留空 | 避免后续误以为门禁已覆盖语言相关要求 | 纯注释容易被忽略；显式占位让缺口可见 |
