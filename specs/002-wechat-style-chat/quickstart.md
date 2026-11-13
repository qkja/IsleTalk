# Quickstart: 端到端验证指南

**Feature**: `specs/002-wechat-style-chat` | **Date**: 2026-11-13

本文件是验证指南，不是实现文档。它回答"怎么证明这个 feature 真的成立了"。

> 本 feature 交付的是**架构与契约**（服务拆分、选型、信封、错误码、proto、数据模型），
> 不含任何服务的业务实现——各服务在 feature `003-*` 落地。因此本文分两部分：
> **A. 现在就能跑**（治理校验），**B. 首段切片落地后可跑**（端到端验证场景，作为 003+ 的验收基线）。

## A. 现在就能跑的验证

本 feature 产物的正确性由治理校验器兜底（`plan.md` 已遵循 spec→plan 依赖）：

```bash
python .github/scripts/check-governance.py
```

- 预期：退出码 0，输出"治理校验通过"。
- 它会校验：宪法无残留、版本行合规、`specs/002-*/` 的 spec.md 存在且 plan.md 引用正确。
- Windows 上必须用 `python`（`python3` 是 Microsoft Store 别名，会静默退出 49），
  中文输出乱码时加 `PYTHONIOENCODING=utf-8`。

## B. 首段切片（MVP）的端到端验证场景

以下场景对应 spec 的 US1（注册登录）、US2（好友）、US3（1v1 聊天）验收标准，是
`identity-svc` + `relationship-svc` + `message-svc` + 双客户端落地后的验收基线。
各场景的**预期结果**必须逐条通过，任一失败 = 对应 SC 不达标。

### 前置条件

- 本地依赖已起（MongoDB / Redis / 队列 / MinIO），见 `deploy/docker-compose.yml`（由首段 feature 建立）。
- 三个服务已启动且通过就绪检查；iOS/Android 客户端可连。
- 两个测试账号 A、B，A 与 B 互为好友。

### 场景与预期

| # | 场景 | 预期 | 对应 |
|---|---|---|---|
| 1 | A 在干净设备手机号注册 → 验证码 → 设昵称头像 | 直接进入空会话列表，无需人工审核 | US1、SC-001/002 |
| 2 | A 与 B 互为好友后，A 发一条文本消息 | B 数秒内可见，A 端状态 发送中→已送达 | US3、SC-003 |
| 3 | B 打开 A 的会话 | A 端该消息变为"已读" | US3、SC-010 |
| 4 | A 撤回一条 2 分钟内的消息 | 双方一致显示"对方撤回了一条消息"，正文不可见 | US3、SC-009 |
| 5 | A 发送时断网，恢复后重发 | 不出现重复消息；失败时 A 端显式可见并可重发 | US3、SC-006 |
| 6 | B 离线时 A 发多条消息，B 重新上线 | 按原顺序全部收到，无遗漏无重复 | US3、SC-004 |
| 7 | A 拉黑 B，B 再发消息 | B 收到"被拒绝"（区别于网络故障），历史仍双方可见 | US2、SC-008/034 |
| 8 | A 按 B 的公开账号名精确搜索 | 命中 B；按手机号搜索不返回任何注册状态信息 | US2、SC-036/037 |
| 9 | A 有 100 个会话，首次进会话列表 | 可见时间 ≤1.5s，按最近活跃倒序 | US5、SC-007 |
| 10 | 万级并发在线时发消息 | 送达 p95 ≤2s | SC-022 |

### 契约验证

- 三份 proto 经 `buf lint` / `buf breaking` 无告警（见 [contracts/proto/README.md](./contracts/proto/README.md)）。
- 提供方与调用方各持基于同一 `.proto` 的契约测试，纳入 CI。
- 消息投递幂等：同 `message_id` 重复提交只产生一条消息（FR-028）。

## 待办提醒（不阻塞本计划，但 MUST 完成）

1. 宪法 MINOR 修订：① A 组登记 Go 工具链；② G 组明确原生移动端可访问性口径。任一服务合入主干前完成。
2. 各服务 feature（`003-identity` / `004-relationship` / `005-message`…）建立后，本文件的场景 B
   落到各服务的 `tasks.md`，作为其 E2E 验收项。
