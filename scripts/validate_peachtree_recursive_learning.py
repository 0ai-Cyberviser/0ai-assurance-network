#!/usr/bin/env python3
"""Validate Peachtree recursive learning fixture records.

This validator is intentionally conservative. It checks schema, policy posture,
and redaction invariants only. It does not execute fuzzing, training, signing, or
network actions.
"""
from __future__ import annotations

import json
import sys
from pathlib import Path
from typing import Any

ALLOWED_SCHEMA_VERSION = "peachtree.recursive_learning.v1"
ALLOWED_POLICY_DECISIONS = {"allow_dry_run", "human_review_required", "blocked"}
ALLOWED_DATASET_ACTIONS = {"retain", "hash_only", "discard", "human_review"}
BLOCKED_SUBSTRINGS = (
    "private_key",
    "secret_key",
    "api_key",
    "bearer ",
    "password",
    "token=",
    "authorization:",
    "cookie:",
)


def _fail(path: Path, message: str) -> str:
    return f"{path}: {message}"


def _contains_blocked_secret(value: Any) -> bool:
    serialized = json.dumps(value, sort_keys=True).lower()
    return any(item in serialized for item in BLOCKED_SUBSTRINGS)


def validate_record(path: Path) -> list[str]:
    errors: list[str] = []
    try:
        record = json.loads(path.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        return [_fail(path, f"invalid JSON: {exc}")]

    if record.get("schema_version") != ALLOWED_SCHEMA_VERSION:
        errors.append(_fail(path, "schema_version must be peachtree.recursive_learning.v1"))

    required_fields = {
        "cycle_id",
        "artifact_type",
        "artifact_path",
        "artifact_sha256",
        "validation_commands",
        "sandbox_required",
        "policy_decision",
        "findings",
        "dataset_candidate",
        "self_improvement",
    }
    missing = sorted(required_fields - set(record))
    if missing:
        errors.append(_fail(path, f"missing required fields: {', '.join(missing)}"))

    if not str(record.get("artifact_sha256", "")).startswith("sha256:"):
        errors.append(_fail(path, "artifact_sha256 must start with sha256:"))

    if record.get("sandbox_required") is not True:
        errors.append(_fail(path, "sandbox_required must be true"))

    if record.get("policy_decision") not in ALLOWED_POLICY_DECISIONS:
        errors.append(_fail(path, "policy_decision is not allowed"))

    commands = record.get("validation_commands", [])
    if not isinstance(commands, list) or not commands:
        errors.append(_fail(path, "validation_commands must be a non-empty list"))
    elif any(not isinstance(command, str) for command in commands):
        errors.append(_fail(path, "validation_commands must contain only strings"))

    dataset = record.get("dataset_candidate", {})
    if dataset.get("dataset_action") not in ALLOWED_DATASET_ACTIONS:
        errors.append(_fail(path, "dataset_candidate.dataset_action is not allowed"))

    if record.get("policy_decision") in {"human_review_required", "blocked"}:
        if record.get("self_improvement", {}).get("requires_human_approval") is not True:
            errors.append(_fail(path, "human-review or blocked records must require human approval"))

    if _contains_blocked_secret(record):
        errors.append(_fail(path, "record appears to contain a blocked secret/token marker"))

    return errors


def main(argv: list[str]) -> int:
    if not argv:
        print("usage: validate_peachtree_recursive_learning.py <file-or-directory> [...]")
        return 2

    paths: list[Path] = []
    for arg in argv:
        path = Path(arg)
        if path.is_dir():
            paths.extend(sorted(path.glob("*.json")))
        else:
            paths.append(path)

    errors: list[str] = []
    for path in paths:
        errors.extend(validate_record(path))

    if errors:
        print("Peachtree recursive learning validation failed:")
        for error in errors:
            print(f"- {error}")
        return 1

    print(f"Validated {len(paths)} Peachtree recursive learning record(s).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
