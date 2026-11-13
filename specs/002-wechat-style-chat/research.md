# Research: IsleTalk 即时通讯应用 — 技术选型决策

**Feature**: `specs/002-wechat-style-chat` | **Date**: 2026-11-13

本文是 Phase 0 输出，把 plan.md Technical Context 里的每个未知项解决为有依据的决策。
格式统一为：**Decision / Rationale / Alternatives considered**。

## R-1 后端实现语言：Go

- **Decision**: 全部 7 个后端服务统一 Go（1.22+）。
- **Rationale**:
  1. 宪法 A 组「同一类能力 MUST 语言一致」——后端是同类能力，统一 Go 满足单一工具链、单一运维手法。
  2. goroutine 并发模型适配万级并发在线长连接与群消息扇出；gRPC/protobuf 在 Go 生态最成熟。
  3. 仓库已有 `go-service-scaffold`（DDD 四层 + Google Wire）、`go-unit-test` 技能，直接复用，零迁移成本。
- **Alternatives considered**:
  - Java/Spring：生态大但并发模型与内存占用不占优，长连接扇出需额外异步框架，运维重。
  - Node.js/TypeScript：动态类型在消息/契约密集场景风险高，单机吞吐与内存效率弱于 Go。
  - Rust：性能强但团队无既有资产、迭代成本高，违背「最适配」而非「最新」原则。
- **结论**: Go 是唯一同时满足"并发适配 + 契约成熟 + 团队已有资产 + 单一工具链"四项的选项。

## R-2 服务框架与依赖注入

- **Decision**: go-zero（或等价）服务框架 + Google Wire 显式依赖注入。
- **Rationale**: 与既有技能模板一致；go-zero 内建限流/熔断/日志/链路，减少自研；Wire 是编译期依赖注入，
  满足宪法「依赖注入 MUST 显式、组合根唯一」。
- **Alternatives considered**: 手写 net/http + grpc-go（框架能力需自研，违反复用优先）；Spring 系（语言不符）。

## R-3 消息队列产品

- **Decision**: 方向确定，最终产品在首个使用它的服务 feature 锁定。候选排序：**NATS JetStream > Kafka > RabbitMQ**。
- **Rationale**: 本项目的异步事件是「好友申请、群成员变更、会话偏好同步、处置生效」等**低吞吐、高可靠、
  需至少一次投递 + 消费者幂等**的场景，且万级并发在线规模远未到 Kafka 的必要量级。NATS JetStream 部署轻、
  Go 原生、够用且运维成本最低；Kafka 留作吞吐真超限时升级路径。
- **Alternatives considered**: Kafka（强但重，运维成本高，本规模杀鸡用牛刀）；RabbitMQ（成熟但路由模型与
  流式回放弱于 NATS JetStream）；Redis Stream（宪法明文禁止用 Redis 当消息队列）。
- **备注**: 宪法要求消费者幂等、事件带 schema 与版本（CloudEvents），与具体产品无关，一律遵守。

## R-4 媒体面（语音/视频通话）

- **Decision**: SFU 用 **LiveKit**；通话信令由 `call-svc` 编排，媒体流经 LiveKit。
- **Rationale**: LiveKit 是 Go 实现、开源、支持 SFU 多路转发 + WebRTC，与后端语言统一，运维心智一致；
  避免自研媒体协商/NAT 穿透/弱网恢复（复杂度远高于消息，见 spec User Story 7 的 Why this priority）。
- **Alternatives considered**: Janus（C，语言不一致）、mediasoup（Node.js，语言不一致）、
  自研 WebRTC 媒体服务（宪法「优先复用」，且媒体子系统自研风险不可接受）。
- **边界**: 通话记录（FR-039）存 `call-svc` 的 MongoDB；媒体流不落盘（FR-042，不提供录制），
  质量排查片段去标识化、≤7 天、不进举报证据。

## R-5 对象存储（图片/文件消息）

- **Decision**: 本地开发用 **MinIO**（S3 兼容），生产接云对象存储（S3 兼容接口），`message-svc` 只存对象 key。
- **Rationale**: 图片/文件消息二进制不进 MongoDB；S3 兼容接口让本地/生产可互换，预签名 URL 上传下载。
- **Alternatives considered**: 直接存 MongoDB GridFS（违反"单文档不逼近 BSON 上限"，且失去 CDN/缩略图能力）。

## R-6 推送与短信

- **Decision**: 系统推送 APNs（iOS）+ FCM（Android）；短信用云短信服务商，本地开发用 mock。
- **Rationale**: 双端系统推送是移动端通知的唯一可靠通道（FR-032/034）；短信仅用于验证码（FR-001），
  本地开发不发真短信。
- **风险**: 境外短信/推送的合规（PIPL/GDPR）在 `plan.md` Assumptions 已留触发条件评估。

## R-7 可观测性

- **Decision**: OpenTelemetry（日志/指标/追踪）+ Prometheus/Grafana + Loki（日志）+ Tempo（链路）。
- **Rationale**: 宪法 D 组要求结构化 JSON 日志、四金信号、SLO 燃烧率告警、trace context 跨服务传递；
  OpenTelemetry 是标准语义约定，避免自造属性。
- **告警**: 燃烧率告警（14.4×/1h+5m 快、6×/6h+30m 慢），每个告警配 runbook。

## R-8 数据保留与可撤回窗口（宪法 B 组 + spec Assumptions 待定项的默认值）

- **Decision**: 以下数值在 plan 阶段给出保守默认值，最终值在 `message-svc` 的 feature plan 锁定并对用户可见：
  - 消息保留期：**默认 180 天**（到期软删除 + 备份处置按宪法 B 组恢复演练要求）
  - 撤回窗口：**2 分钟**（服务端时钟判定，MUST NOT 依赖客户端本地时间，对应 Edge Case）
  - 禁言最长时长：**7 天**（超长禁言必须走 `moderation-svc` 处置）
  - 账号名保留期：**90 天**（FR-078 默认值，冷却期 30 天，FR-077）
  - 身份信息删除：撤回实名或注销后 **30 天**内完成（SC-033，法定留存除外）
- **Rationale**: 这些是可调参数，给默认值避免阻塞；均为"对用户可见"的运营参数，非技术常量。

## R-9 分页与错误语义（宪法 B 组）

- **Decision**: 列表分页遵循 AIP-158（`page_size` 默认 50、上限 1000、超限收敛不报错、负值 `INVALID_ARGUMENT`、
  不透明 `page_token`、翻页排序稳定）；错误遵循 AIP-193 + RFC 9457（统一结构、稳定错误码、简短可执行消息）。
- **Rationale**: 宪法 B 组明文要求，直接落地为 `contracts/envelope.md` 与 `error-registry.md`。

## 未决项（Deferred，不阻塞本计划）

| 项 | 现状 | 何时定 |
|---|---|---|
| 消息队列最终产品 | NATS JetStream 方向 | 首个使用它的服务 feature plan |
| 对象存储生产产品 | MinIO/S3 方向 | `message-svc` feature plan |
| 生产编排/部署平台 | TODO(TECH_STACK) | 首个服务上线前，MINOR 登记宪法 |
| 服务发现与配置方案 | TODO(TECH_STACK) | 首个服务上线前，MINOR 登记宪法 |
| 契约测试工具 | buf | 首段切片 feature plan 锁定 |
| Go 工具链宪法登记 | 见 R-1 | 任一服务合入主干前，MINOR 修订宪法 |
