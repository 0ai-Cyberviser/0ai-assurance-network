"""CLI for the 0AI Assurance Network repo skeleton."""

from __future__ import annotations

import argparse
import json
import sys

from .config import ValidationError, load_config, validate_all
from .inference import (
    infer_governance_drift,
    infer_governance_portfolio_trends,
    infer_governance_queue,
    infer_governance_remediation_plans,
    load_checkpoint_statuses,
    load_checkpoint_signer_policy,
    infer_governance_report,
    load_history,
    load_inference_policy,
    load_proposal,
    load_registry,
    replay_checkpoint_state,
)
from .readiness import build_readiness_report
from .render import write_localnet_artifacts


def _parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(prog="assurancectl")
    parser.add_argument("--root", default=None, help="project root override")
    subparsers = parser.add_subparsers(dest="command", required=True)

    subparsers.add_parser("validate", help="validate repo skeleton config")
    subparsers.add_parser("render-localnet", help="render localnet artifacts")

    readiness = subparsers.add_parser("readiness-report", help="build a launch-readiness report")
    readiness.add_argument("--json", action="store_true", help="emit machine-readable JSON")

    governance = subparsers.add_parser("governance-sim", help="simulate governance inference on a proposal")
    governance.add_argument("--proposal", required=True, help="proposal JSON file path")
    governance.add_argument("--history", help="governance history JSON file path")
    governance.add_argument("--json", action="store_true", help="emit machine-readable JSON")

    queue = subparsers.add_parser("governance-queue", help="score a registry of proposals")
    queue.add_argument("--registry", required=True, help="proposal registry JSON file path")
    queue.add_argument("--history", help="governance history JSON file path")
    queue.add_argument("--json", action="store_true", help="emit machine-readable JSON")

    trends = subparsers.add_parser("governance-trends", help="cluster governance trends across the queue")
    trends.add_argument("--registry", required=True, help="proposal registry JSON file path")
    trends.add_argument("--history", required=True, help="governance history JSON file path")
    trends.add_argument("--json", action="store_true", help="emit machine-readable JSON")

    remediation = subparsers.add_parser(
        "governance-remediation",
        help="emit structured mitigation bundles for the active governance queue",
    )
    remediation.add_argument("--registry", required=True, help="proposal registry JSON file path")
    remediation.add_argument("--history", required=True, help="governance history JSON file path")
    remediation.add_argument("--status", help="checkpoint status JSON file path")
    remediation.add_argument("--json", action="store_true", help="emit machine-readable JSON")

    replay = subparsers.add_parser(
        "governance-replay",
        help="replay a checkpoint event log or snapshot into current checkpoint state",
    )
    replay.add_argument("--status", required=True, help="checkpoint state or event log JSON file path")
    replay.add_argument("--json", action="store_true", help="emit machine-readable JSON")

    drift = subparsers.add_parser("governance-drift", help="compare a proposal against governance history")
    drift.add_argument("--proposal", required=True, help="proposal JSON file path")
    drift.add_argument("--history", required=True, help="governance history JSON file path")
    drift.add_argument("--json", action="store_true", help="emit machine-readable JSON")
    return parser


def _print_readiness(report, emit_json: bool) -> None:
    if emit_json:
        print(
            json.dumps(
                {
                    "status": report.status,
                    "score": report.score,
                    "strengths": report.strengths,
                    "blockers": report.blockers,
                    "next_actions": report.next_actions,
                },
                indent=2,
            )
        )
        return

    print(f"Readiness status: {report.status}")
    print(f"Readiness score: {report.score}/100")
    if report.strengths:
        print("Strengths:")
        for item in report.strengths:
            print(f"  - {item}")
    if report.blockers:
        print("Blockers:")
        for item in report.blockers:
            print(f"  - {item}")
    print("Next actions:")
    for item in report.next_actions:
        print(f"  - {item}")


def _governance_report_payload(report) -> dict[str, object]:
    return {
        "engine": report.engine,
        "proposal_id": report.proposal_id,
        "title": report.title,
        "proposal_class": report.proposal_class,
        "confidence": report.confidence,
        "risk_score": report.risk_score,
        "required_houses": report.required_houses,
        "recommended_disposition": report.recommended_disposition,
        "triggered_signals": report.triggered_signals,
        "rationale": report.rationale,
        "remediation": report.remediation,
        "summary": report.summary,
    }


def _governance_drift_payload(report) -> dict[str, object]:
    return {
        "proposal_id": report.proposal_id,
        "title": report.title,
        "trend_cluster": report.trend_cluster,
        "cluster_size": report.cluster_size,
        "cluster_outcomes": report.cluster_outcomes,
        "stable_pattern": report.stable_pattern,
        "suppressed_signals": report.suppressed_signals,
        "historical_matches": report.historical_matches,
        "baseline_risk_score": report.baseline_risk_score,
        "drift_score": report.drift_score,
        "drift_attention": report.drift_attention,
        "drift_signals": report.drift_signals,
        "precedent_ids": report.precedent_ids,
        "rationale": report.rationale,
        "remediation": report.remediation,
        "summary": report.summary,
    }


def _print_governance_report(report, emit_json: bool) -> None:
    payload = _governance_report_payload(report)
    if emit_json:
        print(json.dumps(payload, indent=2))
        return

    print(f"Proposal: {report.title}")
    print(f"Proposal class: {report.proposal_class}")
    print(f"Confidence: {report.confidence:.2f}")
    print(f"Risk score: {report.risk_score}")
    print(f"Required houses: {', '.join(report.required_houses)}")
    print(f"Disposition: {report.recommended_disposition}")
    print(f"Summary: {report.summary}")
    if report.triggered_signals:
        print("Triggered signals:")
        for item in report.triggered_signals:
            print(f"  - {item}")
    if report.rationale:
        print("Rationale:")
        for item in report.rationale:
            print(f"  - {item}")
    if report.remediation:
        print("Remediation:")
        for item in report.remediation:
            print(f"  - {item}")


def _print_governance_drift(report, emit_json: bool) -> None:
    payload = _governance_drift_payload(report)
    if emit_json:
        print(json.dumps(payload, indent=2))
        return

    print("Historical drift:")
    print(f"  attention: {report.drift_attention}")
    print(f"  drift score: {report.drift_score}")
    print(f"  historical matches: {report.historical_matches}")
    print(f"  cluster: {report.trend_cluster} ({report.cluster_size})")
    baseline = "n/a" if report.baseline_risk_score is None else f"{report.baseline_risk_score:.2f}"
    print(f"  baseline risk: {baseline}")
    print(f"  stable pattern: {'yes' if report.stable_pattern else 'no'}")
    print(f"  summary: {report.summary}")
    if report.cluster_outcomes:
        print("  cluster outcomes:")
        for key, value in report.cluster_outcomes.items():
            print(f"    - {key}: {value}")
    if report.drift_signals:
        print("  signals:")
        for item in report.drift_signals:
            print(f"    - {item}")
    if report.suppressed_signals:
        print("  suppressed:")
        for item in report.suppressed_signals:
            print(f"    - {item}")
    if report.precedent_ids:
        print("  precedents:")
        for item in report.precedent_ids:
            print(f"    - {item}")
    if report.rationale:
        print("  rationale:")
        for item in report.rationale:
            print(f"    - {item}")
    if report.remediation:
        print("  remediation:")
        for item in report.remediation:
            print(f"    - {item}")


def _print_governance_queue(entries, emit_json: bool) -> None:
    payload = [
        {
            "path": entry.path,
            "status": entry.status,
            "priority": entry.priority,
            "owner": entry.owner,
            "proposal_id": entry.report.proposal_id,
            "title": entry.report.title,
            "proposal_class": entry.report.proposal_class,
            "confidence": entry.report.confidence,
            "risk_score": entry.report.risk_score,
            "required_houses": entry.report.required_houses,
            "recommended_disposition": entry.report.recommended_disposition,
            "summary": entry.report.summary,
            "drift": _governance_drift_payload(entry.drift) if entry.drift else None,
        }
        for entry in entries
    ]
    if emit_json:
        print(json.dumps(payload, indent=2))
        return

    print(f"Queue entries: {len(entries)}")
    for index, entry in enumerate(entries, start=1):
        print(
            f"{index}. [{entry.priority}/{entry.status}] {entry.report.title} "
            f"-> {entry.report.proposal_class} ({entry.report.risk_score})"
        )
        print(f"   houses: {', '.join(entry.report.required_houses)}")
        print(f"   disposition: {entry.report.recommended_disposition}")
        print(f"   owner: {entry.owner}")
        print(f"   summary: {entry.report.summary}")
        if entry.drift:
            print(f"   drift: {entry.drift.drift_attention} ({entry.drift.drift_score})")
            print(f"   drift summary: {entry.drift.summary}")


def _print_governance_trends(trends, emit_json: bool) -> None:
    payload = [
        {
            "trend_cluster": trend.trend_cluster,
            "proposal_kind": trend.proposal_kind,
            "owner": trend.owner,
            "active_proposals": trend.active_proposals,
            "proposal_ids": trend.proposal_ids,
            "historical_precedents": trend.historical_precedents,
            "historical_outcomes": trend.historical_outcomes,
            "recent_window_days": trend.recent_window_days,
            "baseline_window_days": trend.baseline_window_days,
            "recent_historical_precedents": trend.recent_historical_precedents,
            "baseline_historical_precedents": trend.baseline_historical_precedents,
            "recent_active_proposals": trend.recent_active_proposals,
            "trend_velocity": trend.trend_velocity,
            "seasonal_kind_recent_total": trend.seasonal_kind_recent_total,
            "seasonal_kind_baseline_total": trend.seasonal_kind_baseline_total,
            "seasonal_expected_recent": trend.seasonal_expected_recent,
            "seasonal_pressure": trend.seasonal_pressure,
            "highest_drift_attention": trend.highest_drift_attention,
            "highest_drift_score": trend.highest_drift_score,
            "proposal_classes": trend.proposal_classes,
            "systemic_signals": trend.systemic_signals,
            "summary": trend.summary,
        }
        for trend in trends
    ]
    if emit_json:
        print(json.dumps(payload, indent=2))
        return

    print(f"Portfolio trends: {len(trends)}")
    for index, trend in enumerate(trends, start=1):
        print(
            f"{index}. {trend.trend_cluster} -> {trend.highest_drift_attention} "
            f"({trend.highest_drift_score})"
        )
        print(f"   velocity: {trend.trend_velocity}")
        print(f"   seasonal pressure: {trend.seasonal_pressure}")
        print(f"   active proposals: {trend.active_proposals}")
        print(f"   historical precedents: {trend.historical_precedents}")
        print(
            f"   recent/baseline: {trend.recent_historical_precedents}+{trend.recent_active_proposals} active over "
            f"{trend.recent_window_days}d vs {trend.baseline_historical_precedents} over {trend.baseline_window_days}d"
        )
        print(
            f"   kind seasonal: {trend.seasonal_kind_recent_total} recent vs {trend.seasonal_kind_baseline_total} baseline "
            f"(expected {trend.seasonal_expected_recent:.2f})"
        )
        print(f"   proposal classes: {', '.join(trend.proposal_classes)}")
        if trend.systemic_signals:
            print(f"   signals: {', '.join(trend.systemic_signals)}")
        print(f"   summary: {trend.summary}")


def _print_governance_replay(replay, emit_json: bool) -> None:
    payload = {
        "source_kind": replay.source_kind,
        "replay_event_count": replay.replay_event_count,
        "invalid_event_count": replay.invalid_event_count,
        "event_alerts": replay.event_alerts,
        "checkpoints": [
            {
                "checkpoint_id": checkpoint_id,
                "previous_status": checkpoint_state.get("previous_status"),
                "updated_at": checkpoint_state.get("updated_at"),
                "recorded_by": checkpoint_state.get("recorded_by"),
                "status": checkpoint_state.get("status"),
            }
            for checkpoint_id, checkpoint_state in sorted(replay.checkpoints.items())
        ],
    }
    if emit_json:
        print(json.dumps(payload, indent=2))
        return

    print(f"Replay source: {replay.source_kind}")
    print(f"Replayed events: {replay.replay_event_count}")
    print(f"Invalid events: {replay.invalid_event_count}")
    if replay.event_alerts:
        print("Event alerts:")
        for item in replay.event_alerts:
            print(f"  - {item}")
    print(f"Checkpoint states: {len(replay.checkpoints)}")
    for checkpoint_id, checkpoint_state in sorted(replay.checkpoints.items()):
        print(
            f"  - {checkpoint_id}: {checkpoint_state.get('previous_status')} -> "
            f"{checkpoint_state.get('status')}"
        )
        if checkpoint_state.get("updated_at"):
            print(f"    updated: {checkpoint_state.get('updated_at')}")
        if checkpoint_state.get("recorded_by"):
            print(f"    by: {checkpoint_state.get('recorded_by')}")


def _print_governance_remediation(plans, emit_json: bool) -> None:
    payload = [
        {
            "trend_cluster": plan.trend_cluster,
            "severity": plan.severity,
            "proposal_kind": plan.proposal_kind,
            "owner": plan.owner,
            "proposal_ids": plan.proposal_ids,
            "triggers": plan.triggers,
            "release_readiness": plan.release_readiness,
            "current_release_readiness": plan.current_release_readiness,
            "owner_roles": plan.owner_roles,
            "checkpoint_status_counts": plan.checkpoint_status_counts,
            "replay_event_count": plan.replay_event_count,
            "invalid_event_count": plan.invalid_event_count,
            "invalid_transition_count": plan.invalid_transition_count,
            "invalid_audit_count": plan.invalid_audit_count,
            "event_alerts": plan.event_alerts,
            "transition_alerts": plan.transition_alerts,
            "audit_alerts": plan.audit_alerts,
            "progress_summary": plan.progress_summary,
            "immediate_actions": plan.immediate_actions,
            "approval_guardrails": plan.approval_guardrails,
            "monitoring_actions": plan.monitoring_actions,
            "release_blockers": plan.release_blockers,
            "checkpoints": [
                {
                    "checkpoint_id": checkpoint.checkpoint_id,
                    "phase": checkpoint.phase,
                    "phase_order": checkpoint.phase_order,
                    "owner_role": checkpoint.owner_role,
                    "title": checkpoint.title,
                    "blocking": checkpoint.blocking,
                    "previous_status": checkpoint.previous_status,
                    "updated_at": checkpoint.updated_at,
                    "recorded_by": checkpoint.recorded_by,
                    "status": checkpoint.status,
                    "transition_valid": checkpoint.transition_valid,
                    "transition_note": checkpoint.transition_note,
                    "audit_valid": checkpoint.audit_valid,
                    "audit_note": checkpoint.audit_note,
                    "ready_to_start": checkpoint.ready_to_start,
                    "depends_on": checkpoint.depends_on,
                    "completion_criteria": checkpoint.completion_criteria,
                }
                for checkpoint in plan.checkpoints
            ],
            "summary": plan.summary,
        }
        for plan in plans
    ]
    if emit_json:
        print(json.dumps(payload, indent=2))
        return

    print(f"Remediation plans: {len(plans)}")
    for index, plan in enumerate(plans, start=1):
        print(f"{index}. {plan.trend_cluster} -> {plan.severity}")
        print(f"   kind: {plan.proposal_kind}")
        print(f"   owner: {plan.owner}")
        print(f"   release readiness: {plan.release_readiness}")
        print(f"   current readiness: {plan.current_release_readiness}")
        print(f"   replayed events: {plan.replay_event_count}")
        if plan.owner_roles:
            print(f"   owner roles: {', '.join(plan.owner_roles)}")
        if plan.triggers:
            print(f"   triggers: {', '.join(plan.triggers)}")
        if plan.event_alerts:
            print(f"   event alerts: {'; '.join(plan.event_alerts)}")
        if plan.transition_alerts:
            print(f"   transition alerts: {'; '.join(plan.transition_alerts)}")
        if plan.audit_alerts:
            print(f"   audit alerts: {'; '.join(plan.audit_alerts)}")
        print(f"   progress: {plan.progress_summary}")
        print(f"   summary: {plan.summary}")
        if plan.release_blockers:
            print("   release blockers:")
            for item in plan.release_blockers:
                print(f"     - {item}")
        if plan.immediate_actions:
            print("   immediate actions:")
            for item in plan.immediate_actions:
                print(f"     - {item}")
        if plan.approval_guardrails:
            print("   approval guardrails:")
            for item in plan.approval_guardrails:
                print(f"     - {item}")
        if plan.monitoring_actions:
            print("   monitoring actions:")
            for item in plan.monitoring_actions:
                print(f"     - {item}")
        if plan.checkpoints:
            print("   checkpoints:")
            for checkpoint in plan.checkpoints:
                block = "blocking" if checkpoint.blocking else "non-blocking"
                print(
                    f"     - {checkpoint.checkpoint_id} [{checkpoint.phase}:{checkpoint.phase_order}/{block}] "
                    f"{checkpoint.owner_role}: {checkpoint.title} ({checkpoint.status})"
                )
                if checkpoint.previous_status:
                    print(f"       from: {checkpoint.previous_status}")
                if checkpoint.updated_at:
                    print(f"       updated: {checkpoint.updated_at}")
                if checkpoint.recorded_by:
                    print(f"       by: {checkpoint.recorded_by}")
                if checkpoint.transition_note and not checkpoint.transition_valid:
                    print(f"       transition: {checkpoint.transition_note}")
                if checkpoint.audit_note and not checkpoint.audit_valid:
                    print(f"       audit: {checkpoint.audit_note}")
                if checkpoint.ready_to_start:
                    print("       ready: yes")
                if checkpoint.depends_on:
                    print(f"       after: {', '.join(checkpoint.depends_on)}")


def main(argv: list[str] | None = None) -> int:
    parser = _parser()
    args = parser.parse_args(argv)

    try:
        config = load_config(args.root)
        validate_all(config)
    except ValidationError as exc:
        print(f"config validation failed: {exc}", file=sys.stderr)
        return 1

    if args.command == "validate":
        print("0AI Assurance Network skeleton config: OK")
        return 0

    if args.command == "render-localnet":
        output = write_localnet_artifacts(config)
        print(f"rendered localnet artifacts in {output}")
        return 0

    if args.command == "readiness-report":
        report = build_readiness_report(config)
        _print_readiness(report, emit_json=args.json)
        return 0 if report.status != "not_ready" else 2

    if args.command == "governance-sim":
        proposal = load_proposal(args.proposal)
        report = infer_governance_report(config, proposal)
        if args.history:
            history = load_history(args.history)
            drift = infer_governance_drift(config, proposal, history)
            if args.json:
                print(
                    json.dumps(
                        {
                            "report": _governance_report_payload(report),
                            "drift": _governance_drift_payload(drift),
                        },
                        indent=2,
                    )
                )
            else:
                _print_governance_report(report, emit_json=False)
                print()
                _print_governance_drift(drift, emit_json=False)
            return 0 if report.recommended_disposition != "hold" and drift.drift_attention != "escalate" else 2
        _print_governance_report(report, emit_json=args.json)
        return 0 if report.recommended_disposition != "hold" else 2

    if args.command == "governance-queue":
        registry = load_registry(args.registry)
        history = load_history(args.history) if args.history else None
        policy = load_inference_policy(config) if history else None
        entries = infer_governance_queue(config, registry, registry_path=args.registry, history=history, policy=policy)
        _print_governance_queue(entries, emit_json=args.json)
        if history and not args.json:
            print()
            trends = infer_governance_portfolio_trends(entries, history=history, policy=policy)
            _print_governance_trends(trends, emit_json=False)
        return (
            0
            if all(
                entry.report.recommended_disposition != "hold"
                and (entry.drift is None or entry.drift.drift_attention != "escalate")
                for entry in entries
            )
            else 2
        )

    if args.command == "governance-trends":
        registry = load_registry(args.registry)
        history = load_history(args.history)
        policy = load_inference_policy(config)
        entries = infer_governance_queue(config, registry, registry_path=args.registry, history=history, policy=policy)
        trends = infer_governance_portfolio_trends(entries, history=history, policy=policy)
        _print_governance_trends(trends, emit_json=args.json)
        return (
            0
            if all(
                trend.highest_drift_attention != "escalate" and trend.trend_velocity != "accelerating"
                for trend in trends
            )
            else 2
        )

    if args.command == "governance-remediation":
        registry = load_registry(args.registry)
        history = load_history(args.history)
        policy = load_inference_policy(config)
        signer_policy = load_checkpoint_signer_policy(config)
        checkpoint_statuses = load_checkpoint_statuses(args.status) if args.status else None
        entries = infer_governance_queue(config, registry, registry_path=args.registry, history=history, policy=policy)
        trends = infer_governance_portfolio_trends(entries, history=history, policy=policy)
        plans = infer_governance_remediation_plans(
            entries,
            trends,
            policy=policy,
            checkpoint_statuses=checkpoint_statuses,
            signature_policy=signer_policy,
        )
        _print_governance_remediation(plans, emit_json=args.json)
        return (
            0
            if all(plan.current_release_readiness in {"monitoring", "complete"} for plan in plans)
            else 2
        )

    if args.command == "governance-replay":
        checkpoint_statuses = load_checkpoint_statuses(args.status)
        signer_policy = load_checkpoint_signer_policy(config)
        replay = replay_checkpoint_state(checkpoint_statuses, signature_policy=signer_policy)
        _print_governance_replay(replay, emit_json=args.json)
        return 0 if replay.invalid_event_count == 0 else 2

    if args.command == "governance-drift":
        proposal = load_proposal(args.proposal)
        history = load_history(args.history)
        drift = infer_governance_drift(config, proposal, history)
        _print_governance_drift(drift, emit_json=args.json)
        return 0 if drift.drift_attention != "escalate" else 2

    parser.error(f"unknown command: {args.command}")
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
