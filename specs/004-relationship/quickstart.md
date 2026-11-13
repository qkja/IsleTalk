# Quickstart: relationship-svc 验证指南

**Feature**: `specs/004-relationship` | **Date**: 2026-11-14

本文件是验证指南，回答"怎么证明 relationship-svc 真的成立了"。实现细节归 tasks.md。

## 前置条件

- 本地依赖已起：MongoDB（`relationship` database）、Redis（Docker Compose）。
- identity-svc 已起（供 `GetUser` 读陌生人开关、`SearchByPublicName` 搜索；或用 mock IdentityClient）。
- 服务已启动并通过就绪检查。

## 端到端验证场景

以下场景对应 spec 的 3 个用户故事验收标准，逐条必须通过。

| # | 场景 | 预期 | 对应 |
|---|---|---|---|
| 1 | A 搜索 B 的公开账号名并发好友申请（含附言） | B 在「新的朋友」看到申请及附言 | US1、SC-004 |
| 2 | B 接受申请 | 双方好友列表同时出现对方，`CanMessage` 返回 allowed | US1、SC-002 |
| 3 | A 再次向 B 发申请（已是好友） | 被拒绝（重复申请） | US1、FR-003 |
| 4 | B 拒绝另一条申请 | 申请移入历史，双方关系维持原状 | US1 |
| 5 | A 给 B 设备注 | A 端显示备注，B 端不受影响 | US2、SC-005 |
| 6 | A 删除 B | 双方关系解除，`CanMessage` 返回 false（历史由 message-svc 保留） | US2、SC-006 |
| 7 | A 拉黑 B | B 发申请/`CanMessage` 被拒且收到原因类别；A/B 历史不受影响 | US3、SC-003 |
| 8 | A 解除拉黑 B | B 恢复发申请能力 | US3 |
| 9 | A 关闭「允许陌生人添加」 | 非好友 B 发申请被拒，提示未开启 | US3、FR-010 |

## 契约验证

- `relationship/v1/relationship.proto` 的 4 个 RPC 全部实现，签名与 002 一致。
- `buf lint` / `buf breaking` 无告警；提供方契约测试纳入 CI。
- `go test -cover` 覆盖率 ≥80%，CI 门禁生效。

## 铁律复核（宪法原则 V / E 组）

- `domain` 层无 MongoDB driver / Redis client / gRPC 类型 / 配置库。
- relationship ↔ identity 调用 100% mTLS。
- Redis 非唯一事实来源：停掉 Redis 后好友关系数据仍可恢复，仅限流/缓存降级。
