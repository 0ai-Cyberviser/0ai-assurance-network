"""Explainable governance inference for the 0AI Assurance Network."""

from __future__ import annotations

import hashlib
import hmac
import json
from dataclasses import dataclass, replace
from datetime import datetime, timedelta, timezone
from pathlib import Path
from typing import Any

from .config import LoadedConfig


@dataclass(frozen=True)
class GovernanceInferenceReport:
    engine: str
    proposal_id: str
    title: str
    proposal_class: str
    confidence: float
    risk_score: int
    required_houses: list[str]
    recommended_disposition: str
    triggered_signals: list[str]
    rationale: list[str]
    remediation: list[str]
    summary: str


@dataclass(frozen=True)
class GovernanceQueueEntry:
    path: str
    status: str
    priority: str
    owner: str
    proposal_kind: str
    submitted_at: str | None
    report: GovernanceInferenceReport
    drift: GovernanceDriftReport | None = None


@dataclass(frozen=True)
class GovernanceDriftReport:
    proposal_id: str
    title: str
    trend_cluster: str
    cluster_size: int
    cluster_outcomes: dict[str, int]
    stable_pattern: bool
    suppressed_signals: list[str]
    historical_matches: int
    baseline_risk_score: float | None
    drift_score: int
    drift_attention: str
    drift_signals: list[str]
    precedent_ids: list[str]
    rationale: list[str]
    remediation: list[str]
    summary: str


@dataclass(frozen=True)
class GovernancePortfolioTrend:
    trend_cluster: str
    proposal_kind: str
    owner: str
    active_proposals: int
    proposal_ids: list[str]
    historical_precedents: int
    historical_outcomes: dict[str, int]
    recent_window_days: int
    baseline_window_days: int
    recent_historical_precedents: int
    baseline_historical_precedents: int
    recent_active_proposals: int
    trend_velocity: str
    seasonal_kind_recent_total: int
    seasonal_kind_baseline_total: int
    seasonal_expected_recent: float
    seasonal_pressure: str
    highest_drift_attention: str
    highest_drift_score: int
    proposal_classes: list[str]
    systemic_signals: list[str]
    summary: str


@dataclass(frozen=True)
class GovernanceRemediationPlan:
    trend_cluster: str
    severity: str
    proposal_kind: str
    owner: str
    proposal_ids: list[str]
    triggers: list[str]
    release_readiness: str
    current_release_readiness: str
    owner_roles: list[str]
    checkpoint_status_counts: dict[str, int]
    replay_event_count: int
    invalid_event_count: int
    invalid_transition_count: int
    invalid_audit_count: int
    event_alerts: list[str]
    transition_alerts: list[str]
    audit_alerts: list[str]
    progress_summary: str
    immediate_actions: list[str]
    approval_guardrails: list[str]
    monitoring_actions: list[str]
    release_blockers: list[str]
    checkpoints: list[GovernanceRemediationCheckpoint]
    summary: str


@dataclass(frozen=True)
class GovernanceRemediationCheckpoint:
    checkpoint_id: str
    phase: str
    phase_order: int
    owner_role: str
    eligible_actors: list[GovernanceIdentityActorRef]
    title: str
    blocking: bool
    previous_status: str | None
    updated_at: str | None
    recorded_by: str | None
    actor_id: str | None
    assigned_actor: GovernanceIdentityActorRef | None
    status: str
    transition_valid: bool
    transition_note: str | None
    audit_valid: bool
    audit_note: str | None
    ready_to_start: bool
    depends_on: list[str]
    completion_criteria: str


@dataclass(frozen=True)
class GovernanceCheckpointReplay:
    source_kind: str
    replay_event_count: int
    invalid_event_count: int
    event_alerts: list[str]
    checkpoints: dict[str, dict[str, str | None]]


@dataclass(frozen=True)
class GovernanceIdentityActorRef:
    actor_id: str
    display_name: str
    actor_type: str
    organization_id: str | None
    organization_display_name: str | None
    role: str
    scopes: list[str]


def _load_json(path: Path) -> dict[str, Any]:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def load_inference_policy(config: LoadedConfig) -> dict[str, Any]:
    return _load_json(config.root / "config" / "governance" / "inference-policy.json")


def load_checkpoint_signer_policy(config: LoadedConfig) -> dict[str, Any]:
    return _load_json(config.root / "config" / "governance" / "checkpoint-signers.json")


def load_proposal(path: str | Path) -> dict[str, Any]:
    return _load_json(Path(path))


def load_registry(path: str | Path) -> dict[str, Any]:
    return _load_json(Path(path))


def load_history(path: str | Path) -> dict[str, Any]:
    return _load_json(Path(path))


def load_checkpoint_statuses(path: str | Path) -> dict[str, Any]:
    return _load_json(Path(path))


def _identity_actor_indexes(
    identity_bootstrap: dict[str, Any] | None,
) -> tuple[dict[str, dict[str, Any]], dict[tuple[str, str], list[str]]]:
    actors_by_id: dict[str, dict[str, Any]] = {}
    actor_role_scopes: dict[tuple[str, str], list[str]] = {}
    if not identity_bootstrap:
        return actors_by_id, actor_role_scopes
    for actor in identity_bootstrap.get("actors", []):
        actor_id = str(actor.get("actor_id", "")).strip()
        if actor_id:
            actors_by_id[actor_id] = dict(actor)
    for binding in identity_bootstrap.get("role_bindings", []):
        if str(binding.get("status", "")).strip() != "active":
            continue
        actor_id = str(binding.get("actor_id", "")).strip()
        role = str(binding.get("role", "")).strip()
        scope = str(binding.get("scope", "")).strip()
        if not actor_id or not role or not scope:
            continue
        actor_role_scopes.setdefault((actor_id, role), [])
        if scope not in actor_role_scopes[(actor_id, role)]:
            actor_role_scopes[(actor_id, role)].append(scope)
    for scopes in actor_role_scopes.values():
        scopes.sort()
    return actors_by_id, actor_role_scopes


def resolve_identity_actor_ref(
    identity_bootstrap: dict[str, Any] | None,
    actor_id: str | None,
    role: str | None,
) -> GovernanceIdentityActorRef | None:
    if not identity_bootstrap or not actor_id or not role:
        return None
    actors_by_id, actor_role_scopes = _identity_actor_indexes(identity_bootstrap)
    actor = actors_by_id.get(actor_id)
    if actor is None or str(actor.get("status", "")).strip() != "active":
        return None
    scopes = actor_role_scopes.get((actor_id, role), [])
    if not scopes:
        return None
    organization_id = str(actor.get("organization_id", "")).strip() or None
    organization_display_name = None
    if organization_id:
        organization = actors_by_id.get(organization_id)
        if organization is not None:
            organization_display_name = str(organization.get("display_name", "")).strip() or None
    return GovernanceIdentityActorRef(
        actor_id=actor_id,
        display_name=str(actor.get("display_name", "")).strip(),
        actor_type=str(actor.get("actor_type", "")).strip(),
        organization_id=organization_id,
        organization_display_name=organization_display_name,
        role=role,
        scopes=list(scopes),
    )


def resolve_role_actor_refs(
    identity_bootstrap: dict[str, Any] | None,
    role: str,
) -> list[GovernanceIdentityActorRef]:
    actors_by_id, actor_role_scopes = _identity_actor_indexes(identity_bootstrap)
    refs: list[GovernanceIdentityActorRef] = []
    for (actor_id, bound_role), scopes in sorted(actor_role_scopes.items()):
        if bound_role != role:
            continue
        actor = actors_by_id.get(actor_id)
        if actor is None or str(actor.get("status", "")).strip() != "active":
            continue
        organization_id = str(actor.get("organization_id", "")).strip() or None
        organization_display_name = None
        if organization_id:
            organization = actors_by_id.get(organization_id)
            if organization is not None:
                organization_display_name = str(organization.get("display_name", "")).strip() or None
        refs.append(
            GovernanceIdentityActorRef(
                actor_id=actor_id,
                display_name=str(actor.get("display_name", "")).strip(),
                actor_type=str(actor.get("actor_type", "")).strip(),
                organization_id=organization_id,
                organization_display_name=organization_display_name,
                role=role,
                scopes=list(scopes),
            )
        )
    return refs


def validate_identity_actor_role_assignment(
    *,
    identity_bootstrap: dict[str, Any] | None,
    actor_id: str | None,
    role: str | None,
    context: str,
) -> str | None:
    if not identity_bootstrap or not actor_id or not role:
        return None
    actors_by_id, actor_role_scopes = _identity_actor_indexes(identity_bootstrap)
    actor = actors_by_id.get(actor_id)
    if actor is None:
        return f"{context}: unknown actor_id {actor_id}."
    if str(actor.get("status", "")).strip() != "active":
        return f"{context}: actor {actor_id} is not active."
    if (actor_id, role) not in actor_role_scopes:
        return f"{context}: actor {actor_id} is not actively bound to role {role}."
    return None


def _contains_any(text: str, keywords: list[str]) -> bool:
    lowered = text.lower()
    return any(keyword.lower() in lowered for keyword in keywords)


def _class_rank(name: str) -> int:
    return {"standard": 0, "high_impact": 1, "safety_critical": 2}.get(name, 0)


def _attention_rank(name: str) -> int:
    return {"normal": 0, "review": 1, "escalate": 2}.get(name, 0)


def _severity_rank(name: str) -> int:
    return {"routine": 0, "elevated": 1, "critical": 2}.get(name, 0)


def _cluster_outcomes(entries: list[tuple[dict[str, Any], dict[str, Any], GovernanceInferenceReport]]) -> dict[str, int]:
    outcomes: dict[str, int] = {}
    for entry, _, _ in entries:
        outcome = str(entry.get("outcome", "unknown")).strip().lower() or "unknown"
        outcomes[outcome] = outcomes.get(outcome, 0) + 1
    return dict(sorted(outcomes.items()))


def _trend_cluster_key(proposal_kind: str, owner: str) -> str:
    return f"{proposal_kind}:{owner}"


def _checkpoint_slug(value: str) -> str:
    return "".join(ch.lower() if ch.isalnum() else "-" for ch in value).strip("-")


def replay_checkpoint_state(
    statuses: dict[str, Any] | None,
    *,
    signature_policy: dict[str, Any] | None = None,
    identity_bootstrap: dict[str, Any] | None = None,
) -> GovernanceCheckpointReplay:
    if not statuses:
        return GovernanceCheckpointReplay(
            source_kind="empty",
            replay_event_count=0,
            invalid_event_count=0,
            event_alerts=[],
            checkpoints={},
        )

    has_events_key = "events" in statuses
    raw_events = statuses.get("events")
    if has_events_key:
        if isinstance(raw_events, list):
            return _replay_checkpoint_events(
                raw_events,
                signature_policy=signature_policy,
                identity_bootstrap=identity_bootstrap,
            )
        return GovernanceCheckpointReplay(
            source_kind="event_log",
            replay_event_count=0,
            invalid_event_count=1,
            event_alerts=["Invalid event log schema: 'events' must be a list."],
            checkpoints={},
        )

    raw_checkpoints = statuses.get("checkpoints", statuses)
    normalized: dict[str, dict[str, str | None]] = {}
    allowed_statuses = {"pending", "in_progress", "completed"}
    if isinstance(raw_checkpoints, dict):
        for checkpoint_id, status in raw_checkpoints.items():
            status_name = str(status).strip().lower()
            normalized[str(checkpoint_id)] = {
                "previous_status": None,
                "updated_at": None,
                "recorded_by": None,
                "actor_id": None,
                "status": status_name if status_name in allowed_statuses else "pending",
            }
        return GovernanceCheckpointReplay(
            source_kind="snapshot",
            replay_event_count=0,
            invalid_event_count=0,
            event_alerts=[],
            checkpoints=normalized,
        )
    if isinstance(raw_checkpoints, list):
        for entry in raw_checkpoints:
            if not isinstance(entry, dict) or "checkpoint_id" not in entry:
                continue
            status = str(entry.get("status", "pending")).strip().lower()
            previous_status = entry.get("previous_status")
            previous_name = None
            if previous_status is not None:
                candidate = str(previous_status).strip().lower()
                previous_name = candidate if candidate in allowed_statuses else None
            normalized[str(entry["checkpoint_id"])] = {
                "previous_status": previous_name,
                "updated_at": (
                    str(entry.get("updated_at")).strip()
                    if entry.get("updated_at") is not None and str(entry.get("updated_at")).strip()
                    else None
                ),
                "recorded_by": (
                    str(entry.get("recorded_by")).strip()
                    if entry.get("recorded_by") is not None and str(entry.get("recorded_by")).strip()
                    else None
                ),
                "actor_id": (
                    str(entry.get("actor_id")).strip()
                    if entry.get("actor_id") is not None and str(entry.get("actor_id")).strip()
                    else None
                ),
                "status": status if status in allowed_statuses else "pending",
            }
    return GovernanceCheckpointReplay(
        source_kind="snapshot",
        replay_event_count=0,
        invalid_event_count=0,
        event_alerts=[],
        checkpoints=normalized,
    )


def _build_checkpoint_signature_message(
    *,
    checkpoint_id: str,
    previous_status: str,
    new_status: str,
    updated_at: str,
    recorded_by: str,
    actor_id: str,
    rationale: str,
    signature_meta: dict[str, str],
) -> str:
    payload = {
        "checkpoint_id": checkpoint_id,
        "previous_status": previous_status,
        "new_status": new_status,
        "updated_at": updated_at,
        "recorded_by": recorded_by,
        "actor_id": actor_id,
        "rationale": rationale,
        "signature": signature_meta,
    }
    return json.dumps(payload, sort_keys=True, separators=(",", ":"))


def _validate_signed_checkpoint_event(
    *,
    event_index: int,
    checkpoint_id: str,
    previous_status: str,
    new_status: str,
    updated_at: str,
    recorded_by: str,
    actor_id: str,
    rationale: str,
    raw_signature: Any,
    signature_policy: dict[str, Any] | None,
    identity_bootstrap: dict[str, Any] | None,
    seen_signature_ids: set[str],
) -> str | None:
    if not signature_policy or not signature_policy.get("require_signatures_for_event_logs", False):
        return None
    if not isinstance(raw_signature, dict):
        return f"Event {event_index} ({checkpoint_id}): missing signature."

    signature_format = str(raw_signature.get("format", "")).strip()
    signer_id = str(raw_signature.get("signer_id", "")).strip()
    key_id = str(raw_signature.get("key_id", "")).strip()
    signature_id = str(raw_signature.get("signature_id", "")).strip()
    signed_at = str(raw_signature.get("signed_at", "")).strip()
    expires_at = str(raw_signature.get("expires_at", "")).strip()
    signature_value = str(raw_signature.get("value", "")).strip()

    if not all([signature_format, signer_id, key_id, signature_id, signed_at, expires_at, signature_value]):
        return f"Event {event_index} ({checkpoint_id}): signature metadata is incomplete."
    if signature_format != str(signature_policy.get("signature_format", "")).strip():
        return (
            f"Event {event_index} ({checkpoint_id}): unsupported signature format "
            f"{signature_format}."
        )
    if signature_id in seen_signature_ids:
        return f"Event {event_index} ({checkpoint_id}): replay attempt detected for signature_id {signature_id}."

    signed_timestamp = _parse_timestamp(signed_at)
    expires_timestamp = _parse_timestamp(expires_at)
    updated_timestamp = _parse_timestamp(updated_at)
    if signed_timestamp is None or expires_timestamp is None or updated_timestamp is None:
        return f"Event {event_index} ({checkpoint_id}): signature timestamps are invalid."
    maximum_validity_seconds = int(signature_policy.get("maximum_signature_validity_seconds", 0))
    validity_window = int((expires_timestamp - signed_timestamp).total_seconds())
    if validity_window <= 0:
        return f"Event {event_index} ({checkpoint_id}): signature expires before it becomes valid."
    if maximum_validity_seconds > 0 and validity_window > maximum_validity_seconds:
        return (
            f"Event {event_index} ({checkpoint_id}): signature validity window exceeds policy limit."
        )
    if not (signed_timestamp <= updated_timestamp <= expires_timestamp):
        return f"Event {event_index} ({checkpoint_id}): signature is expired or not yet valid for updated_at."

    signer_entries = list(signature_policy.get("signers", []))
    signer_entry = next(
        (
            entry
            for entry in signer_entries
            if str(entry.get("signer_id", "")).strip() == signer_id
            and str(entry.get("key_id", "")).strip() == key_id
        ),
        None,
    )
    if signer_entry is None:
        return f"Event {event_index} ({checkpoint_id}): unknown signer_id/key_id pair."
    signer_actor_id = str(signer_entry.get("actor_id", "")).strip()
    if not signer_actor_id:
        return f"Event {event_index} ({checkpoint_id}): signer {signer_id} is missing actor_id."
    if signer_actor_id != actor_id:
        return (
            f"Event {event_index} ({checkpoint_id}): signer {signer_id} actor {signer_actor_id} "
            f"does not match event actor_id {actor_id}."
        )

    signer_roles = [str(role).strip() for role in signer_entry.get("roles", [])]
    if recorded_by not in signer_roles:
        return (
            f"Event {event_index} ({checkpoint_id}): signer {signer_id} is not authorized for role {recorded_by}."
        )
    identity_error = validate_identity_actor_role_assignment(
        identity_bootstrap=identity_bootstrap,
        actor_id=actor_id,
        role=recorded_by,
        context=f"Event {event_index} ({checkpoint_id})",
    )
    if identity_error:
        return identity_error

    signature_meta = {
        "format": signature_format,
        "signer_id": signer_id,
        "key_id": key_id,
        "signature_id": signature_id,
        "signed_at": signed_at,
        "expires_at": expires_at,
    }
    signing_message = _build_checkpoint_signature_message(
        checkpoint_id=checkpoint_id,
        previous_status=previous_status,
        new_status=new_status,
        updated_at=updated_at,
        recorded_by=recorded_by,
        actor_id=actor_id,
        rationale=rationale,
        signature_meta=signature_meta,
    )
    shared_secret = str(signer_entry.get("shared_secret", "")).encode("utf-8")
    expected_value = hmac.new(shared_secret, signing_message.encode("utf-8"), hashlib.sha256).hexdigest()
    if not hmac.compare_digest(expected_value, signature_value):
        return f"Event {event_index} ({checkpoint_id}): invalid signature."

    seen_signature_ids.add(signature_id)
    return None


def _replay_checkpoint_events(
    events: list[Any],
    *,
    signature_policy: dict[str, Any] | None = None,
    identity_bootstrap: dict[str, Any] | None = None,
) -> GovernanceCheckpointReplay:
    allowed_statuses = {"pending", "in_progress", "completed"}
    allowed_transitions = {
        ("pending", "in_progress"),
        ("in_progress", "completed"),
    }
    replayed: dict[str, dict[str, str | None]] = {}
    replayed_timestamps: dict[str, datetime] = {}
    event_alerts: list[str] = []
    replay_event_count = 0
    seen_signature_ids: set[str] = set()

    for index, raw_event in enumerate(events, start=1):
        replay_event_count += 1
        if not isinstance(raw_event, dict):
            event_alerts.append(f"Event {index}: event must be an object.")
            continue

        checkpoint_id = str(raw_event.get("checkpoint_id", "")).strip()
        if not checkpoint_id:
            event_alerts.append(f"Event {index}: missing checkpoint_id.")
            continue

        previous_status = raw_event.get("previous_status")
        previous_name = str(previous_status).strip().lower() if previous_status is not None else None
        new_status = raw_event.get("new_status", raw_event.get("status"))
        new_name = str(new_status).strip().lower() if new_status is not None else ""
        updated_at = str(raw_event.get("updated_at")).strip() if raw_event.get("updated_at") is not None else None
        recorded_by = str(raw_event.get("recorded_by")).strip() if raw_event.get("recorded_by") is not None else None
        actor_id = str(raw_event.get("actor_id")).strip() if raw_event.get("actor_id") is not None else None
        rationale = str(raw_event.get("rationale")).strip() if raw_event.get("rationale") is not None else None

        if previous_name not in allowed_statuses:
            event_alerts.append(
                f"Event {index} ({checkpoint_id}): previous_status must be one of pending, in_progress, completed."
            )
            continue
        if new_name not in allowed_statuses:
            event_alerts.append(
                f"Event {index} ({checkpoint_id}): new_status must be one of pending, in_progress, completed."
            )
            continue
        if previous_name == new_name:
            event_alerts.append(
                f"Event {index} ({checkpoint_id}): duplicate or no-op events are not allowed in the append-only log."
            )
            continue
        if (previous_name, new_name) not in allowed_transitions:
            event_alerts.append(
                f"Event {index} ({checkpoint_id}): illegal lifecycle transition {previous_name} -> {new_name}."
            )
            continue

        parsed_timestamp = _parse_timestamp(updated_at)
        if parsed_timestamp is None:
            event_alerts.append(f"Event {index} ({checkpoint_id}): invalid or missing updated_at timestamp.")
            continue
        if not recorded_by:
            event_alerts.append(f"Event {index} ({checkpoint_id}): missing recorded_by.")
            continue
        if not actor_id:
            event_alerts.append(f"Event {index} ({checkpoint_id}): missing actor_id.")
            continue
        if not rationale:
            event_alerts.append(f"Event {index} ({checkpoint_id}): missing rationale.")
            continue
        signature_error = _validate_signed_checkpoint_event(
            event_index=index,
            checkpoint_id=checkpoint_id,
            previous_status=previous_name,
            new_status=new_name,
            updated_at=updated_at,
            recorded_by=recorded_by,
            actor_id=actor_id,
            rationale=rationale,
            raw_signature=raw_event.get("signature"),
            signature_policy=signature_policy,
            identity_bootstrap=identity_bootstrap,
            seen_signature_ids=seen_signature_ids,
        )
        if signature_error:
            event_alerts.append(signature_error)
            continue

        current_state = replayed.get(checkpoint_id)
        expected_previous = str(current_state.get("status", "pending")) if current_state is not None else "pending"
        if previous_name != expected_previous:
            event_alerts.append(
                f"Event {index} ({checkpoint_id}): contradictory event history, expected previous_status "
                f"{expected_previous} but received {previous_name}."
            )
            continue

        prior_timestamp = replayed_timestamps.get(checkpoint_id)
        if prior_timestamp is not None and parsed_timestamp <= prior_timestamp:
            event_alerts.append(
                f"Event {index} ({checkpoint_id}): out-of-order updated_at {updated_at}."
            )
            continue

        replayed[checkpoint_id] = {
            "previous_status": previous_name,
            "updated_at": updated_at,
            "recorded_by": recorded_by,
            "actor_id": actor_id,
            "status": new_name,
        }
        replayed_timestamps[checkpoint_id] = parsed_timestamp

    return GovernanceCheckpointReplay(
        source_kind="event_log",
        replay_event_count=replay_event_count,
        invalid_event_count=len(event_alerts),
        event_alerts=event_alerts,
        checkpoints=replayed,
    )


def _validate_checkpoint_transition(
    previous_status: str | None,
    status: str,
    allowed_transitions: dict[str, list[str]],
) -> tuple[bool, str | None]:
    if previous_status is None:
        if status == "pending":
            return True, "Checkpoint is still pending with no prior transition recorded."
        return False, "Non-pending checkpoint updates must include previous_status for transition validation."

    allowed = allowed_transitions.get(previous_status, [])
    if status in allowed:
        return True, None
    return False, f"Illegal checkpoint transition: {previous_status} -> {status}"


def _validate_checkpoint_audit(
    *,
    status: str,
    updated_at: str | None,
    recorded_by: str | None,
    actor_id: str | None,
    expected_owner_role: str,
    require_actor_for_non_pending: bool,
    require_timestamp_for_non_pending: bool,
    identity_bootstrap: dict[str, Any] | None,
) -> tuple[bool, str | None, datetime | None]:
    parsed_timestamp = _parse_timestamp(updated_at)
    if updated_at is not None and parsed_timestamp is None:
        return False, f"Invalid checkpoint updated_at timestamp: {updated_at}", None

    if status != "pending":
        if require_actor_for_non_pending and (not recorded_by or not actor_id):
            return (
                False,
                "Non-pending checkpoint updates must include recorded_by and actor_id for audit attribution.",
                parsed_timestamp,
            )
        if require_timestamp_for_non_pending and not updated_at:
            return False, "Non-pending checkpoint updates must include updated_at for audit ordering.", parsed_timestamp
        if recorded_by and recorded_by != expected_owner_role:
            return (
                False,
                f"Checkpoint actor role mismatch: expected {expected_owner_role}, received {recorded_by}.",
                parsed_timestamp,
            )
        identity_error = validate_identity_actor_role_assignment(
            identity_bootstrap=identity_bootstrap,
            actor_id=actor_id,
            role=recorded_by,
            context="Checkpoint audit",
        )
        if identity_error:
            return False, identity_error, parsed_timestamp

    return True, None, parsed_timestamp


def _parse_timestamp(value: str | None) -> datetime | None:
    if not value:
        return None
    normalized = value.replace("Z", "+00:00")
    try:
        return datetime.fromisoformat(normalized).astimezone(timezone.utc)
    except ValueError:
        return None


def infer_governance_report(
    config: LoadedConfig,
    proposal: dict[str, Any],
    policy: dict[str, Any] | None = None,
) -> GovernanceInferenceReport:
    inference_policy = policy or load_inference_policy(config)
    weights = inference_policy["weights"]
    thresholds = inference_policy["thresholds"]
    keywords = inference_policy["keywords"]

    title = str(proposal.get("title", "")).strip()
    summary = str(proposal.get("summary", "")).strip()
    combined_text = f"{title}\n{summary}\n" + "\n".join(str(item) for item in proposal.get("actions", []))

    score = 0
    signals: list[str] = []
    rationale: list[str] = []
    remediation: list[str] = []

    if _contains_any(combined_text, keywords["safety"]):
        score += int(weights["safety_keywords"])
        signals.append("safety_keywords")
        rationale.append("Proposal language touches incident, safety, slashing, or emergency controls.")

    if proposal.get("is_emergency"):
        score += int(weights["emergency_flag"])
        signals.append("emergency_flag")
        rationale.append("Proposal explicitly declares emergency execution pressure.")
        remediation.append("Require a bounded expiry and a public postmortem deadline.")

    if float(proposal.get("treasury_amount_usd", 0)) >= float(thresholds["large_treasury_amount_usd"]):
        score += int(weights["treasury_large"])
        signals.append("treasury_large")
        rationale.append("Treasury amount crosses the large-disbursement threshold.")
        remediation.append("Break funding into staged releases with milestone checkpoints.")

    if proposal.get("affects_validators"):
        score += int(weights["validator_change"])
        signals.append("validator_change")
        rationale.append("Proposal affects validator, staking, or consensus assumptions.")
        remediation.append("Require validator impact simulation before final approval.")

    if proposal.get("changes_governance_rules") or _contains_any(combined_text, keywords["governance"]):
        score += int(weights["governance_change"])
        signals.append("governance_change")
        rationale.append("Proposal changes governance semantics or review controls.")
        remediation.append("Publish a governance diff and quorum analysis.")

    if proposal.get("touches_ai_systems") or _contains_any(combined_text, keywords["ai"]):
        score += int(weights["ai_system_impact"])
        signals.append("ai_system_impact")
        rationale.append("Proposal affects models, agents, inference, or simulation layers.")
        remediation.append("Run AI safety and operational rollback review before execution.")

    if proposal.get("external_dependencies"):
        score += int(weights["external_dependency"])
        signals.append("external_dependency")
        rationale.append("Proposal depends on external vendors or off-chain actors.")
        remediation.append("Document external dependency failure modes and accountability.")

    if score >= int(thresholds["safety_critical_score"]):
        proposal_class = "safety_critical"
    elif score >= int(thresholds["high_impact_score"]):
        proposal_class = "high_impact"
    else:
        proposal_class = "standard"

    minimum_classes = inference_policy.get("minimum_classes", {})
    enforced_class = proposal_class
    if proposal.get("is_emergency"):
        enforced_class = max(
            (enforced_class, str(minimum_classes.get("emergency", enforced_class))),
            key=_class_rank,
        )
    if proposal.get("affects_validators"):
        enforced_class = max(
            (enforced_class, str(minimum_classes.get("validator_change", enforced_class))),
            key=_class_rank,
        )
    if proposal.get("changes_governance_rules"):
        enforced_class = max(
            (enforced_class, str(minimum_classes.get("governance_rule_change", enforced_class))),
            key=_class_rank,
        )
    proposal_class = enforced_class

    required_houses = list(inference_policy["proposal_classes"][proposal_class]["required_houses"])
    minimum_confidence = float(inference_policy["proposal_classes"][proposal_class]["minimum_confidence"])

    confidence = min(0.97, minimum_confidence + (min(score, 80) / 200.0))

    if proposal_class == "safety_critical":
        disposition = "review"
    elif score >= int(thresholds["high_impact_score"]) and proposal.get("treasury_amount_usd", 0):
        disposition = "review"
    else:
        disposition = "advance"

    if proposal.get("is_emergency") and not proposal.get("actions"):
        disposition = "hold"
        remediation.append("Emergency proposals must include explicit actions and stop conditions.")

    unique_remediation = []
    seen = set()
    for item in remediation:
        if item not in seen:
            seen.add(item)
            unique_remediation.append(item)

    summary_line = (
        f"{proposal_class.replace('_', ' ')} proposal with risk score {score} "
        f"and required houses {', '.join(required_houses)}."
    )

    return GovernanceInferenceReport(
        engine=str(inference_policy["engine"]),
        proposal_id=str(proposal.get("proposal_id", "unknown")),
        title=title or "Untitled proposal",
        proposal_class=proposal_class,
        confidence=round(confidence, 2),
        risk_score=score,
        required_houses=required_houses,
        recommended_disposition=disposition,
        triggered_signals=signals,
        rationale=rationale,
        remediation=unique_remediation,
        summary=summary_line,
    )


def infer_governance_drift(
    config: LoadedConfig,
    proposal: dict[str, Any],
    history: dict[str, Any],
    *,
    policy: dict[str, Any] | None = None,
) -> GovernanceDriftReport:
    inference_policy = policy or load_inference_policy(config)
    history_weights = inference_policy.get("history_weights", {})
    history_thresholds = inference_policy.get("history_thresholds", {})

    current_report = infer_governance_report(config, proposal, policy=inference_policy)
    current_kind = str(proposal.get("kind", "unknown"))
    current_requester = str(proposal.get("requested_by", "unknown"))

    history_entries = list(history.get("entries", []))
    historical_reports: list[tuple[dict[str, Any], dict[str, Any], GovernanceInferenceReport]] = []
    for entry in history_entries:
        historical_proposal = dict(entry.get("proposal", {}))
        if not historical_proposal:
            continue
        historical_report = infer_governance_report(config, historical_proposal, policy=inference_policy)
        historical_reports.append((entry, historical_proposal, historical_report))

    same_kind = [
        (entry, historical_proposal, historical_report)
        for entry, historical_proposal, historical_report in historical_reports
        if str(historical_proposal.get("kind", "unknown")) == current_kind
    ]
    same_cluster = [
        (entry, historical_proposal, historical_report)
        for entry, historical_proposal, historical_report in same_kind
        if str(historical_proposal.get("requested_by", "unknown")) == current_requester
    ]
    cluster_entries = same_cluster or same_kind
    trend_cluster = _trend_cluster_key(current_kind, current_requester)
    precedent_ids = [
        str(historical_proposal.get("proposal_id", "unknown"))
        for _, historical_proposal, _ in cluster_entries
    ]
    cluster_outcomes = _cluster_outcomes(cluster_entries)

    score = 0
    signals: list[str] = []
    rationale: list[str] = []
    remediation: list[str] = []
    suppressed_signals: list[str] = []

    baseline_risk_score = None
    if same_kind:
        baseline_risk_score = round(
            sum(historical_report.risk_score for _, _, historical_report in same_kind) / len(same_kind),
            2,
        )

    risk_score_drift = int(history_thresholds.get("risk_score_drift", 0))
    if (
        risk_score_drift > 0
        and baseline_risk_score is not None
        and current_report.risk_score >= (baseline_risk_score + risk_score_drift)
    ):
        score += int(history_weights.get("risk_score_drift", 0))
        signals.append("risk_score_drift")
        rationale.append(
            "Current proposal risk materially exceeds the historical baseline for the same proposal kind."
        )
        remediation.append("Require a comparative risk review against prior proposals of the same kind.")

    current_treasury_amount = float(proposal.get("treasury_amount_usd", 0))
    same_kind_treasury = [
        float(historical_proposal.get("treasury_amount_usd", 0))
        for _, historical_proposal, _ in same_kind
        if float(historical_proposal.get("treasury_amount_usd", 0)) > 0
    ]
    treasury_growth_ratio = float(history_thresholds.get("treasury_growth_ratio", 0))
    if same_kind_treasury:
        historical_max_treasury = max(same_kind_treasury)
        if (
            current_treasury_amount > 0
            and treasury_growth_ratio > 0
            and current_treasury_amount >= (historical_max_treasury * treasury_growth_ratio)
        ):
            score += int(history_weights.get("treasury_growth", 0))
            signals.append("treasury_growth")
            rationale.append("Requested treasury spend is materially larger than the prior observed maximum.")
            remediation.append("Add staged treasury controls and a tighter release schedule.")

    repeated_emergency_threshold = int(history_thresholds.get("repeated_emergency", 0))
    repeated_emergency_count = sum(
        1 for _, historical_proposal, _ in historical_reports if historical_proposal.get("is_emergency")
    )
    if (
        repeated_emergency_threshold > 0
        and proposal.get("is_emergency")
        and repeated_emergency_count >= repeated_emergency_threshold
    ):
        score += int(history_weights.get("repeated_emergency", 0))
        signals.append("repeated_emergency_pattern")
        rationale.append("History shows repeated emergency governance activity, which may indicate control instability.")
        remediation.append("Require an explicit emergency root-cause review before approving another emergency action.")

    requester_high_impact_threshold = int(history_thresholds.get("requester_high_impact", 0))
    requester_high_impact_count = sum(
        1
        for _, historical_proposal, historical_report in historical_reports
        if str(historical_proposal.get("requested_by", "unknown")) == current_requester
        and historical_report.proposal_class in {"high_impact", "safety_critical"}
    )
    if (
        requester_high_impact_threshold > 0
        and requester_high_impact_count >= requester_high_impact_threshold
    ):
        score += int(history_weights.get("requester_concentration", 0))
        signals.append("requester_concentration")
        rationale.append("One requester has accumulated multiple high-impact governance actions in history.")
        remediation.append("Require independent sponsorship or counter-signoff before approval.")

    repeated_validator_change_threshold = int(history_thresholds.get("repeated_validator_change", 0))
    validator_change_count = sum(
        1 for _, historical_proposal, _ in historical_reports if historical_proposal.get("affects_validators")
    )
    if (
        repeated_validator_change_threshold > 0
        and proposal.get("affects_validators")
        and validator_change_count >= repeated_validator_change_threshold
    ):
        score += int(history_weights.get("repeated_validator_change", 0))
        signals.append("repeated_validator_change")
        rationale.append("Validator-impacting proposals are recurring often enough to justify structural review.")
        remediation.append("Run validator stability simulation and publish liveness assumptions before approval.")

    adverse_outcomes = {
        "failed",
        "rolled_back",
        "incident",
        "rejected",
    }
    adverse_precedent_count = sum(
        1
        for entry, _, _ in same_kind
        if str(entry.get("outcome", "")).strip().lower() in adverse_outcomes
    )
    if adverse_precedent_count > 0:
        score += int(history_weights.get("adverse_precedent", 0))
        signals.append("adverse_precedent")
        rationale.append("There is adverse execution history for this proposal kind that should not be ignored.")
        remediation.append("Attach the prior failure analysis and explain how this proposal avoids the same outcome.")

    stable_pattern_window = int(history_thresholds.get("stable_pattern_window", 0))
    stable_pattern_completed_ratio = float(history_thresholds.get("stable_pattern_completed_ratio", 0))
    suppression_policy = inference_policy.get("history_suppression", {})
    completed_count = cluster_outcomes.get("completed", 0)
    cluster_size = len(cluster_entries)
    completed_ratio = (completed_count / cluster_size) if cluster_size else 0.0
    stable_pattern_requires_non_emergency = bool(
        suppression_policy.get("stable_pattern_requires_non_emergency", True)
    )
    stable_pattern = (
        cluster_size >= stable_pattern_window > 0
        and adverse_precedent_count == 0
        and completed_ratio >= stable_pattern_completed_ratio
        and (not stable_pattern_requires_non_emergency or not proposal.get("is_emergency"))
    )
    if stable_pattern:
        for signal_name in suppression_policy.get("suppressible_signals", []):
            if signal_name in signals:
                signals.remove(signal_name)
                suppressed_signals.append(signal_name)
                score = max(0, score - int(history_weights.get(signal_name, 0)))
        if suppressed_signals:
            rationale.append("Stable historical precedent cluster justified suppressing recurring-pattern drift signals.")
            remediation.append("Publish a short precedent diff to confirm this proposal still fits the stable pattern.")

    review_drift_score = int(history_thresholds.get("review_drift_score", 0))
    escalate_drift_score = int(history_thresholds.get("escalate_drift_score", 0))
    if escalate_drift_score > 0 and score >= escalate_drift_score:
        drift_attention = "escalate"
    elif review_drift_score > 0 and score >= review_drift_score:
        drift_attention = "review"
    else:
        drift_attention = "normal"

    unique_remediation = []
    seen = set()
    for item in remediation:
        if item not in seen:
            seen.add(item)
            unique_remediation.append(item)

    if signals:
        summary = (
            f"{drift_attention} drift attention: {len(signals)} historical governance signals triggered "
            f"with drift score {score} for cluster {trend_cluster}."
        )
    else:
        summary = f"No unsuppressed historical governance drift signals triggered for cluster {trend_cluster}."

    return GovernanceDriftReport(
        proposal_id=str(proposal.get("proposal_id", "unknown")),
        title=str(proposal.get("title", "Untitled proposal")),
        trend_cluster=trend_cluster,
        cluster_size=cluster_size,
        cluster_outcomes=cluster_outcomes,
        stable_pattern=stable_pattern,
        suppressed_signals=suppressed_signals,
        historical_matches=len(same_kind),
        baseline_risk_score=baseline_risk_score,
        drift_score=score,
        drift_attention=drift_attention,
        drift_signals=signals,
        precedent_ids=precedent_ids,
        rationale=rationale,
        remediation=unique_remediation,
        summary=summary,
    )


def infer_governance_queue(
    config: LoadedConfig,
    registry: dict[str, Any],
    *,
    registry_path: str | Path,
    history: dict[str, Any] | None = None,
    policy: dict[str, Any] | None = None,
) -> list[GovernanceQueueEntry]:
    registry_root = Path(registry_path).resolve().parent
    entries: list[GovernanceQueueEntry] = []

    for item in registry.get("proposals", []):
        raw_path = Path(str(item["path"]))
        if raw_path.is_absolute():
            proposal_path = raw_path
        else:
            project_relative = Path(config.root) / raw_path
            registry_relative = registry_root / raw_path
            proposal_path = project_relative if project_relative.exists() else registry_relative
        proposal = load_proposal(proposal_path)
        report = infer_governance_report(config, proposal, policy=policy)
        drift = infer_governance_drift(config, proposal, history, policy=policy) if history else None
        entries.append(
            GovernanceQueueEntry(
                path=str(proposal_path),
                status=str(item.get("status", "pending")),
                priority=str(item.get("priority", "normal")),
                owner=str(item.get("owner", "unknown")),
                proposal_kind=str(proposal.get("kind", "unknown")),
                submitted_at=str(item.get("submitted_at")) if item.get("submitted_at") else None,
                report=report,
                drift=drift,
            )
        )

    priority_rank = {"urgent": 0, "high": 1, "normal": 2, "low": 3}
    class_rank = {"safety_critical": 0, "high_impact": 1, "standard": 2}

    return sorted(
        entries,
        key=lambda entry: (
            priority_rank.get(entry.priority, 9),
            class_rank.get(entry.report.proposal_class, 9),
            -_attention_rank(entry.drift.drift_attention) if entry.drift else 0,
            -(entry.drift.drift_score if entry.drift else 0),
            -entry.report.risk_score,
            entry.report.proposal_id,
        ),
    )


def infer_governance_portfolio_trends(
    entries: list[GovernanceQueueEntry],
    *,
    history: dict[str, Any] | None = None,
    policy: dict[str, Any] | None = None,
) -> list[GovernancePortfolioTrend]:
    grouped: dict[str, list[GovernanceQueueEntry]] = {}
    for entry in entries:
        cluster = (
            entry.drift.trend_cluster
            if entry.drift is not None
            else _trend_cluster_key(entry.proposal_kind, entry.owner)
        )
        grouped.setdefault(cluster, []).append(entry)

    inference_policy = policy or {}
    history_windows = inference_policy.get("history_windows", {})
    recent_window_days = int(history_windows.get("recent_days", 45))
    baseline_window_days = int(history_windows.get("baseline_days", 180))
    acceleration_ratio = float(history_windows.get("acceleration_ratio", 1.75))
    minimum_recent_events = int(history_windows.get("minimum_recent_events", 2))
    history_entries = list((history or {}).get("entries", []))

    trends: list[GovernancePortfolioTrend] = []
    for cluster, cluster_entries in grouped.items():
        first = cluster_entries[0]
        drifts = [entry.drift for entry in cluster_entries if entry.drift is not None]
        highest_drift_score = max((drift.drift_score for drift in drifts), default=0)
        highest_drift_attention = "normal"
        for drift in drifts:
            if _attention_rank(drift.drift_attention) > _attention_rank(highest_drift_attention):
                highest_drift_attention = drift.drift_attention

        proposal_classes = sorted(
            {entry.report.proposal_class for entry in cluster_entries},
            key=_class_rank,
            reverse=True,
        )
        systemic_signals = sorted({signal for drift in drifts for signal in drift.drift_signals})
        if len(cluster_entries) > 1:
            systemic_signals = ["queue_cluster_repetition", *systemic_signals]

        historical_precedents = max((drift.historical_matches for drift in drifts), default=0)
        historical_outcomes = drifts[0].cluster_outcomes if drifts else {}
        proposal_ids = [entry.report.proposal_id for entry in cluster_entries]

        cluster_history = []
        for history_entry in history_entries:
            proposal = dict(history_entry.get("proposal", {}))
            if not proposal:
                continue
            if _trend_cluster_key(
                str(proposal.get("kind", "unknown")),
                str(proposal.get("requested_by", "unknown")),
            ) == cluster:
                cluster_history.append(history_entry)

        history_timestamps = [
            _parse_timestamp(str(history_entry.get("recorded_at", "")))
            for history_entry in cluster_history
        ]
        active_timestamps = [_parse_timestamp(entry.submitted_at) for entry in cluster_entries]
        anchor_candidates = [timestamp for timestamp in [*history_timestamps, *active_timestamps] if timestamp is not None]
        anchor_time = max(anchor_candidates) if anchor_candidates else datetime.now(timezone.utc)
        recent_start = anchor_time - timedelta(days=recent_window_days)
        baseline_start = anchor_time - timedelta(days=baseline_window_days)

        recent_historical_precedents = sum(
            1
            for timestamp in history_timestamps
            if timestamp is not None and timestamp >= recent_start
        )
        baseline_historical_precedents = sum(
            1
            for timestamp in history_timestamps
            if timestamp is not None and baseline_start <= timestamp < recent_start
        )
        recent_active_proposals = sum(
            1
            for timestamp in active_timestamps
            if timestamp is not None and timestamp >= recent_start
        )

        recent_total = recent_historical_precedents + recent_active_proposals
        recent_rate = recent_total / max(recent_window_days, 1)
        baseline_span_days = max(baseline_window_days - recent_window_days, 1)
        baseline_rate = baseline_historical_precedents / baseline_span_days

        if (
            recent_total >= minimum_recent_events
            and (baseline_rate == 0 or recent_rate >= (baseline_rate * acceleration_ratio))
        ):
            trend_velocity = "accelerating"
            systemic_signals = ["trend_acceleration", *systemic_signals]
        elif recent_total > 0:
            trend_velocity = "elevated"
            systemic_signals = ["recent_activity", *systemic_signals]
        else:
            trend_velocity = "stable"
        systemic_signals = list(dict.fromkeys(systemic_signals))

        kind_history_timestamps = [
            _parse_timestamp(str(history_entry.get("recorded_at", "")))
            for history_entry in history_entries
            if str(dict(history_entry.get("proposal", {})).get("kind", "unknown")) == first.proposal_kind
        ]
        kind_active_timestamps = [
            _parse_timestamp(entry.submitted_at)
            for entry in entries
            if entry.proposal_kind == first.proposal_kind
        ]
        seasonal_kind_recent_total = sum(
            1 for timestamp in kind_history_timestamps if timestamp is not None and timestamp >= recent_start
        ) + sum(
            1 for timestamp in kind_active_timestamps if timestamp is not None and timestamp >= recent_start
        )
        seasonal_kind_baseline_total = sum(
            1
            for timestamp in kind_history_timestamps
            if timestamp is not None and baseline_start <= timestamp < recent_start
        )
        seasonal_expected_recent = round(
            (seasonal_kind_baseline_total / baseline_span_days) * recent_window_days,
            2,
        )
        if seasonal_kind_recent_total == 0:
            seasonal_pressure = "quiet"
        elif seasonal_kind_recent_total < minimum_recent_events:
            seasonal_pressure = "watch"
        elif seasonal_expected_recent == 0 or seasonal_kind_recent_total >= (
            seasonal_expected_recent * acceleration_ratio
        ):
            seasonal_pressure = "above_norm"
            systemic_signals = ["seasonal_pressure", *systemic_signals]
        else:
            seasonal_pressure = "in_band"
        systemic_signals = list(dict.fromkeys(systemic_signals))

        summary = (
            f"cluster {cluster} has {len(cluster_entries)} active proposal(s), "
            f"{historical_precedents} historical precedent(s), highest drift "
            f"{highest_drift_attention} ({highest_drift_score}), velocity {trend_velocity}, "
            f"and seasonal pressure {seasonal_pressure}."
        )

        trends.append(
            GovernancePortfolioTrend(
                trend_cluster=cluster,
                proposal_kind=first.proposal_kind,
                owner=first.owner,
                active_proposals=len(cluster_entries),
                proposal_ids=proposal_ids,
                historical_precedents=historical_precedents,
                historical_outcomes=historical_outcomes,
                recent_window_days=recent_window_days,
                baseline_window_days=baseline_window_days,
                recent_historical_precedents=recent_historical_precedents,
                baseline_historical_precedents=baseline_historical_precedents,
                recent_active_proposals=recent_active_proposals,
                trend_velocity=trend_velocity,
                seasonal_kind_recent_total=seasonal_kind_recent_total,
                seasonal_kind_baseline_total=seasonal_kind_baseline_total,
                seasonal_expected_recent=seasonal_expected_recent,
                seasonal_pressure=seasonal_pressure,
                highest_drift_attention=highest_drift_attention,
                highest_drift_score=highest_drift_score,
                proposal_classes=proposal_classes,
                systemic_signals=systemic_signals,
                summary=summary,
            )
        )

    return sorted(
        trends,
        key=lambda trend: (
            -_attention_rank(trend.highest_drift_attention),
            -trend.highest_drift_score,
            trend.trend_velocity != "accelerating",
            -trend.active_proposals,
            trend.trend_cluster,
        ),
    )


def infer_governance_remediation_plans(
    entries: list[GovernanceQueueEntry],
    trends: list[GovernancePortfolioTrend],
    *,
    policy: dict[str, Any] | None = None,
    checkpoint_statuses: dict[str, Any] | None = None,
    signature_policy: dict[str, Any] | None = None,
    identity_bootstrap: dict[str, Any] | None = None,
) -> list[GovernanceRemediationPlan]:
    remediation_policy = (policy or {}).get("remediation", {})
    severity_actions = remediation_policy.get("severity_actions", {})
    class_actions = remediation_policy.get("class_actions", {})
    signal_actions = remediation_policy.get("signal_actions", {})
    kind_actions = remediation_policy.get("kind_actions", {})
    execution_defaults = remediation_policy.get("execution_defaults", {})
    phase_owners = execution_defaults.get("phase_owners", {})
    phase_completion = execution_defaults.get("phase_completion", {})
    owner_overrides = execution_defaults.get("owner_overrides", {})
    phase_order = execution_defaults.get("phase_order", {})
    phase_dependencies = execution_defaults.get("phase_dependencies", {})
    allowed_transitions = execution_defaults.get("allowed_transitions", {})
    audit_requirements = execution_defaults.get("audit_requirements", {})
    require_actor_for_non_pending = bool(audit_requirements.get("require_actor_for_non_pending", True))
    require_timestamp_for_non_pending = bool(audit_requirements.get("require_timestamp_for_non_pending", True))
    enforce_dependency_timestamp_order = bool(audit_requirements.get("enforce_dependency_timestamp_order", True))
    checkpoint_replay = replay_checkpoint_state(
        checkpoint_statuses,
        signature_policy=signature_policy,
        identity_bootstrap=identity_bootstrap,
    )
    normalized_statuses = checkpoint_replay.checkpoints

    cluster_entries: dict[str, list[GovernanceQueueEntry]] = {}
    for entry in entries:
        cluster = (
            entry.drift.trend_cluster
            if entry.drift is not None
            else _trend_cluster_key(entry.proposal_kind, entry.owner)
        )
        cluster_entries.setdefault(cluster, []).append(entry)

    plans: list[GovernanceRemediationPlan] = []
    for trend in trends:
        if (
            trend.highest_drift_attention == "escalate"
            or trend.trend_velocity == "accelerating"
            or trend.seasonal_pressure == "above_norm"
        ):
            severity = "critical"
        elif (
            trend.highest_drift_attention == "review"
            or trend.trend_velocity == "elevated"
            or trend.seasonal_pressure == "watch"
        ):
            severity = "elevated"
        else:
            severity = "routine"

        plan_entries = cluster_entries.get(trend.trend_cluster, [])
        immediate_actions: list[str] = []
        approval_guardrails: list[str] = []
        monitoring_actions: list[str] = []
        release_blockers: list[str] = []
        phase_checkpoint_ids: dict[str, list[str]] = {}
        raw_checkpoints: list[dict[str, Any]] = []

        def extend_unique(target: list[str], values: list[str] | None) -> None:
            if not values:
                return
            for item in values:
                if item not in target:
                    target.append(item)

        def owner_for_phase(phase: str) -> str:
            return str(
                owner_overrides.get(trend.proposal_kind, {}).get(phase)
                or phase_owners.get(phase, "governance-ops")
            )

        def build_checkpoints(phase: str, actions: list[str], *, blocking: bool) -> None:
            owner_role = owner_for_phase(phase)
            completion_criteria = str(
                phase_completion.get(
                    phase,
                    "Record completion in the governance execution log before advancing.",
                )
            )
            order = int(phase_order.get(phase, 99))
            checkpoint_ids: list[str] = []
            for index, action in enumerate(actions, start=1):
                checkpoint_id = f"{_checkpoint_slug(trend.trend_cluster)}-{phase}-{index}"
                checkpoint_ids.append(checkpoint_id)
                raw_checkpoints.append(
                    {
                        "checkpoint_id": checkpoint_id,
                        "phase": phase,
                        "phase_order": order,
                        "owner_role": owner_role,
                        "title": action,
                        "blocking": blocking,
                        "completion_criteria": completion_criteria,
                    }
                )
            phase_checkpoint_ids[phase] = checkpoint_ids

        def dependency_completed(checkpoint_id: str) -> bool:
            checkpoint_state = normalized_statuses.get(checkpoint_id)
            if isinstance(checkpoint_state, dict):
                return str(checkpoint_state.get("status", "pending")) == "completed"
            return False

        severity_bundle = severity_actions.get(severity, {})
        extend_unique(immediate_actions, severity_bundle.get("immediate_actions"))
        extend_unique(approval_guardrails, severity_bundle.get("approval_guardrails"))
        extend_unique(monitoring_actions, severity_bundle.get("monitoring_actions"))
        extend_unique(release_blockers, severity_bundle.get("release_blockers"))

        kind_bundle = kind_actions.get(trend.proposal_kind, {})
        extend_unique(immediate_actions, kind_bundle.get("immediate_actions"))
        extend_unique(approval_guardrails, kind_bundle.get("approval_guardrails"))
        extend_unique(monitoring_actions, kind_bundle.get("monitoring_actions"))
        extend_unique(release_blockers, kind_bundle.get("release_blockers"))

        for proposal_class in trend.proposal_classes:
            class_bundle = class_actions.get(proposal_class, {})
            extend_unique(immediate_actions, class_bundle.get("immediate_actions"))
            extend_unique(approval_guardrails, class_bundle.get("approval_guardrails"))
            extend_unique(monitoring_actions, class_bundle.get("monitoring_actions"))
            extend_unique(release_blockers, class_bundle.get("release_blockers"))

        for signal in trend.systemic_signals:
            signal_bundle = signal_actions.get(signal, {})
            extend_unique(immediate_actions, signal_bundle.get("immediate_actions"))
            extend_unique(approval_guardrails, signal_bundle.get("approval_guardrails"))
            extend_unique(monitoring_actions, signal_bundle.get("monitoring_actions"))
            extend_unique(release_blockers, signal_bundle.get("release_blockers"))

        for entry in plan_entries:
            extend_unique(immediate_actions, entry.report.remediation)
            if entry.drift is not None:
                extend_unique(immediate_actions, entry.drift.remediation)

        build_checkpoints("release_blocker", release_blockers, blocking=True)
        build_checkpoints("immediate_action", immediate_actions, blocking=False)
        build_checkpoints("approval_guardrail", approval_guardrails, blocking=False)
        build_checkpoints("monitoring", monitoring_actions, blocking=False)

        checkpoints: list[GovernanceRemediationCheckpoint] = []
        event_alerts = list(checkpoint_replay.event_alerts)
        transition_alerts: list[str] = []
        audit_alerts: list[str] = []
        checkpoint_timestamps: dict[str, datetime | None] = {}
        for raw_checkpoint in raw_checkpoints:
            depends_on: list[str] = []
            for dependency_phase in phase_dependencies.get(raw_checkpoint["phase"], []):
                depends_on.extend(phase_checkpoint_ids.get(str(dependency_phase), []))
            checkpoint_id = str(raw_checkpoint["checkpoint_id"])
            checkpoint_state = normalized_statuses.get(
                checkpoint_id,
                {
                    "previous_status": None,
                    "updated_at": None,
                    "recorded_by": None,
                    "actor_id": None,
                    "status": "pending",
                },
            )
            previous_status = (
                str(checkpoint_state.get("previous_status"))
                if checkpoint_state.get("previous_status") is not None
                else None
            )
            updated_at = (
                str(checkpoint_state.get("updated_at"))
                if checkpoint_state.get("updated_at") is not None
                else None
            )
            recorded_by = (
                str(checkpoint_state.get("recorded_by"))
                if checkpoint_state.get("recorded_by") is not None
                else None
            )
            actor_id = (
                str(checkpoint_state.get("actor_id"))
                if checkpoint_state.get("actor_id") is not None
                else None
            )
            status = str(checkpoint_state.get("status", "pending"))
            eligible_actors = resolve_role_actor_refs(identity_bootstrap, str(raw_checkpoint["owner_role"]))
            assigned_actor = resolve_identity_actor_ref(identity_bootstrap, actor_id, recorded_by)
            transition_valid, transition_note = _validate_checkpoint_transition(
                previous_status,
                status,
                allowed_transitions,
            )
            if not transition_valid and transition_note and transition_note not in transition_alerts:
                transition_alerts.append(transition_note)
            audit_valid, audit_note, parsed_timestamp = _validate_checkpoint_audit(
                status=status,
                updated_at=updated_at,
                recorded_by=recorded_by,
                actor_id=actor_id,
                expected_owner_role=str(raw_checkpoint["owner_role"]),
                require_actor_for_non_pending=require_actor_for_non_pending,
                require_timestamp_for_non_pending=require_timestamp_for_non_pending,
                identity_bootstrap=identity_bootstrap,
            )
            if not audit_valid and audit_note and audit_note not in audit_alerts:
                audit_alerts.append(audit_note)
            checkpoint_timestamps[checkpoint_id] = parsed_timestamp
            ready_to_start = transition_valid and audit_valid and status != "completed" and all(
                dependency_completed(dependency_id)
                for dependency_id in dict.fromkeys(depends_on)
            )
            checkpoints.append(
                GovernanceRemediationCheckpoint(
                    checkpoint_id=checkpoint_id,
                    phase=str(raw_checkpoint["phase"]),
                    phase_order=int(raw_checkpoint["phase_order"]),
                    owner_role=str(raw_checkpoint["owner_role"]),
                    eligible_actors=eligible_actors,
                    title=str(raw_checkpoint["title"]),
                    blocking=bool(raw_checkpoint["blocking"]),
                    previous_status=previous_status,
                    updated_at=updated_at,
                    recorded_by=recorded_by,
                    actor_id=actor_id,
                    assigned_actor=assigned_actor,
                    status=status,
                    transition_valid=transition_valid,
                    transition_note=transition_note,
                    audit_valid=audit_valid,
                    audit_note=audit_note,
                    ready_to_start=ready_to_start,
                    depends_on=list(dict.fromkeys(depends_on)),
                    completion_criteria=str(raw_checkpoint["completion_criteria"]),
                )
            )

        if enforce_dependency_timestamp_order:
            updated_checkpoints: list[GovernanceRemediationCheckpoint] = []
            for checkpoint in checkpoints:
                audit_valid = checkpoint.audit_valid
                audit_note = checkpoint.audit_note
                if (
                    audit_valid
                    and checkpoint.status != "pending"
                    and checkpoint.depends_on
                ):
                    completed_dependencies = [
                        dependency_id
                        for dependency_id in checkpoint.depends_on
                        if dependency_completed(dependency_id)
                    ]
                    missing_dependency_timestamps = [
                        dependency_id
                        for dependency_id in completed_dependencies
                        if checkpoint_timestamps.get(dependency_id) is None
                    ]
                    if missing_dependency_timestamps:
                        audit_valid = False
                        audit_note = (
                            "Completed dependency lacks auditable updated_at timestamp: "
                            + ", ".join(missing_dependency_timestamps)
                        )
                    else:
                        current_timestamp = checkpoint_timestamps.get(checkpoint.checkpoint_id)
                        dependency_timestamps = [
                            checkpoint_timestamps[dependency_id]
                            for dependency_id in completed_dependencies
                            if checkpoint_timestamps.get(dependency_id) is not None
                        ]
                        if (
                            current_timestamp is not None
                            and dependency_timestamps
                            and current_timestamp < max(dependency_timestamps)
                        ):
                            latest_dependency = max(dependency_timestamps)
                            dependency_iso = latest_dependency.astimezone(timezone.utc).isoformat().replace("+00:00", "Z")
                            audit_valid = False
                            audit_note = (
                                "Checkpoint update predates dependency completion: "
                                f"{checkpoint.updated_at} < {dependency_iso}"
                            )
                if not audit_valid and audit_note and audit_note not in audit_alerts:
                    audit_alerts.append(audit_note)
                ready_to_start = checkpoint.transition_valid and audit_valid and checkpoint.status != "completed" and all(
                    dependency_completed(dependency_id)
                    for dependency_id in checkpoint.depends_on
                )
                updated_checkpoints.append(
                    replace(
                        checkpoint,
                        audit_valid=audit_valid,
                        audit_note=audit_note,
                        ready_to_start=ready_to_start,
                    )
                )
            checkpoints = updated_checkpoints

        owner_roles = list(dict.fromkeys(checkpoint.owner_role for checkpoint in checkpoints))
        checkpoint_status_counts = {
            "pending": sum(1 for checkpoint in checkpoints if checkpoint.status == "pending"),
            "in_progress": sum(1 for checkpoint in checkpoints if checkpoint.status == "in_progress"),
            "completed": sum(1 for checkpoint in checkpoints if checkpoint.status == "completed"),
        }
        invalid_event_count = checkpoint_replay.invalid_event_count
        invalid_transition_count = sum(1 for checkpoint in checkpoints if not checkpoint.transition_valid)
        invalid_audit_count = sum(1 for checkpoint in checkpoints if not checkpoint.audit_valid)
        if invalid_event_count > 0 or invalid_transition_count > 0 or invalid_audit_count > 0:
            current_release_readiness = "invalid"
        elif any(checkpoint.blocking and checkpoint.status != "completed" for checkpoint in checkpoints):
            current_release_readiness = "blocked"
        elif any(
            checkpoint.phase in {"immediate_action", "approval_guardrail"}
            and checkpoint.status != "completed"
            for checkpoint in checkpoints
        ):
            current_release_readiness = "guarded"
        elif any(checkpoint.phase == "monitoring" and checkpoint.status != "completed" for checkpoint in checkpoints):
            current_release_readiness = "monitoring"
        else:
            current_release_readiness = "complete"

        if release_blockers:
            release_readiness = "blocked"
        elif severity in {"critical", "elevated"} or approval_guardrails:
            release_readiness = "guarded"
        else:
            release_readiness = "monitor_only"

        progress_summary = (
            f"{checkpoint_status_counts['completed']}/{len(checkpoints)} checkpoints completed; "
            f"{checkpoint_status_counts['in_progress']} in progress, "
            f"{checkpoint_status_counts['pending']} pending."
        )
        if invalid_event_count > 0:
            progress_summary += f" {invalid_event_count} invalid event(s) detected."
        if invalid_transition_count > 0:
            progress_summary += f" {invalid_transition_count} invalid transition(s) detected."
        if invalid_audit_count > 0:
            progress_summary += f" {invalid_audit_count} invalid audit record(s) detected."
        summary = (
            f"{severity} remediation for {trend.trend_cluster}: baseline {release_readiness}, "
            f"current {current_release_readiness}, "
            f"{len(release_blockers)} blocker(s), {len(immediate_actions)} immediate action(s), "
            f"and {len(monitoring_actions)} monitoring step(s)."
        )

        plans.append(
            GovernanceRemediationPlan(
                trend_cluster=trend.trend_cluster,
                severity=severity,
                proposal_kind=trend.proposal_kind,
                owner=trend.owner,
                proposal_ids=trend.proposal_ids,
                triggers=trend.systemic_signals,
                release_readiness=release_readiness,
                current_release_readiness=current_release_readiness,
                owner_roles=owner_roles,
                checkpoint_status_counts=checkpoint_status_counts,
                replay_event_count=checkpoint_replay.replay_event_count,
                invalid_event_count=invalid_event_count,
                invalid_transition_count=invalid_transition_count,
                invalid_audit_count=invalid_audit_count,
                event_alerts=event_alerts,
                transition_alerts=transition_alerts,
                audit_alerts=audit_alerts,
                progress_summary=progress_summary,
                immediate_actions=immediate_actions,
                approval_guardrails=approval_guardrails,
                monitoring_actions=monitoring_actions,
                release_blockers=release_blockers,
                checkpoints=checkpoints,
                summary=summary,
            )
        )

    return sorted(
        plans,
        key=lambda plan: (
            -_severity_rank(plan.severity),
            -len(plan.release_blockers),
            -len(plan.immediate_actions),
            plan.trend_cluster,
        ),
    )
