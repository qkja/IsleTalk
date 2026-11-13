# Tasks: IsleTalk 即时通讯应用 — 架构落地与门禁

**Input**: Design documents from `specs/002-wechat-style-chat/`

**Prerequisites**: plan.md（本 feature 的架构设计）、spec.md（用户故事）、research.md、data-model.md、contracts/

**Feature 性质说明**：本 feature 是**架构规划 feature**，交付的是服务拆分、技术选型、契约与数据模型，**不实现任何服务的业务代码**。依宪法原则 V「新增微服务 MUST 有对应 spec.md」，7 个服务由后续 feature（`003-identity`、`004-relationship`、`005-message` …）分别承接。因此本 tasks 覆盖的是**架构资产的验证 + 两道阻塞门禁（宪法修订）**，不含业务实现。

**Tests**: 本 feature 无业务代码，不生成单元测试任务；正确性由治理校验器与 proto 契约校验覆盖（与 001 相同的处理）。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行（不同文件、无依赖）
- **[Story]**: 仅用户故事阶段任务需要；本 feature 不实现用户故事，故不加
- 每个任务含明确文件路径或命令

## Phase 1: Setup（架构资产验证）

**Purpose**: 确认已产出的架构资产自洽、可执行

- [X] T001 运行治理校验，确认 002 产物（spec→plan→tasks）依赖关系成立（命令：`python .github/scripts/check-governance.py`，Windows 用 `python` 非 `python3`）
- [X] T002 [P] 校验 `contracts/proto/` 下三份 proto（`identity/v1/identity.proto`、`relationship/v1/relationship.proto`、`message/v1/message.proto`）语法正确、`relationship` 与 `message` 对 `identity/v1/identity.proto` 的 import 依赖成立
- [X] T003 [P] 复核 `data-model.md` 的 16 个实体与 `spec.md`「Key Entities」一一对应，无遗漏、无跨服务数据归属冲突（每实体归单一服务，引用以业务 ID 而非他方外键）

---

## Phase 2: Foundational（宪法修订——阻塞所有服务 feature）

**Purpose**: 两道宪法 MINOR 修订是任一服务合入主干的前置门禁（plan.md「待办」第 1 条）

**⛔ CRITICAL**: 未完成本阶段，任何 `003-*` 服务 feature 都不得合入主干

- [X] T004 通过 `/speckit-constitution` 以 **MINOR** 修订宪法 A 组：登记 Go 工具链（gofmt / golangci-lint / go vet / go mod 依赖锁定 / go test 覆盖率 / buf 契约测试），解除 `.specify/memory/constitution.md` 中 `TODO(TECH_STACK)` 的「实现语言」待定项
- [X] T005 通过 `/speckit-constitution` 以 **MINOR** 修订宪法 G 组：明确原生移动端 WCAG 2.2 AA 落地口径（「平台原生无障碍规范 + WCAG 判据逐条复核，自动化可访问性检查覆盖 Web 端页面」），落地 `.specify/memory/constitution.md`
- [X] T006 完成 T004/T005 后，更新 `specs/002-wechat-style-chat/plan.md` 的「Constitution Check」表中 G 组 ⚠️ 项为 ✅，并同步「待办」清单移除已登记项

**Checkpoint**: 宪法门禁就绪，服务 feature 方可启动

---

## Phase 3: 验证与收尾

**Purpose**: 三份产物一致性复核 + 可重复验证

- [X] T007 运行 `/speckit-analyze` 复核 `spec.md` / `plan.md` / `tasks.md` 的一致性与质量，发现项在合并前处置（记录于该命令输出）
- [X] T008 执行 `quickstart.md` 场景 A（治理校验），确认退出码 0 且可重复通过

---

## Phase 4: Polish & Cross-Cutting Concerns

**Purpose**: 影响后续所有 feature 的收尾

- [X] T009 复核 `plan.md`「服务拆分」7 个限界上下文与 `data-model.md` 实体归属、`contracts/proto/README.md` 包名约定三者一致，无一处数据越界或契约归属矛盾
- [X] T010 [P] 复核 `research.md` 的「未决项」是否已在 plan.md「待办」逐一登记，无遗漏（消息队列/对象存储/编排平台/服务发现等）

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: 无依赖，可立即开始；T001 必须最先（它是后续所有验证的基线）
- **Foundational (Phase 2)**: 依赖 Phase 1 完成；T004/T005 相互独立可并行，T006 依赖二者
- **验证与收尾 (Phase 3)**: 依赖 Phase 2（宪法修订后 spec/plan 引用才最终一致）
- **Polish (Phase 4)**: 依赖 Phase 1–3

### 后续 feature 依赖（本 tasks 之外）

- 各服务 feature（`003-identity` 起）MUST 在 Phase 2（宪法修订）完成后启动，且各走完整 specify→clarify→plan→tasks→implement→analyze
- 首段切片 = `003-identity` + `004-relationship` + `005-message`（对应 spec 的 US1/US2/US3，1v1 场景）
- 群聊（US4）、通知（US6）、通话（US7）、治理（US8/9）为后续切片

### Parallel Opportunities

- Phase 1：T002、T003 可与 T001 之后的验证并行（不同文件）
- Phase 2：T004、T005 相互独立可并行（同一文件不同组，但 `/speckit-constitution` 是命令驱动，实际串行执行避免版本冲突）
- Phase 4：T009、T010 可并行（不同文件）

---

## Implementation Strategy

### 本 feature 的交付（MVP = 架构 + 门禁就绪）

1. 完成 Phase 1：验证架构资产自洽
2. 完成 Phase 2：宪法修订（Go 工具链 + WCAG 口径）—— 这是本 feature 的实质交付
3. 完成 Phase 3/4：analyze 复核 + 收尾
4. **STOP and VALIDATE**: 治理校验通过、三份 proto 可 lint、宪法已修订、plan 与 spec 一致

### 后续交付（各服务 feature，非本 tasks）

1. `003-identity` 落地账号/会话/实名 → 独立验证
2. `004-relationship` 落地好友/拉黑 → 独立验证
3. `005-message` 落地 1v1 消息 → 三者合成首段切片 MVP
4. 群聊/通知/通话/治理逐一切片接入

---

## Notes

- 本 tasks 遵循「过程中不提交、feature 完成再提交」的工作约定，全部改动留工作区直至本 feature 收尾
- 宪法修订（T004/T005）MUST 走 `/speckit-constitution`，不得手改 constitution.md；修订后顶部 Sync Impact Report MUST 提交前删除
- 每个任务完成后可独立验证；T001 的治理校验是贯穿始终的回归基线
