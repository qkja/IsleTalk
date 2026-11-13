# Implementation Plan: identity-svc 身份服务

**Branch**: `feature/003-identity-svc` | **Date**: 2026-11-13 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/003-identity-svc/spec.md`（承接 `specs/002-wechat-style-chat` 的 identity-svc 职责）

## Summary

本 feature 实现 IsleTalk 的第一个微服务 **identity-svc（账号与身份）**，是首段切片（MVP）的第一块基石。
它承接 002「FR→服务承接表」中 identity-svc 的职责：手机号注册登录、多端会话、公开账号名与不可变内部 ID、
资料隐私、可选实名与信任等级、注销。按宪法原则 V，服务内部采用 DDD 四层（api → application → domain → infrastructure），
服务间经 gRPC 契约（002 已定义 `identity/v1/identity.proto`），数据服务独占（MongoDB `identity` database）。

**本 feature 产出真实业务代码**——这是 IsleTalk 首次落地服务实现，区别于 001/002 的治理与架构资产。

## Technical Context

**Language/Version**: Go 1.22+（宪法 v2.7.0 已登记工具链）。

**Primary Dependencies**: go-zero（服务框架）、Google Wire（编译期依赖注入）、gRPC + protobuf、
MongoDB Go driver、Redis Go client、Argon2id 密码哈希库、OpenTelemetry（日志/指标/追踪）。

**Storage**: MongoDB（`identity` 独占 database）；Redis（会话令牌、验证码、限流——非唯一事实来源）。

**Testing**: go test + testify（单元/集成）；go test -cover（覆盖率 CI 门禁 ≥80%）；buf（契约 lint）。

**Target Platform**: Linux 容器（本地 Docker Compose；生产编排待定，见 002 待办）。

**Project Type**: gRPC 微服务（DDD 四层分层）。

**Performance Goals**: 注册/登录接口 p95 ≤ 500ms；账号名精确搜索 p95 ≤ 200ms；万级并发在线（002 SC-022）。

**Constraints**: 手机号/身份信息不落明文、不进日志/指标/追踪（FR-022、宪法 E 组）；密码 Argon2id（宪法 E 组）；
令牌时限（access ≤15min、refresh ≤7day，宪法 E 组）；账号名全局唯一（FR-013）。

**Scale/Scope**: 十万级注册用户（002 SC-023）；单服务，6 个聚合根/实体。

## Constitution Check

*GATE: 已通过。*

| 宪法条目 | 判定 | 落地方式 |
|---|---|---|
| 原则 V DDD 四层 + 依赖单向 | ✅ | `api → application → domain`；`infrastructure` 实现 port 由外层注入；domain 零框架依赖 |
| 原则 V 数据服务独占 | ✅ | MongoDB `identity` database，集合带 `identity` 前缀 |
| 原则 II 覆盖率 ≥80% | ✅ | domain/application 无 IO 单测；go test -cover CI 门禁 |
| 原则 III 统一信封 | ✅ | 客户端接口走 `{success,data,message,meta}`；gRPC 对内 |
| A 组 Go 工具链 | ✅ | 已登记宪法 v2.7.0（gofmt/golangci-lint/go vet/buf） |
| E 组 密码 Argon2id | ✅ | Argon2id m=19456/t=2/p=1（<1s） |
| E 组 令牌时限 | ✅ | access ≤15min、refresh ≤7day 且用后轮换 |
| E 组 手机号/身份信息脱敏 | ✅ | 手机号哈希存储；身份信息 AEAD 加密；日志不落 |
| E 组 登录/找回限流 | ✅ | 账号维度 + IP 维度双限流（Redis） |
| B 组 错误码注册表 | ✅ | 复用 002 `contracts/error-registry.md` |

**无宪法违规需要豁免。**

## 服务内 DDD 四层结构

```
services/identity/
├── api/                       # 协议适配：gRPC handler、入参校验、错误码映射
│   └── grpc/                  #   identity_grpc.go —— 实现 identity.v1.IdentityService
├── application/               # 用例编排、事务边界、鉴权入口、信封映射
│   ├── command/               #   命令对象（Register、Login、SetPublicName…）
│   ├── query/                 #   查询对象（GetUser、SearchByPublicName…）
│   └── service/               #   application service（每个用例一个）
├── domain/                    # 实体、值对象、领域服务、仓储接口 port（零框架依赖）
│   ├── entity/                #   User、Session、Verification、TrustLevel、NameClaim
│   ├── valueobject/           #   UserId、PublicName、PhoneHash、PasswordHash
│   ├── service/               #   领域服务（如账号名唯一性判定）
│   └── port/                  #   UserRepository、SessionRepository、VerificationRepository
│                              #   TokenIssuer、VerificationCodeSender（port 接口）
└── infrastructure/            # 仓储实现、第三方客户端、配置装配、组合根
    ├── persistence/           #   MongoDB 仓储实现 + 索引定义
    ├── cache/                 #   Redis 会话/验证码/限流适配器
    ├── auth/                  #   Argon2id 哈希、令牌签发/校验、短信适配器
    └── di/                    #   Wire 组合根（唯一）
```

**分层铁律**（宪法原则 V）：
- `domain` 层 MUST NOT 出现 MongoDB driver、Redis client、gRPC 类型、HTTP 框架、配置库。
- 领域实体 MUST 用自有值类型（如 `UserId string`），MUST NOT 直接使用 `ObjectId` 作标识。
- `application` 层定义事务边界（一个用例 = 一个本地事务），MUST NOT 开跨服务事务。
- `api` 层只做协议适配与入参校验，MUST NOT 含业务逻辑。

## 数据存储（MongoDB）

**database**: `identity`（本服务独占）。集合与索引：

| 集合 | 对应实体 | 关键索引 |
|---|---|---|
| `identity.users` | 用户账号 | `public_name` 唯一（部分索引，仅 active）、`phone_hash` 唯一 |
| `identity.sessions` | 设备会话 | `user_id` 普通、`session_id` 唯一 |
| `identity.user_prefs` | 界面语言偏好 | `user_id` 唯一 |
| `identity.verifications` | 实名信息 | `user_id` 唯一 |
| `identity.trust` | 信任等级 | `user_id` 唯一 |
| `identity.name_claims` | 账号名占用 | `name` 唯一（含保留期条目） |

**关键设计**：
- `phone_hash`：手机号以加盐哈希存储，MUST NOT 落明文；哈希算法与盐在 `infrastructure/persistence` 内定，domain 只见 `PhoneHash` 值类型。
- `public_name`：唯一性由 `identity.users.public_name` 部分唯一索引 + `identity.name_claims.name` 唯一索引双重保证；改名先写 `name_claims`（旧名进保留期），再更新 `users`，事务边界在 application。
- 身份信息（`identity.verifications`）字段以 AEAD（AES-256-GCM）加密，密钥经密钥管理注入，MUST NOT 硬编码。

## 缓存与协调（Redis）

| 用途 | key | TTL | 说明 |
|---|---|---|---|
| 会话令牌 | `identity:session:{sid}` | 与令牌一致 | 撤销时删除；MongoDB 会话记录为持久事实，Redis 为热路径 |
| 验证码 | `identity:verify:{phone_hash}` | 5min（默认） | 一次性：校验后删除；重发冷却 60s（默认） |
| 登录限流 | `identity:ratelimit:login:{phone_hash}` / `:{ip}` | 窗口 | 账号 + IP 双维度（宪法 E 组） |
| 账号名占用锁 | `identity:lock:name:{name}` | 30s | 改名/注册时的分布式锁，防并发占用 |

**铁律**：Redis 一律非唯一事实来源；核心写路径不依赖 Redis 可用（Redis 故障只降级限流/验证码，不丢用户数据）。

## 契约（gRPC + 对外信封）

- **服务间 gRPC**：实现 002 已定义的 `contracts/proto/identity/v1/identity.proto` 的 5 个 RPC
  （GetUser / SearchByPublicName / CheckUser / ListSessions / RevokeSession），RPC 签名与消息编号 MUST 与 002 一致，
  不得在本 feature 改动。
- **对外（客户端）**：统一信封 `{success,data,message,meta}`（002 `contracts/envelope.md`）；错误码复用 002
  `contracts/error-registry.md`（AUTH_* / ARG_* / RATE_*）。
- **契约测试**：基于 `identity/v1/identity.proto` 的提供方契约测试纳入 CI；proto 兼容性用 buf breaking 检测。
- **防枚举**：登录/注册/找回/验证码接口 MUST NOT 区分"手机号是否已注册"（FR-003、SC-003）；`CheckUser` 响应统一。

## 测试策略与覆盖率预期

- **单元测试**：`domain` 与 `application` 层无数据库/无网络/无框架容器，覆盖实体不变量、值对象、
  账号名唯一性领域服务、用例编排（含负向用例）。这是覆盖率主体。
- **集成测试**：针对容器化 MongoDB/Redis，覆盖仓储实现、索引约束（唯一性）、会话撤销、验证码一次性。
- **契约测试**：identity.proto 提供方契约测试纳入 CI。
- **覆盖率 ≥80%**（CI 硬门禁）：`go test -cover` 汇入 CI；关键路径（注册/登录/改名/实名撤回/注销）接近 100% 分支覆盖。
- **测试隔离**：不依赖执行顺序、真实短信/网络、生产数据；短信用 mock。

## 项目结构（本 feature 产出）

```text
specs/003-identity-svc/
├── plan.md              # 本文件
├── research.md          # Phase 0：identity 特有决策（令牌/密码/验证码/账号名）
├── data-model.md        # Phase 1：6 实体详细字段、索引、状态机
└── quickstart.md        # Phase 1：端到端验证场景

services/identity/       # 服务代码（本 feature 实现，见上方四层结构）
├── api/ application/ domain/ infrastructure/
├── go.mod / go.sum
└── ...                  # 由 tasks.md 逐条落地
```

> 注：`contracts/` 不重复——identity.proto 已在 002 定义，本 feature 实现并持有契约测试，不改契约。

## Complexity Tracking

无宪法违规需要论证豁免。

## 待办（本 plan 遗留，不阻塞实现但 MUST 锁定）

1. **可调参数锁定**（在 tasks 落地时）：验证码有效期/冷却、账号名冷却/保留期、令牌时长——均已有默认值，在实现时以配置项落地并对用户可见。
2. **信任等级额度**（FR-026）：✅ 已落地（2026-11-14）——见 `data-model.md`「信任等级」额度表：消息发送频率 30→60 条/分、新增好友 20→40 个/天、创建群聊 5→10 个/天；由各消费服务按 `level` 执行。
3. **短信/密钥管理服务**：本地 mock 短信；生产短信服务商与密钥管理服务选型属 002 待办，本服务以 port 接口隔离。
