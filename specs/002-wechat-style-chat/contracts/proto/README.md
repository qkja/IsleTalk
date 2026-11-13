# 服务间契约：proto 约定

**Feature**: `specs/002-wechat-style-chat` | **Date**: 2026-11-13

服务间同步调用一律 gRPC + protobuf（宪法原则 V）。本文是契约组织与版本化约定。

## 包名与稳定性（AIP-181）

- 包名格式：`isletalk.{service}.{version}`，`version` = `v1` / `v2` …，目录对应 `proto/{service}/{version}/`。
- 稳定性用包内注释标注：`alpha` / `beta` / `stable`，对应不同兼容性承诺。
  - `alpha`：可破坏性变更；`beta`：尽力兼容；`stable`：MUST 向后兼容，破坏性变更 MUST 升大版本。
- 首段切片的 proto 一律标 `alpha`，待契约测试与双跑稳定后升 `beta`/`stable`。

## 版本化与兼容

- 破坏性变更（改字段号、改语义、删字段）MUST 升 `version` 并给出双跑迁移方案，MUST NOT 原地破坏调用方。
- 字段号一经使用不复用；已废弃字段保留编号、标注 `deprecated`。
- 消费者 MUST 容忍新增可选字段；事件（异步）遵循 CloudEvents 1.0 必需属性。

## 契约测试（宪法原则 V）

- 提供方与调用方各自持有基于同一份 `.proto` 的契约测试，纳入 CI。
- 破坏性变更检测：CI MUST 运行 proto 兼容性检查（如 buf breaking），检测到破坏性变更且无版本方案 MUST 阻断合并。

## 首段切片 proto

以下三个 proto 是首段切片（identity / relationship / message）的服务间契约骨架，定义 RPC 与消息形状。
完整字段与实现细节在各服务 feature 的 `plan.md` 细化，但 RPC 签名与消息编号一经此处确定即作为契约基准。

| proto | 服务 | 稳定性 |
|---|---|---|
| [identity/v1/identity.proto](./identity/v1/identity.proto) | identity-svc | alpha |
| [relationship/v1/relationship.proto](./relationship/v1/relationship.proto) | relationship-svc | alpha |
| [message/v1/message.proto](./message/v1/message.proto) | message-svc | alpha |
