# IsleTalk Constitution

## Core Principles 核心原则

### I. Code Quality First 代码质量优先

代码质量是首要目标，不以交付速度牺牲可读性、可测试性与安全性。

- 每项功能在编码前 MUST 有明确的需求规格与实现计划；需求未澄清时 MUST 先澄清，不得凭猜测编码。
- 代码 MUST 通过仓库配置的格式化、静态检查、类型检查与构建；所有自动化检查全绿是提交的前提。
- 函数 MUST 保持单一职责：单函数不超过 50 行、嵌套不超过 4 层；文件 SHOULD 控制在 200–400 行，800 行为硬上限，超限 MUST 拆分模块。
- 错误 MUST 显式处理，不得静默吞掉；面向使用者的错误信息要可读可行动，日志侧保留完整上下文。
- 所有外部输入（请求参数、第三方响应、文件内容）MUST 在系统边界校验后再处理，不得信任外部数据。
- 密钥、令牌、口令 MUST 通过环境变量或密钥管理注入，MUST NOT 硬编码；启动时 MUST 校验必需密钥存在。
- MUST NOT 提交调试残留（调试打印、断点语句、注释掉的死代码、临时测试开关）。
- 新增抽象、新增依赖、新增模式 MUST 有书面理由，说明它解决的真实痛点；无对应痛点时按 YAGNI 拒绝引入。
- 动手实现前 MUST 先检索仓库内既有实现与可用依赖，优先复用，禁止重复造轮子（DRY）。
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
- 列表与查询接口 MUST 分页，MUST NOT 无界返回；单次返回量 MUST 有上限。
- 对外接口 MUST 有入参校验与速率限制；状态变更接口 MUST 有鉴权与 CSRF 防护（Web 场景）。
- 响应结构的任何变更都视为契约变更：MUST 先更新契约测试（纳入 CI），破坏性变更 MUST 版本化并在 plan 中写明迁移方案。
- 契约 MUST 有可执行的契约测试或 schema 校验，防止信封结构在实现漂移中走形。

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

**二、服务边界与形态规则**

- **服务拆分 MUST 按业务边界（限界上下文 / 业务域）**：一个服务 MUST 对应一项清晰的业务能力；MUST NOT 以"user-service / order-dao / api-gateway"这类技术角色切分。
- **服务粒度 MUST 满足单一职责**：服务 MUST 只暴露一组内聚的业务能力；MUST NOT 出现"所有业务都往里塞"的通用服务。
- **每个服务 MUST 独立自治**：独立进程、独立部署、独立扩缩容；MUST NOT 依赖其他服务同时在线才能启动或就绪。
- **MUST NOT 为"将来可能需要"而预先拆分服务**：新增微服务 MUST 有对应的 `spec.md` 支撑，写明其独立的业务边界与拆分理由。
- **MUST NOT 共享内部状态**：服务间只经由契约交互，MUST NOT 共享内存、共享库表或共享缓存对象。

**三、数据边界规则**

- **数据边界 MUST 服务独占（database per service）**：每个服务 MUST 独占自己的数据库或 schema；MUST NOT 跨服务直接读写对方的库或表；跨服务取数 MUST 经由对方接口或事件。
- **在 MongoDB 上 MUST 以 database 为隔离单位**：每个服务 MUST 使用自己独占的 database（必要时独立实例）；MUST NOT 跨服务访问对方的 database 或集合，即便同一实例亦然。集合命名 MUST 带服务前缀（`{service}_{entity}`），MUST NOT 出现无归属前缀的集合名。
- **跨服务数据一致性 MUST 走最终一致性**：状态变更 SHOULD 通过事件/消息传播；订阅方 MUST 实现幂等消费；MUST NOT 用跨服务同步调用模拟分布式事务。

**四、服务间通信与契约规则**

- **同步调用 MUST 使用 gRPC**：契约 MUST 以 `.proto` 定义并纳入版本控制；契约 MUST 由提供方与调用方共同持有，且双方 MUST 基于同一份 `.proto` 生成代码。
- **契约变更 MUST 向后兼容或显式版本化**：破坏性变更 MUST 通过 proto package 版本管理，并 MUST 给出双跑迁移方案与下线计划；MUST NOT 直接原地破坏调用方。
- **契约 MUST 有契约测试**：服务提供方与调用方 MUST 各自持有基于同一份 `.proto` 的契约测试并纳入 CI；MUST NOT 在无契约测试的情况下变更服务间接口。
- **跨服务调用 MUST 具备容错语义**：每个调用 MUST 设置超时；重试 MUST 仅用于幂等操作；MUST 具备熔断、降级与限流策略；调用失败 MUST 显式处理，不得静默吞掉。
- **写操作 MUST 幂等**：跨服务的状态变更 MUST 携带幂等键（幂等 ID），以抵御重复投递与重试。
- **MUST NOT 有跨服务的编译期或发布期强耦合**：一个服务的发布 MUST NOT 要求其他服务同步发布。

**五、可观测性与工程约束**

- **每个服务 MUST 可独立观测**：MUST 暴露健康检查（health check）与就绪检查（readiness），MUST 输出结构化日志、指标与链路追踪；MUST NOT 在日志或指标中泄露敏感信息。
- **架构边界 MUST 有自动化检查纳入 CI**：服务内依赖方向检查（禁止内层导入外层）+ 禁止跨服务直连他方数据库的检查；具体工具随技术栈选型确定（见 `TODO(TECH_STACK)`）。
- **若某 feature 需要跨层捷径（如为性能绕过用例层直连存储），MUST 在 `plan.md` 中书面论证、记录豁免与回归验证方式，MUST NOT 成为常态。**

**Rationale**：微服务把边界从代码目录提升为进程与网络边界——好处是各服务能独立演进、独立伸缩、独立失败；代价是边界一旦含糊，代价会以分布式事务、循环依赖和无法独立发布的连锁故障的形式一次性兑现。因此本原则的重心不在"用了多少个服务"，而在三条可检查的底线：按业务边界拆分、数据独占、服务间只经显式契约交互。三者共同保证任何单个服务的故障不会扩散成全局故障。服务内保留 DDD 四层，是因为跨进程的成本越高，层内边界就越要靠纪律维持；反过来，若服务内已经是"随手写"，拆成十个服务只会把十个随手写变成十个不可调试的分布式黑盒。

## Additional Constraints 附加约束

- **技术栈中立**：本宪法不预设编程语言、框架与部署形态。
  - TODO(TECH_STACK): **存储与缓存已定**（2026-11-13：MongoDB 为默认持久化、Redis 承担缓存/锁/限流/会话，本地依赖以 Docker 编排启动）；**架构范式已定**（2026-11-13：微服务 + 服务内 DDD 四层分层，服务间 gRPC/proto，数据服务独占）。剩余待定项：具体实现语言（按服务逐个确定并登记）、依赖注入方案、依赖方向检查工具、服务发现与配置方案、消息队列具体产品、编排与部署平台、契约测试工具、格式化/静态检查/覆盖率工具链、测试框架选型。每引入一门语言或一项关键组件，MUST 在本章节登记其工具链与配置细则，登记动作以 MINOR 版本修订本宪法；未登记的服务 MUST NOT 合入主干。

- **服务目录与模块 MUST 与分层对应**：每个微服务 MUST 具备可识别的 `api` / `application` / `domain` / `infrastructure` 四层目录（命名可按技术栈惯例微调，职责边界 MUST 不变）；每层内部 MUST 按业务能力进一步划分，MUST NOT 再按 `models` / `services` / `controllers` 这类技术角色横向切分。
- **服务间契约 MUST 版本化**：`.proto` MUST 纳入版本控制并按包名管理版本；契约变更 MUST 有对应的契约测试与迁移说明；废弃契约 MUST 有明确的下线计划而非长期并存。
- **跨服务调用 MUST 有超时、幂等与降级**：MUST NOT 存在无超时的跨服务调用；MUST NOT 对非幂等写操作做自动重试。
- **CI MUST 包含两类架构检查**：服务内依赖方向检查（禁止内层导入外层）+ 跨服务数据访问检查（禁止直连他方数据库/schema）。
- 依赖引入 MUST 评估维护活跃度、许可证兼容性与已知安全漏洞；MUST NOT 引入功能重叠的重复依赖。
- 敏感信息（密钥、令牌、个人身份信息）MUST NOT 进入代码库或日志；日志 MUST NOT 输出完整身份证号、手机号、银行卡号等。
- 并发场景 MUST 有明确的同步/隔离策略；共享状态优先通过不可变数据与纯函数消除。
- 查询与列表 MUST 分页并限制单次返回量；MUST NOT 出现 N+1 查询，必要时 MUST 改写为批量查询或 JOIN。
- 错误处理 MUST 保留可诊断信息：错误码、关键上下文、来源链路。

### 数据与中间件选型规则

**持久化（MongoDB 为默认）**

- **持久化 MUST 默认使用 MongoDB**：每个服务 MUST 使用自己独占的 database（必要时独立实例）。选用其他存储形态属例外，MUST 在该 feature 的 `plan.md` 中论证理由、替代方案与退出成本，并经评审确认。
- **MUST NOT 跨服务直连他方数据库**：即便同实例不同 database，MUST NOT 由一个服务直接读写另一个服务的 database 或集合；跨服务取数 MUST 经由对方接口或事件。
- **集合命名 MUST 带服务前缀**（`{service}_{entity}`），MUST NOT 出现无归属前缀的集合名。
- **领域模型 MUST 与 BSON 文档解耦**：领域实体 MUST NOT 直接使用驱动文档类型（如 `ObjectId`）作为领域标识；MUST 定义自己的值类型（如自有 ID 类型），并在 `infrastructure` 层显式做映射。
- **MUST NOT 用文档结构掩盖关系建模**：MUST NOT 把本应独立建模的数据塞进一个不断膨胀的嵌套文档；MUST NOT 依赖无界数组累积，单文档 MUST NOT 逼近 BSON 大小上限。
- **查询 MUST 有索引**：高频查询路径 MUST 建索引；MUST NOT 依赖无索引的全集合扫描；MUST NOT 在查询条件中做无法走索引的计算。
- **列表 MUST 有上限**：MUST 使用稳定排序游标；MUST NOT 无界返回，skip 分页 MUST 有上限。
- **MUST 以单文档原子性为设计依据**：需要多文档原子性时 MUST 明确说明并评估 replica set 事务代价；MUST NOT 用应用层"先查后写"模拟原子性。

**缓存与协调（Redis）**

- **Redis MUST NOT 成为唯一事实来源**：持久化事实 MUST 存于 MongoDB；Redis 中的数据 MUST 能在缓存失效后重建。
- **缓存 MUST 采用 cache-aside**：读时先查缓存，未命中回源并回填；MUST NOT 把 Redis 作为写入主路径。
- **每个缓存条目 MUST 有 TTL**，MUST NOT 出现无过期时间的键；MUST NOT 缓存无限增长的集合。
- **回源 MUST 有并发抑制**：MUST NOT 允许缓存失效瞬间大量并发回源（须有并发上限或过期时间抖动）；MUST NOT 把缓存当作限流的唯一手段。
- **Redis key MUST 按服务命名空间隔离**（`{service}:{entity}:{id}`），MUST NOT 出现无前缀的共享键空间；MUST NOT 把本服务的键直接交由其他服务读写。
- **分布式锁 MUST 带 TTL 与持有者标识**：MUST NOT 依赖进程内互斥充当分布式锁；MUST NOT 在锁可能已失效后继续假设其有效；MUST NOT 用锁掩盖幂等设计的缺失。
- **限流 MUST 口径统一**：限流策略 MUST 由服务自身或网关统一实施，MUST NOT 各服务各写一套口径。
- **会话/令牌 MUST 明确有效期与撤销策略**（若由 Redis 承担）。
- **Redis 故障 MUST NOT 导致数据丢失**：核心写路径 MUST NOT 依赖 Redis 可用；降级行为 MUST 明确（缓存不可用时回源；锁不可用时 fail-closed 还是 fail-open MUST 按场景显式决定并记录）。
- **异步事件 MUST 走专用消息队列**：MUST NOT 把 Redis 当作消息队列使用（缓存/协调与事件投递职责 MUST 分离）。

**本地环境与依赖编排（Docker）**

- **本地开发依赖 MUST 通过 Docker 编排启动**：MongoDB、Redis 与消息中间件等 MUST 以 Docker Compose（或等价编排）方式启动；镜像版本 MUST 固定（MUST NOT 使用 `latest` 标签）。
- **MUST NOT 连接生产或他人共享环境**：本地与测试环境的连接信息 MUST 与生产隔离；凭据 MUST NOT 硬编码进编排文件，MUST 经环境变量或密钥注入。
- **集成测试 MUST 针对容器化依赖运行**：测试依赖的版本 MUST 与生产同大版本；测试数据 MUST 可重复初始化与清理；测试 MUST NOT 依赖手工预置的数据状态。

### 语言选型规则

- **选型 MUST 以"最适配"为准绳**：结合该服务的业务特性、并发模型、性能与资源约束、团队现有能力、既有基础设施在目标语言下的成熟度（gRPC/protobuf 支持、消息中间件客户端、数据库驱动与 ORM、可观测性 SDK）择优。 MUST NOT 以个人偏好、"顺手"或"这个语言写起来快"为由选型。
- **选型 MUST 在该 feature 的 `plan.md` 中留痕**：MUST 记录候选语言、对照维度、最终选择与理由，以及被否决方案及其否决原因。
- **选型 MUST 经评审确认**：MUST NOT 无理由引入一门新语言；评审者 MUST 检查选型理由是否指向该服务的真实约束。
- **一个微服务 MUST 使用单一语言**：服务内部 MUST NOT 混用多门实现语言；跨语言需求 MUST 通过契约委托给拥有该能力的其他服务。
- **多语言栈 MUST 有边界**：不同服务可用不同语言，但同一类能力（例如同一类基础设施适配器、同一侧的写服务）MUST 保持语言一致，避免同类逻辑出现多套实现与多套运维手法。
- **每种语言 MUST 具备完整工具链**：格式化、静态检查、类型检查、测试、覆盖率、依赖版本锁定与依赖管理 MUST 按该语言的官方标准工具链配置。 MUST NOT 只配测试而不配静态检查，也 MUST NOT 依赖人工约定代替工具链。
- **跨语言 MUST 只经由契约交互**：服务间 MUST 只通过 `.proto` 契约与事件契约交互（原则 V），MUST NOT 出现语言特有的进程内耦合或私有协议。
- **MUST NOT 重复实现同一能力**：当目标语言生态缺少某能力的成熟实现时，MUST 通过契约委托给具备该能力的语言服务，MUST NOT 在多语言栈中各写一份。
- **语言运行时与工具链版本 MUST 锁定并记录**：版本写入项目配置与构建文件，MUST NOT 依赖开发机上的浮动版本。
- **每引入一门语言 MUST 评估运维代价**：MUST 同时具备构建、部署、监控与排障手段；无法满足该条件的语言 MUST NOT 引入。

## Development Workflow & Quality Gates 开发流程与质量门禁

标准流程：**Research → Specify → Clarify → Plan → Tasks → Implement → Analyze → Ship**。该流程由原则 IV 强制，MUST NOT 跳步。

1. **Research & Reuse**：先检索仓库内既有实现与可用依赖库，再决定动手方式。
2. **Specify**：`/speckit-specify` 产出 `spec.md`，含验收标准与范围边界。
3. **Clarify**：`/speckit-clarify` 消除规格歧义并回写 `spec.md`；无歧义时 MUST 显式记录"无需澄清"。
4. **Plan**：`/speckit-plan` 产出 `plan.md`，MUST 明确 API 契约落地形式（含信封字段名）、测试策略与覆盖率预期；MUST 声明本次改动归属哪个微服务、落在该服务的哪一层（`api` / `application` / `domain` / `infrastructure`）、是否触及 `domain` 层、是否新增或变更服务间契约（`.proto`）、是否引入新的跨服务调用或数据访问（原则 V）。若本次新增微服务或为某服务首次选定实现语言，MUST 同时在该 `plan.md` 中记录语言选型（候选语言、对照维度、最终选择与理由、被否决方案），并按附加约束的"语言选型规则"登记该语言的工具链。若本次涉及持久化或缓存，MUST 在该 `plan.md` 中记录所选 database（默认 MongoDB）、collection 划分、索引与查询路径、以及缓存键设计与 TTL；选用 MongoDB 以外的存储 MUST 附例外论证。
5. **Tasks**：`/speckit-tasks` 产出 `tasks.md`，每条任务 MUST 有可验证的完成标准。
6. **Implement**：`/speckit-implement` 按 `tasks.md` 顺序实现；实现业务逻辑前 MUST 先写失败测试（RED）。
7. **Analyze**：`/speckit-analyze` 校验 spec / plan / tasks 的一致性与质量，发现项 MUST 在合并前处置。
8. **Review**：代码评审 MUST 逐条核对本宪法合规性。

合并前质量门禁（全部通过方可合并）：

- **产物门禁**：`spec.md`、`plan.md`、`tasks.md` 三份产物齐备，且与实现一致。
- **流程门禁**：各阶段均由对应的 Spec Kit 命令产生，可从提交历史追溯。
- **架构门禁**：依赖方向检查通过（无内层导入外层）；无跨服务直连他方数据库的访问；改动落在正确的服务与正确的层；`domain` 与 `application` 层可在无数据库、无网络条件下测试（原则 V）。
- **服务边界门禁**：新增微服务有 `spec.md` 支撑并写明业务边界与拆分理由；服务间契约变更有对应契约测试与迁移说明；新增跨服务调用具备超时、幂等与降级语义。
- **语言选型门禁**：新增微服务或新语言已有 `plan.md` 中的选型论证；该语言的工具链（格式化/静态检查/类型检查/测试/覆盖率）已配置并通过；服务内部未混用多门实现语言。
- **数据与中间件门禁**：服务使用自己独占的 database 且集合名带服务前缀；无跨服务直连他方集合；高频查询路径有索引；缓存条目有 TTL 且键空间按服务隔离；核心写路径不依赖 Redis 可用性。
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

**Version**: 2.2.0 | **Ratified**: 2026-11-12 | **Last Amended**: 2026-11-13
