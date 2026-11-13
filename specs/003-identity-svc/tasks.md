# Tasks: identity-svc 身份服务

**Input**: Design documents from `specs/003-identity-svc/`

**Prerequisites**: plan.md（DDD 四层 + MongoDB/Redis + gRPC）、spec.md（6 用户故事）、data-model.md（6 实体）、research.md（决策）

**Tests**: 宪法原则 II 强制 TDD——每个行为 MUST 先写失败测试（RED）再实现（GREEN）。故每个用户故事的测试任务先于实现任务。

**Organization**: 按用户故事分组，每个故事可独立实现、独立测试。

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行（不同文件、无依赖）
- **[Story]**: 所属用户故事（US1–US6）
- 描述含文件路径（相对 `services/identity/`）

## Phase 1: Setup（项目初始化）

**Purpose**: 建立 Go 服务骨架与工具链

- [X] T001 初始化 Go module 与基础依赖，创建 `services/identity/go.mod`（module `isletalk/identity`，Go 1.22，依赖 go-zero、Wire、MongoDB driver、Redis client、Argon2id、OpenTelemetry）
- [X] T002 [P] 配置 `services/identity/.golangci.yml`（gofmt/golangci-lint 规则）与 `services/identity/buf.yaml`（契约 lint）
- [X] T003 [P] 建立 DDD 四层目录骨架：`api/grpc/`、`application/{command,query,service}/`、`domain/{entity,valueobject,service,port}/`、`infrastructure/{persistence,cache,auth,di}/`

## Phase 2: Foundational（阻塞所有用户故事）

**Purpose**: 值类型、仓储接口、基础设施适配器、组合根——所有故事的地基

- [X] T004 定义领域值类型：`domain/valueobject/user_id.go`（`UserId string`）、`public_name.go`、`phone_hash.go`、`password_hash.go`（不可变、含校验）
- [X] T005 定义仓储接口 port：`domain/port/user_repository.go`、`session_repository.go`、`verification_repository.go`、`name_claim_repository.go`
- [X] T006 定义外部能力接口 port：`domain/port/token_issuer.go`（签发/校验/撤销）、`verification_code_sender.go`（下发验证码）
- [X] T007 实现 Argon2id 密码哈希：`infrastructure/auth/password.go`（m=19456/t=2/p=1，带参数与盐的哈希格式）
- [X] T008 实现令牌签发/校验：`infrastructure/auth/token.go`（access ≤15min、refresh ≤7day 且用后轮换，JWT + JWKS 校验 `aud`）
- [X] T009 实现 MongoDB 连接与集合索引：`infrastructure/persistence/mongo.go`（6 集合，`public_name`/`phone_hash`/`name` 唯一索引）
- [X] T010 实现 Redis 适配器：`infrastructure/cache/redis.go`（会话/验证码/限流 key，TTL 见 research）
- [X] T011 实现 Wire 组合根：`infrastructure/di/wire.go`（唯一装配入口，注入仓储与适配器）

**Checkpoint**: 地基就绪——domain 可在无数据库/网络下编译并测试

## Phase 3: User Story 1 - 手机号注册并登录 (Priority: P1) 🎯 MVP

**Goal**: 用户手机号注册、验证码校验、密码登录、找回密码

**Independent Test**: 干净环境注册 → 验证码校验 → 登录 → 退出，不依赖其他服务

### Tests (RED first)

- [X] T012 [P] [US1] 写 User 实体测试 `domain/entity/user_test.go`（不变量：`id` 不可变、`public_name` 非空、状态机）
- [X] T013 [P] [US1] 写注册用例测试 `application/service/register_test.go`（验证码一次性、重复手机号拒绝、防枚举）
- [X] T014 [P] [US1] 写登录用例测试 `application/service/login_test.go`（密码正确/错误、防枚举统一响应）

### Implementation (GREEN)

- [X] T015 [US1] 实现 User 实体 `domain/entity/user.go`（字段按 data-model，`PhoneHash`/`PublicName` 值类型）
- [X] T016 [US1] 实现验证码 mock 发送 `infrastructure/auth/sms_mock.go`（实现 `VerificationCodeSender` port）
- [X] T017 [US1] 实现注册用例 `application/service/register.go`（验证码校验 → 建账号 → 发令牌；依赖 T015/T016）
- [X] T018 [US1] 实现登录用例 `application/service/login.go`（密码 Argon2id 校验 → 发令牌；失败统一响应）
- [X] T019 [US1] 实现找回密码用例 `application/service/reset_password.go`（手机号验证 → 重置 → 旧会话全失效）
- [X] T020 [US1] 实现 gRPC handler `api/grpc/identity_grpc.go`（实现 identity.proto 的注册/登录相关 RPC，入参校验 + 错误码映射）

**Checkpoint**: US1 可独立验证——注册登录闭环可跑

## Phase 4: User Story 2 - 多端会话管理 (Priority: P1)

**Goal**: 多设备同时在线、远程下线、改密全失效

**Independent Test**: 双设备登录 → 列会话 → 远程下线其一 → 改密后全失效

- [X] T021 [P] [US2] 写 Session 实体测试 `domain/entity/session_test.go`（撤销后拒绝、状态机）
- [X] T022 [P] [US2] 写会话用例测试 `application/service/session_test.go`（双写 Redis/MongoDB 一致、改密全失效）
- [X] T023 [US2] 实现 Session 实体 `domain/entity/session.go`
- [X] T024 [US2] 实现会话用例 `application/service/session.go`（创建/列表/撤销，撤销=删 Redis + 标 MongoDB，以 MongoDB 为准收敛）
- [X] T025 [US2] 实现改密全失效逻辑 `application/service/session.go`（并入会话用例，改密触发全量撤销，5 分钟内收敛）
- [X] T026 [US2] 实现 gRPC `ListSessions`/`RevokeSession` RPC（`api/grpc/identity_grpc.go` 扩展）

**Checkpoint**: US2 可独立验证——多端会话管理闭环可跑

## Phase 5: User Story 3 - 公开账号名与不可变标识 (Priority: P1)

**Goal**: 设置/改名公开账号名（唯一 + 保留期 + 冷却期），内部 ID 不可变

**Independent Test**: 设账号名 → 改名 → 旧名保留期不可占 → 内部 ID 不变

- [X] T027 [P] [US3] 写 NameClaim 与账号名唯一性领域服务测试 `domain/service/name_uniqueness_test.go`（并发冲突、保留期占位）
- [X] T028 [US3] 实现账号名唯一性领域服务 `domain/service/name_uniqueness.go`（唯一索引 + `name_claims` 保留期 + 分布式锁）
- [X] T029 [US3] 实现设置/改名用例 `application/service/public_name.go`（冷却期 30 天、保留期 90 天、内部 ID 不变）
- [X] T030 [US3] 实现 gRPC `SearchByPublicName` RPC `api/grpc/identity_grpc.go`（精确搜索，未命中返回空）

**Checkpoint**: US3 可独立验证——账号名生命周期闭环可跑

## Phase 6: User Story 4 - 资料与隐私可见性 (Priority: P2)

**Goal**: 改昵称头像签名、资料可见范围、陌生人添加开关、界面语言

**Independent Test**: 改资料 → 调可见范围 → 切陌生人开关 → 设语言

- [X] T031 [P] [US4] 写资料修改用例测试 `application/service/profile_test.go`
- [X] T032 [US4] 实现资料修改用例 `application/service/profile.go`（昵称/头像/签名/可见范围/陌生人开关）
- [X] T033 [US4] 实现界面语言偏好用例 `application/service/locale.go`（zh-CN/en-US 持久化）

**Checkpoint**: US4 可独立验证

## Phase 7: User Story 5 - 可选实名与信任等级 (Priority: P2)

**Goal**: 实名登记（收集 + 加密存储 + 单独同意）、撤回与信任降级

**Independent Test**: 提交身份信息（含单独同意）→ 获信任等级 → 撤回 → 降级

- [X] T034 [P] [US5] 写实名登记用例测试 `application/service/verification_test.go`（单独同意必填、撤回降级、信息不落日志）
- [X] T035 [US5] 实现 Verification 实体与 AEAD 加密 `domain/entity/verification.go` + `infrastructure/persistence/crypto.go`
- [X] T036 [US5] 实现实名登记用例 `application/service/verification.go`（单独同意记录 → 加密存储 → 授予信任等级）
- [X] T037 [US5] 实现撤回实名用例 `application/service/verification.go`（删除身份信息 + 信任降级 + 额度回落）

**Checkpoint**: US5 可独立验证

## Phase 8: User Story 6 - 注销与数据处置 (Priority: P2)

**Goal**: 主动注销（二次确认 + 重验证）、数据处置告知

**Independent Test**: 注销 → 二次确认 → 账号失效 + 会话全失效

- [X] T038 [P] [US6] 写注销用例测试 `application/service/deactivate_test.go`
- [X] T039 [US6] 实现注销用例 `application/service/deactivate.go`（重验证 → 状态置 deactivated → 会话全失效 → 数据保留策略）

**Checkpoint**: US6 可独立验证——全 6 个故事闭环

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: 对外服务能力、契约测试、集成测试、覆盖率门禁

- [X] T040 实现 gRPC `GetUser`/`CheckUser` RPC `api/grpc/identity_grpc.go`（对外服务能力 FR-029/030，响应不含手机号明文/身份信息）
- [ ] T041 [P] 写契约测试 `services/identity/test/contract/identity_contract_test.go`（基于 identity.proto，5 RPC 全覆盖）
- [X] T042 [P] 写集成测试 `services/identity/test/integration/`（容器化 MongoDB/Redis，覆盖仓储索引约束、会话双写、验证码一次性）
- [ ] T043 配置 CI 覆盖率门禁 `.github/workflows/ci.yml`（go test -cover ≥80%，排除生成代码）并执行 `quickstart.md` 全部 14 个验证场景

---

## Phase 10: Analyze 补遗（覆盖缺口修复）

**Purpose**: `/speckit-analyze` 发现的 4 个覆盖缺口，追加为可执行任务。各任务标注其本应归属的阶段。

- [X] T044 实现服务间 gRPC mTLS `infrastructure/auth/mtls.go`（宪法 E 组：100% 启用、校验对端证书 SAN；应在 Foundational T011 之后执行）
- [ ] T045 实现注册/登录风控的异常行为识别 `application/service/ratelimit.go`（账号 + IP 双维度限流、刷号/暴力尝试自动阻断，对应 FR-006/SC-010；应在 US1 内执行）
- [X] T046 实现身份信息零泄露验证 `services/identity/test/security/leak_check_test.go`（抽样断言日志/指标/追踪/错误响应/对外接口中出现次数为 0，对应 SC-007；应在 US5 后执行）
- [X] T047 实现非界面文案国际化（验证码短信、错误提示、注销告知的 zh-CN/en-US 双语资源，对应 FR-019；应在 US4 的 T033 之后执行）

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (P1)**: 无依赖
- **Foundational (P2)**: 依赖 Setup，**阻塞所有用户故事**
- **US1–US3 (P1 故事)**: 依赖 Foundational，可并行（不同文件）
- **US4–US6 (P2 故事)**: 依赖 Foundational；US5 依赖 US1（实名基于账号）；US4/US6 相对独立
- **Polish (P9)**: 依赖全部故事

### User Story Dependencies

- US1（注册登录）：无依赖，MVP 核心
- US2（会话）：依赖 US1（会话挂账号）
- US3（账号名）：依赖 US1（账号名挂账号）
- US4（资料）：依赖 US1
- US5（实名）：依赖 US1
- US6（注销）：依赖 US1/US2（注销需会话全失效）

### Within Each User Story

- 测试（RED）→ 实体/值对象 → 用例（application service）→ gRPC handler → 集成
- 依赖 T004–T011（Foundational）的接口，MUST 先于故事实现

### Parallel Opportunities

- Foundational 内 T004/T005/T006（port 与值类型）可并行；T007/T008/T009/T010 可并行（不同文件）
- US1 的 T012/T013/T014（测试）可并行
- US2/US3/US4 三个故事在 Foundational 完成后可并行（不同文件，不同聚合根）
- Polish 的 T041/T042 可并行

---

## Implementation Strategy

### MVP First（US1）

1. Phase 1 Setup + Phase 2 Foundational
2. Phase 3 US1（注册登录）
3. **STOP and VALIDATE**: 跑 quickstart 场景 1–3、5，覆盖率 ≥80%
4. 这是可独立交付的最小闭环

### Incremental Delivery

1. US1 → 验证 → 交付（身份注册登录可用）
2. US2 + US3 → 验证（多端会话 + 账号名，relationship-svc 可依赖）
3. US4 + US5 + US6 → 验证（资料/实名/注销）
4. Polish → 契约测试 + 覆盖率门禁 → 全 14 场景通过

### 与其他服务的关系

- identity-svc 是首段切片第一块；完成后 `004-relationship`、`005-message` 才可依赖其 `GetUser`/`CheckUser` 接口
- gRPC 契约（identity.proto）在 002 已定义，本 feature 只实现不改契约

---

## Notes

- [P] 任务 = 不同文件、无依赖，可并行
- 每个任务完成后可独立验证；TDD 强制先 RED 后 GREEN
- 宪法原则 II：覆盖率 ≥80% 是 CI 硬门禁，不达标阻断合并
- 手机号/身份信息不落明文、不进日志/指标/追踪（FR-022，评审已回写）
- 会话撤销双写（Redis/MongoDB）以 MongoDB 为准收敛（Edge Case 已回写）
