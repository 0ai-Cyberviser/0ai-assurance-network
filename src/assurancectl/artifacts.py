"""Versioned governance artifact helpers."""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any


GOVERNANCE_ARTIFACT_SCHEMA = "0ai.assurance.governance.artifact"
GOVERNANCE_ARTIFACT_SCHEMA_VERSION = "1.0.0"
GOVERNANCE_ARTIFACT_COMPATIBILITY = {
    "breaking_change": "increment major",
    "additive_change": "increment minor",
    "clarification_change": "increment patch",
}


def governance_artifact(
    *,
    artifact_type: str,
    command: str,
    payload: Any,
    sources: dict[str, Any] | None = None,
) -> dict[str, Any]:
    return {
        "schema": GOVERNANCE_ARTIFACT_SCHEMA,
        "schema_version": GOVERNANCE_ARTIFACT_SCHEMA_VERSION,
        "compatibility": GOVERNANCE_ARTIFACT_COMPATIBILITY,
        "artifact_type": artifact_type,
        "command": command,
        "sources": sources or {},
        "payload": payload,
    }


def write_artifact(path: str | Path, artifact: dict[str, Any]) -> Path:
    target = Path(path)
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(json.dumps(artifact, indent=2) + "\n", encoding="utf-8")
    return target
