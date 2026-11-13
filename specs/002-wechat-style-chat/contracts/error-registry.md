# 错误码注册表

**Feature**: `specs/002-wechat-style-chat` | **Date**: 2026-11-13

宪法原则 III 与 B 组要求错误码来自唯一注册表、同义错误同码、废弃码标记不复用。本文是注册表起点。

## 结构

每个错误码 = `{code}`（字符串，稳定机器码）+ 语义 + 是否可重试 + 对应 HTTP 状态。注册表随代码演进；
**废弃的错误码标记 `DEPRECATED`，绝不复用**。客户端 MUST 用 `code` 分支，不得解析 `message` 文本。

## 编码约定

- 分类前缀：`AUTH_`（认证/授权）、`ARG_`（参数）、`RES_`（资源/状态）、`RATE_`（限流）、
  `SYS_`（系统/依赖）、`MOD_`（治理处置）。
- 同义错误全项目同码：例如"对象不存在"统一 `RES_NOT_FOUND`，不因接口而异。

## 注册表

| code | 语义 | 可重试 | HTTP |
|---|---|---|---|
| `AUTH_UNAUTHENTICATED` | 未认证或令牌无效/过期 | 否 | 401 |
| `AUTH_FORBIDDEN` | 已认证但无权访问（含对象不存在与无权限的统一响应） | 否 | 403 |
| `AUTH_TOKEN_EXPIRED` | 访问令牌过期，需刷新 | 是 | 401 |
| `AUTH_REAUTH_REQUIRED` | 高风险操作需重新验证身份（FR-005） | 否 | 403 |
| `AUTH_ACCOUNT_DISABLED` | 账号被处置（封禁/限制登录） | 否 | 403 |
| `ARG_INVALID` | 参数非法（类型/枚举/格式） | 否 | 400 |
| `ARG_INVALID_ARGUMENT` | 分页负值等特定非法参数（AIP-158） | 否 | 400 |
| `ARG_PAYLOAD_TOO_LARGE` | 请求体/文件超上限（FR-018） | 否 | 413 |
| `RES_NOT_FOUND` | 资源不存在（对外统一，不区分无权限，见 API1） | 否 | 404 |
| `RES_CONFLICT` | 并发冲突（账号名占用、好友申请重复等） | 否 | 409 |
| `RES_ALREADY_FRIEND` | 已是好友 | 否 | 409 |
| `RES_BLOCKED` | 被对方拉黑，行为被拒（FR-011，明确区分于网络故障） | 否 | 403 |
| `RES_GROUP_MEMBER_LIMIT` | 群成员达上限（FR-045） | 否 | 409 |
| `RES_GROUP_MUTED` | 群成员被禁言（FR-044） | 否 | 403 |
| `RES_GROUP_NOT_MEMBER` | 非群成员（含已被移除） | 否 | 403 |
| `RATE_LIMITED` | 触发限流（登录/验证码/邀请/搜索试探） | 是 | 429 |
| `RATE_VERIFY_CODE_COOLDOWN` | 验证码冷却期内 | 是 | 429 |
| `SYS_UNAVAILABLE` | 依赖服务不可用（熔断/超时） | 是 | 503 |
| `SYS_MESSAGE_SEND_FAILED` | 消息发送失败，可重发（FR-029，显式可见） | 是 | 503 |
| `MOD_DISPOSED` | 已受平台处置，含处置详情入口 | 否 | 403 |
| `MOD_APPEAL_PENDING` | 申诉处理中 | 否 | 409 |
| `CALL_BUSY` | 对方忙线（FR-038） | 否 | 409 |
| `CALL_ABORTED` | 通话异常中断 | 是 | 500 |

## 维护规则

- 新增错误码前 MUST 先查注册表，同义复用；确需新增 MUST 在本表登记并同步更新契约测试。
- 废弃：改标 `DEPRECATED` 并在 `message` 指向替代码，保留一个版本周期后从代码移除（表内永不删除记录）。
- 对外 `message` 文案 MUST 外置为可翻译资源（FR-066），`code` 不变、文案按语言切换。
