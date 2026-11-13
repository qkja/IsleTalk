# Tasks: relationship-svc 好友关系服务

**Input**: Design documents from `specs/004-relationship/`

**Prerequisites**: plan.md（DDD 四层 + MongoDB/Redis + gRPC）、spec.md（3 用户故事）、data-model.md（4 实体）、research.md（6 决策）

**Tests**: 宪法原则 II 强制 TDD——每个行为 MUST 先写失败测试（RED）再实现（GREEN）。故每个用户故事的测试任务先于实现任务。

**Organization**: 按用户故事分组，每个故事可独立实现、独立测试。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行（不同文件、无依赖）
- **[Story]**: 所属用户故事（US1–US3）
- 描述含文件路径（相对 `services/relationship/`）

## Phase 1: Setup（项目初始化）

**Purpose**: 建立 Go 服务骨架与工具链

- [X] T001 初始化 Go module 与基础依赖，创建 `services/relationship/go.mod`（module `isletalk/relationship`，Go 1.22，依赖 go-zero、Wire、MongoDB driver、Redis client、gRPC）
- [X] T002 [P] 配置 `services/relationship/.golangci.yml` 与 `services/relationship/buf.yaml`（契约 lint，指向 002 的 `relationship/v1/relationship.proto`）
- [X] T003 [P] 建立 DDD 四层目录骨架：`api/grpc/`、`application/service/`、`domain/{entity,valueobject,service,port}/`、`infrastructure/{persistence,cache,grpcclient,di}/`

## Phase 2: Foundational（阻塞所有用户故事）

**Purpose**: 值类型、仓储接口、基础设施适配器、组合根——所有故事的地基

- [X] T004 定义领域值类型：`domain/valueobject/user_id.go`（`UserId string`）、`remark_text.go`（备注文本，非空、有长度上限）
- [X] T005 定义仓储接口 port：`domain/port/friendship_repository.go`、`friend_request_repository.go`、`block_repository.go`、`remark_repository.go`
- [X] T006 定义 IdentityClient port：`domain/port/identity_client.go`（`GetUser` 读陌生人开关、`SearchByPublicName` 搜索）
- [X] T007 实现 MongoDB 连接与集合索引：`infrastructure/persistence/mongo.go`（4 集合：`relationship.friendships` 的 `(user_id, friend_id)` 唯一、`friend_requests` 的 `(initiator_id, target_id)` 唯一、`blocks` 的 `(owner_id, target_id)` 唯一、`remarks` 的 `(owner_id, friend_id)` 唯一）
- [X] T008 实现 Redis 适配器：`infrastructure/cache/redis.go`（申请限流 key `relationship:ratelimit:request:{id}/{ip}`、好友列表缓存 key `relationship:friends:{id}`，TTL 60s）
- [X] T009 实现 identity-svc gRPC 客户端：`infrastructure/grpcclient/identity.go`（封装 `GetUser`/`SearchByPublicName`，100% mTLS）
- [X] T010 实现 Wire 组合根：`infrastructure/di/wire.go`（唯一装配入口，注入仓储与 identity 客户端）

**Checkpoint**: 地基就绪——domain 可在无数据库/网络下编译并测试

## Phase 3: User Story 1 - 找到人并成为好友 (Priority: P1) 🎯 MVP

**Goal**: 搜索 → 发好友申请（附言）→ 接受/拒绝 → 互为好友

**Independent Test**: A 发申请 → B 接受 → 双方好友列表出现对方，可立即发起会话；不依赖消息服务

### Tests (RED first)

- [X] T011 [P] [US1] 写 Friendship 实体测试 `domain/entity/friendship_test.go`（不变量：一对好友 = 两条文档、状态机 accepted→deleted）
- [X] T012 [P] [US1] 写 FriendRequest 实体测试 `domain/entity/friend_request_test.go`（状态机 pending→accepted/rejected、附言非空）
- [X] T013 [P] [US1] 写发申请用例测试 `application/service/friend_request_test.go`（重复申请拒绝、陌生人开关关闭拒绝）
- [X] T014 [P] [US1] 写接受/拒绝用例测试 `application/service/respond_test.go`（接受后双向一致、拒绝维持原状）

### Implementation (GREEN)

- [X] T015 [US1] 实现 Friendship 实体 `domain/entity/friendship.go`（字段按 data-model：`user_id`/`friend_id`/`created_at`/`status`）
- [X] T016 [US1] 实现 FriendRequest 实体 `domain/entity/friend_request.go`（`initiator_id`/`target_id`/`request_message`/`status`/`created_at`/`responded_at`）
- [X] T017 [US1] 实现发申请用例 `application/service/friend_request.go`（账号+IP 双维度限流 → 唯一索引防重复申请 → 调 IdentityClient 读陌生人开关 → 建申请）
- [X] T018 [US1] 实现接受/拒绝用例 `application/service/respond.go`（查询待处理申请列表 → 接受时本地事务双写两条 Friendship 文档保证双向一致；拒绝标记 rejected）
- [X] T019 [US1] 实现 gRPC handler `api/grpc/relationship_grpc.go`（`SendFriendRequest`/`RespondFriendRequest`，入参校验 + 错误码映射）

**Checkpoint**: US1 可独立验证——发申请/接受/拒绝闭环可跑

## Phase 4: User Story 2 - 好友备注与删除 (Priority: P2)

**Goal**: 单向备注、好友列表、删除好友（双向解除）

**Independent Test**: A 给 B 设备注 → A 端显示备注、B 端不受影响 → A 删除 B → 关系解除

### Tests (RED first)

- [X] T020 [P] [US2] 写 Remark 实体测试 `domain/entity/remark_test.go`（单向、备注文本上限）
- [X] T021 [P] [US2] 写备注/列表/删除用例测试 `application/service/friend_list_test.go`（备注单向、列表分页、删除双删）

### Implementation (GREEN)

- [X] T022 [US2] 实现 Remark 实体 `domain/entity/remark.go`（`owner_id`/`friend_id`/`content`/`updated_at`）
- [X] T023 [US2] 实现备注用例 `application/service/remark.go`（设置/修改备注，单向）
- [X] T024 [US2] 实现好友列表与删除用例 `application/service/friend_list.go`、`delete.go`（列表分页默认 50 最大 1000；删除双删两条 Friendship）

**Checkpoint**: US2 可独立验证——备注/列表/删除闭环可跑

## Phase 5: User Story 3 - 拉黑与陌生人开关 (Priority: P2)

**Goal**: 拉黑/解除拉黑（单向拦截）、陌生人添加开关、CanMessage 授权

**Independent Test**: A 拉黑 B → B 发申请/消息被拒 → A 解除拉黑 → 恢复；A 关陌生人开关 → 非好友发申请被拒

### Tests (RED first)

- [X] T025 [P] [US3] 写 Block 实体测试 `domain/entity/block_test.go`（单向、状态机 blocked↔unblocked）
- [X] T026 [P] [US3] 写拉黑/CanMessage 用例测试 `application/service/block_test.go`、`can_message_test.go`（拉黑拦截申请/消息、历史保留、授权判定）

### Implementation (GREEN)

- [X] T027 [US3] 实现 Block 实体 `domain/entity/block.go`（`owner_id`/`target_id`/`created_at`/`status`）
- [X] T028 [US3] 实现拉黑用例 `application/service/block.go`（拉黑/解除拉黑，单向；拉黑后既有待处理申请失效）
- [X] T029 [US3] 实现 CanMessage 授权用例 `application/service/can_message.go`（判定「互为好友且未被任一方拉黑」，返回 FriendshipStatus）
- [X] T030 [US3] 实现陌生人开关检查 `application/service/friend_request.go`（发申请前调 IdentityClient.GetUser 读 `allow_stranger_add`）
- [X] T031 [US3] 实现 gRPC handler `api/grpc/relationship_grpc.go`（`SetBlock`/`CanMessage`，扩展 T019）

**Checkpoint**: US3 可独立验证——拉黑/CanMessage/陌生人开关闭环可跑

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: 契约测试、集成测试、客户端信封接口、日志、覆盖率门禁

- [X] T032 [P] 写契约测试 `services/relationship/test/contract/relationship_contract_test.go`（基于 relationship.proto，4 RPC 全覆盖）
- [X] T033 [P] 写集成测试 `services/relationship/test/integration/`（容器化 MongoDB/Redis，覆盖双向一致、唯一索引约束、拉黑拦截）
- [X] T034 实现客户端信封接口 `api/http/`（发申请/接受/备注/拉黑/好友列表/删除/「新的朋友」申请列表（待处理+历史），统一信封 `{success,data,message,meta}` 与错误码注册表）
- [X] T035 接入分级日志（slog JSON，debug/info/warn/error，每个用例入口打 info、错误路径打 error）
- [X] T036 配置 CI 覆盖率门禁 `.github/workflows/ci.yml`（relationship 服务 job，mongo/redis 容器 + -race + cover ≥80%）

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (P1)**: 无依赖
- **Foundational (P2)**: 依赖 Setup，**阻塞所有用户故事**
- **US1 (P1 故事)**: 依赖 Foundational，MVP 核心
- **US2/US3 (P2 故事)**: 依赖 Foundational；US3 依赖 US1（CanMessage/拉黑基于好友关系）

### User Story Dependencies

- US1（发申请/接受）：无依赖，MVP 核心
- US2（备注/删除）：依赖 US1（备注/删除作用于好友关系）
- US3（拉黑/CanMessage）：依赖 US1（拉黑与授权判定基于好友关系与申请）

### Within Each User Story

- 测试（RED）→ 实体/值对象 → 用例（application service）→ gRPC handler → 集成
- 依赖 T004–T010（Foundational）的接口，MUST 先于故事实现

### Parallel Opportunities

- Foundational 内 T004/T005/T006（值类型与 port）可并行；T007/T008/T009 可并行（不同文件）
- US1 的 T011/T012/T013/T014（测试）可并行
- US2 与 US3 在 US1 完成后可并行（不同文件，不同聚合根）
- Polish 的 T032/T033 可并行

---

## Implementation Strategy

### MVP First（US1）

1. Phase 1 Setup + Phase 2 Foundational
2. Phase 3 US1（发申请 + 接受/拒绝）
3. **STOP and VALIDATE**: 跑 quickstart 场景 1–4，覆盖率 ≥80%
4. 这是可独立交付的最小闭环（「加好友」能力）

### Incremental Delivery

1. US1 → 验证 → 交付（好友申请/接受可用，message-svc 可依赖 CanMessage）
2. US2 → 验证（备注/删除）
3. US3 → 验证（拉黑/陌生人开关/CanMessage）
4. Polish → 契约测试 + 客户端接口 + 覆盖率门禁 → 全 9 场景通过

### 与其他服务的关系

- relationship-svc 是首段切片第二块；完成后 `005-message` 可依赖其 `CanMessage` 授权接口
- gRPC 契约（relationship.proto）在 002 已定义，本 feature 只实现不改契约

---

## Notes

- [P] 任务 = 不同文件、无依赖，可并行
- 每个任务完成后可独立验证；TDD 强制先 RED 后 GREEN
- 宪法原则 II：覆盖率 ≥80% 是 CI 硬门禁，不达标阻断合并
- 好友关系双向一致（SC-002）：接受时本地事务双写两条 Friendship 文档
- 拉黑单向（FR-008）：Block 只存 owner→target 一条，历史保留
- 删除双向解除（FR-007）：双删两条 Friendship 文档
- 性能基线测试（SC-001 送达率、p95 响应时间）属 002 待办（负载测试工具未选型），本 feature 不实现
