# Data Model: IsleTalk 即时通讯应用

**Feature**: `specs/002-wechat-style-chat` | **Date**: 2026-11-13

本文件把 spec.md 的 16 个 Key Entity 映射到 7 个微服务（限界上下文），给出字段、关系与状态机。
这是**领域级数据模型**：定义"数据归谁、有哪些关键字段、怎么流转"，不写具体存储 schema——
每服务的 collection 划分、索引、BSON 映射细节在各服务 feature 的 `plan.md` 落地。

> 约定：每个服务独占一个 MongoDB database，集合名 `{service前缀}.{entity}`。跨服务不共享集合，
> 引用以对方服务暴露的业务 ID 进行，不持有对方表外键。

## 服务 → 实体归属总表

| 实体 | 归属服务 | 数据库前缀 | 首段切片 |
|---|---|---|---|
| 用户账号 | identity-svc | identity | ✅ |
| 设备会话 | identity-svc | identity | ✅ |
| 界面语言偏好 | identity-svc | identity | ✅ |
| 实名信息 | identity-svc | identity | 后续 |
| 信任等级 | identity-svc | identity | 后续 |
| 账号名占用 | identity-svc | identity | ✅ |
| 好友关系 | relationship-svc | rel | ✅ |
| 会话 | message-svc | msg | ✅ |
| 消息 | message-svc | msg | ✅ |
| 阅读位置 | message-svc | msg | ✅ |
| 通知订阅（会话级） | message-svc | msg | ✅ |
| 群聊 | group-svc | grp | 后续 |
| 群成员身份 | group-svc | grp | 后续 |
| 通话 | call-svc | call | 后续 |
| 举报与处置记录 | moderation-svc | mod | 后续 |
| 运营人员账号 | moderation-svc | mod | 后续 |

## identity-svc（账号与身份）

### 用户账号（`identity.users`）
- 字段：`id`（不可变内部 ID，UUIDv7）、`public_name`（公开账号名，唯一）、`phone_hash`（手机号哈希，
  明文不落库或仅加密存储）、`nickname`、`avatar`、`signature`、`visibility`（资料可见范围）、
  `allow_stranger_add`（是否允许陌生人添加）、`status`（active / deactivated）、`moderation_status`（风控处置状态）。
- 约束：手机号（FR-051/071）绝不进日志/指标/追踪；`public_name` 唯一性由「账号名占用」保证。
- 状态机：`active → deactivated`（用户主动注销，FR-006）；处置状态由 moderation-svc 下发，见下。

### 设备会话（`identity.sessions`）
- 字段：`id`、`user_id`、`device_id`、`created_at`、`last_active_at`、`revoked`（是否被远程下线）。
- 关系：一个账号多会话（FR-003 多端同时在线、FR-005 改密后全失效）。
- 状态机：`active → revoked`（主动退出/改密/注销/远程下线）。

### 界面语言偏好（`identity.user_prefs`）
- 字段：`user_id`、`locale`（zh-CN / en-US）、`source`（system / user）、`updated_at`。
- 约束：FR-065/066/067 的驱动数据。

### 实名信息 + 信任等级（`identity.verifications` / `identity.trust`）
- 实名字段：`user_id`、`id_type`、`id_digest`（加密）、`status`（pending / verified / revoked）、
  `consent_at`（单独同意时间）、`retention_until`。
- 信任等级字段：`user_id`、`level`、`quota_overrides`（放宽额度）。
- 状态机：实名 `pending → verified → revoked`（FR-073）；撤回触发信任降级 + 额度回落（Edge Case）。
- 敏感约束：FR-070（单独同意）、FR-071（静态加密、脱敏）、FR-072（用途限定）、FR-074（保留期删除）。

### 账号名占用（`identity.name_claims`）
- 字段：`name`、`owner_user_id`、`released_at`、`reserved_until`。
- 关系：与用户账号的 `public_name` 一一对应；释放后进入保留期（FR-078，默认 90 天）。
- 状态机：`occupied → reserved（保留期）→ available`；并发占用靠唯一索引 + 分布式锁（Edge Case 账号名冲突）。

## relationship-svc（好友与关系）

### 好友关系（`rel.friendships`）
- 字段：`id`、`initiator_user_id`、`target_user_id`、`status`（pending / accepted / rejected / blocked）、
  `remark`（备注，仅对设置方可见）、`request_message`（申请附言）、`created_at`、`updated_at`。
- 关系：双向关系但状态单向存储（发起方/接收方视角可派生）；拉黑是单向（FR-011）。
- 状态机：`pending → accepted | rejected`；`accepted → blocked → accepted`（解除拉黑）；删除好友即移除关系
  （FR-010，但不删消息历史——历史归 message-svc）。
- 跨服务：接受好友申请后向 message-svc 发事件以允许建立单聊会话（FR-009）。

## message-svc（会话与消息）

### 会话（`msg.conversations`）
- 字段：`id`、`type`（direct / group）、`participants[]`（direct 为两人，group 引用 group-svc 的群 ID）、
  `last_message_preview`、`last_active_at`、`per_participant_state`（未读、置顶、免打扰、标记未读）。
- 关系：direct 会话与好友关系弱耦合——删除好友后会话保留但禁发（FR-013/011）。
- 状态机：会话不删除（删除会话仅清本地视图，服务端保留历史，FR-024 语义）。

### 消息（`msg.messages`）
- 字段：`id`（幂等去重键，FR-028）、`conversation_id`、`sender_user_id`、`type`（text / image / file）、
  `content`（文本正文或附件 key）、`seq`（会话内单调递增，保证固有时序 FR-031）、`sent_at`、
  `status`（sent / delivered / read / failed）、`recalled`、`forward_of`（转发来源）、`quote_of`（引用目标）。
- 约束：正文不进日志/指标/追踪（FR-051）；撤回置 `recalled`，正文对双方一致不可见（FR-020）。
- 关系：附件二进制在对象存储，此处只存 key 与元数据。

### 阅读位置（`msg.read_cursors`）
- 字段：`conversation_id`、`user_id`、`last_read_seq`、`updated_at`。
- 约束：群聊用于汇总已读人数（FR-083/084），逐人位置 MUST NOT 暴露给其他成员；成员退群后保留位置
  但不再计入新增（Edge Case）。

### 通知订阅（会话级，`msg.conversation_settings`）
- 字段：`conversation_id`、`user_id`、`pinned`、`muted`、`marked_unread`。
- 关系：会话级偏好（FR-024）；与 notification-svc 的推送级偏好（隐私模式）经事件同步，最终一致。

## group-svc（群组）

### 群聊（`grp.groups`）
- 字段：`id`、`name`、`announcement`、`owner_user_id`、`created_at`、`member_cap`（默认 500，FR-045）。

### 群成员身份（`grp.memberships`）
- 字段：`group_id`、`user_id`、`role`（owner / admin / member）、`muted_until`（禁言解除时间，FR-044）、
  `joined_at`。
- 状态机：`member → muted → member`（禁言到期）；`owner` 转让/退出自动转移（FR-015 + Edge Case）。
- 跨服务：message-svc 发群消息前经 gRPC 校验成员资格与禁言状态；成员变更发事件同步。

## notification-svc（通知推送）

### 通知订阅（推送级）
- 字段：`user_id`、`privacy_mode`（通知是否隐藏正文，FR-033）、`device_tokens[]`。
- 关系：与会话级偏好（message-svc）分工见 plan.md「边界上的两条显式决策」。

## call-svc（通话）

### 通话（`call.calls`）
- 字段：`id`、`type`（audio / video / group_audio / group_video）、`participants[]`、`caller_user_id`、
  `started_at`、`ended_at`、`duration`、`end_reason`（hung_up / declined / busy / aborted / missed）。
- 约束：通话记录属敏感数据（FR-052）；不存储媒体流（FR-042）；举报核查仅用元数据（FR-050）。

## moderation-svc（治理与运营）

### 举报与处置记录（`mod.reports` / `mod.dispositions`）
- 举报字段：`id`、`reporter_user_id`、`target_type`（user / group / message / call）、`target_id`、
  `reason`、`evidence`、`status`（received / reviewing / resolved）。
- 处置字段：`id`、`target_user_id`、`action`（warn / mute / login_restrict / ban）、`reason`、`scope`、
  `operator_id`、`applied_at`、`expires_at`、`appeal_status`。
- 约束：审计留痕（操作人/时间/对象/动作/依据）不可删改（FR-063）；处置由运营人员执行，申诉由非原处置人复核
  （FR-064）。

### 运营人员账号（`mod.operators`）
- 字段：`id`、`role`、`permissions[]`、`status`（active / disabled）。
- 约束：独立鉴权与授权（FR-062）；职责结束时权限及时失效（Edge Case）。

## 跨服务关系图（关键引用）

```text
identity-svc ──(好友申请/关系建立)──> relationship-svc
relationship-svc ──(接受好友→允许建会话)──> message-svc
message-svc ──(群成员校验/扇出)──> group-svc
message-svc ──(会话偏好事件)──> notification-svc
message-svc/group-svc/call-svc ──(举报)──> moderation-svc
moderation-svc ──(处置下发)──> identity-svc（封禁）/ group-svc（禁言）/ message-svc（限制）
call-svc ──(通话记录)──> 自有库（元数据）
```

所有跨服务引用均以业务 ID 进行，经 gRPC 或事件，禁止直连他方 database。
