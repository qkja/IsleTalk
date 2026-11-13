# Implementation Plan: message-svc 会话消息服务

**Branch**: `feature/005-message` | **Date**: 2026-11-14 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/005-message/spec.md`（承接 `specs/002-wechat-style-chat` 的 message-svc 职责）

## Summary

本 feature 实现 IsleTalk 的第三个微服务 **message-svc（会话与消息）**，是首段切片（MVP）的最后一块基石。
它承接 002「FR→服务承接表」中 message-svc 的职责：1v1 单聊会话、文本/图片/文件消息、发送状态（送达/已读）、
撤回（时间窗）、会话列表与管理、多端一致。按宪法原则 V，服务内部采用 DDD 四层，服务间经 gRPC 契约
（002 已定义 `message/v1/message.proto`，4 个 RPC），数据服务独占（MongoDB `message` database）。

**首段切片范围**：仅 1v1 单聊；群聊（FR-014/015）、转发（FR-021）、引用（FR-022）不在本切片。

## Technical Context

**Language/Version**: Go 1.22+（宪法 v2.7.0 已登记工具链）。

**Primary Dependencies**: go-zero、Google Wire、gRPC + protobuf、MongoDB Go driver、Redis Go client、OpenTelemetry。

**Storage**: MongoDB（`message` 独占 database）；Redis（未读计数、会话缓存——非唯一事实来源）。

**Testing**: go test + testify；go test -cover（覆盖率 CI 门禁 ≥80%）；buf（契约 lint）。

**Target Platform**: Linux 容器。

**Project Type**: gRPC 微服务（DDD 四层分层）。

**Performance Goals**: 发送消息 p95 ≤ 200ms；消息送达 p95 ≤ 500ms；万级并发在线（002 SC-022）。

**Constraints**: 消息幂等去重（message_id 唯一）；会话内 seq 单调递增保证时序；撤回由服务端时钟判定；
发消息前 MUST 校验好友关系（调 relationship CanMessage）；多端消息顺序一致（FR-026）。

**Scale/Scope**: 十万级注册用户；单服务，2 个聚合根（Conversation、Message）。

## Constitution Check

*GATE: 已通过。*

| 宪法条目 | 判定 | 落地方式 |
|---|---|---|
| 原则 V DDD 四层 + 依赖单向 | ✅ | `api → application → domain`；`infrastructure` 实现 port 由外层注入 |
| 原则 V 数据服务独占 | ✅ | MongoDB `message` database，集合带 `message` 前缀 |
| 原则 II 覆盖率 ≥80% | ✅ | domain/application 无 IO 单测；go test -cover CI 门禁 |
| 原则 III 统一信封 | ✅ | 客户端接口走 `{success,data,message,meta}`；gRPC 对内 |
| A 组 Go 工具链 | ✅ | 已登记宪法 v2.7.0 |
| E 组 服务间 mTLS | ✅ | message ↔ relationship 调用 100% mTLS |
| E 组 消息授权 | ✅ | 发消息前 MUST 校验互为好友且未拉黑（CanMessage） |
| B 组 分页 | ✅ | ListMessages 分页（page_size 默认 50、上限 1000，page_token 游标） |
| B 组 幂等 | ✅ | message_id 唯一索引去重，重发不产生重复 |

**无宪法违规需要豁免。**

## 服务内 DDD 四层结构

```
services/message/
├── api/                       # gRPC handler（4 RPC）+ 客户端信封接口
│   └── grpc/                  #   message_grpc.go —— 实现 message.v1.MessageService
├── application/               # 用例编排、事务边界
│   └── service/               #   send_message.go / list_messages.go / recall.go / mark_read.go / conversation.go
├── domain/                    # 实体、值对象、领域服务、port（零框架依赖）
│   ├── entity/                #   Conversation、Message
│   ├── valueobject/           #   MessageId、MessageContent
│   ├── service/               #   领域服务（时序 seq、撤回判定）
│   └── port/                  #   MessageRepository、ConversationRepository、RelationshipClient
└── infrastructure/            # MongoDB 仓储、Redis、gRPC 客户端、Wire
    ├── persistence/           #   mongo_repo.go（消息 + 会话）
    ├── cache/                 #   redis.go（未读计数）
    ├── grpcclient/            #   relationship 的 CanMessage 客户端
    └── di/                    #   wire.go
```

**分层铁律**（宪法原则 V）：
- `domain` 层 MUST NOT 出现 MongoDB driver、Redis client、gRPC 类型、配置库。
- `application` 层定义事务边界；`api` 层只做协议适配与入参校验。

## 数据存储（MongoDB）

**database**: `message`（本服务独占）。集合与索引：

| 集合 | 对应实体 | 关键索引 |
|---|---|---|
| `message.conversations` | 会话 | `conversation_id` 唯一、`participant` 普通 |
| `message.messages` | 消息 | `message_id` 唯一（幂等）、`(conversation_id, seq)` 唯一（时序） |

**关键设计**：
- **消息幂等**（FR-007）：`message_id` 唯一索引，重发相同 message_id 时返回已有消息，MUST NOT 重复。
- **时序**（FR-010）：`seq` 为会话内单调递增序号，`(conversation_id, seq)` 唯一索引保证无乱序；seq 由服务端分配。
- **撤回**（FR-006）：`recalled` 标记 + 占位提示；撤回窗口（2 分钟）由服务端时钟判定。
- **未读计数**：会话文档存未读数量，Redis 缓存热路径，MongoDB 为持久事实。

## 缓存与协调（Redis）

| 用途 | key | TTL | 说明 |
|---|---|---|---|
| 未读计数 | `message:unread:{conversation_id}` | 永久（MongoDB 为源） | 热路径缓存，失效回源 |
| 会话列表缓存 | `message:conversations:{user_id}` | 60s | cache-aside |

**铁律**：Redis 非唯一事实来源；消息本体 MUST 存 MongoDB，Redis 故障只降级未读计数/会话缓存，不丢消息。

## 契约（gRPC + 对外信封）

- **服务间 gRPC**：实现 002 已定义的 `contracts/proto/message/v1/message.proto` 的 4 个 RPC
  （SendMessage / ListMessages / RecallMessage / MarkRead），RPC 签名与消息编号 MUST 与 002 一致，不得改动。
- **跨服务调用**：发消息前调 relationship-svc 的 `CanMessage`（校验互为好友且未拉黑）；调用 MUST 100% mTLS。
- **对外（客户端）**：统一信封 `{success,data,message,meta}`；错误码复用 002 `contracts/error-registry.md`。
- **契约测试**：基于 message.proto 的提供方契约测试纳入 CI；proto 兼容性用 buf breaking 检测。

## 测试策略与覆盖率预期

- **单元测试**：domain/application 无 IO 单测，覆盖实体不变量、时序、撤回判定、幂等去重、授权校验（mock RelationshipClient）。
- **集成测试**：容器化 MongoDB/Redis，覆盖唯一索引约束（message_id 幂等、seq 时序）、撤回、未读计数。
- **契约测试**：message.proto 提供方契约测试纳入 CI。
- **覆盖率 ≥80%**（CI 硬门禁）：关键路径（发送/拉取/撤回/已读/授权）接近 100% 分支覆盖。
- **测试隔离**：不依赖真实 relationship 服务（mock）；不依赖生产数据。

## 项目结构（本 feature 产出）

```text
specs/005-message/
├── plan.md              # 本文件
├── research.md          # Phase 0：消息特有决策（幂等/时序/撤回/未读）
├── data-model.md        # Phase 1：2 实体字段、索引、状态机
└── quickstart.md        # Phase 1：端到端验证场景

services/message/        # 服务代码（本 feature 实现）
├── api/ application/ domain/ infrastructure/
├── go.mod / go.sum
└── ...                  # 由 tasks.md 逐条落地
```

> 注：`contracts/` 不重复——message.proto 已在 002 定义，本 feature 实现并持有契约测试，不改契约。

## Complexity Tracking

无宪法违规需要论证豁免。

## 待办（本 plan 遗留，不阻塞实现但 MUST 锁定）

1. **可调参数锁定**（在 tasks 落地时）：撤回时间窗（默认 2 分钟）、文件大小限制（图片 10MB/文件 100MB）——均已有默认值，实现时以配置项落地。
2. **附件存储**：图片/文件正文的对象存储选型属 002 待办；本服务消息只存附件引用 key，上传/下载接口由后续 feature 或对象存储直传。
3. **消息正文全文检索**（FR-085）：另开 feature，本切片不实现。
