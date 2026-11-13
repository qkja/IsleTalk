# Feature Specification: 仓库治理基线

**Feature Branch**: `feature/001-repo-governance-baseline`

**Created**: 2026-11-13

**Status**: Draft

**Input**: User description: "仓库治理基线：补齐许可证、贡献指南、CODEOWNERS、安全策略、PR 模板与 CI"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 仓库必备文件齐备 (Priority: P1)

新加入 IsleTalk 的工程师或自动化工具打开仓库，能在根目录找到许可证、贡献指南、归属规则与安全策略，
不必靠口头约定或翻聊天记录。

**Why this priority**: 宪法 F 组「仓库卫生与协作入口」把这几项列为 MUST 必备；
缺任何一项，该组要求即处于不合规状态，且其他所有条款的落地方式都依赖它们。

**Independent Test**: 检查根目录存在 `LICENSE`、`CONTRIBUTING.md`、`CODEOWNERS`、`SECURITY.md`
四个文件且非空，即可验证，不依赖其他改动。

**Acceptance Scenarios**:

1. **Given** 仓库根目录，**When** 列出其中文件，**Then** 存在 LICENSE、CONTRIBUTING.md、CODEOWNERS、SECURITY.md
2. **Given** 任一必��文件，**When** 读其内容，**Then** 内容指向宪法并说明该文件在流程中的位置，不是占位符
3. **Given** CODEOWNERS，**When** 核对规则，**Then** 有覆盖全部路径的默认归属，且给出按服务细化的示例

---

### User Story 2 - 变更提交前自动校验治理合规 (Priority: P1)

维护者提交变更时，CI 自动检查宪法是否完整、临时材料是否已清除、规格产物依赖关系是否成立，
不合规时立刻阻断。

**Why this priority**: 宪法把「Sync Impact Report 提交前必须删除」等要求写成了 MUST，
但靠人工记忆执行必然遗漏。把它变成 CI 检查，是唯一能保证该条长期成立的执行方式。

**Independent Test**: 在本地运行治理校验脚本，人为在宪法中插入 Sync Impact Report 标记，
脚本应当以非零退出码报出该项；移除后应通过。

**Acceptance Scenarios**:

1. **Given** 完整且合规的仓库，**When** 运行治理校验，**Then** 退出码为 0 并输出通过信息
2. **Given** 宪法中残留 Sync Impact Report，**When** 运行治理校验，**Then** 退出码为 1 且明确指出该项
3. **Given** 宪法中残留 `[ALL_CAPS]` 占位符，**When** 运行治理校验，**Then** 退出码为 1 且列出占位符
4. **Given** 某 feature 目录有 tasks.md 但缺 plan.md，**When** 运行治理校验，**Then** 退出码为 1 且指出该目录
5. **Given** 宪法版本行格式不合规，**When** 运行治理校验，**Then** 退出码为 1 并给出正确格式

---

### User Story 3 - 提交与评审入口规范化 (Priority: P2)

开发者提 PR 时拿到带检查清单的模板，评审者按统一格式判断变更意图、影响面与验证方式。

**Why this priority**: 宪法 F 组要求 PR 描述必须说明变更意图、影响面与验证方式，
且每个 PR 至少一名非作者评审。模板把这三项固定下来，降低评审摩擦；优先级低于前两项，
因为它是提效项而非合规底线。

**Independent Test**: 新建 PR 时模板自动带出检查清单，且清单条目与合并门禁逐条对应。

**Acceptance Scenarios**:

1. **Given** 打开新建 PR 页面，**When** 编辑描述，**Then** 模板已包含变更意图、影响面、验证方式三段
2. **Given** 模板中的合并前自查清单，**When** 逐条比对，**Then** 覆盖宪法规定的各道合并门禁

---

### User Story 4 - 仓库合并策略与宪法一致 (Priority: P2)

仓库设置只允许 squash 合并，PR 标题默认作为提交信息，合并后自动删除分支，
使仓库行为与宪法 F 组一致，而不是靠纪律维持。

**Why this priority**: 宪法把合并方式、分支清理写成 MUST。若仓库设置允许 merge commit
或保留已合并分支，则宪法要求形同虚设。改设置成本极低、收益立竿见影。

**Independent Test**: 查询仓库设置，四个开关的取值与宪法要求一致；新建一个测试 PR 确认只有 squash 可选。

**Acceptance Scenarios**:

1. **Given** 仓库设置，**When** 查询合并选项，**Then** squash 允许，merge commit 与 rebase merge 关闭
2. **Given** 仓库设置，**When** 查询 squash 提交信息默认值，**Then** 为 PR 标题
3. **Given** 仓库设置，**When** 查询合并后行为，**Then** 已合并分支自动删除

---

### Edge Cases

- **CI 首次运行且无 feature 产物**：`specs/` 目录不存在时应跳过产物检查而不是报错——仓库当前正处于这个状态。
- **宪法尚未创建**：治理校验必须报出缺失而不是静默通过。
- **文件以 UTF-8 无 BOM 保存**：Windows 上 PowerShell 写文件可能带 BOM，读取时必须显式指定编码。
- **私有仓库的 GitHub 能力受限**：当前套餐下分支保护、规则集、Dependabot、secret scanning 均不可用，
  CI 需自行实现密钥扫描，且分支保护要求暂时无法由平台强制。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 仓库 MUST 包含 `LICENSE`、`CONTRIBUTING.md`、`CODEOWNERS`、`SECURITY.md` 四个文件且内容非空。
- **FR-002**: `CONTRIBUTING.md` MUST 说明 Spec Kit 标准流程、分支模型、提交与 PR 规范、合并前自查清单。
- **FR-003**: `CODEOWNERS` MUST 至少有一条覆盖全部路径的规则，并给出按服务细化的示例。
- **FR-004**: `SECURITY.md` MUST 给出私有漏洞报告渠道、分级与修复时限、响应承诺、处置流程。
- **FR-005**: 仓库 MUST 提供 `.github/pull_request_template.md`，包含变更意图、影响面、验证方式与合并前自查清单。
- **FR-006**: 仓库 MUST 提供 `.github/ISSUE_TEMPLATE/`（缺陷与需求模板）。
- **FR-007**: 仓库 MUST 提供 `.github/workflows/ci.yml`，至少包含治理校验与密钥扫描两个作业。
- **FR-008**: 治理校验 MUST 检查：宪法无 Sync Impact Report 残留、版本行格式合规、必需章节齐全、
  无未替换占位符、必备文件存在、`specs/` 下产物依赖关系成立。
- **FR-009**: 治理校验 MUST 在不合规时以非零退出码结束并逐条列出问题。
- **FR-010**: 仓库设置 MUST 满足：仅允许 squash 合并、squash 提交信息默认取 PR 标题、合并后自动删除分支。

### Key Entities

- **治理校验脚本** (`.github/scripts/check-governance.py`)：唯一判定治理合规的自动化执行者。
- **CI 作业** (`.github/workflows/ci.yml`)：PR 与主干推送时触发；语言相关门禁待技术栈确定后补入。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `LICENSE`、`CONTRIBUTING.md`、`CODEOWNERS`、`SECURITY.md` 四个文件全部存在且治理校验通过。
- **SC-002**: 治理校验在人为注入 Sync Impact Report 时以退出码 1 失败，移除后以退出码 0 通过。
- **SC-003**: 仓库设置的四个合并相关开关全部符合 FR-010。
- **SC-004**: 新建 PR 时模板自动带出检查清单，覆盖宪法的各道合并门禁。
- **SC-005**: CI 在 main 与 develop 的推送及 PR 上触发，治理校验作业为绿灯。
- **SC-006**: 仓库根目录不再残留工具抓取的临时文件。

## Assumptions

- 仓库已于 2026-11-13 由私有改为公开。随之解锁并已启用：分支保护（main / develop）、
  仓库规则集（标签保护）、GitHub 内置 secret scanning 与 push protection、Dependabot 漏洞告警
  与安全更新。CI 中的 gitleaks 作业作为纵深防御保留。
- **评审人数的限时例外**：GitHub 禁止作者批准自己的 PR，单一维护者仓库无法满足"≥1 名非作者评审"。
  期间以 CI 全绿 + CODEOWNERS 审查请求 + 禁直接推送 + 禁 force push + 删除保护替代，
  已在宪法 v2.6.0 的「协作与知识」中写明例外与恢复条件。
- LICENSE 采用 Apache-2.0；若项目实际不打算开源，可改为专有声明或 MIT，法律含义由维护者确认。
- 语言相关的 CI 门禁（覆盖率、依赖方向检查、SAST/SCA、契约破坏检测）待首个服务确定实现语言后补入，
  在 `.github/workflows/ci.yml` 中已留占位说明。
- 本仓库当前无应用代码，故本 feature 的"实现"全部为仓库治理资产，无编译产物、无单元测试；
  其正确性由治理校验脚本自身覆盖（反向测试：注入违规后必须失败）。
