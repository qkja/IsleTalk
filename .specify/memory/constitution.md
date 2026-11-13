# IsleTalk Constitution

## Core Principles 核心原则

### I. Code Quality First 代码质量优先

代码质量是首要目标，不以交付速度牺牲可读性、可测试性与安全性。

- 每项功能在编码前 MUST 有明确的需求规格与实现计划；需求未澄清时 MUST 先澄清，不得凭猜测编码。
- 代码 MUST 通过仓库配置的格式化、静态检查、类型检查与构建；所有自动化检查全绿是提交的前提。
- 函数 MUST 保持单一职责：单函数不超过 50 行、嵌套不超过 4 层；文件 SHOULD 控制在 200–400 行，800 行为硬上限，超限 MUST 拆分模块。
- 错误 MUST 显式处理，不得静默吞掉；面向使用者的错误信息要可读可行动，日志侧保留完整上下文。
- 所有外部输入（请求参数、第三方响应、文件内容）MUST 在系统边界校验后再处理，不得信任外部数据。
- 密钥、令牌、口令 MUST 通过环境变量或密钥管理注入，MUST NOT 硬编码；启动时 MUST 校验必需密钥存在（细则见 E 组）。
- MUST NOT 提交调试残留（调试打印、断点语句、注释掉的死代码、临时测试开关）。
- 新增抽象、新增依赖、新增模式 MUST 有书面理由，说明它解决的真实痛点；无对应痛点时按 YAGNI 拒绝引入。
- 数据优先采用不可变模型与纯函数；确需共享可变状态时 MUST 显式声明同步与隔离策略。

**Rationale**：IsleTalk 是一个新项目，代码库将长期演进。早期在质量上的妥协会以指数级复利——低可测试性的代码会让后续每个功能都变慢，低覆盖率的代码会让缺陷在生产环境而非测试环境暴露。上述每一条都是可被 lint、测试或评审自动验证的硬约束，不依赖个人自觉。

### II. Test Coverage Floor 测试覆盖率不低于 80%

测试覆盖率是硬性下限门禁，不是参考指标。

- 全仓库测试覆盖率 MUST ≥ 80%（按行/语句统计），由 CI 强制执行；低于阈值即阻断合并。
- 覆盖率是下限而非上限。关键路径（鉴权、权限、支付、金额、数据写入、外部集成、错误处理分支）REQUIRE 接近 100% 的分支覆盖，且 MUST 包含负向用例（错误输入、边界值、依赖失败）。
- 每个缺陷修复 MUST 附带回归测试：先写能复现缺陷的失败测试（RED），再改实现（GREEN），最后重构（IMPROVE）。
- 新增或修改的行为 MUST 先有测试再实现，严格遵循 Arrange-Act-Assert 结构与描述性测试命名。
- 测试类型 MUST 齐全：单元测试（函数与工具）、集成测试（接口与数据访问）、端到端测试（关键用户流程）。
- 覆盖率统计 MUST 显式排除生成代码、mock 桩、数据库迁移脚本等非业务代码，排除规则写在 CI 配置中，不接受手工豁免。
- MUST NOT 通过删除测试、弱化断言强度、添加 `skip`/`ignore`、或扩大排除规则来提高覆盖率数字。
- 测试 MUST 相互隔离：不得依赖执行顺序、真实网络、真实生产数据或本地机器状态。

**Rationale**：80% 阈值常被当作可协商的目标，实际结果是无人协商即被突破。把阈值挂在 CI 门禁上，并同时禁止"靠删测试刷数字"这条最常见的作弊路径，才能让覆盖率数字持续反映真实风险。关键路径的高覆盖要求补足了行覆盖率对分支缺陷不敏感的固有盲区。

### III. Unified API Response Format API 返回格式统一

所有对外接口 MUST 使用同一套返回信封与错误语义，禁止接口各自为政。

- 每个对外接口 MUST 返回同一信封结构，顶层字段集合保持一致，不允许个别接口自定义顶层字段：

  ```json
  {
    "success": true,
    "data": { },
    "message": null,
    "meta": { "total": 0, "page": 1, "limit": 20 }
  }
  ```

  - `success` (bool) — 成功标识，必填。
  - `data` (object | null) — 业务数据，必填；失败时为 `null`。
  - `message` (string | null) — 面向调用方的错误或提示信息，必填；成功时可为 `null`。
  - `meta` (object | null) — 分页与元信息，分页接口必填（至少含 `total`、`page`、`limit`），非分页接口可为 `null`。

- 传输协议与字段命名的最终形态由每个 feature 的 spec / plan 阶段确定并在计划中写明；一经确定即全项目统一，MUST NOT 并存多种响应风格。
- 传输层成功与业务成功 MUST 分离：传输层返回成功状态时，业务失败仍 MUST 通过 `success=false` 加稳定的错误码表达。
- 错误码 MUST 来自唯一注册表，同一语义的错误在不同接口 MUST 使用相同 code 与 message；注册表随代码演进，废弃的错误码 MUST 标记而非复用。
- 列表与查询接口 MUST 分页，MUST NOT 无界返回；单次返回量 MUST 有上限（分页细则见 B 组）。
- 对外接口 MUST 有入参校验与速率限制；状态变更接口 MUST 有鉴权与 CSRF 防护（Web 场景）。
- 契约演进 MUST 有测试与版本管理：MUST 先更新契约测试（纳入 CI），破坏性变更 MUST 版本化并在 plan 中写明迁移方案（细则见 B 组）。

**Rationale**：统一的返回格式让客户端可以写一次解析逻辑处理所有接口，也让错误处理集中在一处；错误码注册表则把"这个错误是什么意思"从各接口的实现细节提升为可检索的系统级事实。传输层不绑定具体协议，是为了让原则在任何技术栈下都成立——协议一旦由 feature 计划确定，就必须严格统一。

### IV. Spec-Driven Development (NON-NEGOTIABLE) 规格驱动开发

所有需求 MUST 严格走 Spec Kit 标准流程，MUST NOT 跳过任何阶段。

- 每一个需求（功能、缺陷修复、重构、以及任何非平凡改动）MUST 以 `/speckit-specify` 生成 `spec.md` 为起点，并归属到独立的 feature 目录。
- 标准流程顺序 MUST 严格为：**specify → clarify → plan → tasks → implement → analyze**。MUST NOT 跳步；MUST NOT 在缺少上游产物的情况下进入下游阶段。
- 每个阶段 MUST 使用对应的 Spec Kit 命令，产出受版本管理的产物：

  | 阶段 | 命令 | 产物 |
  |---|---|---|
  | Specify | `/speckit-specify` | `spec.md` — 功能规格与验收标准 |
  | Clarify | `/speckit-clarify` | 回写至 `spec.md` — 消除歧义 |
  | Plan | `/speckit-plan` | `plan.md` — 技术方案、契约、测试策略 |
  | Tasks | `/speckit-tasks` | `tasks.md` — 依赖有序、可验证的任务清单 |
  | Implement | `/speckit-implement` | 按 `tasks.md` 执行的实现 |
  | Analyze | `/speckit-analyze` | spec / plan / tasks 一致性与质量复核报告 |

- 规格中存在歧义时 MUST 执行 `/speckit-clarify`；未澄清完毕 MUST NOT 进入 plan 阶段。
- `/speckit-analyze` MUST 在合并前运行一次；其发现的问题 MUST 在合并前修复，或按豁免流程书面说明理由。
- `spec.md` MUST 含可验证的验收标准（Given/When/Then 或等价形式），并显式写明范围内与范围外。
- `plan.md` MUST 明确技术栈选型、API 契约落地形式（信封字段名）、测试策略与覆盖率预期。
- `tasks.md` 中每条任务 MUST 有可独立验证的完成标准；MUST NOT 存在"实现某功能"这类无法验证的条目。
- 实现阶段 MUST 逐条推进 `tasks.md`；MUST NOT 执行 `tasks.md` 之外的改动。范围外的内容 MUST 记为新任务并重新走流程。
- MUST NOT 以"先写代码、后补文档"的方式绕过本流程。因线上紧急修复需要先落地时，MUST 在同一分支内补齐 `spec.md` / `plan.md` / `tasks.md` 后方可合并。
- 实现与产物不一致时，以产物为准修正实现；确需变更产物 MUST 走修订流程并留下记录。
- 需求本身的澄清（`/speckit-clarify`）、检查清单（`/speckit-checklist`）、实现收敛（`/speckit-converge`）、任务转 Issue（`/speckit-taskstoissues`）按需选用，但 MUST NOT 替代上述主流程。

**Rationale**：Spec Kit 的价值不在于多写三份文档，而在于让"需求—方案—任务—代码"四者之间存在可机器校验的一致性约束。跳步最大的代价不是流程仪式感，而是实现与真实需求脱节后无法判断哪些代码是多余的、哪些行为无人验证过。把流程顺序写成不可协商的原则，意味着任何一次偏离都会在评审和 `/speckit-analyze` 环节被显式暴露，而不是被"顺手先做了"的惯性吞掉。

### V. Standard Layered Business Architecture 标准分层业务架构

业务架构 MUST 采用「微服务 + 服务内 DDD 四层分层」：系统由多个可独立部署的微服务组成，每个服务内部 MUST 采用 DDD 分层架构（api → application → domain → infrastructure）。

**一、服务内分层规则**

- **依赖方向 MUST 单向向内**：`api → application → domain`；`infrastructure` 实现 `domain` / `application` 定义的接口并由外层注入。`domain` 层 MUST NOT 依赖任何外层或任何框架（ORM、HTTP 框架、消息队列 SDK、缓存客户端、配置库不得出现在 `domain`）。
- **每层职责 MUST 明确且 MUST NOT 越界**：
  - `domain` — 实体、值对象、领域服务、领域事件、仓储与外部能力的接口（port）定义。 MUST NOT 包含 IO 调用、框架类型、跨用例编排。
  - `application` — 用例编排（application service）、事务边界、命令/查询对象、鉴权与权限判定入口、API 信封映射。 MUST NOT 实现业务规则本身（业务规则属于 `domain`）。
  - `infrastructure` — 仓储实现、第三方客户端、消息/缓存/文件适配器、序列化与存储细节、配置装配。
  - `api` — 协议适配：路由/handler、入参校验、响应封装、鉴权中间件、错误码映射。 MUST NOT 包含业务逻辑。
- 跨层调用 MUST 通过显式定义的接口（port）进行；MUST NOT 在上层直接 `new` 具体的基础设施实现。
- 依赖注入 MUST 显式；组合根（composition root）MUST 唯一且位于最外层（`api` 或 `infrastructure` 的装配入口）。 MUST NOT 使用服务定位器、全局单例容器或框架隐式注入来掩盖层间依赖。
- 领域模型 MUST 与存储结构解耦：持久化实体 MUST NOT 直接作为 API 响应体返回，ORM/表结构细节 MUST NOT 泄漏进 `domain`。
- 事务边界 MUST 定义在 `application` 层（一个用例 = 一个本地事务）；MUST NOT 开启跨服务分布式事务（禁止 2PC / XA）。
- 分层 MUST 可被测试验证：`domain` 与 `application` 层 MUST 能在无数据库、无网络、无框架容器的条件下完成单元测试（原则 II 的覆盖率统计必须覆盖这两层）。
- 新增能力 MUST 落在正确的层；MUST NOT 以"方便"为理由把业务逻辑写进 `api` 或 `infrastructure`，也 MUST NOT 把 IO 编排写进 `domain`。
- 架构范式在项目内 MUST 保持一致；MUST NOT 在同一仓库并存多套分层约定。

**二、服务边界与数据边界规则**

- **服务拆分 MUST 按业务边界（限界上下文 / 业务域）**：一个服务 MUST 对应一项清晰的业务能力；MUST NOT 以"user-service / order-dao / api-gateway"这类技术角色切分。
- **服务粒度 MUST 满足单一职责**：服务 MUST 只暴露一组内聚的业务能力；MUST NOT 出现"所有业务都往里塞"的通用服务。
- **每个服务 MUST 独立自治**：独立进程、独立部署、独立扩缩容；MUST NOT 依赖其他服务同时在线才能启动或就绪。
- **MUST NOT 为"将来可能需要"而预先拆分服务**：新增微服务 MUST 有对应的 `spec.md` 支撑，写明其独立的业务边界与拆分理由。
- **MUST NOT 共享内部状态**：服务间只经由契约交互，MUST NOT 共享内存、共享库表或共享缓存对象。

- **数据边界 MUST 服务独占（database per service）**：跨服务取数 MUST 经由对方接口或事件，MUST NOT 直连他方存储；隔离单位与命名规则见 A 组。
- **跨服务数据一致性 MUST 走最终一致性**：状态变更 SHOULD 通过事件/消息传播；订阅方 MUST 实现幂等消费；MUST NOT 用跨服务同步调用模拟分布式事务。

**三、服务间通信、观测与工程约束规则**

- **同步调用 MUST 使用 gRPC**：契约 MUST 以 `.proto` 定义并纳入版本控制；契约 MUST 由提供方与调用方共同持有，且双方 MUST 基于同一份 `.proto` 生成代码。
- **契约变更 MUST 向后兼容或显式版本化**：破坏性变更 MUST 通过 proto package 版本管理，并 MUST 给出双跑迁移方案与下线计划；MUST NOT 直接原地破坏调用方。
- **契约 MUST 有契约测试**：服务提供方与调用方 MUST 各自持有基于同一份 `.proto` 的契约测试并纳入 CI；MUST NOT 在无契约测试的情况下变更服务间接口。
- **跨服务调用 MUST 具备容错语义与幂等设计**：调用 MUST 设超时、熔断、降级与限流；重试 MUST 仅用于幂等操作；写操作 MUST 携带幂等键以抵御重复投递与重试；调用失败 MUST 显式处理，不得静默吞掉（细则见 D 组）。
- **MUST NOT 有跨服务的编译期或发布期强耦合**：一个服务的发布 MUST NOT 要求其他服务同步发布。

- **每个服务 MUST 可独立观测**：MUST 暴露健康检查与就绪检查，MUST 输出结构化日志、指标与链路追踪（细则见 D 组）。
- **架构边界 MUST 有自动化检查纳入 CI**：服务内依赖方向检查（禁止内层导入外层）+ 禁止跨服务直连他方数据库的检查；具体工具随技术栈选型确定（见 `TODO(TECH_STACK)`）。
- **若某 feature 需要跨层捷径（如为性能绕过用例层直连存储），MUST 在 `plan.md` 中书面论证、记录豁免与回归验证方式，MUST NOT 成为常态。**

**Rationale**：微服务把边界从代码目录提升为进程与网络边界——好处是各服务能独立演进、独立伸缩、独立失败；代价是边界一旦含糊，代价会以分布式事务、循环依赖和无法独立发布的连锁故障的形式一次性兑现。因此本原则的重心不在"用了多少个服务"，而在三条可检查的底线：按业务边界拆分、数据独占、服务间只经显式契约交互。三者共同保证任何单个服务的故障不会扩散成全局故障。服务内保留 DDD 四层，是因为跨进程的成本越高，层内边界就越要靠纪律维持；反过来，若服务内已经是"随手写"，拆成十个服务只会把十个随手写变成十个不可调试的分布式黑盒。

## Additional Constraints 附加约束

本节是对五条核心原则的展开。原则给出不可协商的底线，本节给出可执行细则。同一要求 MUST 只在本节出现一次，原则与其他小节以引用为准，MUST NOT 重复规定或给出宽严不一的两套说法。

### A. 技术栈与依赖选型规则

**语言**

- 语言选型 MUST 以"最适配"为准绳：结合服务业务特性、并发模型、性能与资源约束、团队能力、既有基础设施在目标语言下的成熟度择优；MUST NOT 以个人偏好、"顺手"或"写起来快"为由。
- 选型 MUST 在该 feature 的 `plan.md` 留痕（候选语言、对照维度、最终选择与理由、被否决方案及原因），并经评审确认；MUST NOT 无理由引入新语言。
- 一个微服务 MUST 使用单一语言，服务内部 MUST NOT 混用多门实现语言；跨语言需求 MUST 通过契约委托给持有该能力的其他服务。
- 不同服务可用不同语言，但同一类能力 MUST 保持语言一致，避免同类逻辑出现多套实现与多套运维手法。
- 每种语言 MUST 具备完整工具链（格式化、静态检查、类型检查、测试、覆盖率、依赖版本锁定与依赖管理）；MUST NOT 只配测试不配静态检查，MUST NOT 靠人工约定代替工具链。
- 运行时与工具链版本 MUST 锁定并写入项目配置，MUST NOT 依赖开发机上的浮动版本。
- 每引入一门语言 MUST 评估运维代价（构建、部署、监控、排障手段齐备）；无法满足者 MUST NOT 引入。
- **后端实现语言已定（2026-11-13）**：后端统一 Go（1.22+）。工具链：gofmt（格式化）、golangci-lint（静态检查）、go vet（类型检查）、go mod + lockfile（依赖锁定）、go test + testify（单元/集成测试）、go test -cover（覆盖率，CI 门禁 ≥80%）、buf（protobuf 契约 lint 与 breaking 检测）；服务框架 go-zero，依赖注入 Google Wire。客户端为 Swift（iOS）、Kotlin（Android）、TypeScript（Web 账号安全与申诉页）。服务内 MUST 单一语言；新引入语言 MUST 按本节重新登记并经评审确认。

**依赖引入**

- 新增依赖 MUST 先检索仓库内既有实现与可用库，优先复用（DRY）；MUST NOT 引入功能重叠的重复依赖。
- 依赖引入 MUST 评估维护活跃度、许可证兼容性与已知安全漏洞；组件 MUST 只从官方源获取并优先使用已签名包；MUST NOT 长期依赖已停止维护的组件而无替换计划（见 F 组）。

**持久化（MongoDB 为默认）**

- 持久化 MUST 默认使用 MongoDB，每个服务 MUST 使用自己独占的 database（必要时独立实例）；集合命名 MUST 带服务前缀（`{service}_{entity}`），MUST NOT 出现无归属前缀的集合名；选用其他存储 MUST 在 `plan.md` 论证理由、替代方案与退出成本并经评审确认。
- MUST NOT 跨服务直连他方 database 或集合，即便同实例亦然；跨服务取数 MUST 经由对方接口或事件。
- 领域模型 MUST 与 BSON 文档解耦：领域实体 MUST NOT 直接使用驱动文档类型（如 `ObjectId`）作为标识，MUST 定义自有值类型并在 `infrastructure` 层显式映射。
- MUST NOT 用嵌套文档掩盖关系建模；MUST NOT 依赖无界数组累积，单文档 MUST NOT 逼近 BSON 大小上限。
- 查询 MUST 有索引：高频查询路径 MUST 建索引，MUST NOT 依赖无索引的全集合扫描，MUST NOT 在查询条件中做无法走索引的计算；MUST NOT 出现 N+1 查询，必要时 MUST 改写为批量查询。
- MUST 以单文档原子性为设计依据：需要多文档原子性时 MUST 明确说明并评估 replica set 事务代价，MUST NOT 用应用层"先查后写"模拟原子性。

**缓存与协调（Redis）**

- Redis MUST NOT 成为唯一事实来源：持久化事实 MUST 存于 MongoDB，Redis 中的数据 MUST 能在缓存失效后重建；MUST NOT 把 Redis 作为写入主路径。
- 缓存 MUST 采用 cache-aside；每个缓存条目 MUST 有 TTL，MUST NOT 缓存无限增长的集合；MUST NOT 允许缓存失效瞬间大量并发回源（须有并发上限或过期抖动）。
- Redis key MUST 按服务命名空间隔离，MUST NOT 出现无前缀的共享键空间，MUST NOT 把本服务键交由他服务读写。
- 分布式锁 MUST 带 TTL 与持有者标识，MUST NOT 依赖进程内互斥充当分布式锁，MUST NOT 用锁掩盖幂等设计的缺失。
- 限流 MUST 口径统一；会话/令牌 MUST 明确有效期与撤销策略。
- Redis 故障 MUST NOT 导致数据丢失：核心写路径 MUST NOT 依赖 Redis 可用；锁不可用时 fail-closed 还是 fail-open MUST 按场景显式决定并记录。
- 异步事件 MUST 走专用消息队列，MUST NOT 把 Redis 当消息队列使用。

**本地环境与依赖编排（Docker）**

- 本地开发依赖 MUST 通过 Docker Compose（或等价编排）启动 MongoDB、Redis 与消息中间件；镜像版本 MUST 固定，MUST NOT 使用 `latest` 标签。
- MUST NOT 连接生产或他人共享环境；凭据 MUST NOT 硬编码进编排文件，MUST 经环境变量或密钥注入。
- 集成测试 MUST 针对容器化依赖运行，版本 MUST 与生产同大版本，测试数据 MUST 可重复初始化与清理。

**待登记项**

- TODO(TECH_STACK): **存储与缓存已定**（2026-11-13：MongoDB 为默认持久化、Redis 承担缓存/锁/限流/会话，本地依赖以 Docker 编排启动）；**架构范式已定**（2026-11-13：微服务 + 服务内 DDD 四层分层，服务间 gRPC/proto，数据服务独占）；**后端实现语言与工具链已定**（2026-11-13：Go + gofmt/golangci-lint/go vet/go mod/testify/go test -cover/buf + go-zero/Wire，见「语言」小节）；**生产就绪标准已定**（2026-11-13：见 D、E、F、G 组）。剩余待定项：依赖方向检查工具、服务发现与配置方案、消息队列具体产品、编排与部署平台、安全工具链（SAST / SCA / 密钥扫描 / 镜像扫描 / IaC 扫描 / SBOM 生成 / 构件签名与 provenance）、密钥管理服务与内部证书体系、日志与追踪后端、指标与看板后端、告警与 burn-rate 计算平台、负载测试与性能剖析工具、可访问性自动检查工具、数据校验与对账工具、缺陷跟踪与成本归集平台。每引入一门语言或一项关键组件，MUST 在本节登记其工具链与配置细则，登记动作以 MINOR 版本修订本宪法；未登记的服务 MUST NOT 合入主干。

### B. API 与数据契约治理规则

**对外接口**

- 对外 API MUST 以资源为中心，MUST NOT 用 CRUD 动词命名方法；MUST 使用标准方法 + 命名自定义方法。（来源：AIP-121 / AIP-123 / AIP-131）
- 列表 MUST 支持分页：MUST 提供 `page_size`（默认 50、最大 1000）与不透明 `page_token`；MUST NOT 用 offset 分页；`page_size` 超上限 MUST 收敛到上限而非报错，负值 MUST 返回 `INVALID_ARGUMENT`；翻页期间排序 MUST 稳定；MUST NOT 实现"先取全量再切片"的分页。（来源：AIP-158）
- 错误 MUST 使用统一错误结构与稳定错误码，错误信息 MUST 简短可执行，MUST NOT 泄露内部实现细节，额外上下文 MUST 放独立字段承载。（来源：AIP-193 / RFC 9457）
- 错误码 MUST 集中定义与登记并标注可重试/不可重试；MUST NOT 复用已废弃的错误码。
- 对外写接口 SHOULD 支持 `Idempotency-Key`（服务端保留 24 小时）；MUST NOT 让调用方靠"重试看看"承担一致性。（来源：AIP-192）
- MUST 明确版本策略；废弃 MUST 提前公告（建议 ≥90 天）并通过 Sunset 头标记下线时间，MUST NOT 无限期双版本并存。（来源：RFC 8594）

**服务间契约**

- `.proto` 包 MUST 标注 alpha / beta / stable 稳定性级别，对应不同兼容性承诺。（来源：AIP-181）
- CI MUST 运行契约破坏性变更检测工具；MUST NOT 合并破坏性契约变更而无版本方案与迁移说明；MUST NOT 复用已废弃的字段号或改变字段语义。
- 事件 MUST 有 schema 与版本；MUST NOT 静默改变已有事件语义；消费者 MUST 容忍新增可选字段；投递语义 MUST 明确且消费者 MUST 幂等。（来源：CloudEvents）
- `.proto` 中 RPC 名 MUST NOT 与其输入/输出 message 名同名（同名时 protoc 无法把 RPC 名解析为 message 类型）；冲突时 MUST 改 RPC 名、message 名保持不变。
- 共享 `.proto` 被各服务 vendoring（各自生成代码）时 MUST 本地化 `go_package`（`{module}/api/{service}/v1`）与 import 路径，MUST NOT 改动字段号、类型或 RPC 签名等契约语义；仅允许改 import 路径、`go_package` 这类构建层信息，改动 MUST 记录在对应 research.md。

**数据治理**

- MUST 建立数据分类分级与 PII 清单，标注每类数据的敏感级别、用途与保留期限。
- MUST 明确数据保留期与到期删除机制；删除 MUST 在各服务生效，且 MUST 明确备份中数据如何处置。
- MUST 验证恢复：备份 MUST 定期做恢复演练，MUST NOT 只检查备份任务是否成功。

**数据完整性与数据管道**

- 数据完整性 MUST 有独立于可用性的目标（"读到的就是写下的"）；MUST NOT 用可用性 SLO 代替完整性目标——两者失败后果完全不同，最糟的分歧时机是用户数据没了之后。（来源：Google SRE Book Ch.26《Data Integrity》）
- 跨服务与跨系统传递的数据 MUST 有校验和或等价机制，MUST NOT 只依赖传输层 TLS；MUST 有定期抽样比对与对账。
- 损坏 MUST 被主动检测并在用户受影响前修复，MUST NOT 依赖用户反馈才发现；检测到后 MUST 能在约 30 分钟内完成修复与恢复。
- 恢复出的数据 MUST 与源数据做一致性校验。
- 管道步骤 MUST 幂等，MUST NOT 依赖"恰好一次"投递，MUST 用唯一标识去重。
- MUST NOT 串联过深的处理链；MUST 记录管道深度与各阶段吞吐差异，对吞吐不匹配做限流或重构；MUST 支持按分区或时间范围回填与重算，MUST NOT 以全量重跑作为常规手段。（来源：Google SRE Book Ch.25《Data Processing Pipelines》）
- 关键数据集 MUST 有来源与去向的记录，MUST NOT 无法回答"这个数字从哪来"。

### C. 质量、测试与性能规则

- MUST 遵循测试金字塔：单元测试 MUST 占主体且 MUST 能在数分钟内跑完数千个；集成测试次之；端到端与 UI 测试 MUST 保持在少数。层级越高测试数量 MUST 越少；高层级测试 MUST 只覆盖高层级才能发现的问题，MUST NOT 重复验证单元测试已覆盖的逻辑。（来源：Martin Fowler《The Practical Test Pyramid》）
- 覆盖率统计 MUST 禁用测试缓存（Go 用 `go test -count=1`），MUST NOT 因测试缓存导致覆盖率数字波动或被低估。
- MUST 以零 MTTR 缺陷为目标：凡是监控能发现的问题，MUST 尽量由一个能在上线前发现它的测试覆盖，从而直接阻断发布而不是上线后再修；零 MTTR 缺陷发现得越多，MTBF 越高。（来源：Google SRE Book Ch.24《Testing Reliability》）
- 影响运行行为的配置 MUST 纳入版本控制，MUST 有测试比对线上实际配置与仓库期望配置并报告差异（配置漂移检测）。（来源：Ch.24）
- 每个服务 MUST 有可自动执行的冒烟测试，覆盖启动与核心路径，作为发布后第一道验证。
- 关键路径 MUST 有响应时间与资源占用的基线测试并每次发布执行，防止性能静默劣化（如 10ms → 50ms → 100ms）。（来源：Ch.24）
- 流水线 MUST 有时长预算（建议单元与集成合计 ≤10 分钟，端到端单独流水线 ≤30 分钟）；超出 MUST 拆分或标记为非阻塞并记录原因。
- CI MUST 统计并隔离 flaky 测试（隔离而非忽略）；隔离项 MUST 有负责人与修复期限，MUST NOT 长期挂起。
- MUST 用消费者驱动的契约测试替代共享集成环境，MUST NOT 依赖大规模共享环境保证跨服务正确性。
- 测试数据 MUST 可重复生成与清理，MUST NOT 依赖生产数据，MUST NOT 依赖手工预置状态。
- MUST 追踪缺陷逃逸率（上线后才发现的缺陷占比）与缺陷密度；覆盖率 MUST NOT 作为唯一质量信号。
- MUST 有合成探测或影子流量类的受控生产验证，MUST NOT 对生产执行破坏性测试。
- 每个对外接口 MUST 有性能预算，MUST 以 p95/p99 表达，MUST NOT 只看平均值；超出预算 MUST 阻断发布或需书面豁免。
- MUST 监控 p95 与 p99 延迟，MUST NOT 只监控平均值（平均值会掩盖长尾）。
- 关键路径 MUST 有可重复执行的负载测试，MUST 在发布前或定期执行，MUST 有吞吐与延迟的回归基线。
- MUST 有慢查询与无索引查询的监控和阈值，MUST NOT 依赖人工定期检查。
- MUST 记录容量水位与扩容触发点，MUST NOT 在容量耗尽后才扩容。
- 性能问题 MUST 以剖析数据定位，MUST NOT 以猜测改代码。

### D. 可靠性、可观测性与运行保障规则

**可靠性与弹性**

- 每个对外服务 MUST 有书面的 SLI（延迟、可用性、正确性、吞吐）与 SLO（目标值），并 MUST 有对应看板。（来源：Google SRE Book《Service Level Objectives》）
- SLO MUST 配套错误预算（= 1 − SLO）；错误预算耗尽时 MUST 冻结新功能发布，转入可靠性工作。（来源：SRE Book《Error Budget Policy》）
- 每个外部依赖 MUST 定义失败行为（fail-open / fail-closed），MUST NOT 让依赖失败直接升级为整体不可用。（来源：Azure Well-Architected 韧性模式）
- 每次跨服务调用 MUST 设置超时；重试 MUST ≤2 次且带指数退避+抖动，MUST 仅用于幂等操作；MUST 有熔断与舱壁隔离；重试总耗时 MUST 小于调用方剩余 deadline。（来源：Google Cloud 韧性设计模式）
- 跨服务调用 MUST 向下传递剩余超时时间，MUST NOT 在下游使用比上游更长的超时；MUST NOT 无退避地放大下游压力。
- 服务 MUST 处理终止信号，停止接收新请求并在宽限期内完成在途请求，MUST NOT 强杀造成半完成事务。
- 存活（进程活着）、就绪（可接流量）、启动（启动完成）三种健康检查 MUST 语义分离；MUST NOT 用存活检查探测依赖可用性（会造成级联重启）。
- 服务 MUST 记录资源基线与扩容触发条件。
- 关键服务 MUST 有书面的 RTO/RPO 与备份策略；备份 MUST 至少每季度做一次恢复验证，MUST NOT 只验证备份存在。（来源：SRE Book《Disaster Recovery》）
- 事故 MUST 分级、MUST 有值班、MUST 做无责复盘并产出可追踪的改进项；事件响应计划 MUST 成文并每年演练 ≥1 次。（来源：SRE Book《Postmortem Culture》《On-Call》；NIST SP 800-61）
- 每次上线 MUST 有回滚路径，或明确记录"仅可前滚"及其代价。（来源：SRE Book《Release Engineering》）
- MUST 记录并定期回顾变更前置时间、部署频率、失败部署恢复时间、变更失败率、部署返工率；MUST NOT 把指标本身当作目标（Goodhart 定律）。（来源：DORA 官方指南）

**可观测性**

- trace 与 span MUST 遵循 OpenTelemetry 语义约定的标准命名与属性，MUST NOT 自造同义属性；跨服务调用 MUST 传递 trace context，MUST NOT 中断。（来源：opentelemetry.io 语义约定）
- 日志 MUST 为 JSON 结构化，MUST 有固定字段集（UTC 时间、级别、服务名、版本、trace_id、span_id、请求或事件 ID、错误码），MUST NOT 做自由文本拼接。
- 日志 MUST 分级：`debug`（调试细节，默认关闭）、`info`（关键操作与状态）、`warn`（可恢复异常）、`error`（不可恢复错误）。每个对外接口与关键用例的入口 MUST 打 `info`；错误与异常路径 MUST 打 `warn`/`error` 并携带完整上下文（输入摘要、错误堆栈、关联 ID），MUST NOT 静默吞错——程序必有缺陷，分级日志是可排查性的底线。
- 日志、指标、追踪与错误响应 MUST NOT 包含秘密或完整 PII（身份证号、手机号、银行卡号、令牌、口令）；MUST 脱敏或哈希；MUST NOT 把用户输入未净化地拼入日志字段（防日志注入）。
- MUST NOT 把用户 ID、请求 ID、时间戳等高基数字段用作指标标签；每个指标的标签基数 MUST 有上限。（来源：Prometheus cardinality 实践）
- 每个服务 MUST 有延迟、流量、错误、饱和度四个黄金信号的指标与看板。（来源：SRE Book《Monitoring Distributed Systems》）
- 告警 MUST 基于 SLO 错误预算燃烧率，MUST NOT 基于静态阈值；快速燃烧（14.4×，1 小时窗口 + 5 分钟长窗口）与慢速燃烧（6×，6 小时 + 30 分钟）MUST 分别告警；每个告警 MUST 可操作且 MUST 有对应 runbook，MUST NOT 存在无 runbook 的告警。（来源：SRE Workbook《Alerting on SLOs》）
- 追踪采样率 MUST 显式设定（常态建议 1%–10%），错误请求 MUST 保留完整追踪，MUST NOT 默认全量高成本采样。
- 日志、指标、追踪的留存期 MUST 明确，MUST NOT 无限制保留；MUST NOT 依赖人工翻日志来发现问题。

### E. 安全、隐私与供应链合规规则

引用版本：OWASP Top 10 当前版本为 **2025**（A03 = Software Supply Chain Failures，A07 = Authentication Failures）；OWASP ASVS 稳定版为 **5.0.0**（345 条要求，L1 70 / L2 253 / L3 345，本项目以 L2 为基线）；NIST SP 800-218（SSDF）；SLSA Build L0–L3。

- 新服务与重大接口变更 MUST 有威胁建模记录（信任边界、数据流、滥用场景）并纳入设计评审。（来源：SSDF PW.1.1；ASVS V1）
- MUST 逐条防御 OWASP API Security Top 10 (2023)，落地映射见下表。
- 对外 API MUST 只接受 OAuth 2.1 授权码 + PKCE 取得的令牌，MUST NOT 支持隐式流与 ROPC 密码模式。（来源：draft-ietf-oauth-v2-1-11 §10.1；RFC 9700 §2.1）
- 令牌 MUST 校验签名、`alg` 白名单、`iss`、`aud`、`exp`、`nbf`，MUST NOT 接受 `alg=none`；校验密钥 MUST 取自预配置的受信 JWKS 端点，MUST NOT 使用令牌自带的 `jku`/`x5u`/`jwk` 头；每服务 MUST 校验 `aud` 含自身标识，同一令牌 MUST NOT 跨服务复用。（来源：ASVS V9.1.1–V9.2.3；RFC 9700 §4.10.2）
- 访问令牌 MUST ≤15 分钟；刷新令牌 MUST ≤7 天且每次使用后轮换、旧令牌立即失效；授权码 MUST ≤10 分钟；MUST NOT 存在无期限令牌；会话 MUST 同时设空闲超时与绝对超时；改密或禁用账号 MUST 终止该用户全部会话；会话 ID 熵 MUST ≥64 bit，参考令牌 MUST ≥128 bit。（来源：OAuth 2.1 §3.2.3；RFC 9700 §4.9.3；ASVS V7）
- 改密、改 MFA、换绑手机或邮箱、查看敏感信息 MUST 强制重新认证。（来源：ASVS V7.5.1）
- 密码 MUST 用 Argon2id（推荐 m=19456(19 MiB)/t=2/p=1，或 m=47146(46 MiB)/t=1/p=1；或 scrypt N=2^17,r=8,p=1）；遗留 bcrypt MUST cost ≥10；FIPS 场景 PBKDF2-HMAC-SHA256 MUST ≥600,000 轮；MUST NOT 使用 MD5 / SHA-1 / 裸 SHA-256；单次哈希耗时 MUST <1 秒。（来源：OWASP Password Storage Cheat Sheet；RFC 9106）
- 登录、找回密码、验证码端点 MUST 同时按账号维度与 IP 维度限流并失败锁定；找回密码 MUST 与登录同等防护。（来源：API2:2023；OWASP Authentication Cheat Sheet）
- 服务间 gRPC 与事件总线连接 MUST 100% 启用 mTLS，MUST 校验对端证书 SAN 与服务身份，MUST NOT 关闭证书校验。（来源：ASVS V12.3.4–V12.3.5）
- 授权 MUST 默认拒绝，仅显式授予的角色或策略可访问，MUST NOT 采用"未命中即放行"；管理类接口 MUST 独立路由并独立鉴权。（来源：ASVS V8.2.1；API5:2023）
- 凡以 ID 访问存储的接口 MUST 在服务端校验主体对该对象的所有权或权限，MUST NOT 仅比对路径 ID 与令牌 `sub`（OWASP 明确此法不足）；查询条件 MUST 带上主体或租户过滤，MUST NOT 先按 ID 查出再判断；对象 ID MUST 使用不可预测的 UUIDv4/v7，MUST NOT 使用自增序号。（来源：API1:2023；ASVS V8.2.2）
- 响应 MUST 由显式白名单 DTO 构造，MUST NOT 直接序列化 MongoDB 文档；写入 MUST 只接受允许字段，MUST NOT 反序列化整个请求体。（来源：API3:2023；ASVS V8.2.3）
- 每个服务 MUST 使用独立 MongoDB 账号且仅授权其自有 database 的必要集合；Redis MUST 按服务划分 key 前缀并用 ACL 限制可见 key 模式与命令集；分布式锁 MUST 使用带 fencing token 的实现。（来源：OWASP Top 10:2025 A01；ASVS V8.2.2）
- 授权 MUST 基于原始主体权限，MUST NOT 基于中间服务的权限；多租户过滤 MUST 下推到查询层，MUST NOT 依赖调用方传入的租户字段。（来源：ASVS V8.3.3、V8.4.1）
- 对外 API MUST 有按主体与 IP 的限流、请求体大小上限、查询结果条数上限；昂贵操作 MUST 单独限流；异步事件消费 MUST 有背压与最大并发。（来源：API4:2023；ASVS V13.1.2）
- 用户可控 URL 触发的出站请求 MUST 经统一出站代理与目标白名单；MUST 禁用自动跟随重定向；MUST 在 DNS 解析后校验目标 IP 不属于内网、回环或云元数据段（169.254.169.254）。（来源：API7:2023；ASVS V13.2.4–V13.2.5）
- 生产 MUST 关闭 debug / pprof / 管理端点，MUST NOT 返回后端精确版本号，MUST NOT 暴露目录列表与 TRACE；CORS MUST 用显式来源白名单，MUST NOT 反射 Origin，MUST NOT 与凭据同用 `*`；跳转目标 MUST 为白名单相对路径或已登记域名；MUST 拒绝同时出现 `Content-Length` 与 `Transfer-Encoding` 的请求；文件上传 MUST 校验类型与大小并重命名、存放于非可执行目录。（来源：API8:2023；ASVS V4、V5、V13）
- 对外暴露的接口与依赖 MUST 有清单（路径、方法、认证要求、敏感级别、废弃状态）并与实现自动比对，下线即移除。（来源：API9:2023）
- 消费外部 API MUST 全程 TLS 并校验证书；返回数据 MUST 做 schema 校验，MUST NOT 直接反序列化；MUST 设超时、重定向白名单与响应大小上限；第三方凭据 MUST 走密钥管理并轮换。（来源：API10:2023；ASVS V13.2.1）
- MUST 仅启用 TLS 1.3（TLS 1.2 仅在明确评估后作兼容回退），MUST NOT 启用 TLS 1.0/1.1，MUST NOT 提供明文 HTTP 业务端点；静态敏感数据 MUST 用 AEAD 加密（AES-256-GCM），nonce/IV MUST NOT 复用；MUST NOT 使用 MD5 / SHA-1 / ECB / PKCS#1v1.5 用于安全用途，RSA MUST ≥3072 位或 ECC ≥256 位；加密密钥 MUST 有版本号、用途、过期时间与轮换计划，MUST 支持新旧密钥并存以零停机轮换；内部 CA 签发证书 MUST 自动续期并有过期告警。（来源：ASVS V11、V12；RFC 8996；NIST SP 800-57）
- 源码、测试夹具、示例配置、CI 变量与镜像层中 MUST NOT 出现密码、密钥或令牌；一律运行时注入，MUST 以最小权限只读取本服务所需秘密；秘密 MUST 有轮换周期并自动执行，MUST 支持新旧并存过渡。（来源：ASVS V13.3；OWASP Top 10:2025 A08）
- 密钥扫描（gitleaks）扫的是提交历史而非工作区快照：MUST NOT 以"本地开发占位"为由把密钥/令牌硬编码进源码（占位值同样会被 `generic-api-key` 规则命中）；开发用密钥 MUST 走环境变量注入并在缺失时 fail-fast（原则 I）；已进入提交历史的硬编码密钥 MUST 通过 squash/重写历史彻底清除，MUST NOT 仅在新提交中删除（否则扫描仍命中旧提交）。
- MUST 维护"哪个事件记到哪"的日志清单（格式、存储位置、访问控制、留存期）；认证成功与失败、授权拒绝、账号与权限变更、配置变更、密钥使用 MUST 全部记录；每条安全日志 MUST 含 UTC 时间、主体、来源 IP、动作、目标对象与结果，且 MUST NOT 含敏感数据本体；审计日志 MUST 写入 append-only 存储并与业务库分离，MUST NOT 提供应用侧删除或修改接口；安全审计日志 MUST 留存 ≥6 个月（《网络安全法》第二十一条）；MUST 对爆破、越权尝试、密钥误用与异常批量导出建立检测与告警。（来源：ASVS V16；OWASP Top 10:2025 A09；PIPL 第19条）
- MUST 建立个人信息与敏感个人信息清单（生物识别、宗教信仰、特定身份、医疗健康、金融账户、行踪轨迹、未成年人信息）；采集与返回字段 MUST 与处理目的直接相关且只取最小必要；处理敏感个人信息 MUST 取得单独同意；MUST 按处理目的实现自动过期删除；MUST 提供主体查阅、复制、更正、删除的受理入口；向境外提供个人信息 MUST 满足安全评估、专业机构认证或标准合同三者之一并取得单独同意；处理敏感个人信息、自动化决策、委托处理、对外提供与跨境传输 MUST 事前做影响评估并保存 ≥3 年；MUST 有泄露响应的书面应急预案；测试与分析场景 MUST 使用去标识化或假名化数据，MUST NOT 直接使用生产个人数据。（来源：PIPL 第6/17/19/28/29/38/39/44/51/55/56/57 条；GDPR Art.5/17/25/32/35）
- 每次发布 MUST 生成 CycloneDX 1.6 或 SPDX 2.3 格式 SBOM，覆盖直接与传递依赖并与制品一同归档；CI MUST 扫描直接依赖、传递依赖与构建工具的已知漏洞；MUST 提交 lockfile 并使用不可变版本或哈希，MUST NOT 在构建时解析浮动版本；MUST 有允许与禁止许可证清单并由 CI 阻断未批准许可证；发布构件 MUST 签名并附 SLSA provenance，部署前 MUST 校验签名（目标 SLSA Build L2 —— 有来源证明、托管构建平台）；构建 MUST 在隔离环境执行，开发、测试、生产 MUST 分离；MUST 定期移除未使用依赖；MUST 一次构建多环境晋级，MUST NOT 每个环境重新构建；MUST NOT 让同一人从写代码到发布生产全程无第二人监督；CI 配置、制品库、SBOM 工具与镜像仓库的变更 MUST 有版本控制。（来源：OWASP Top 10:2025 A03；SSDF PO/PS/PW/RV；SLSA；OpenSSF Scorecard）
- 漏洞 MUST 按等级限期修复 —— CVSS ≥9.0 或已列入 CISA KEV 目录 MUST 7 天内修复，CVSS ≥7.0 MUST 30 天内，中危 MUST 90 天内；无法修复 MUST 有书面风险接受、责任人、到期日与补偿控制。（来源：SSDF RV.2；CISA KEV 目录）
- CI MUST 具备静态分析（SAST）、依赖扫描（SCA）、密钥扫描、容器镜像扫描、基础设施即代码扫描、契约兼容性检查；CVSS ≥7.0 MUST 阻断合并与发布；SBOM 生成、构件签名与 provenance 缺失 MUST 直接阻断；MUST 有 BOLA 与功能级授权的自动化回归用例，失败 MUST 阻断发布；对外 API MUST 跑动态扫描（DAST）；CI 安全检查的执行记录 MUST 保留 ≥12 个月；MUST 每季度复核全部安全控制的实际有效性并记录结论。（来源：SSDF PS.3/RV.1；OWASP Top 10:2025 A03/A08）
- 所有外部输入 MUST 在系统边界做类型、长度、范围、枚举的 schema 校验；MUST NOT 用字符串拼接构造 MongoDB 查询，MUST 用强类型查询构造器，MUST 在用户可控字段上禁用 `$where` 与 `$regex`；MUST 按输出上下文编码；MUST NOT 反序列化不可信数据（禁用 Java 原生反序列化、Python `pickle`、不安全的 YAML 加载）；MUST NOT 拼接用户输入到系统命令；文件读写 MUST 先规范化路径并校验其位于允许根目录内。（来源：ASVS V1.2、V2、V5；OWASP Top 10:2025 A05；OWASP NoSQL Injection / Deserialization Cheat Sheet）
- 统一信封的 `message` MUST 是面向用户的安全文案，MUST NOT 含堆栈、查询语句、密钥、内部路径或主机名；MUST 提供机器可读错误码，客户端 MUST NOT 解析 `message` 文本；MUST 返回语义正确的 HTTP 状态码，MUST NOT 用 200 包装全部错误；响应 MUST 带与内容匹配的 `Content-Type`，敏感响应 MUST 带 `Cache-Control: no-store`；令牌、密钥与敏感查询条件 MUST NOT 出现在 URL 或查询串中；未授权响应 MUST NOT 区分"对象不存在"与"无权访问"。（来源：ASVS V4.1、V16.5；RFC 9110）
- 异步事件 MUST 带来源标识与签名或 HMAC，消费者 MUST 验签并校验时间戳窗口；事件总线 MUST 强制 TLS/mTLS 与身份认证，MUST NOT 存在匿名可写主题；事件 MUST 遵循 CloudEvents 1.0 的必需属性（`id`、`source`、`type`、`specversion`、`time`）；消费者 MUST 以 `(source, id)` 去重，生产者 MUST 保证该二元组唯一；事件载荷 MUST 只含下游必需字段，MUST NOT 塞入整份文档。（来源：CloudEvents 1.0.3；ASVS V12.3.5）
- 容器 MUST 以非 root 用户运行，MUST NOT 挂载 `docker.sock`，MUST NOT 使用特权模式；MUST 使用最小基础镜像并定期重建以获取补丁；CI MUST 扫描镜像漏洞；生产配置 MUST 由环境或配置中心注入，MUST NOT 打进镜像；MongoDB 与 Redis MUST NOT 暴露到公网，MUST NOT 使用 `network_mode: host`；容器 MUST 设 CPU 与内存上限及健康检查。（来源：OWASP Docker Security Cheat Sheet；OWASP Top 10:2025 A03/A08）
- MUST 有公开的漏洞披露通道与响应流程；安全风险评估 MUST 随版本更新；安全例外 MUST 有书面记录、责任人与到期日；第三方组件 MUST 有引入审批流程与清单。（来源：SSDF RV.1.3、PO.4.2；ISO 30111）

**OWASP API Security Top 10 (2023) 逐条落地**

| 项 | 本项目落地规则 |
|---|---|
| API1 对象级授权越权 | 以对象 ID 为入口的操作 MUST 在服务端校验主体与对象的关系；查询 MUST 带主体/租户过滤；对象 ID MUST 用不可预测 UUID |
| API2 认证失效 | 对外 MUST 走 OAuth 2.1 授权码 + PKCE；令牌 MUST 有 ≤15 分钟 `exp`、`aud`、`iss` 与算法白名单；登录与找回密码 MUST 双维度限流 |
| API3 属性级授权越权 | 响应 MUST 由显式白名单 DTO 构造，MUST NOT 直接序列化 MongoDB 文档；写入 MUST 只接受允许字段 |
| API4 资源消耗无界 | MUST 有按主体与 IP 的限流、请求体上限、分页上限、超时上限；事件消费 MUST 有背压 |
| API5 功能级授权越权 | 授权 MUST 默认拒绝；管理端点 MUST 独立路由与独立鉴权；gRPC 每个方法 MUST 显式声明所需角色 |
| API6 敏感业务流滥用 | 高价值流程（登录、发券、下单、批量导出）MUST 加人机识别或二次验证、速率限制、幂等键与异动告警 |
| API7 SSRF | 出站调用 MUST 经统一出站代理与目标白名单；MUST 禁自动重定向；MUST 校验解析后 IP 非内网/回环/云元数据段 |
| API8 安全配置错误 | 生产 MUST 关闭 debug/pprof/管理端点；MUST NOT 返回精确版本号；CORS MUST 显式白名单；配置 MUST 全部外置并按环境分离 |
| API9 资产管理不当 | 对外接口 MUST 有清单并与实现自动比对；非生产环境 MUST NOT 使用生产数据；废弃接口 MUST 有下线日期 |
| API10 不安全消费第三方 API | 外部 API 响应 MUST 做 schema 校验，MUST NOT 直接反序列化；MUST 设超时与重定向白名单；凭据 MUST 走密钥管理并轮换 |

### F. 交付、分支与协作治理规则

**分支模型（三常驻分支）**

- 仓库 MUST 采用三常驻分支模型：`main`（生产）、`release`（预发布）、`develop`（集成）为三个常驻分支；`feature/`（需求）、`bug/`（缺陷）、`hotfix/`（紧急上线）为工作分支，一个 Spec Kit feature 对应一个工作分支。
- 合入方向 MUST 单向且固定，MUST NOT 反向合入：
  - `main` MUST 只接受 `release` 与 `hotfix` 分支合入。
  - `release` MUST 只接受 `develop` 与 `bug` 分支合入。
  - `develop` MUST 只接受 `feature` 分支合入。
- 分支切出点 MUST 固定：
  - 特性分支（`feature/`）MUST 从最新的 `develop` 切出，合回 `develop`。
  - 缺陷分支（`bug/`）MUST 从 `release` 切出，合回 `release`（修复发布候选中的缺陷）。
  - 紧急修复分支（`hotfix/`）MUST 从 `main` 上一个发布 tag 切出，合回 `main`，并 MUST 反合 `release` 与 `develop`，MUST NOT 遗漏（否则下个发布会把该缺陷带回）。
- `main` MUST 始终等于"当前生产运行状态"；`release` MUST 始终等于"下一发布候选"；`develop` MUST 始终可集成、可部署。
- 发布流程：`develop` 集成完成 → 合入 `release` 冻结候选 → `release` 上只做缺陷修复与发布准备 → `release` 验收通过 → 合入 `main` 并打 tag。

**命名与关联**

- 分支名 MUST 为 `{type}/{编号}-{kebab-case-slug}`，工作分支 `type` ∈ `feature` / `bug` / `hotfix`，维护分支 `type` ∈ `chore` / `docs`，编号为任务或 issue 号，slug 用小写连字符。常驻分支 `main` / `release` / `develop` 不作为工作分支名。
- MUST NOT 使用含义不明的分支名（个人名、`test`、`tmp`、`dev2` 等）。
- 每个特性分支 MUST 对应一个 Spec Kit feature 目录（含 `spec.md` / `plan.md` / `tasks.md`）；分支名 MUST 含 feature 编号，使提交可回溯到规格产物。spec 目录名（`specs/{编号}-{slug}`）与 git 分支名相互独立，由 `.specify/feature.json` 关联；两者的编号 MUST 一致。
- PR MUST 引用任务或 issue 号；提交信息 MUST 遵循 Conventional Commits，MUST NOT 使用 `fix`、`update`、`临时` 等无信息量描述。
- 每个提交 MUST 只做一件事，MUST NOT 把格式化、依赖升级与功能改动混在同一提交。

**保护规则**

- `main`、`release` 与 `develop` MUST 开启分支保护：MUST 要求 CI 全部通过（治理校验、密钥扫描与各服务覆盖率门禁）；MUST 要求分支与目标分支同步后才可合并；MUST NOT 允许直接推送；MUST NOT 允许 force push；MUST 启用删除保护；MUST 启用线性历史。（评审人数要求及其例外见「协作与知识」。）（来源：GitHub Docs 分支保护）
- 多服务仓库的 CI 覆盖率门禁 MUST 按服务目录变更过滤（如 dorny/paths-filter）触发，MUST NOT 让一个服务的 PR 触发并阻塞其他服务的覆盖率 job；服务数增长时 MUST 同步维护各服务的 path-filter 与对应覆盖率 job。
- 受保护分支 MUST 要求"分支与目标分支同步"后才允许合并，MUST NOT 允许带着落后主干的分支合并。
- MUST 启用 squash merge 与已合并分支自动删除。

**合并策略**

- 所有合入 `develop`、`release` 与 `main` 的 PR MUST 使用 squash merge，主干上每个 PR MUST 恰好产生一个提交。
- PR 标题 MUST 即为主干提交信息（Conventional Commits 格式）；MUST NOT 把多个不相关任务合入一个 PR。
- 合并前 MUST 先把目标分支（特性分支合 `develop` 时合 `develop`，缺陷分支合 `release` 时合 `release`，紧急修复合 `main` 时合 `main`）同步进工作分支。
- 受保护分支 MUST NOT 被重写历史；已推送的共享分支 MUST NOT force push。

**工作分支生命周期**

- 工作分支（`feature` / `bug` / `hotfix`）MUST 短命：建议 ≤3 天或一个迭代，MUST NOT 跨越迭代长期存在。
- 工作分支 MUST 定期同步其目标分支（`feature` 同步 `develop`、`bug` 同步 `release`、`hotfix` 同步 `main`）；落后超过 3 天 MUST 先同步再继续工作。
- PR MUST 在可观察的时限内被首次评审（建议 1 个工作日；紧急修复 ≤2 小时），超时 MUST 催办。
- 已合并的分支 MUST 立即删除；未合并且超过 14 天无进展的分支 MUST 关闭或重新评估。
- MUST NOT 在同一个分支上并行开发两个任务。

**标签与发布**

- 发布 MUST 打 annotated tag，标签名 MUST 为 `v{MAJOR}.{MINOR}.{PATCH}`（遵循 SemVer），MUST NOT 复用或移动已发布标签。
- 标签 MUST 受保护，MUST NOT 被强制推送或删除。
- `release`（常驻预发布分支）合入 `main` 后 MUST 在 `main` 的 squash 提交上打 tag；紧急修复发布 MUST 同时打 tag 并反合 `release` 与 `develop`。
- `release` 合入 `main` 的合并 MUST 使用 Conventional Commits 风格的版本提交信息（如 `chore(release): v1.2.0`）。
- `release` 采用 squash 合入 `main` 后，同步回 `develop` MUST 以独立提交带回净变化（如 CHANGELOG 新版本段），MUST NOT 直接 merge `main`→`develop`（squash 提交与 develop 的独立提交会产生 add/add 冲突）。

**仓库卫生与协作入口**

- 仓库 MUST 具备：`README`、`LICENSE`、`CONTRIBUTING.md`、`CODEOWNERS`、`SECURITY.md`、`.gitignore`、`.gitattributes` 与 CI 配置。
- 每个服务与关键目录 MUST 指定 CODEOWNERS 归属人，对其变更强制评审。
- 仓库 MUST NOT 提交密钥、凭据、构建产物、IDE 本地配置与超大二进制文件（建议阈值 10MB）；MUST 通过 `.gitignore` 与 CI 检查阻止。
- 外部贡献 MUST 经 PR 合入，MUST NOT 直接推送到受保护分支。

**变更与发布**

- 对外契约与客户端库 MUST 遵循 SemVer，每次发布 MUST 更新 CHANGELOG。（来源：SemVer / Keep a Changelog）
- 重大架构决策 MUST 写 ADR（背景 / 决策 / 后果 / 备选方案），并与 `plan.md` 分开存放。（来源：adr.github.io）
- 新行为 MUST 通过 feature flag 逐步放量；每个 flag MUST 有负责人与下线期限（建议 ≤90 天），MUST NOT 留下永不过期的 flag。
- 数据库与契约变更 MUST 用 expand-contract 模式（先扩展、再回填双写、最后收缩），MUST NOT 在同一次发布内同时做破坏性变更与依赖它的代码。
- MUST 保持变更粒度小，MUST NOT 把多个不相关变更塞进一次发布。（来源：DORA）
- 生产发布 MUST 分批放量（建议 5% → 25% → 100%），每档 MUST 设观察时长与自动回滚阈值；MUST NOT 假设所有变更都能回滚，不可回滚的变更 MUST 前滚修复并在 plan 中标注。
- 依赖升级 MUST 由机器人提 PR 并经完整 CI；MUST NOT 自动合并高危安全更新；MUST 定期清理积压。
- 高风险时段 MUST 冻结发布；紧急发布 MUST 事后复盘。

**协作与知识**

- 每个任务 MUST 有明确的完成定义（DoD）——代码、测试、文档、契约、观测与门禁全部满足；MUST NOT 以"能跑起来"作为完成标准。
- 每个变更 MUST 至少一名非作者评审通过，MUST NOT 作者自审；评审 MUST 检查正确性与宪法合规，MUST NOT 仅检查格式。
- **评审人数例外（限时）**：GitHub 禁止 PR 作者批准自己的 PR，因此当仓库只有单一维护者时该要求在平台上无法满足。此期间该条暂以下列组合替代：CI 全绿 + CODEOWNERS 自动审查请求 + 禁止直接推送 + 禁止 force push + 删除保护 + 必须同步 + 线性历史 + 公开的全过程记录。MUST 在引入第二名评审者后立即恢复为 ≥1 名（新增服务、数据库迁移、破坏性契约变更 MUST ≥2 名），并以 MINOR 版本修订本宪法。
- 新服务上线前 MUST 由另一名工程师按附加约束 A–G 组做一次评审——像接手一个已有服务那样评审它。（来源：Google SRE Book Ch.17《Software Engineering in SRE》）
- PR 描述 MUST 说明变更意图、影响面与验证方式；PR MUST 控制规模，MUST NOT 把无关变更混入（分支命名与合并规则见上方「分支模型」各节）。
- 每个服务 MUST 有 README（职责、启动方式、依赖、配置项）；每个告警 MUST 有 runbook；每个对外接口 MUST 有文档；文档 MUST 与代码同仓并随变更更新。
- on-call 交接 MUST 有书面记录（本周事故、待观察项、已知问题）。
- 重复性、可预测的运维操作 MUST 被自动化或消除；运维类工作占团队时间的比例 MUST 记录并设上限（参考 Google 的 50% 上限）。（来源：Google SRE Workbook Ch.6《Eliminating Toil》）
- 已知技术债 MUST 登记（影响、触发条件、计划处理时间），MUST NOT 只存在于口头交流或 issue 标题里；重大事故与重大回滚 MUST 产出改进项，改进项 MUST 有负责人与截止时间。

**架构与成本治理**

- 架构约束 MUST 以可持续验证的方式落地（依赖方向检查与跨服务数据访问检查 MUST 纳入 CI），MUST NOT 依赖人工记忆维持架构。（来源：原则 V）
- MUST 每季度或每个大版本做一次架构现状评审，与目标架构对照并产出差距清单与整改项；已知偏离目标架构之处 MUST 登记，MUST NOT 静默偏离。
- MUST 定期评估依赖与技术的采用状态（采用 / 试验 / 评估 / 暂缓）；试验性技术 MUST 有退出条件，MUST NOT 无期限停留在主干。（来源：Thoughtworks Technology Radar）
- 服务 MUST 能把资源成本（计算、存储、网络、第三方 API）归集到可归属的维度；MUST 有成本预算或用量上限并在接近上限时告警；MUST 定义并监控单位经济性指标（如每请求成本、每租户成本）。
- 降本措施 MUST NOT 以牺牲 SLO 或数据完整性为代价；任何以性能或可靠性换成本的取舍 MUST 书面记录理由；孤儿数据、无引用对象、过期快照与备份 MUST 有清理策略与保留期。

### G. 用户体验与运营支持规则

**可访问性、国际化与本地化**

- 面向终端用户的前端界面 MUST 满足 WCAG 2.2 的 A 级与 AA 级成功判据，MUST NOT 只满足 A 级；自动化可访问性检查 MUST 纳入 CI。（来源：W3C《WCAG 2.2》）
- **原生移动端的可访问性口径（2026-11-13）**：本项目的终端用户界面为 iOS / Android 原生客户端（Web 仅承担账号安全与申诉）。原生客户端以「平台原生无障碍规范 + 逐条复核 WCAG 2.2 AA 成功判据」为落地方式；「自动化可访问性检查纳入 CI」适用于 Web 端页面。原生客户端 MUST 满足：系统字体放大到设备支持的最大档时关键内容不截断、读屏软件可完整朗读关键控件与消息状态、状态与错误有颜色之外的表达；MUST NOT 以"原生无 Web 自动化检查工具"为由豁免 WCAG 判据的逐条复核。
- MUST NOT 存在只能靠鼠标或触屏操作的关键路径；键盘焦点 MUST 可见。（WCAG 2.1.1 / 2.4.7）
- 非文本内容 MUST 有替代文本；表单控件 MUST 有关联标签；MUST NOT 用无语义元素模拟按钮。（WCAG 1.1.1 / 4.1.2 / 3.3.2）
- 状态与错误 MUST 有颜色之外的表达（文本、图标或形状），并满足对比度要求。（WCAG 1.4.1 / 1.4.3）
- 用户可见文案 MUST 外置为可翻译资源，MUST NOT 硬编码在代码中；MUST 支持 Unicode；MUST NOT 拼接句子再做翻译。
- 时间 MUST 以明确时区存储与展示；日期、数字、货币 MUST 按 locale 格式化，MUST NOT 硬编码格式。
- 改动用户界面的 feature MUST 复核可访问性与国际化，MUST NOT 只验证功能可用。

**运营与支持**

- 对外服务 MUST 有面向用户的可用性说明与状态页，MUST NOT 让用户只能通过工单得知故障。（来源：Google SRE Book Ch.14《Postmortem Culture》）
- 缺陷 MUST 分级并有响应与修复时限，MUST NOT 全部等同处理。
- 用户问题与反馈 MUST 有收集入口与处理记录，MUST NOT 只在工单系统里丢失。
- 事故 MUST 有对外沟通负责人与口径模板，MUST NOT 由工程师临场对外解释。
- 服务下线 MUST 有数据归档与销毁方案、调用方通知与下线期限，MUST NOT 悄悄关停。

**当前阶段明确不引入的标准**

以下项已确认当前引入只会产生成本而无收益，MUST NOT 写入当前强制清单；触发条件达成时 MUST 以 MINOR 版本修订本宪法补入。

| 项 | 引入触发条件 |
|---|---|
| ISO/IEC 27001 认证 | 进入政企采购目录、被客户安全问卷点名要求，或需向外部证明体系有效性 |
| 定期第三方渗透测试 | 首次公网生产上线前做 1 次基线；此后每年 1 次，或发生重大架构变更 |
| SLSA Build L3 | 开始向外部客户分发构件，或供应链审计提出要求 |
| TUF / Notary | 需在离线或内网环境安全更新依赖元数据，或需分发私有构件给外部方 |
| DPoP / 发送方约束令牌（RFC 9449） | 对外 API 开放给不可信的第三方原生 App，或发生 Bearer 令牌大规模泄露 |
| 端到端加密 / 隐私增强技术 | 需与外部组织交换敏感数据，或监管对特定场景提出加密要求 |
| HSM / 硬件密钥根 / FIPS 140-3 | 需通过等保三级或金融行业 FIPS 要求，或合同明确要求密钥不出 HSM |
| 等保 2.0 三级测评 | 系统成为关键信息基础设施，或作为政务系统上线 |
| PIPL 个人信息保护认证 / GDPR 认证 | 达到规定的处理个人信息数量阈值，或客户采购要求认证 |
| SOC / 安全运营中心 / 攻击模拟演练 | 服务数 >10 或团队 >10 人，或日均请求量进入需要 7×24 值守的量级 |
| 商业 WAF / API Gateway | 面向公网开放且流量规模大，需要厂商级 DDoS 清洗与高级 bot 防护 |
| 后量子密码迁移 | NIST 发布正式标准且出现"先窃取后解密"的现实威胁；当前 MUST 做的是保持密码敏捷性 |
| OWASP ASVS L3 全量条款 | 接入支付、金融、医疗等高敏数据，或客户合同指定 L3 |
| SPIFFE / SPIRE | 证书生命周期管理成为运维瓶颈，或需跨组织、跨云的服务身份 |
| 变异测试（mutation testing） | 覆盖率已稳定达标但线上缺陷逃逸率仍高时引入，作为覆盖率之外的补充指标 |
| 全量模糊测试（fuzzing） | 存在解析不可信输入的代码路径时引入；当前先依赖 schema 校验与属性测试 |
| 100% 分布式追踪采样 | SLO 排障确需完整链路时临时提升采样率，而非常态全量 |

## Development Workflow & Quality Gates 开发流程与质量门禁

标准流程：**Research → Specify → Clarify → Plan → Tasks → Implement → Analyze → Ship**。该流程由原则 IV 强制，MUST NOT 跳步。

1. **Research & Reuse**：先检索仓库内既有实现与可用依赖库，再决定动手方式。
2. **Specify**：`/speckit-specify` 产出 `spec.md`，含验收标准与范围边界。
3. **Clarify**：`/speckit-clarify` 消除规格歧义并回写 `spec.md`；无歧义时 MUST 显式记录"无需澄清"。
4. **Plan**：`/speckit-plan` 产出 `plan.md`，MUST 明确 API 契约落地形式（含信封字段名）、测试策略与覆盖率预期；MUST 声明本次改动归属哪个微服务、落在该服务的哪一层（`api` / `application` / `domain` / `infrastructure`）、是否触及 `domain` 层、是否新增或变更服务间契约（`.proto`）、是否引入新的跨服务调用或数据访问（原则 V）。若本次新增微服务或为某服务首次选定实现语言，MUST 同时在该 `plan.md` 中记录语言选型（候选语言、对照维度、最终选择与理由、被否决方案），并按 A 组「语言」的规则登记该语言的工具链。若本次涉及持久化或缓存，MUST 在该 `plan.md` 中记录所选 database（默认 MongoDB）、collection 划分、索引与查询路径、以及缓存键设计与 TTL；选用 MongoDB 以外的存储 MUST 附例外论证。MUST 声明本次变更是否触发附加约束 A–G 组的自检（新增服务、新增对外接口、新增外部依赖、新增存储形态，或改动触及安全、可靠性、可观测性标准）；触发时 MUST 在 `plan.md` 中列出受影响的标准条目与落实方式。
5. **Tasks**：`/speckit-tasks` 产出 `tasks.md`，每条任务 MUST 有可验证的完成标准。
6. **Implement**：`/speckit-implement` 按 `tasks.md` 顺序实现；实现业务逻辑前 MUST 先写失败测试（RED）。
7. **Analyze**：`/speckit-analyze` 校验 spec / plan / tasks 的一致性与质量，发现项 MUST 在合并前处置。
8. **Review**：代码评审 MUST 逐条核对本宪法合规性。

合并前质量门禁（全部通过方可合并）：

- **产物门禁**：`spec.md`、`plan.md`、`tasks.md` 三份产物齐备，且与实现一致。
- **流程门禁**：各阶段均由对应的 Spec Kit 命令产生，可从提交历史追溯。
- **架构门禁**：依赖方向检查通过（无内层导入外层）；无跨服务直连他方数据库的访问；改动落在正确的服务与正确的层；`domain` 与 `application` 层可在无数据库、无网络条件下测试（原则 V）。
- **服务边界门禁**：新增微服务有 `spec.md` 支撑并写明业务边界与拆分理由；服务间契约变更有对应契约测试与迁移说明；新增跨服务调用具备超时、幂等与降级语义。
- **选型与数据门禁**：新增微服务或新语言已有 `plan.md` 选型论证且该语言工具链已配置通过；服务内部未混用多门实现语言；服务使用自己独占的 database 且集合名带服务前缀；无跨服务直连他方集合；高频查询路径有索引；缓存条目有 TTL 且键空间按服务隔离；核心写路径不依赖 Redis 可用性。
- **分支管理门禁**：PR 从正确源分支切出（特性分支自 `develop`，紧急修复自 `main` 的 tag）；分支名符合 `{type}/{编号}-{slug}`；目标分支已保护且 CI 全绿；合入方式为 squash；PR 关联任务或 issue；已合并分支已删除。
- **安全门禁**：CI 的 SAST / SCA / 密钥扫描 / 镜像扫描 / IaC 扫描已执行且无 CVSS ≥7.0 未处置项；SBOM、构件签名与 provenance 已生成；对外接口有 BOLA 与功能级授权的回归用例；无硬编码密钥。
- **生产就绪门禁**：新增服务或新增对外接口已按附加约束 A–G 组逐组自检并留下书面结果；SLI/SLO 与错误预算已定义；告警均有 runbook；回滚路径已确认；SLO 燃烧率告警已配置。
- **协作与文档门禁**：变更由非作者评审通过；新增服务有 README，新增告警有 runbook，新增对外接口有文档；PR 规模受控且描述含验证方式。
- **性能与可访问性门禁**：改动对外接口或关键路径时有性能预算校验（p95/p99 不劣化）；改动用户界面时通过可访问性自动检查且满足 WCAG 2.2 AA。
- 格式化、静态检查、类型检查、构建全部通过。
- 覆盖率 ≥ 80%（原则 II），CI 门禁生效。
- 代码评审无 CRITICAL 与 HIGH 级问题；MEDIUM 级问题已修复或已记录原因。
- 无硬编码密钥、无调试残留、无被注释掉的死代码。
- 新增或变更的对外接口具备契约测试与接口文档。
- 任何对原则的例外 MUST 在 PR 描述中书面说明理由并获得评审者确认。

## Governance

- 本宪法优先级高于其他工程实践、团队习惯与既有文档；与之冲突时以本宪法为准，例外只能通过本节的修订流程获得，不接受口头豁免。
- **修订流程**：提出修订 → 评估影响并确定版本位（MAJOR/MINOR/PATCH）→ 更新本文件 → 在文件顶部的 Sync Impact Report 记录变更 → 通知团队 → 生效。
- **版本策略**（语义化版本）：
  - **MAJOR**：删除或重定义既有原则、破坏性治理变更（既有实现将不再合规）。
  - **MINOR**：新增原则或章节、对既有约束的实质性扩展。
  - **PATCH**：措辞澄清、错别字修正、非语义性修订。
- **合规复核**：`/speckit-analyze` 与代码评审 MUST 核对本宪法合规性；每个 feature 的 plan 与 tasks MUST 显式声明其遵循的原则编号；合规性检查 MUST 同时覆盖"spec/plan/tasks 三份产物齐备且与实现一致"（原则 IV）与"分层依赖方向、层职责与服务边界合规"（原则 V）；覆盖率门禁、依赖方向检查与跨服务数据访问检查是唯一可自动判定的合规项，其余由评审者负责。
- **架构决策的变更 MUST 走 MAJOR**：服务形态、服务间协议、数据边界等架构级决策的变更 MUST 以 MAJOR 版本发布，MUST NOT 以 PATCH 或 MINOR 掩盖不兼容变更。
- **复杂度豁免**：任何偏离 MUST 有书面理由并经评审确认，且 MUST 记录在对应 feature 的 plan 或 PR 描述中，不得默认延续。
- 本文件顶部的 Sync Impact Report 是供人工审阅修订的临时材料，**提交前 MUST 删除**。
- 运行时开发指引参见各 feature 的 spec / plan / tasks 文档；本宪法只定义不可协商的底线。

**Version**: 3.0.0 | **Ratified**: 2026-11-12 | **Last Amended**: 2026-11-14
