#!/usr/bin/env python3
"""治理校验：检查宪法完整性与规格产物的一致性。

对应宪法要求：
- Sync Impact Report 是临时审阅材料，提交前必须删除
- 宪法 MUST 无未解释的占位符，版本行 MUST 合规
- 核心原则与 A–G 组 MUST 齐全
- spec.md / plan.md / tasks.md 三份产物的依赖关系 MUST 成立

退出码 0 = 通过；1 = 有校验失败项。
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[2]
CONSTITUTION = REPO_ROOT / ".specify" / "memory" / "constitution.md"
SPECS_DIR = REPO_ROOT / "specs"

REQUIRED_FILES = ["README.md", "CLAUDE.md", "CONTRIBUTING.md", "CODEOWNERS", "SECURITY.md"]
REQUIRED_SECTIONS = [
    "### I. Code Quality First",
    "### II. Test Coverage Floor",
    "### III. Unified API Response Format",
    "### IV. Spec-Driven Development",
    "### V. Standard Layered Business Architecture",
    "### A. 技术栈与依赖选型规则",
    "### B. API 与数据契约治理规则",
    "### C. 质量、测试与性能规则",
    "### D. 可靠性、可观测性与运行保障规则",
    "### E. 安全、隐私与供应链合规规则",
    "### F. 交付、分支与协作治理规则",
    "### G. 用户体验与运营支持规则",
    "## Governance",
]
VERSION_LINE = re.compile(
    r"^\*\*Version\*\*:\s*(\d+\.\d+\.\d+)\s*\|\s*"
    r"\*\*Ratified\*\*:\s*(\d{4}-\d{2}-\d{2})\s*\|\s*"
    r"\*\*Last Amended\*\*:\s*(\d{4}-\d{2}-\d{2})\s*$",
    re.MULTILINE,
)
PLACEHOLDER = re.compile(r"\[([A-Z][A-Z0-9_]{2,})\]")
TASK_ENTRY = re.compile(r"^\s*-\s*\[[ xX]\]\s+\S", re.MULTILINE)

failures: list[str] = []


def fail(message: str) -> None:
    failures.append(message)


def check_constitution() -> None:
    if not CONSTITUTION.is_file():
        fail("缺少 .specify/memory/constitution.md")
        return

    text = CONSTITUTION.read_text(encoding="utf-8")

    if "SYNC IMPACT REPORT" in text:
        fail(
            "宪法中残留 Sync Impact Report 临时块 —— 提交前必须删除"
            "（治理条款：它是供人工审阅修订的临时材料）"
        )

    version_match = VERSION_LINE.search(text)
    if not version_match:
        fail(
            "宪法的版本行格式不合规，应为："
            "**Version**: X.Y.Z | **Ratified**: YYYY-MM-DD | **Last Amended**: YYYY-MM-DD"
        )

    for section in REQUIRED_SECTIONS:
        if section not in text:
            fail(f"宪法缺少必需章节：{section}")

    placeholders = sorted(set(PLACEHOLDER.findall(text)))
    if placeholders:
        fail(f"宪法残留未替换的占位符：{', '.join(placeholders)}")


def check_required_files() -> None:
    for relative in REQUIRED_FILES:
        if not (REPO_ROOT / relative).is_file():
            fail(f"缺少仓库必备文件：{relative}")


def check_specs() -> None:
    """产物依赖关系：plan 需要 spec，tasks 需要 plan。"""
    if not SPECS_DIR.is_dir():
        return

    for feature_dir in sorted(p for p in SPECS_DIR.iterdir() if p.is_dir()):
        has_spec = (feature_dir / "spec.md").is_file()
        has_plan = (feature_dir / "plan.md").is_file()
        has_tasks = (feature_dir / "tasks.md").is_file()

        if has_plan and not has_spec:
            fail(f"{feature_dir.name}/ 有 plan.md 但缺少 spec.md")
        if has_tasks and not has_plan:
            fail(f"{feature_dir.name}/ 有 tasks.md 但缺少 plan.md")
        if has_tasks:
            body = (feature_dir / "tasks.md").read_text(encoding="utf-8")
            if not TASK_ENTRY.search(body):
                fail(f"{feature_dir.name}/tasks.md 中没有可执行的任务条目")


def main() -> int:
    check_constitution()
    check_required_files()
    check_specs()

    if failures:
        print("治理校验失败：")
        for item in failures:
            print(f"  ✗ {item}")
        return 1

    print("治理校验通过：宪法完整、占位符已替换、版本行合规、必备文件齐全、产物依赖关系成立。")
    return 0


if __name__ == "__main__":
    sys.exit(main())
