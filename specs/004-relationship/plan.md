# Implementation Plan: relationship-svc 好友关系服务

**Branch**: `feature/004-relationship` | **Date**: 2026-11-14 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/004-relationship/spec.md`（承接 `specs/002-wechat-style-chat` 的 relationship-svc 职责）

## Summary

本 feature 实现 IsleTalk 的第二个微服务 **relationship-svc（好友与关系）**，是首段切片（MVP）的第二块基石。
它承接 002「FR→服务承接表」中 relationship-svc 的职责：好友申请（含附言）、接受/拒绝、单向备注、
拉黑/解除拉黑、删除好友、陌生人添加开关，并对外提供「两人是否可互发消息」的授权校验（供 message-svc 消费）。
按宪法原则 V，服务内部采用 DDD 四层（api → application → domain → infrastructure），服务间经 gRPC 契约
（002 已定义 `relationship/v1/relationship.proto`，4 个 RPC），数据服务独占（MongoDB `relationship` database）。

**本 feature 产出真实业务代码**，与 003 identity-svc 同构（Go + DDD 四层 + MongoDB/Redis + gRPC）。

## Technical Context

**Language/Version**: Go 1.22+（宪法 v2.7.0 已登记工具链）。

**Primary Dependencies**: go-zero（服务框架）、Google Wire（编译期依赖注入）、gRPC + protobuf、
MongoDB Go driver、Redis Go client、OpenTelemetry（日志/指标/追踪）。

**Storage**: MongoDB（`relationship` 独占 database）；Redis（申请限流、好友列表缓存——非唯一事实来源）。

**Testing**: go test + testify（单元/集成）；go test -cover（覆盖率 CI 门禁 ≥80%）；buf（契约 lint）。

**Target Platform**: Linux 容器（本地 Docker Compose；生产编排待定，见 002 待办）。

**Project Type**: gRPC 微服务（DDD 四层分层）。

**Performance Goals**: 好友申请/接受 p95 ≤ 500ms；`CanMessage` 授权校验 p95 ≤ 100ms；万级并发在线（002 SC-022）。

**Constraints**: 好友关系 MUST 双向一致（接受后双方好友列表同时出现，SC-002）；拉黑 MUST 单向（002 FR-011）；
拉黑 MUST NOT 影响既有历史记录；服务间 gRPC MUST 100% mTLS（宪法 E 组）；删除好友 MUST 双向解除。

**Scale/Scope**: 十万级注册用户（002 SC-023）；单服务，4 个聚合根/实体。

## Constitution Check

*GATE: 已通过。*

| 宪法条目 | 判定 | 落地方式 |
|---|---|---|
| 原则 V DDD 四层 + 依赖单向 | ✅ | `api → application → domain`；`infrastructure` 实现 port 由外层注入；domain 零框架依赖 |
| 原则 V 数据服务独占 | ✅ | MongoDB `relationship` database，集合带 `relationship` 前缀 |
| 原则 II 覆盖率 ≥80% | ✅ | domain/application 无 IO 单测；go test -cover CI 门禁 |
| 原则 III 统一信封 | ✅ | 客户端接口走 `{success,data,message,meta}`；gRPC 对内（002 契约已定） |
| A 组 Go 工具链 | ✅ | 已登记宪法 v2.7.0（gofmt/golangci-lint/go vet/buf） |
| E 组 服务间 mTLS | ✅ | relationship ↔ identity 调用 MUST 100% mTLS，校验对端 SAN |
| E 组 拉黑拦截 | ✅ | 拉黑后 CanMessage/发申请 MUST 拒绝，历史保留 |
| E 组 限流 | ✅ | 好友申请按账号 + IP 双维度限流（防骚扰刷申请） |
| B 组 错误码注册表 | ✅ | 复用 002 `contracts/error-registry.md` |

**无宪法违规需要豁免。**

## 服务内 DDD 四层结构

```
services/relationship/
├── api/                       # 协议适配：gRPC handler、入参校验、错误码映射
│   └── grpc/                  #   relationship_grpc.go —— 实现 relationship.v1.RelationshipService
├── application/               # 用例编排、事务边界、鉴权入口、信封映射
│   └── service/               #   friend_request.go / respond.go / block.go / remark.go
│                              #   friend_list.go / delete.go / can_message.go
├── domain/                    # 实体、值对象、领域服务、仓储接口 port（零框架依赖）
│   ├── entity/                #   Friendship、FriendRequest、Block、Remark
│   ├── valueobject/           #   UserId、RemarkText
│   ├── service/               #   领域服务（如 CanMessage 判定、双向一致性）
│   └── port/                  #   FriendshipRepository、FriendRequestRepository、BlockRepository
│                              #   RemarkRepository、IdentityClient（调 identity-svc）
└── infrastructure/            # 仓储实现、gRPC 客户端、配置装配、组合根
    ├── persistence/           #   MongoDB 仓储实现 + 索引定义
    ├── cache/                 #   Redis 限流/缓存适配器
    ├── grpcclient/            #   identity-svc 的 SearchByPublicName/CheckUser 客户端
    └── di/                    #   Wire 组合根（唯一）
```

**分层铁律**（宪法原则 V）：
- `domain` 层 MUST NOT 出现 MongoDB driver、Redis client、gRPC 类型、HTTP 框架、配置库。
- 领域实体 MUST 用自有值类型（如 `UserId string`），MUST NOT 直接使用 `ObjectId` 作标识。
- `application` 层定义事务边界（一个用例 = 一个本地事务），MUST NOT 开跨服务事务。
- `api` 层只做协议适配与入参校验，MUST NOT 含业务逻辑。

## 数据存储（MongoDB）

**database**: `relationship`（本服务独占）。集合与索引：

| 集合 | 对应实体 | 关键索引 |
|---|---|---|
| `relationship.friendships` | 好友关系 | `user_id` 普通、`(user_id, friend_id)` 唯一 |
| `relationship.friend_requests` | 好友申请 | `(initiator_id, target_id)` 唯一、`target_id` 普通（「新的朋友」入口） |
| `relationship.blocks` | 拉黑关系 | `(owner_id, target_id)` 唯一、`owner_id` 普通 |
| `relationship.remarks` | 好友备注 | `(owner_id, friend_id)` 唯一 |

**关键设计**：
- **好友关系双向一致**（SC-002）：接受申请时在 application 层一个本地事务内写入两条 Friendship 文档（A→B 与 B→A），保证双方好友列表同时出现对方；MUST NOT 出现单向好友。
- **拉黑单向**（FR-011）：Block 文档只记录 `owner_id → target_id` 一条，被拉黑方的可见性由查询时判断。
- **申请唯一性**（FR-003）：`friend_requests` 的 `(initiator_id, target_id)` 唯一索引保证同一对用户同一时刻只有一条待处理申请。
- **备注单向**（FR-005）：Remark 文档只存 `owner_id → friend_id → content`，对方不可见。

## 缓存与协调（Redis）

| 用途 | key | TTL | 说明 |
|---|---|---|---|
| 申请限流 | `relationship:ratelimit:request:{user_id}` / `:{ip}` | 窗口 | 账号 + IP 双维度（防骚扰刷申请） |
| 好友列表缓存 | `relationship:friends:{user_id}` | 60s | cache-aside，失效回源 MongoDB（非唯一事实来源） |

**铁律**：Redis 一律非唯一事实来源；核心写路径不依赖 Redis 可用（Redis 故障只降级限流/缓存，不丢好友关系）。

## 契约（gRPC + 对外信封）

- **服务间 gRPC**：实现 002 已定义的 `contracts/proto/relationship/v1/relationship.proto` 的 4 个 RPC
  （SendFriendRequest / RespondFriendRequest / SetBlock / CanMessage），RPC 签名与消息编号 MUST 与 002 一致，
  不得在本 feature 改动。
- **跨服务调用**：搜索能力由 identity-svc 的 `SearchByPublicName` 提供（客户端先调 identity 搜索拿 UserId，
  再调 relationship 发申请）；`CanMessage` 供 message-svc 做发单聊前的授权校验；relationship → identity 调用
  MUST 100% mTLS。
- **对外（客户端）**：统一信封 `{success,data,message,meta}`（002 `contracts/envelope.md`）；错误码复用 002
  `contracts/error-registry.md`。
- **契约测试**：基于 `relationship/v1/relationship.proto` 的提供方契约测试纳入 CI；proto 兼容性用 buf breaking 检测。

## 测试策略与覆盖率预期

- **单元测试**：`domain` 与 `application` 层无数据库/无网络/无框架容器，覆盖实体不变量、双向一致性领域服务、
  用例编排（含负向用例：重复申请、拉黑拦截、删除后重申请）。这是覆盖率主体。
- **集成测试**：针对容器化 MongoDB/Redis，覆盖仓储实现、唯一索引约束（申请唯一、好友唯一）、双向一致性。
- **契约测试**：relationship.proto 提供方契约测试纳入 CI。
- **覆盖率 ≥80%**（CI 硬门禁）：`go test -cover` 汇入 CI；关键路径（发申请/接受/拉黑/删除/CanMessage）接近 100% 分支覆盖。
- **测试隔离**：不依赖执行顺序、真实 identity 服务（用 mock IdentityClient）；不依赖生产数据。

## 项目结构（本 feature 产出）

```text
specs/004-relationship/
├── plan.md              # 本文件
├── research.md          # Phase 0：relationship 特有决策（双向一致/拉黑语义/申请时效）
├── data-model.md        # Phase 1：4 实体详细字段、索引、状态机
└── quickstart.md        # Phase 1：端到端验证场景

services/relationship/   # 服务代码（本 feature 实现，见上方四层结构）
├── api/ application/ domain/ infrastructure/
├── go.mod / go.sum
└── ...                  # 由 tasks.md 逐条落地
```

> 注：`contracts/` 不重复——relationship.proto 已在 002 定义，本 feature 实现并持有契约测试，不改契约。

## Complexity Tracking

无宪法违规需要论证豁免。

## 待办（本 plan 遗留，不阻塞实现但 MUST 锁定）

1. **可调参数锁定**（在 tasks 落地时）：申请时效（默认 30 天过期）、申请限流阈值——均已有默认值，在实现时以配置项落地并对用户可见。
2. **拉黑原因类别**（FR-009）：被拒绝原因类别的具体枚举（如「对方已拉黑」「对方未开启陌生人添加」）在实现时锁定，写错误码注册表。
3. **好友列表分页**：好友列表 MUST 分页（宪法 B 组），单次上限在实现时锁定。
