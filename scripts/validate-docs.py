#!/usr/bin/env python3
"""Validate EchoLine documentation against repo facts. Exit 1 on errors."""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

# Paths that must not appear in living docs (docs/, root agent files)
GHOST_PATHS = [
    "backend/internal/api/",
    "backend/api/main.go",
    "backend/worker/",
    "backend/internal/tracing/",
    "backend/internal/group/",
    "backend/internal/channel/handler",
    "backend/internal/social/",
    "backend/internal/conversation/roles.go",
    "docs/adr/0003-cache-and-mq.md",
    "docs/research/",
]

LIVING_GLOBS = [
    "docs/**/*.md",
    "AGENTS.md",
    "CLOUD_AGENT_PROMPT.md",
    "CONTEXT_COMPACTION.md",
    "CURRENT_STATE.md",
    "NEXT_ACTIONS.md",
    "RESEARCH_PLAN.md",
    "README.md",
]

# Allowed in historical manifests/reports if header contains disclaimer keyword
HISTORICAL_FILES = {
    "BATCH_100_MANIFEST.md",
    "BATCH_120_MANIFEST.md",
    "BATCH_NEXT_120_MANIFEST.md",
    "BATCH_NEXT_200_MANIFEST.md",
    "reports/review-api-consistency.md",
    "reports/review-reliability.md",
    "reports/review-performance.md",
    "reports/review-security.md",
    "reports/review-test-coverage.md",
}

WRONG_API_PATTERNS = [
    (re.compile(r"GET /api/sync"), "POST /api/sync (JSON body with device_id + cursors)"),
    (re.compile(r"POST /api/conversations/dm\b"), "POST /api/conversations/direct"),
    (re.compile(r"POST /api/conversations/group\b"), "POST /api/conversations/groups"),
    (re.compile(r"message\.updated"), "message.edited"),
]


def living_files() -> list[Path]:
    out: list[Path] = []
    for pattern in LIVING_GLOBS:
        out.extend(ROOT.glob(pattern))
    return sorted(set(out))


def check_ghost_paths() -> list[str]:
    errors: list[str] = []
    for path in living_files():
        rel = path.relative_to(ROOT).as_posix()
        text = path.read_text(encoding="utf-8")
        for ghost in GHOST_PATHS:
            if ghost in text:
                errors.append(f"{rel}: contains ghost path `{ghost}`")
    return errors


def check_wrong_api_docs() -> list[str]:
    errors: list[str] = []
    targets = list((ROOT / "docs").rglob("*.md")) + [
        ROOT / "reports/review-performance.md",
        ROOT / "reports/review-api-consistency.md",
    ]
    for path in targets:
        rel = path.relative_to(ROOT).as_posix()
        for i, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
            for pat, hint in WRONG_API_PATTERNS:
                if pat.search(line):
                    errors.append(f"{rel}:{i}: use {hint}")
    return errors


def check_openapi_routes() -> list[str]:
    import yaml

    spec = yaml.safe_load((ROOT / "docs/openapi.yaml").read_text())
    openapi_paths = set(spec.get("paths", {}).keys())

    server_go = (ROOT / "backend/internal/server/server.go").read_text()
    routes = set(re.findall(r'"(GET|POST|PUT|PATCH|DELETE) ([^"]+)"', server_go))
    # login route in options.go
    options = (ROOT / "backend/internal/server/options.go").read_text()
    routes.update(re.findall(r'"(GET|POST|PUT|PATCH|DELETE) ([^"]+)"', options))

    errors: list[str] = []
    for method, route in sorted(routes):
        path = route.split(",")[0].strip()
        if path not in openapi_paths:
            errors.append(f"openapi.yaml missing path {path} ({method})")
    return errors


def main() -> int:
    errors: list[str] = []
    errors.extend(check_ghost_paths())
    errors.extend(check_wrong_api_docs())
    try:
        errors.extend(check_openapi_routes())
    except Exception as e:
        errors.append(f"openapi check failed: {e}")

    if errors:
        print("Documentation validation FAILED:\n")
        for e in errors:
            print(f"  - {e}")
        return 1

    print("Documentation validation OK")
    return 0


if __name__ == "__main__":
    sys.exit(main())
