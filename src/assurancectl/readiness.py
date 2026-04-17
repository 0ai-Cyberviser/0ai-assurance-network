"""Launch-readiness reporting for the 0AI Assurance Network skeleton."""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

from .config import LoadedConfig


@dataclass(frozen=True)
class ReadinessReport:
    status: str
    score: int
    strengths: list[str]
    blockers: list[str]
    next_actions: list[str]


def _exists(root: Path, relative_path: str) -> bool:
    return (root / relative_path).exists()


def build_readiness_report(config: LoadedConfig) -> ReadinessReport:
    strengths: list[str] = []
    blockers: list[str] = []
    next_actions: list[str] = []

    if config.topology["mode"] == "permissioned_testnet":
        strengths.append("permissioned testnet mode is enforced")
    if config.genesis["governance"]["dual_house_enabled"]:
        strengths.append("dual-house governance is enabled in genesis")
    if not config.policy["safe_mode_defaults"]["public_transferability"]:
        strengths.append("public transferability is disabled by default")
    if "external_security_audit" in config.policy["required_before_public_launch"]:
        strengths.append("external security audit is a declared launch gate")

    required_artifacts = {
        "docs/threat-model.md": "missing threat model",
        "docs/security-assumptions.md": "missing security assumptions",
        "docs/token-disclosure.md": "missing token disclosure baseline",
        "docs/legal-questions.md": "missing legal questions tracker",
    }
    for relative_path, message in required_artifacts.items():
        if not _exists(config.root, relative_path):
            blockers.append(message)

    if len(config.topology["validators"]) < 7:
        blockers.append("validator set is below the recommended seven-validator launch floor")

    if blockers:
        next_actions.extend(
            [
                "write the missing governance and legal operator docs",
                "expand the validator plan to at least seven launch candidates",
                "prepare an external legal and security review package",
            ]
        )
        status = "not_ready"
        score = max(10, 70 - (10 * len(blockers)))
    else:
        next_actions.extend(
            [
                "promote the repo skeleton into a dedicated repository",
                "add chain binary scaffolding and deterministic key tooling",
                "schedule external review and a permissioned testnet exercise",
            ]
        )
        status = "incubating"
        score = 82

    return ReadinessReport(
        status=status,
        score=score,
        strengths=strengths,
        blockers=blockers,
        next_actions=next_actions,
    )
