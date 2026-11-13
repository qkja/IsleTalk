# Quickstart: message-svc 验证指南

**Feature**: `specs/005-message` | **Date**: 2026-11-14

本文件是验证指南，回答"怎么证明 message-svc 真的成立了"。实现细节归 tasks.md。

## 前置条件

- 本地依赖已起：MongoDB（`message` database）、Redis（Docker Compose）。
- relationship-svc 已起（供 `CanMessage` 授权校验，或用 mock RelationshipClient）。
- 服务已启动并通过就绪检查。

## 端到端验证场景

以下场景对应 spec 的 3 个用户故事验收标准，逐条必须通过。

| # | 场景 | 预期 | 对应 |
|---|---|---|---|
| 1 | A 与 B 是好友，A 发文本消息 | B 数秒内看到，A 端状态「发送中」→「已送达」 | US1、SC-001 |
| 2 | B 打开 A 会话 | A 端消息状态变「已读」 | US1、SC-002 |
| 3 | A 发图片消息 | B 收到并可见图片 | US1、FR-003 |
| 4 | A 发超限文件 | 发送前拦截并说明限制 | US1、FR-004 |
| 5 | A 撤回时间窗内的消息 | 双方看到一致占位提示，正文不再可见 | US2、SC-003 |
| 6 | A 重发同 message_id | 不产生重复消息（幂等） | US2、SC-006 |
| 7 | B 离线期间 A 连发多条，B 上线 | 按原顺序收到全部，无遗漏无重复 | US2、SC-004 |
| 8 | 会话列表 | 最新摘要/未读/活跃时间倒序 | US3、FR-008 |
| 9 | 置顶/免打扰/删除会话 | 逐项生效 | US3、FR-009 |

## 契约验证

- `message/v1/message.proto` 的 4 个 RPC 全部实现，签名与 002 一致。
- `buf lint` / `buf breaking` 无告警；提供方契约测试纳入 CI。
- `go test -cover` 覆盖率 ≥80%，CI 门禁生效。

## 铁律复核（宪法原则 V / E 组）

- `domain` 层无 MongoDB driver / Redis client / gRPC 类型 / 配置库。
- message ↔ relationship 调用 100% mTLS。
- Redis 非唯一事实来源：消息本体存 MongoDB，Redis 故障只降级未读计数，不丢消息。
