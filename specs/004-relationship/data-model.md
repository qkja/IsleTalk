# Data Model: relationship-svc 好友关系服务

**Feature**: `specs/004-relationship` | **Date**: 2026-11-14

本文件定义 relationship-svc 的 4 个实体的字段、索引与状态机。领域模型与 BSON 文档解耦（宪法原则 V）：
领域实体用自有值类型，BSON 映射在 `infrastructure/persistence` 完成。

## 实体总表

| 实体 | 集合 | 聚合根 | 关键不变量 |
|---|---|---|---|
| 好友关系 | `relationship.friendships` | ✅ | 双向一致（一对好友 = 两条文档）；`(user_id, friend_id)` 唯一 |
| 好友申请 | `relationship.friend_requests` | ✅ | `(initiator_id, target_id)` 唯一；状态机 pending → accepted/rejected |
| 拉黑关系 | `relationship.blocks` | ✅ | 单向；`(owner_id, target_id)` 唯一 |
| 好友备注 | `relationship.remarks` | ✅ | 单向；`(owner_id, friend_id)` 唯一 |

## 好友关系（`relationship.friendships`）

- 字段：`user_id`（关系归属方）、`friend_id`（对方）、`created_at`、`status`（`accepted` / `deleted`）。
- 索引：`(user_id, friend_id)` 唯一、`user_id` 普通（列好友）。
- 状态机：`accepted → deleted`（删除好友，双向解除）。
- 说明：一对好友 = 两条文档（A→B 与 B→A），保证 SC-002 双向一致。

## 好友申请（`relationship.friend_requests`）

- 字段：`initiator_id`、`target_id`、`request_message`（附言）、`status`（`pending` / `accepted` / `rejected`）、
  `created_at`、`responded_at`。
- 索引：`(initiator_id, target_id)` 唯一、`target_id` 普通（「新的朋友」待处理列表）。
- 状态机：`pending → accepted / rejected`；pending 默认 30 天过期（FR-003）。

## 拉黑关系（`relationship.blocks`）

- 字段：`owner_id`（拉黑方）、`target_id`（被拉黑方）、`created_at`、`status`（`blocked` / `unblocked`）。
- 索引：`(owner_id, target_id)` 唯一、`owner_id` 普通。
- 状态机：`blocked ↔ unblocked`（解除拉黑）；单向，历史记录不受影响（FR-011）。

## 好友备注（`relationship.remarks`）

- 字段：`owner_id`（设置方）、`friend_id`、`content`、`updated_at`。
- 索引：`(owner_id, friend_id)` 唯一。
- 约束：单向，仅 owner 可见（FR-005）。

## 跨服务关系

- relationship-svc **被** message-svc（发单聊前 CanMessage 授权）、客户端（发申请/接受/备注/拉黑/删除）调用。
- relationship-svc **依赖** identity-svc：`GetUser`（读 allow_stranger_add）、`SearchByPublicName`（客户端搜索）。
- 对外不暴露 phone_hash 等身份敏感信息；好友列表返回公开账号名/昵称/备注。
