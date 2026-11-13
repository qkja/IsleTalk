# Changelog

本项目的所有显著变更记录在此。

格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，
版本号遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [未发布]

尚未发布的变更记录在此。

## [0.1.0] - 2026-11-13

首个版本。建立项目治理基线与平台保护，尚无应用代码。

### 新增

- **项目宪法** v2.6.0（`.specify/memory/constitution.md`）：五条核心原则与附加约束 A–G 七组
  - I. 代码质量优先
  - II. 测试覆盖率 ≥ 80%
  - III. API 返回格式统一（信封 `{success, data, message, meta}`）
  - IV. 规格驱动开发（不可协商）
  - V. 标准分层业务架构（微服务 + 服务内 DDD 四层 + gRPC/proto + 服务独占数据）
  - A. 技术栈与依赖选型 · B. API 与数据契约治理 · C. 质量、测试与性能
  - D. 可靠性、可观测性与运行保障 · E. 安全、隐私与供应链合规
  - F. 交付、分支与协作治理 · G. 用户体验与运营支持
- **仓库治理资产**：`LICENSE`（Apache-2.0）、`CONTRIBUTING.md`、`CODEOWNERS`、`SECURITY.md`
- **协作入口**：PR 模板（含合并前自查清单）、Issue 模板（缺陷 / 需求 / 配置）
- **CI**：治理校验作业 + gitleaks 密钥扫描作业
- **治理校验器**（`.github/scripts/check-governance.py`）：检查宪法无临时块、版本行合规、
  必需章节齐全、无未替换占位符、必备文件存在、`specs/` 产物依赖关系成立
- **依赖更新**：`.github/dependabot.yml` 覆盖 GitHub Actions / pip / docker
- **`CLAUDE.md`**：Claude Code 工作指引

### 变更

- 仓库由私有改为公开
- 仓库设置对齐宪法：仅 squash 合并、PR 标题作默认提交信息、合并后自动删除分支
- `main` / `develop` 分支保护：CI 必须全绿、必须同步、禁直接推送、禁 force push、
  删除保护、线性历史、管理员同样受限
- 仓库规则集：`v*` 发布标签禁删除、禁改写
- 启用 GitHub 内置 secret scanning 与 push protection、Dependabot 漏洞告警与安全更新

### 已知限制

- **评审人数的限时例外**：GitHub 禁止 PR 作者批准自己的 PR，单一维护者仓库无法满足
  "≥1 名非作者评审"。当前 `required_approving_review_count` 暂为 0，其余保护全部生效；
  引入第二名评审者后立即恢复为 ≥1 名并修订宪法。
- **语言相关 CI 门禁未就绪**：覆盖率 ≥80%、依赖方向检查、SAST/SCA、契约破坏检测
  待首个服务确定实现语言后补入 `.github/workflows/ci.yml` 的预留位置。
- **待登记项**（`TODO(TECH_STACK)`）：具体实现语言、依赖注入方案、服务发现与配置方案、
  消息队列产品、编排与部署平台、契约测试工具、格式化/静态检查/覆盖率工具链、测试框架选型等。

[未发布]: https://github.com/qkja/IsleTalk/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/qkja/IsleTalk/releases/tag/v0.1.0
