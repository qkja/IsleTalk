# identity-svc 身份服务

IsleTalk 的账号与身份限界上下文服务，承接 `specs/002-wechat-style-chat` 架构中 identity-svc 的职责，是首段切片（MVP）的第一块基石。

## 职责

- 手机号注册、密码登录、验证码登录（辅）、找回密码、退出登录
- 多设备会话管理与远程下线、改密全失效
- 公开账号名（全局唯一、保留词校验、改名冷却 30 天、保留期 90 天）与不可变内部账号 ID
- 资料隐私可见性（昵称/头像/签名/可见范围/陌生人添加开关）与界面语言偏好
- 可选实名登记与信任等级（收集 + AEAD 加密存储，不比对权威数据源做真实验证）
- 主动注销与数据处置

## 架构

服务内 DDD 四层，依赖单向向内：`api → application → domain → infrastructure`。

| 层 | 目录 | 内容 |
|---|---|---|
| api | `api/grpc` | 服务间 gRPC（identity.proto 5 个 RPC） |
| api | `api/http` | 客户端信封接口 + 健康检查 |
| application | `application/service` | 用例编排（注册/登录/会话/账号名/资料/实名/注销） |
| domain | `domain/` | 实体、值对象、领域服务、port 接口（零框架依赖） |
| infrastructure | `infrastructure/` | MongoDB 仓储、Redis 缓存/限流、Argon2id、令牌、mTLS、日志 |

## 启动方式

```bash
cd services/identity
go build ./...
# 运行需注入 MongoDB/Redis 连接与密钥（见下方配置项）
```

本地依赖（MongoDB + Redis）用 Docker Compose 启动，镜像版本固定。

## 依赖

- **MongoDB**：`identity` 独占 database（6 个集合，集合名带 `identity` 前缀）
- **Redis**：会话令牌、验证码、限流——非唯一事实来源，故障只降级不丢数据
- **短信**：本地 mock（`infrastructure/auth/sms_mock.go`），生产接云短信（经 port 隔离）

## 配置项（环境变量）

| 变量 | 说明 | 默认 |
|---|---|---|
| `MONGODB_URI` | MongoDB 连接串 | `mongodb://localhost:27017` |
| `REDIS_ADDR` | Redis 地址 | `localhost:6379` |
| `IDENTITY_ID_ENC_KEY` | 身份信息 AEAD 加密密钥（32 字节） | 本地开发占位，勿用于生产 |
| `IDENTITY_TOKEN_SECRET` | 令牌签发密钥 | — |

密钥 MUST 经环境变量或密钥管理注入，MUST NOT 硬编码（宪法 E 组）。

## 健康检查

- `GET /healthz` — 存活探针（进程活着即 200，MUST NOT 探测依赖）
- `GET /readyz` — 就绪探针（可接流量即 200，依赖探测待装配时注入）

## 测试

```bash
cd services/identity
go test ./...          # 单元 + 契约 + 安全；集成测试无容器时自动 skip
go test -cover ./...   # 覆盖率，CI 门禁 ≥80%
```

集成测试需本地 MongoDB/Redis 容器（`MONGODB_URI` / `REDIS_ADDR` 指向本地实例）。

## 契约

- **服务间 gRPC**：`contracts/proto/identity/v1/identity.proto`（002 定义，本服务实现不改契约）
- **客户端信封**：`{success, data, message, meta}`（统一信封，错误码见 `contracts/error-registry.md`）

## 相关文档

- 规格：`specs/003-identity-svc/spec.md`
- 技术方案：`specs/003-identity-svc/plan.md`
- 数据模型：`specs/003-identity-svc/data-model.md`
- 验证指南：`specs/003-identity-svc/quickstart.md`
