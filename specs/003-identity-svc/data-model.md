# Data Model: identity-svc 身份服务

**Feature**: `specs/003-identity-svc` | **Date**: 2026-11-13

本文件定义 identity-svc 的 6 个实体的字段、索引与状态机。领域模型与 BSON 文档解耦（宪法原则 V）：
领域实体用自有值类型（`UserId`、`PublicName`、`PhoneHash` 等），BSON 映射在 `infrastructure/persistence` 完成。

## 实体总表

| 实体 | 集合 | 聚合根 | 关键不变量 |
|---|---|---|---|
| 用户账号 | `identity.users` | ✅ | `public_name` 全局唯一；`phone_hash` 唯一；`id` 不可变 |
| 设备会话 | `identity.sessions` | ✅ | 一个用户多会话；撤销后立即拒绝 |
| 界面语言偏好 | `identity.user_prefs` | ❌（属于 User） | 每用户一条 |
| 实名信息 | `identity.verifications` | ✅ | 每用户至多一条有效；加密存储 |
| 信任等级 | `identity.trust` | ❌（派生自 Verification） | 撤回实名同步降级 |
| 账号名占用 | `identity.name_claims` | ✅ | `name` 唯一（含保留期） |

## 用户账号（`identity.users`）

- 字段：
  - `id`：不可变内部账号 ID（UUIDv7 字符串），MUST NOT 展示为可搜索标识（FR-012）
  - `public_name`：公开账号名，全局唯一（FR-013）
  - `phone_hash`：手机号加盐哈希（R-3），唯一；明文不落库
  - `nickname` / `avatar_url` / `signature`：资料字段（FR-017）
  - `visibility`：资料可见范围 `public / friends_only / private`（FR-018）
  - `allow_stranger_add`：是否允许陌生人添加（FR-018）
  - `locale`：界面语言 `zh-CN / en-US`（FR-019；也可拆分到 user_prefs，此处冗余一份热路径）
  - `status`：`active / deactivated`（FR-027/028）
  - `moderation_status`：风控处置状态（由 moderation-svc 下发，本服务只读）
- 索引：`id` 唯一、`public_name` 部分唯一（仅 active）、`phone_hash` 唯一。
- 状态机：`active → deactivated`（注销，二次确认 + 重验证，FR-027）。

## 设备会话（`identity.sessions`）

- 字段：`session_id`（唯一）、`user_id`、`device_id`、`created_at`、`last_active_at`、`revoked`。
- 索引：`session_id` 唯一、`user_id` 普通（列会话）。
- 状态机：`active → revoked`（退出 / 远程下线 / 改密 / 注销，FR-008~010）。

## 界面语言偏好（`identity.user_prefs`）

- 字段：`user_id`（唯一）、`locale`、`source`（`system / user`）、`updated_at`。
- 索引：`user_id` 唯一。

## 实名信息（`identity.verifications`）

- 字段：`user_id`（唯一）、`id_type`、`id_digest`（AEAD 加密，FR-022）、`status`（`pending / submitted / revoked`）、
  `consent_at`（单独同意时间，FR-021）、`retention_until`（保留期限，FR-025）。
- 索引：`user_id` 唯一。
- 状态机：`pending → submitted → revoked`（撤回，FR-024）；注销时按 `retention_until` 删除。

## 信任等级（`identity.trust`）

- 字段：`user_id`（唯一）、`level`、`quota_overrides`（放宽额度，FR-026）。
- 索引：`user_id` 唯一。
- 约束：派生自 Verification；撤回实名 MUST 同步降级并回落额度（FR-024）。

**额度数值**（FR-026，对用户可见；首段切片按「已提交」授予 level=1，不基于真实验证）：

| 受限维度 | 未实名（level 0） | 已实名（level 1） | 执行方 |
|---|---|---|---|
| 消息发送频率 | 30 条/分钟 | 60 条/分钟 | message-svc |
| 新增好友数 | 20 个/天 | 40 个/天 | relationship-svc |
| 创建群聊数 | 5 个/天 | 10 个/天 | group-svc |

撤回实名后 `level` 归零、`quota_overrides` 清空，各消费服务按 level 0 回落（FR-024）。

## 账号名占用（`identity.name_claims`）

- 字段：`name`（唯一）、`owner_user_id`、`released_at`、`reserved_until`（保留期，FR-015）。
- 索引：`name` 唯一（含保留期条目，使旧名保留期内仍占位）。
- 状态机：`occupied → reserved（保留期，默认 90 天）→ available`；改名先写新 claim、旧名转 reserved。
- 保留词（FR-013，锁定于 `domain/valueobject/reserved_words.go`）：`admin`、`administrator`、`root`、
  `system`、`moderator`、`official`、`support`、`help`、`service`、`staff`、`bot`、`isletalk`、
  `官方`、`管理员`、`系统`、`客服`、`机器人`；英文大小写不敏感。

## 跨服务关系

- identity-svc 不依赖其他服务；它**被** relationship-svc（好友）、message-svc（消息授权）调用。
- 提供内部接口（002 `identity.proto`）：`GetUser` / `SearchByPublicName` / `CheckUser` / `ListSessions` / `RevokeSession`。
- 对外不暴露 `phone_hash`、`id_digest` 明文（FR-022/029）。
