# Research: identity-svc 身份服务 — 技术决策

**Feature**: `specs/003-identity-svc` | **Date**: 2026-11-13

本文是 Phase 0 输出，解决 identity-svc 特有的技术决策。语言/框架/存储已由 002 与宪法 v2.7.0 确定，此处只补 identity 域内的决策。

## R-1 密码哈希

- **Decision**: Argon2id，参数 m=19456（19 MiB）/ t=2 / p=1，单次哈希 <1s；哈希格式带参数与盐（如 PHC 字符串格式），便于未来调参。
- **Rationale**: 宪法 E 组明文要求 Argon2id（推荐 m=19456/t=2/p=1）；带参数的格式使未来强度提升无需迁移全库。
- **Alternatives considered**: bcrypt（宪法允许 cost≥10，但 Argon2id 是首选）；scrypt（次选）；裸 SHA-256/MD5（宪法禁止）。

## R-2 令牌方案

- **Decision**: 访问令牌（access）≤15min、刷新令牌（refresh）≤7day 且每次使用后轮换、旧令牌立即失效；授权码 ≤10min。对外（客户端）走 OAuth 2.1 授权码 + PKCE；对内（服务间 gRPC）用 mTLS + 短时令牌校验 `aud` 含服务标识。
- **Rationale**: 宪法 E 组明文时限；OAuth 2.1 + PKCE 是宪法唯一允许的对外流（禁隐式流与 ROPC）。
- **Alternatives considered**: 无期限令牌（宪法禁止）；隐式流/ROPC（宪法禁止）；自研令牌格式（用标准 JWT + JWKS 校验，宪法要求）。

## R-3 手机号存储

- **Decision**: 手机号以加盐哈希（如 HMAC-SHA256，盐经密钥管理注入）存储为 `phone_hash`，MUST NOT 落明文；明文仅在验证码下发瞬间使用后即弃。
- **Rationale**: FR-022/051 与宪法 E 组要求手机号不进日志/指标/追踪、对外脱敏；哈希保留"同号唯一"能力（注册去重、登录定位）而不暴露明文。
- **Alternatives considered**: 明文存储（宪法禁止）；对称加密可逆（仍需管理密钥且解密面更大，哈希更简且满足唯一性需求）。

## R-4 账号名唯一性

- **Decision**: 三重保障——① `identity.users.public_name` 部分唯一索引；② `identity.name_claims.name` 唯一索引（含保留期条目，使旧名在保留期内仍占位）；③ 改名/注册时以 Redis 分布式锁（`identity:lock:name:{name}`，TTL 30s）串行化占用判定。
- **Rationale**: FR-013/014/015 要求全局唯一、保留期（默认 90 天）不可被抢占；唯一索引是硬保证，锁避免并发窗口，`name_claims` 承载保留期语义。
- **Alternatives considered**: 仅靠唯一索引（无法表达"保留期"）；仅靠应用层查重（并发窗口下会漏）。

## R-5 验证码

- **Decision**: 验证码存 Redis（`identity:verify:{phone_hash}`，TTL 5min 默认），校验后删除（一次性，FR-001）；重发冷却 60s（默认）；下发经 port 接口 `VerificationCodeSender`，本地 mock、生产接云短信。
- **Rationale**: 验证码是一次性、短时数据，适合 Redis TTL；一次性由"校验后删除"保证（Edge Case 重放）。
- **Alternatives considered**: 存 MongoDB（需额外过期清理，Redis TTL 更简）；无冷却（可被刷，违反 FR-006 风控）。

## R-6 会话管理

- **Decision**: 双写——MongoDB `identity.sessions` 为持久事实（会话记录、远程下线），Redis `identity:session:{sid}` 为热路径（鉴权快速校验、撤销即时生效）；撤销=删除 Redis + 标记 MongoDB；改密/注销=批量撤销该用户全部会话。
- **Rationale**: 多端同时在线 + 远程下线 + 改密全失效（FR-008~010）；Redis 提供低延迟鉴权，MongoDB 提供审计与恢复。
- **Alternatives considered**: 仅 Redis（重启丢会话记录，且审计无凭）；仅 MongoDB（鉴权热路径延迟高）。

## R-7 防枚举与限流

- **Decision**: 登录/注册/找回/验证码接口统一响应（不区分"未注册/密码错"）；限流按账号（phone_hash）+ IP 双维度（Redis 滑动窗口）；验证码下发独立限流。
- **Rationale**: SC-003 与宪法 E 组"登录、找回密码、验证码端点 MUST 按账号与 IP 维度限流"；防手机号枚举是 FR-016 的延伸。
- **Alternatives considered**: 单维度限流（宪法禁止）；区分注册状态响应（FR-003 禁止）。

## 未决项（Deferred）

| 项 | 现状 | 何时定 |
|---|---|---|
| 生产短信服务商 | 本地 mock | 002 待办，服务上线前 |
| 密钥管理服务 | 环境变量注入 | 002 待办，服务上线前 |
| 信任等级额度数值 | 默认值待定 | tasks 实现时论证（FR-026） |
