# Research: relationship-svc 好友关系服务

**Feature**: `specs/004-relationship` | **Date**: 2026-11-14

本文件记录 relationship-svc 的特有决策（承接 002 的好友关系语义；语言/存储选型已在 002 research 与宪法 v2.7.0 锁定，不重复）。

## 决策记录

### R-1 好友关系双向一致（SC-002）

- **Decision**: 接受好友申请时，在 application 层一个本地事务内双写两条 Friendship 文档（A→B 与 B→A）。
- **Rationale**: SC-002 要求「双方好友列表出现对方的一致性 100%」。单文档记录一对关系（pair）会导致查某方好友列表需双向扫描，无法高效索引；双写则双方各有一条 `user_id` 索引直接命中。
- **Alternatives considered**: 单 pair 文档 → 拒绝（查询复杂、`OR` 双字段索引低效）。

### R-2 拉黑单向 + 历史保留（FR-008/009/011）

- **Decision**: Block 文档只记录 `owner_id → target_id` 一条；拉黑拦截申请/会话/消息，但 MUST NOT 删除或隐藏既有历史聊天记录（历史记录由 message-svc 持有，relationship 不触碰）。
- **Rationale**: 002 FR-011 明确「拉黑 MUST 是单向」「拉黑 MUST NOT 影响双方既有历史聊天记录」。
- **Alternatives considered**: 双向屏蔽 → 拒绝（违反 002 单向语义）。

### R-3 申请唯一性 + 时效（FR-003）

- **Decision**: `friend_requests` 的 `(initiator_id, target_id)` 唯一索引保证同一对用户同一时刻只有一条待处理申请；待处理申请默认 30 天过期（可调）。
- **Rationale**: FR-003 要求「同一对用户同一时刻 MUST NOT 存在多条待处理申请」；30 天时效避免陈年申请堆积。
- **Alternatives considered**: 允许多条 → 拒绝（「新的朋友」入口会重复，接受语义歧义）。

### R-4 删除双向解除（FR-007）

- **Decision**: 删除好友时，双删两条 Friendship 文档（A→B 与 B→A）；既有单聊会话由 message-svc 保留历史但禁止新消息（relationship 通过 CanMessage 返回 false 实现）。
- **Rationale**: 002 US2 AC7 明确「删除后双方好友关系解除」。
- **Alternatives considered**: 单向移除 → 拒绝（违反双向解除语义）。

### R-5 CanMessage 授权判定（FR-004/009）

- **Decision**: `CanMessage(sender, recipient)` 判定「是否互为好友且未被拉黑」：双方均存在 ACCEPTED 的 Friendship 且无任一方 Block 另一方 → allowed=true；否则 false 并返回具体 FriendshipStatus。
- **Rationale**: 供 message-svc 在发单聊前做授权校验；判定逻辑内聚在 relationship 领域服务。
- **Alternatives considered**: 由 message-svc 自行判断 → 拒绝（越界，且 relationship 数据对 message 不透明）。

### R-6 陌生人添加开关（FR-010）

- **Decision**: 发申请前，relationship 调 identity-svc 的 `GetUser` 读 `allow_stranger_add`；false 且非好友 → 拒绝申请，提示「对方未开启陌生人添加」。
- **Rationale**: FR-010 要求「关闭后非好友 MUST NOT 能发起申请」；开关数据在 identity-svc（FR-049）。
- **Alternatives considered**: relationship 复制一份开关 → 拒绝（数据冗余、不一致风险）。

### R-7 契约修正：relationship.proto 的 RPC 名冲突（实现时发现）

- **Decision**: 002 定义的 `relationship.proto` 中 `rpc RespondFriendRequest(RespondFriendRequest)` 的 RPC 名与输入 message 名相同，protoc 无法解析（RPC 名遮蔽 message 名，报 "RespondFriendRequest is not a message type"）。本 feature 将 RPC 名改为 `RespondToFriendRequest`，message 名保持 `RespondFriendRequest` 不变。
- **Rationale**: 这是 002 契约的一个编译级 bug，不修复无法生成任何代码；RPC 名改动词、message 名保留名词，语义更清晰。proto 为 alpha（002 已标注），此类破坏性修正可接受，但 MUST 在合入后同步回 002 契约文档。
- **Alternatives considered**: 改 message 名 → 拒绝（message 名被客户端/测试引用更多）；换 protoc 版本 → 拒绝（不解决根本冲突）。
