"""Zero-day threat detection for governance proposals."""

from __future__ import annotations

import json
from dataclasses import dataclass
from pathlib import Path
from typing import Any

from .config import LoadedConfig


@dataclass(frozen=True)
class ThreatDetectionReport:
    """Report from threat detection analysis."""

    proposal_id: str
    title: str
    threat_level: str
    threat_score: int
    vulnerability_categories: list[str]
    attack_vectors: list[str]
    triggered_patterns: list[str]
    security_signals: list[str]
    rationale: list[str]
    security_remediation: list[str]
    requires_security_review: bool
    blocks_execution: bool
    requires_escalation: bool
    summary: str


def load_threat_detection_policy(config: LoadedConfig) -> dict[str, Any]:
    """Load threat detection policy configuration."""
    policy_path = config.root / "config" / "governance" / "threat-detection-policy.json"
    if not policy_path.exists():
        return {"enabled": False}
    with policy_path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def _contains_pattern(text: str, patterns: list[str]) -> bool:
    """Check if text contains any of the patterns (case-insensitive)."""
    lowered = text.lower()
    return any(pattern.lower() in lowered for pattern in patterns)


def _detect_vulnerability_category(
    text: str,
    patterns: dict[str, list[str]]
) -> list[str]:
    """Detect vulnerability categories based on pattern matching."""
    detected = []
    for category, category_patterns in patterns.items():
        if _contains_pattern(text, category_patterns):
            detected.append(category)
    return detected


def infer_threat_detection(
    config: LoadedConfig,
    proposal: dict[str, Any],
    *,
    policy: dict[str, Any] | None = None,
) -> ThreatDetectionReport | None:
    """Analyze proposal for zero-day threats and vulnerabilities.

    Returns None if threat detection is disabled or not configured.
    """
    threat_policy = policy or load_threat_detection_policy(config)

    if not threat_policy.get("enabled", False):
        return None

    proposal_id = str(proposal.get("proposal_id", "unknown"))
    title = str(proposal.get("title", "Untitled proposal")).strip()
    summary = str(proposal.get("summary", "")).strip()
    combined_text = f"{title}\n{summary}\n" + "\n".join(
        str(item) for item in proposal.get("actions", [])
    )

    vulnerability_patterns = threat_policy.get("vulnerability_patterns", {})
    threat_weights = threat_policy.get("threat_weights", {})
    threat_thresholds = threat_policy.get("threat_thresholds", {})
    threat_classifications = threat_policy.get("threat_classifications", {})
    detection_rules = threat_policy.get("detection_rules", {})
    remediation_templates = threat_policy.get("remediation_templates", {})

    score = 0
    vulnerability_categories = []
    attack_vectors = []
    triggered_patterns = []
    security_signals = []
    rationale = []
    security_remediation = []

    # Detect vulnerability categories
    detected_categories = _detect_vulnerability_category(
        combined_text,
        vulnerability_patterns
    )
    vulnerability_categories.extend(detected_categories)

    # Apply detection rules
    for rule_name, rule_config in detection_rules.items():
        rule_patterns = rule_config.get("patterns", [])
        if not _contains_pattern(combined_text, rule_patterns):
            continue

        triggered_patterns.append(rule_name)
        threat_categories = rule_config.get("threat_categories", [])

        for threat_cat in threat_categories:
            if threat_cat not in attack_vectors:
                attack_vectors.append(threat_cat)
                weight_key = f"{threat_cat}_vector"
                score += int(threat_weights.get(weight_key, 10))
                rationale.append(
                    f"Detected {threat_cat} attack vector via {rule_name} pattern matching."
                )

        # Special escalation rules
        if rule_config.get("escalate_if_emergency") and proposal.get("is_emergency"):
            score += int(threat_weights.get("privilege_escalation", 15))
            security_signals.append("emergency_escalation_risk")
            rationale.append(
                "Emergency execution combined with validator/infrastructure changes increases privilege escalation risk."
            )

        if rule_config.get("escalate_if_large_amount"):
            treasury_amount = float(proposal.get("treasury_amount_usd", 0))
            if treasury_amount >= 100000:
                score += int(threat_weights.get("economic_exploit_pattern", 10))
                security_signals.append("large_treasury_exploit_risk")
                rationale.append(
                    f"Large treasury amount (${treasury_amount:,.0f}) increases economic exploit risk."
                )

        if rule_config.get("always_review"):
            security_signals.append(f"{rule_name}_requires_review")

        if rule_config.get("check_dependency_audit"):
            if not proposal.get("dependency_audit_completed"):
                score += int(threat_weights.get("unpatched_dependency", 15))
                security_signals.append("unaudited_dependency")
                rationale.append(
                    "External dependencies detected without completed security audit."
                )

    # Add vulnerability-specific scoring
    for vuln_category in detected_categories:
        if vuln_category == "smart_contract":
            score += int(threat_weights.get("smart_contract_vulnerability", 20))
            rationale.append("Detected smart contract vulnerability patterns.")
        elif vuln_category == "governance_attack":
            score += int(threat_weights.get("governance_attack_vector", 25))
            rationale.append("Detected governance attack patterns.")
        elif vuln_category == "data_integrity":
            score += int(threat_weights.get("data_integrity_risk", 18))
            rationale.append("Detected data integrity risk patterns.")

    # Determine threat level
    critical_threshold = int(threat_thresholds.get("critical_threat_score", 50))
    high_threshold = int(threat_thresholds.get("high_threat_score", 30))
    elevated_threshold = int(threat_thresholds.get("elevated_threat_score", 15))

    if score >= critical_threshold:
        threat_level = "critical"
    elif score >= high_threshold:
        threat_level = "high"
    elif score >= elevated_threshold:
        threat_level = "elevated"
    else:
        threat_level = "low"

    # Get classification requirements
    classification = threat_classifications.get(threat_level, {})
    requires_security_review = bool(classification.get("required_security_review", False))
    blocks_execution = bool(classification.get("block_execution", False))
    requires_escalation = bool(classification.get("escalation_required", False))

    # Build remediation from templates
    template_remediations = remediation_templates.get(threat_level, [])
    security_remediation.extend(template_remediations)

    # Add specific remediations based on detected threats
    if "governance_attack" in attack_vectors:
        security_remediation.append(
            "Implement multi-party approval with time delays to prevent governance attacks."
        )
    if "economic_exploit" in attack_vectors:
        security_remediation.append(
            "Conduct economic simulation and game theory analysis before execution."
        )
    if "unaudited_dependency" in security_signals:
        security_remediation.append(
            "Complete third-party security audit of all external dependencies."
        )

    # Deduplicate remediation
    unique_remediation = []
    seen = set()
    for item in security_remediation:
        if item not in seen:
            seen.add(item)
            unique_remediation.append(item)

    if score > 0:
        summary = (
            f"{threat_level} threat level with score {score}: "
            f"{len(vulnerability_categories)} vulnerability categories, "
            f"{len(attack_vectors)} attack vectors detected."
        )
    else:
        summary = "No significant security threats detected."

    return ThreatDetectionReport(
        proposal_id=proposal_id,
        title=title,
        threat_level=threat_level,
        threat_score=score,
        vulnerability_categories=vulnerability_categories,
        attack_vectors=attack_vectors,
        triggered_patterns=triggered_patterns,
        security_signals=security_signals,
        rationale=rationale,
        security_remediation=unique_remediation,
        requires_security_review=requires_security_review,
        blocks_execution=blocks_execution,
        requires_escalation=requires_escalation,
        summary=summary,
    )
