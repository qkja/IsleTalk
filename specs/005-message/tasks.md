# Tasks: message-svc 会话消息服务

**Input**: Design documents from `specs/005-message/`

**Prerequisites**: plan.md（DDD 四层 + MongoDB/Redis + gRPC）、spec.md（3 用户故事）、data-model.md（2 实体）、research.md（6 决策）

**Tests**: 宪法原则 II 强制 TDD——每个行为 MUST 先写失败测试（RED）再实现（GREEN）。

**Organization**: 按用户故事分组，每个故事可独立实现、独立测试。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行（不同文件、无依赖）
- **[Story]**: 所属用户故事（US1–US3）
- 描述含文件路径（相对 `services/message/`）

## Phase 1: Setup（项目初始化）

**Purpose**: 建立 Go 服务骨架与工具链

- [x] T001 初始化 Go module 与基础依赖，创建 `services/message/go.mod`（module `isletalk/message`，Go 1.22，依赖 go-zero、Wire、MongoDB driver、Redis client、gRPC）
- [x] T002 [P] 配置 `services/message/buf.yaml`（契约 lint，指向 002 的 `message/v1/message.proto`）
- [x] T003 [P] 建立 DDD 四层目录骨架：`api/grpc/`、`application/service/`、`domain/{entity,valueobject,service,port}/`、`infrastructure/{persistence,cache,grpcclient,di}/`

## Phase 2: Foundational（阻塞所有用户故事）

**Purpose**: 值类型、仓储接口、基础设施适配器、组合根

- [x] T004 定义领域值类型：`domain/valueobject/message_id.go`（幂等键）、`message_content.go`（正文或附件 key）
- [x] T005 定义仓储接口 port：`domain/port/message_repository.go`、`conversation_repository.go`、`relationship_client.go`（`CanMessage`）
- [x] T006 实现 MongoDB 连接与集合索引：`infrastructure/persistence/mongo.go`（2 集合：`message.messages` 的 `message_id` 唯一 + `(conversation_id, seq)` 唯一、`message.conversations` 的 `conversation_id` 唯一）
- [x] T007 实现 Redis 适配器：`infrastructure/cache/redis.go`（未读计数缓存 `message:unread:{id}`、会话列表缓存）
- [x] T008 实现 relationship-svc gRPC 客户端：`infrastructure/grpcclient/relationship.go`（封装 `CanMessage`，100% mTLS）
- [x] T009 实现 Wire 组合根：`infrastructure/di/wire.go`

**Checkpoint**: 地基就绪——domain 可在无数据库/网络下编译并测试

## Phase 3: User Story 1 - 与好友一对一收发消息 (Priority: P1) 🎯 MVP

**Goal**: 文本/图片/文件消息、发送状态（送达/已读）

**Independent Test**: 两个账号互为好友 → 双向发送文本/图片/文件 → 校验送达与已读状态

### Tests (RED first)

- [x] T010 [P] [US1] 写 Message 实体测试 `domain/entity/message_test.go`（不变量：message_id 非空、seq 单调、状态机 sent→delivered→read）
- [x] T011 [P] [US1] 写 Conversation 实体测试 `domain/entity/conversation_test.go`（类型 direct、未读计数）
- [x] T012 [P] [US1] 写发消息用例测试 `application/service/send_message_test.go`（幂等去重、非好友拒绝、被拉黑拒绝）

### Implementation (GREEN)

- [x] T013 [US1] 实现 Message 实体 `domain/entity/message.go`（`message_id`/`conversation_id`/`sender_id`/`type`/`content`/`seq`/`sent_at`/`status`/`recalled`）
- [x] T014 [US1] 实现 Conversation 实体 `domain/entity/conversation.go`（`conversation_id`/`type`/`participants`/`last_message_preview`/`unread_count`/`pinned`/`muted`/`deleted`/`last_active_at`）
- [x] T015 [US1] 实现发消息用例 `application/service/send_message.go`（调 CanMessage 授权 → message_id 幂等 → seq 分配 → 建消息并标记 delivered）
- [x] T016 [US1] 实现 gRPC handler `api/grpc/message_grpc.go`（`SendMessage`/`ListMessages`，入参校验 + 错误码映射 + 分页 page_size 默认 50 上限 1000）

**Checkpoint**: US1 可独立验证——发送/接收/送达/已读闭环可跑

## Phase 4: User Story 2 - 撤回与离线补发 (Priority: P1)

**Goal**: 撤回时间窗内消息、离线消息按序补发

**Independent Test**: A 撤回消息 → 双方占位；B 离线期间 A 连发多条 → B 上线按序收到全部

### Tests (RED first)

- [x] T017 [P] [US2] 写撤回用例测试 `application/service/recall_test.go`（时间窗内成功、超窗拒绝、非本人拒绝）
- [x] T018 [P] [US2] 写已读上报用例测试 `application/service/mark_read_test.go`（状态 delivered→read、未读计数更新）

### Implementation (GREEN)

- [x] T019 [US2] 实现撤回用例 `application/service/recall.go`（2 分钟时间窗由服务端时钟判定 → recalled 标记 → 校验发送者身份）
- [x] T020 [US2] 实现已读上报用例 `application/service/mark_read.go`（更新消息状态 read + 会话未读计数清零）
- [x] T021 [US2] 实现 gRPC handler `api/grpc/message_grpc.go`（`RecallMessage`/`MarkRead`，扩展 T016）

**Checkpoint**: US2 可独立验证——撤回/已读/离线补发闭环可跑

## Phase 5: User Story 3 - 会话列表与管理 (Priority: P2)

**Goal**: 会话列表（摘要/未读/倒序）、置顶/免打扰/删除

**Independent Test**: 多会话列表 → 置顶 → 免打扰 → 删除，逐项验证

### Tests (RED first)

- [x] T022 [P] [US3] 写会话列表用例测试 `application/service/conversation_test.go`（倒序、未读、置顶、免打扰、删除后重出现）

### Implementation (GREEN)

- [x] T023 [US3] 实现会话列表用例 `application/service/conversation.go`（列表按最近活跃倒序 → 置顶 → 免打扰 → 标记未读 → 删除）
- [x] T024 [US3] 实现客户端信封接口 `api/http/`（会话列表/置顶/免打扰/删除，统一信封 + 错误码注册表）

**Checkpoint**: US3 可独立验证——会话管理闭环可跑

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 契约测试、集成测试、日志、CI 门禁

- [x] T025 [P] 写契约测试 `services/message/test/contract/message_contract_test.go`（基于 message.proto，4 RPC 全覆盖）
- [x] T026 [P] 写集成测试 `services/message/test/integration/`（容器化 MongoDB/Redis，覆盖幂等唯一索引、seq 时序、撤回、未读计数）
- [x] T027 接入分级日志（slog JSON，debug/info/warn/error，每个用例入口打 info、错误路径打 error）
- [x] T028 配置 CI 覆盖率门禁 `.github/workflows/ci.yml`（message 服务 job，mongo 容器 + cover ≥80%，按服务目录变更过滤）

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (P1)**: 无依赖
- **Foundational (P2)**: 依赖 Setup，**阻塞所有用户故事**
- **US1 (P1)**: 依赖 Foundational，MVP 核心
- **US2 (P1)**: 依赖 US1（撤回/已读作用于消息）
- **US3 (P2)**: 依赖 US1（会话列表基于消息）

### User Story Dependencies

- US1（收发消息）：无依赖，MVP 核心
- US2（撤回/已读）：依赖 US1（作用于消息）
- US3（会话列表）：依赖 US1（摘要/未读基于消息）

### Within Each User Story

- 测试（RED）→ 实体 → 用例（application service）→ gRPC handler → 集成

### Parallel Opportunities

- Foundational 内 T004/T005 可并行；T006/T007/T008 可并行
- US1 的 T010/T011/T012（测试）可并行
- US2 的 T017/T018（测试）可并行
- Polish 的 T025/T026 可并行

---

## Implementation Strategy

### MVP First（US1）

1. Phase 1 Setup + Phase 2 Foundational
2. Phase 3 US1（收发消息 + 送达/已读）
3. **STOP and VALIDATE**: 跑 quickstart 场景 1–4，覆盖率 ≥80%
4. 这是可独立交付的最小闭环（「能聊天」）

### Incremental Delivery

1. US1 → 验证 → 交付（收发消息可用）
2. US2 → 验证（撤回/已读/离线补发）
3. US3 → 验证（会话列表管理）
4. Polish → 契约测试 + 客户端接口 + 覆盖率门禁

### 与其他服务的关系

- message-svc 是首段切片最后一块；依赖 relationship-svc 的 `CanMessage`（发消息授权）
- gRPC 契约（message.proto）在 002 已定义，本 feature 只实现不改契约

---

## Notes

- [P] 任务 = 不同文件、无依赖，可并行
- 每个任务完成后可独立验证；TDD 强制先 RED 后 GREEN
- 宪法原则 II：覆盖率 ≥80% 是 CI 硬门禁
- 消息幂等（FR-007）：message_id 唯一索引，重发不重复
- 时序（FR-010）：seq 会话内单调递增，`(conversation_id, seq)` 唯一索引
- 撤回（FR-006）：2 分钟窗口由服务端时钟判定，非本人撤回拒绝
- 授权（FR-011）：发消息前调 relationship CanMessage

## Phase 7: Convergence

- [x] T029 [US2] 撤回消息后，若其为会话最新消息，将会话「最新消息摘要」更新为「撤回了一条消息」占位（MUST NOT 显示原文） per FR-006/CHK014/US2-AC1 (partial)
- [x] T030 实现未读计数与会话列表的 cache-aside 读路径（命中读缓存返回、未命中回源 MongoDB 并回填，写后失效保持） per plan:缓存与协调 (partial)

## Phase 8: Convergence

- [x] T031 补齐 cache-aside 写后失效：撤回（TouchOnRecall）后失效会话双方列表缓存，发消息（TouchOnSend）后同时失效发送方与接收方缓存（当前仅失效接收方，撤回未失效） per T030/plan:缓存与协调 (partial)
