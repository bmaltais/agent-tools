#!/usr/bin/env python3
"""Strict registry validation for local use and CI parity."""

from __future__ import annotations

import json
import sys
from pathlib import Path

from jsonschema import validate


def main() -> int:
    root = Path(__file__).resolve().parent.parent
    registry_path = root / "tools" / "registry.json"
    schema_path = root / "tools" / "registry.schema.json"
    overlay_path = root / "tools" / "overlays" / "copilot.json"

    registry = json.loads(registry_path.read_text())
    schema = json.loads(schema_path.read_text())
    overlay = json.loads(overlay_path.read_text())

    validate(instance=registry, schema=schema)

    if registry["sm"] != overlay["sm"]:
        raise SystemExit(
            f"schema major mismatch: registry sm={registry['sm']} overlay sm={overlay['sm']}"
        )

    seen: set[str] = set()
    for tool in registry["t"]:
        if tool["id"] in seen:
            raise SystemExit(f"duplicate tool id: {tool['id']}")
        seen.add(tool["id"])

        bin_name = tool["bin"]
        main_go = root / "cmd" / bin_name / "main.go"
        if not main_go.exists():
            raise SystemExit(f"tool references unknown binary path: {main_go}")

    overlay_tools = set(overlay["t"].keys())
    registry_tools = set(tool["id"] for tool in registry["t"])
    unknown_overlay = overlay_tools - registry_tools
    if unknown_overlay:
        raise SystemExit(f"overlay references unknown tools: {sorted(unknown_overlay)}")

    print("registry validation passed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
