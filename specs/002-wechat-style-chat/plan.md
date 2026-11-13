# Implementation Plan: IsleTalk 即时通讯应用（类微信）

**Branch**: `feature/002-wechat-style-chat` | **Date**: 2026-11-13 | **Spec**: [spec.md](./spec.md)

**Input**: Feature specification from `specs/002-wechat-style-chat/spec.md`

## Summary

本 feature 将"类微信"的即时通讯应用从产品规格（85 条 FR、10 个用户故事）落地为**可实施的微服务架构**。
本计划做四件事：

1. **拆分服务**：按业务边界把产品能力集合拆成 7 个微服务（限界上下文），并给出拆分理由与数据边界。
2. **确定技术栈**：统一选定实现语言与关键组件（gRPC、消息队列、对象存储、推送、短信、媒体服务），
   满足宪法 A 组「语言」的选型留痕要求。
3. **固定契约与数据模型**：统一信封、错误码注册表、proto 包约定、实体→服务归属映射、存储与缓存设计。
4. **划定首段落地切片**：按 P1 优先级定出第一个可独立交付的纵向切片（注册登录 + 好友 + 1v1 单聊）。

**关键结论**：本计划只产出架构与契约，不直接实现任何单个服务。依宪法原则 V「新增微服务 MUST 有对应
spec.md 支撑」，7 个服务将各开一个 feature 目录（`003-*` 起）分别承接，本 feature 的 `tasks.md`
只覆盖架构资产与首段切片的服务编排，不含各服务的业务实现。

## Technical Context

**Language/Version**: Go 1.22+（后端全部服务）；Swift（iOS 客户端）；Kotlin（Android 客户端）；
TypeScript（Web 账号安全与申诉页，仅此一处前端，非后端语言）。

**Primary Dependencies**:
- 后端：gRPC + protobuf、go-zero 或等价服务框架（见 research.md）、Google Wire（依赖注入）、
  OpenTelemetry（日志/指标/追踪）
- 客户端：iOS 用原生 Swift + WebRTC；Android 用原生 Kotlin + WebRTC
- 媒体：LiveKit（SFU，Go 实现）承载语音/视频通话媒体面；信令由 `call-svc` 编排

**Storage**: MongoDB（每服务独占 database）；对象存储（图片/文件消息）；Redis（缓存/锁/限流/会话）。

**Testing**: Go 用 `go test` + testify + 契约测试（buf/grpc）；覆盖率 `go test -cover` 汇入 CI（≥80%）。
iOS 用 XCTest，Android 用 JUnit + Espresso。契约测试以 `.proto` 为唯一事实来源。

**Target Platform**: 后端 Linux 容器（Docker Compose 本地、生产编排见 TODO(TECH_STACK)）；
客户端 iOS 15+ / Android 8.0+；Web 现代浏览器（仅账号安全与申诉）。

**Project Type**: 微服务后端 + 双原生移动端 + 极薄 Web 前端。

**Performance Goals**（对齐 SC）: 在线消息送达 p95 ≤ 2s（万级并发在线）；1v1 语音接通率 ≥95%；
语音端到端延迟 p95 ≤ 400ms、视频 ≤ 500ms；会话列表（100 会话）可见 ≤1.5s。

**Constraints**:
- 消息零丢失、零重复、零乱序（SC-004/005/024）
- 万级并发在线 + 十万级注册（SC-022/023）
- 敏感数据（消息正文、手机号、身份信息、通话记录）不出现在日志/指标/追踪/错误响应（FR-051/071）
- 不做事前正文风控（FR-079），不提供通话录制（FR-042），不提供消息全文检索（FR-085）

**Scale/Scope**: 7 个后端微服务、2 个原生客户端、1 个薄 Web 前端；首段切片覆盖 3 个服务 + 2 客户端。

## Constitution Check

*GATE: 已通过。逐条核对如下；无未论证的违规。*

| 宪法条目 | 判定 | 落地方式 |
|---|---|---|
| 原则 V 服务拆分按业务边界 | ✅ | 7 个限界上下文，见「服务拆分」；禁止技术角色切分 |
| 原则 V 数据服务独占 | ✅ | database per service，集合名带服务前缀；跨服务经 gRPC/事件 |
| 原则 V gRPC 同步 + proto 契约 | ✅ | 全部服务间调用 gRPC；`.proto` 版本化 + 契约测试 |
| 原则 II 覆盖率 ≥80% | ✅ | Go 覆盖率 CI 门禁；domain/application 无 IO 单测 |
| 原则 III 统一信封 | ✅ | `{success,data,message,meta}`；错误码注册表见 contracts/ |
| A 组 语言选型留痕 + 登记 | ✅ | Go 全后端统一；工具链已登记于宪法 v2.7.0（gofmt/golangci-lint/go vet/go mod/testify/go test -cover/buf + go-zero/Wire） |
| A 组 MongoDB 默认 + 索引 | ✅ | 每服务独占 database；高频查询路径建索引 |
| B 组 分页/错误/幂等 | ✅ | AIP-158 分页、AIP-193 错误、`Idempotency-Key` |
| C 组 测试金字塔 + 性能预算 | ✅ | 单测为主；消息/通话性能预算见 SC |
| D 组 SLI/SLO + 燃烧率告警 | ✅ | 每个服务定义四金信号 + SLO；告警带 runbook |
| E 组 安全（OAuth2.1/Argon2id/mTLS） | ✅ | 见 contracts/ 与 research.md；服务间 mTLS 100% |
| F 组 分支/squash/发布 | ✅ | 沿用仓库既有保护；各服务实现走独立 feature |
| G 组 WCAG 2.2 AA | ✅ | 原生移动端口径已登记于宪法 v2.7.0（平台原生无障碍规范 + WCAG 判据逐条复核，自动化检查覆盖 Web 端） |

## 服务拆分（限界上下文）

按业务边界拆为 7 个服务。每个服务 = 一个清晰业务能力 + 独占数据 + 独立部署。

| # | 服务名 | 业务边界 | 归属的 spec 实体 | 首段切片 |
|---|---|---|---|---|
| 1 | `identity-svc` | 账号与身份 | 用户账号、设备会话、界面语言偏好、实名信息、信任等级、账号名占用 | ✅ 首段 |
| 2 | `relationship-svc` | 好友与关系 | 好友关系 | ✅ 首段 |
| 3 | `message-svc` | 会话与消息 | 会话、消息、阅读位置、通知订阅（会话级偏好） | ✅ 首段（仅 1v1） |
| 4 | `group-svc` | 群组 | 群聊、群成员身份 | 后续 |
| 5 | `notification-svc` | 通知推送 | 通知订阅（推送级）、离线提醒 | 后续 |
| 6 | `call-svc` | 语音/视频通话 | 通话 | 后续 |
| 7 | `moderation-svc` | 治理与运营 | 举报与处置记录、运营人员账号 | 后续 |

**拆分理由**（为什么不是 3 个、也不是 10 个）：

- **身份 / 关系 / 消息必须分开**：三者的读写频率、一致性要求和生命周期完全不同。身份是低频高安全
  （登录、令牌、实名），关系是中等频率的图结构，消息是高频低延迟。绑在一起会互相拖累扩缩容与故障域。
- **群组独立于消息**：群成员/角色/禁言是"谁在群里、能不能说话"的治理域，与"消息怎么传输"是两件事。
  `message-svc` 投递群消息时经 gRPC 向 `group-svc` 查成员与禁言状态，而不是自己维护一份成员表——
  否则成员变更会同时存在于两处，违反单一事实来源。
- **治理必须独立**：`moderation-svc` 持有举报/处置/审计，其访问者（运营人员）与普通用户完全不同，
  审计留痕不可由应用侧篡改（FR-063）。并入任何业务服务都会把"治理权限"混进"业务权限"。
- **通话独立**：媒体面（SFU）与信令是独立子系统，生命周期（会话建立/结束）与消息无关，且故障域必须隔离
  （通话挂了不能带崩消息）。
- **通知独立**：推送依赖 APNs/FCM 外部通道与离线逻辑，是与消息投递不同的责任；但会话级免打扰偏好与
  `message-svc` 的会话列表强相关，边界见下。

**边界上的两条显式决策**（避免含糊）：

1. **会话级通知偏好放 `message-svc`**（置顶、免打扰、标记未读，FR-024），因为它与会话列表同一读路径；
   推送投递与隐私模式（FR-033）放 `notification-svc`。两服务以「会话偏好」事件同步，最终一致。
2. **群消息的授权与扇出**：`message-svc` 收群消息时向 `group-svc` 校验「发送者是否为成员、是否被禁言」，
   取得成员列表做扇出。该调用带超时 + 熔断，成员列表可短时缓存（TTL ≤30s），缓存失效不得放行越权发送。

**共享基础设施（不是业务服务，不入微服务清单）**：对象存储（图片/文件）、SFU（LiveKit）、消息队列、
Redis、推送通道（APNs/FCM）、短信通道。它们的接入方式见 research.md。

## FR → 服务承接表（需求逐条归属）

本节是 `/speckit-analyze` 发现 C1/I1 的修复：把 spec 的 85 条 FR **逐条**归属到承接方，
供 `003-*` 各服务 feature 作为 spec 输入清单，消除"某条 FR 无人认领"的遗漏风险。
「横切」= 全项目约束，由各服务共同遵守，不是独立服务。

| FR 编号 | 主责服务 | 跨服务协作 / 备注 |
|---|---|---|
| FR-001~006 账号与身份 | identity-svc | — |
| FR-007~012 好友与关系 | relationship-svc | FR-007 的"按公开账号名搜索"动作由 identity-svc 的 `SearchByPublicName` 提供；FR-011 拉黑后消息拦截需 message-svc 配合 |
| FR-013 单聊会话 | message-svc | 建会话依赖 relationship-svc 的好友校验（FR-009） |
| FR-014~015 群创建/成员/角色/转让 | group-svc | 群聊会话容器由 message-svc 建；成员变更经事件同步 |
| FR-016~026 消息与会话列表 | message-svc | FR-025 联系人搜索跨 relationship-svc；FR-026 多端同步为 message-svc 职责 |
| FR-027~031 送达/可靠/离线 | message-svc | — |
| FR-032~034 通知 | notification-svc | 依赖 message-svc 的会话级免打扰偏好 |
| FR-035~042 语音视频通话 | call-svc | 媒体面走 LiveKit（共享基础设施）；FR-042 质量片段去标识化 |
| FR-043~045 群角色/禁言/上限 | group-svc | — |
| FR-046~048 处置/举报校验 | moderation-svc | FR-046 群内禁言动作下发给 group-svc；FR-048 昵称签名校验走 identity-svc、群名公告走 group-svc |
| FR-049~050 隐私开关/举报 | identity-svc（FR-049）、moderation-svc（FR-050） | FR-049 陌生人添加开关由 relationship-svc 应用 |
| FR-051~052 日志不泄露/敏感数据 | 横切（全部服务） | FR-052 主责 moderation/call/message |
| FR-053~055 范围边界 | 横切（全部服务） | 不端到端加密（FR-054）、不并列多套实现（FR-055）为全项目约束 |
| FR-056~058 客户端形态 | clients/ios + clients/android + clients/web | 不在 7 个后端服务内 |
| FR-059~061 接入与规模 | identity-svc（FR-059/060）、横切（FR-061） | 注册登录风控主责 identity；FR-061 为全服务容量要求 |
| FR-062~064 运营后台/审计/申诉 | moderation-svc | — |
| FR-065~068 国际化/无障碍 | 客户端（横切） | FR-066 服务端文案同样外置 |
| FR-069~074 实名与敏感个人信息 | identity-svc | — |
| FR-075~078 标识与发现 | identity-svc | — |
| FR-079~080 内容风控口径 | message-svc（FR-079 执行）、moderation-svc（口径）、横切（FR-080 告知） | FR-079"正文不事前拦截"由 message-svc 保证 |
| FR-081~082 群成员准入 | group-svc | — |
| FR-083~084 已读回执 | message-svc | — |
| FR-085 检索范围 | 横切（另开独立 feature） | 本版本不做，未来 feature 承接 |

**覆盖校验**：85 条 FR 全部归属（6+6+1+2+11+5+3+8+3+3+2+2+3+3+3+3+4+6+4+2+2+2+1 = 85）。
每个 `003-*` 服务 feature 的 spec MUST 以本表对应行作为承接范围，不得漏认领。

## 数据边界与存储

- **database per service**：每服务独占一个 MongoDB database，集合名带服务前缀
  （如 `identity.users`、`msg.messages`、`rel.friendships`）。跨服务取数 MUST 经对方 gRPC 接口或事件，
  禁止直连他方集合（宪法原则 V）。
- **消息正文存储**：`message-svc` 独占 `msg.messages`（按会话分片），正文与附件（对象存储 key）分离，
  正文不进日志/指标/追踪。
- **图片/文件**：走对象存储，`message-svc` 只存对象 key 与元数据；上传经预签名 URL，下载带短期签名。
- **Redis 用途**：`message-svc` 用缓存（会话列表/最近消息）、`identity-svc` 用会话令牌与会话状态、
  全服务用分布式锁（账号名占用、群主转让等）与限流。Redis 一律不是唯一事实来源。
- **缓存 TTL 与键隔离**：按服务命名空间分 key（`msg:`, `identity:` 等）；会话列表缓存 TTL ≤5s（保证未读
  计数偏差为 0，SC-010）；群成员列表缓存 TTL ≤30s 且失效后 fail-closed 于越权操作。

## 技术选型与工具链

**实现语言：Go（全部 7 个后端服务）**

- **决策**：后端统一 Go。理由：① 宪法 A 组要求"同一类能力语言一致"，后端全用 Go 满足单一工具链；
  ② Go 的 goroutine 天然适配万级并发长连接与消息扇出；③ gRPC/protobuf 在 Go 生态最成熟；
  ④ 仓库既有 `go-service-scaffold` / `go-unit-test` 技能与 go-zero + Wire 分层模板，直接复用。
- **被否决**：Java/Spring（并发模型与内存占用不占优）、Node.js（长连接与扇出的类型安全弱）、
  Rust（学习与迭代成本高，团队无既有资产）。单服务内严禁混用多语言。
- **工具链登记（待办）**：格式化 gofmt、静态检查 golangci-lint、类型检查 go vet、
  依赖管理 go mod + vendor/lock、测试 go test + testify、覆盖率 go test -cover、
  契约测试 buf + grpcurl。**该登记 MUST 以 MINOR 版本修订宪法 A 组后，各服务才可合入主干**。

**关键组件选型**（详见 research.md）：

| 组件 | 选择 | 说明 |
|---|---|---|
| 服务框架 | go-zero（或等价） | 与既有技能模板一致，内建限流/熔断/日志 |
| 服务间通信 | gRPC + protobuf | 宪法原则 V 强制 |
| 消息队列 | 待定（Kafka/RabbitMQ/NATS 三选一，research.md 定） | 异步事件：好友申请、成员变更、会话偏好同步 |
| 媒体 SFU | LiveKit | Go 实现，与后端语言统一；承载通话媒体面 |
| 对象存储 | 待定（MinIO 本地 / 云对象存储生产） | 图片/文件消息 |
| 推送 | APNs + FCM | 双端系统推送 |
| 短信 | 云短信服务商（本地开发用 mock） | 验证码下发 |
| 可观测 | OpenTelemetry + Prometheus/Grafana + Loki/Tempo | 四金信号 + 链路 |

**客户端**：iOS（Swift + WebRTC）、Android（Kotlin + WebRTC）、Web（TypeScript，仅账号安全与申诉）。

## 契约与信封（落地形式）

- **对外（客户端↔后端）**：采用统一 JSON 信封 `{success, data, message, meta}`（宪法原则 III），
  经 HTTP/2 + TLS；字段命名 camelCase。移动端与 Web 共用同一信封与错误码。
- **对内（服务↔服务）**：gRPC + protobuf，`api/*.proto` 版本化，包名标注稳定性（alpha/beta/stable）。
- **错误码**：唯一注册表，同义错误同 code；可重试/不可重试标注。详见 `contracts/error-registry.md`。
- **幂等**：对外写接口支持 `Idempotency-Key`（保留 24h）；消息投递以消息 ID 去重（FR-028）。
- 详见 `contracts/`。

## 测试策略与覆盖率预期

- **测试金字塔**：单元测试为主（domain/application 无 IO、数分钟内跑完数千条）；集成测试针对容器化
  依赖；E2E 覆盖 P1 四条核心路径（注册登录/加好友/收发消息/会话列表）。
- **覆盖率 ≥80%**（宪法原则 II）：Go 服务 `go test -cover` 汇入 CI 硬门禁；关键路径（鉴权、权限、
  消息写入、撤回、拉黑）接近 100% 分支覆盖，含负向用例。
- **契约测试**：每个服务提供方与调用方各自持有基于同一份 `.proto` 的契约测试，纳入 CI（宪法原则 V）。
- **幂等与可靠性测试**：消息重投不重复、离线补发不遗漏、撤回一致（SC-004/009/024）作为 E2E 必测项。

## 首段落地切片（MVP）

按 spec 的 P1 优先级，首段交付一个可独立验证的纵向切片：

**范围**：`identity-svc` + `relationship-svc` + `message-svc`（仅 1v1 会话，不含群聊）+ iOS/Android 客户端
+ 对象存储 + 短信 + 推送（`notification-svc` 仅交付"新消息通知"这一最薄能力，其余 P2 通知能力后续）。

**对应验收**：US1（注册登录）、US2（好友）、US3（1v1 聊天）全部验收场景 + SC-001~SC-014、SC-022。

**后续切片**（各开独立 feature）：群聊（US4）、通知完整能力（US6）、语音视频通话（US7）、
治理与运营后台（US8/9）、国际化与无障碍完整化（US10）。

## 项目结构（本 feature 产出）

```text
specs/002-wechat-style-chat/
├── plan.md              # 本文件
├── research.md          # Phase 0：选型决策与备选
├── data-model.md        # Phase 1：实体→服务映射、字段、关系、状态机
├── contracts/
│   ├── envelope.md      # 统一信封与分页/错误语义
│   ├── error-registry.md# 错误码注册表
│   └── proto/           # 服务间契约约定 + 首段切片的 proto 定义
├── quickstart.md        # Phase 1：可运行的端到端验证场景
└── tasks.md             # Phase 2（/speckit-tasks 产出，本计划不生成）
```

```text
# 仓库根（各服务实现落地位置，由后续 feature 003+ 建立）
services/
├── identity/            # identity-svc（Go，DDD 四层）
├── relationship/
├── message/
├── group/
├── notification/
├── call/
└── moderation/

clients/
├── ios/                 # Swift
├── android/             # Kotlin
└── web/                 # 账号安全与申诉（薄前端）

deploy/
├── docker-compose.yml   # 本地依赖编排（MongoDB/Redis/队列/SFU/MinIO）
└── ...                  # 生产编排（TODO(TECH_STACK) 待定平台）
```

## Complexity Tracking

> 无宪法违规需要论证豁免。G 组 WCAG 自动化检查口径差异已在宪法 v2.7.0 登记解决，不构成本 feature 的复杂度豁免。

## 待办（本计划遗留、不阻塞后续但 MUST 完成）

1. ~~宪法 MINOR 修订~~ **已完成（v2.7.0）**：Go 工具链与原生移动端 WCAG 口径均已登记。
2. **消息队列 / 对象存储 / 编排平台 / 服务发现与配置方案选型**：research.md 给出方向与备选，最终产品在
   对应服务的 feature `plan.md` 里锁定，并在宪法 A 组 `TODO(TECH_STACK)` 逐项登记。
3. **各服务独立 feature**：`003-identity`、`004-relationship`、`005-message`（首段）……
   每个各产 spec/plan/tasks，作为该服务的实现入口。
