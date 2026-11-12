# IsleTalk

## 项目状态

项目尚未定义具体产品需求。当前仓库只包含**开发治理与流程基础设施**——Spec Kit 工具链与项目宪法。

一切功能开发 MUST 遵循 `/speckit-specify` 起的标准流程（见下文）。

## 开发原则

项目宪法位于 [`.specify/memory/constitution.md`](.specify/memory/constitution.md)，是本项目的最高规则来源，与其他文档冲突时以宪法为准。

| 原则 | 核心约束 |
|---|---|
| I. 代码质量优先 | 函数 ≤50 行、文件 ≤800 行；错误显式处理；输入边界校验；密钥走环境变量 |
| II. 测试覆盖率 ≥ 80% | CI 硬门禁；关键路径接近 100% 分支覆盖；禁止删测试刷数字 |
| III. API 返回格式统一 | 统一信封 `{success, data, message, meta}`；错误码唯一注册表；列表必须分页 |
| IV. 规格驱动开发（不可协商） | specify → clarify → plan → tasks → implement → analyze，禁止跳步 |
| V. 标准分层业务架构 | 微服务 + 服务内 DDD 四层（api → application → domain → infrastructure）；gRPC/proto 契约；服务独占数据 |

## 技术架构决策

| 维度 | 决策 |
|---|---|
| 系统形态 | 微服务，每个服务内部 DDD 四层分层 |
| 服务间通信 | gRPC 同步为主，契约以 `.proto` 定义并版本化；异步事件走专用消息队列 |
| 数据边界 | database per service，禁止跨服务直连他方数据库 |
| 持久化 | MongoDB 为默认（例外须论证）；集合名带服务前缀 |
| 缓存与协调 | Redis：cache-aside 缓存、分布式锁、限流、会话；不得作唯一事实来源 |
| 实现语言 | 多语言栈——每个服务可选最适配语言，服务内部统一；选型须在 `plan.md` 论证 |
| 本地依赖 | Docker Compose 编排，镜像版本固定 |

## 标准流程

```
/speckit-specify    → spec.md    功能规格与验收标准
/speckit-clarify    → 回写 spec   消除歧义
/speckit-plan       → plan.md    技术方案、契约、测试策略
/speckit-tasks      → tasks.md   依赖有序、可验证的任务清单
/speckit-implement  → 实现        按 tasks 顺序，先写失败测试
/speckit-analyze    → 复核        spec/plan/tasks 一致性
```

辅助命令：`/speckit-checklist`、`/speckit-converge`、`/speckit-taskstoissues`。

## 仓库结构

```
.specify/          Spec Kit 工具链：模板、脚本、流程定义、宪法
  memory/constitution.md    项目宪法（最高规则来源）
  templates/               各阶段产物模板
  scripts/powershell/      模板解析与阶段准备脚本
  workflows/speckit/       完整 SDD 流程定义
.claude/skills/    Spec Kit 技能（Claude Code 侧的命令入口）
```

## 待补充

- `TODO(TECH_STACK)`：具体实现语言、依赖注入方案、服务发现与配置方案、消息队列产品、编排与部署平台、契约测试工具等，按服务逐个确定后登记入宪法。
- 对外（北向）API 协议尚未绑定，由各 feature 的 `plan.md` 确定。
