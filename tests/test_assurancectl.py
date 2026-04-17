from __future__ import annotations

import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


PROJECT_ROOT = Path(__file__).resolve().parents[1]
PYTHONPATH = str(PROJECT_ROOT / "src")


class AssuranceCtlTests(unittest.TestCase):
    def run_cli(self, *args: str) -> subprocess.CompletedProcess[str]:
        env = {"PYTHONPATH": PYTHONPATH}
        return subprocess.run(
            [sys.executable, "-m", "assurancectl.cli", *args],
            cwd=PROJECT_ROOT,
            env=env,
            text=True,
            capture_output=True,
            check=False,
        )

    def test_validate_passes(self) -> None:
        result = self.run_cli("validate")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("skeleton config: OK", result.stdout)

    def test_render_localnet_creates_artifacts(self) -> None:
        result = self.run_cli("render-localnet")
        self.assertEqual(result.returncode, 0, result.stderr)
        build = PROJECT_ROOT / "build" / "localnet"
        self.assertTrue((build / "docker-compose.yml").exists())
        self.assertTrue((build / "network-summary.json").exists())
        self.assertTrue((build / "genesis.rendered.json").exists())

    def test_readiness_report_json(self) -> None:
        result = self.run_cli("readiness-report", "--json")
        self.assertEqual(result.returncode, 0, result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["status"], "incubating")
        self.assertGreaterEqual(payload["score"], 80)
        self.assertIn("permissioned testnet mode is enforced", payload["strengths"])

    def test_readiness_report_fails_when_docs_missing(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            root = Path(tmpdir)
            config_dir = root / "config"
            (config_dir / "genesis").mkdir(parents=True)
            (config_dir / "policy").mkdir(parents=True)

            for source, target in (
                ("config/network-topology.json", "config/network-topology.json"),
                ("config/genesis/base-genesis.json", "config/genesis/base-genesis.json"),
                ("config/policy/release-guards.json", "config/policy/release-guards.json"),
            ):
                target_path = root / target
                target_path.write_text((PROJECT_ROOT / source).read_text(encoding="utf-8"), encoding="utf-8")

            env = {"PYTHONPATH": PYTHONPATH}
            result = subprocess.run(
                [sys.executable, "-m", "assurancectl.cli", "--root", str(root), "readiness-report", "--json"],
                cwd=PROJECT_ROOT,
                env=env,
                text=True,
                capture_output=True,
                check=False,
            )
            self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
            payload = json.loads(result.stdout)
            self.assertEqual(payload["status"], "not_ready")
            self.assertTrue(any("missing threat model" == item for item in payload["blockers"]))

    def test_governance_sim_treasury_grant(self) -> None:
        result = self.run_cli(
            "governance-sim",
            "--proposal",
            "examples/proposals/treasury-grant.json",
            "--json",
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["proposal_class"], "high_impact")
        self.assertEqual(payload["recommended_disposition"], "review")
        self.assertIn("treasury_large", payload["triggered_signals"])

    def test_governance_sim_emergency_pause(self) -> None:
        result = self.run_cli(
            "governance-sim",
            "--proposal",
            "examples/proposals/emergency-pause.json",
            "--json",
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["proposal_class"], "safety_critical")
        self.assertIn("safety_council", payload["required_houses"])

    def test_governance_sim_validator_change(self) -> None:
        result = self.run_cli(
            "governance-sim",
            "--proposal",
            "examples/proposals/validator-set-change.json",
            "--json",
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["proposal_class"], "high_impact")
        self.assertIn("token_house", payload["required_houses"])

    def test_governance_drift_emergency_pause_escalates(self) -> None:
        result = self.run_cli(
            "governance-drift",
            "--proposal",
            "examples/proposals/emergency-pause.json",
            "--history",
            "examples/proposals/history.json",
            "--json",
        )
        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["drift_attention"], "escalate")
        self.assertIn("repeated_emergency_pattern", payload["drift_signals"])
        self.assertIn("adverse_precedent", payload["drift_signals"])

    def test_governance_sim_with_history_embeds_drift(self) -> None:
        result = self.run_cli(
            "governance-sim",
            "--proposal",
            "examples/proposals/treasury-grant.json",
            "--history",
            "examples/proposals/history.json",
            "--json",
        )
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertIn("report", payload)
        self.assertIn("drift", payload)
        self.assertEqual(payload["drift"]["drift_attention"], "review")
        self.assertIn("treasury_growth", payload["drift"]["drift_signals"])

    def test_governance_queue_orders_by_priority_and_class(self) -> None:
        result = self.run_cli(
            "governance-queue",
            "--registry",
            "examples/proposals/registry.json",
            "--json",
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(len(payload), 3)
        self.assertEqual(payload[0]["proposal_id"], "draft-pause-001")
        self.assertEqual(payload[0]["proposal_class"], "safety_critical")
        self.assertEqual(payload[1]["proposal_id"], "draft-validator-001")
        self.assertEqual(payload[1]["proposal_class"], "high_impact")
        self.assertEqual(payload[2]["proposal_id"], "draft-grant-001")

    def test_governance_queue_with_history_includes_drift(self) -> None:
        result = self.run_cli(
            "governance-queue",
            "--registry",
            "examples/proposals/registry.json",
            "--history",
            "examples/proposals/history.json",
            "--json",
        )
        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload[0]["proposal_id"], "draft-pause-001")
        self.assertEqual(payload[0]["drift"]["drift_attention"], "escalate")
        self.assertEqual(payload[1]["proposal_id"], "draft-validator-001")
        self.assertEqual(payload[1]["drift"]["drift_attention"], "escalate")
        self.assertEqual(payload[2]["proposal_id"], "draft-grant-001")
        self.assertEqual(payload[2]["drift"]["drift_attention"], "review")

    def test_governance_trends_clusters_portfolio_signals(self) -> None:
        result = self.run_cli(
            "governance-trends",
            "--registry",
            "examples/proposals/registry.json",
            "--history",
            "examples/proposals/history.json",
            "--json",
        )
        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload[0]["trend_cluster"], "emergency_pause:safety-council")
        self.assertEqual(payload[0]["highest_drift_attention"], "escalate")
        self.assertEqual(payload[0]["trend_velocity"], "accelerating")
        self.assertEqual(payload[0]["seasonal_pressure"], "above_norm")
        self.assertEqual(payload[1]["trend_cluster"], "validator_set_change:token-house")
        self.assertIn("adverse_precedent", payload[1]["systemic_signals"])
        self.assertEqual(payload[1]["trend_velocity"], "accelerating")
        self.assertEqual(payload[1]["seasonal_pressure"], "above_norm")
        self.assertEqual(payload[2]["trend_cluster"], "treasury_grant:0ai-core")
        self.assertEqual(payload[2]["highest_drift_attention"], "review")
        self.assertEqual(payload[2]["trend_velocity"], "elevated")
        self.assertEqual(payload[2]["seasonal_pressure"], "watch")

    def test_governance_remediation_emits_structured_plans(self) -> None:
        result = self.run_cli(
            "governance-remediation",
            "--registry",
            "examples/proposals/registry.json",
            "--history",
            "examples/proposals/history.json",
            "--json",
        )
        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload[0]["trend_cluster"], "emergency_pause:safety-council")
        self.assertEqual(payload[0]["severity"], "critical")
        self.assertEqual(payload[0]["release_readiness"], "blocked")
        self.assertIn("safety-council-chair", payload[0]["owner_roles"])
        self.assertIn(
            "Do not execute proposals in this cluster until the incident review and mitigation bundle are approved.",
            payload[0]["release_blockers"],
        )
        self.assertIn("trend_acceleration", payload[0]["triggers"])
        self.assertTrue(any(checkpoint["blocking"] for checkpoint in payload[0]["checkpoints"]))
        self.assertTrue(
            any(
                checkpoint["owner_role"] == "incident-commander"
                and checkpoint["phase"] == "immediate_action"
                and checkpoint["depends_on"]
                for checkpoint in payload[0]["checkpoints"]
            )
        )
        self.assertTrue(
            any(
                checkpoint["phase"] == "release_blocker"
                and checkpoint["phase_order"] == 1
                and not checkpoint["depends_on"]
                for checkpoint in payload[0]["checkpoints"]
            )
        )
        self.assertEqual(payload[1]["trend_cluster"], "validator_set_change:token-house")
        self.assertEqual(payload[1]["severity"], "critical")
        self.assertEqual(payload[2]["trend_cluster"], "treasury_grant:0ai-core")
        self.assertEqual(payload[2]["severity"], "elevated")
        self.assertEqual(payload[2]["release_readiness"], "guarded")
        self.assertIn(
            "Require an explicit reviewer note summarizing why the cluster remains within acceptable bounds.",
            payload[2]["approval_guardrails"],
        )
        self.assertTrue(
            any(
                checkpoint["owner_role"] == "treasury-review-chair"
                and checkpoint["phase"] == "approval_guardrail"
                and checkpoint["depends_on"]
                for checkpoint in payload[2]["checkpoints"]
            )
        )
        self.assertTrue(
            any(
                checkpoint["phase"] == "monitoring"
                and checkpoint["phase_order"] == 4
                and checkpoint["depends_on"]
                for checkpoint in payload[2]["checkpoints"]
            )
        )

    def test_governance_remediation_status_rollup_updates_current_readiness(self) -> None:
        baseline = self.run_cli(
            "governance-remediation",
            "--registry",
            "examples/proposals/registry.json",
            "--history",
            "examples/proposals/history.json",
            "--json",
        )
        self.assertEqual(baseline.returncode, 2, baseline.stdout + baseline.stderr)
        baseline_payload = json.loads(baseline.stdout)
        treasury = next(item for item in baseline_payload if item["trend_cluster"] == "treasury_grant:0ai-core")
        status_payload = {
            "version": "checkpoint-status-test",
            "checkpoints": [
                {
                    "checkpoint_id": checkpoint["checkpoint_id"],
                    "previous_status": (
                        "in_progress"
                        if checkpoint["phase"] in {"immediate_action", "approval_guardrail"}
                        else (
                            "pending"
                            if checkpoint["phase"] == "monitoring" and index == len(treasury["checkpoints"]) - 1
                            else None
                        )
                    ),
                    "updated_at": (
                        f"2026-04-16T12:{index:02d}:00Z"
                        if checkpoint["phase"] in {"immediate_action", "approval_guardrail"}
                        else (
                            "2026-04-16T13:00:00Z"
                            if checkpoint["phase"] == "monitoring" and index == len(treasury["checkpoints"]) - 1
                            else None
                        )
                    ),
                    "recorded_by": (
                        "treasury-program-manager"
                        if checkpoint["phase"] == "immediate_action"
                        else (
                            "treasury-review-chair"
                            if checkpoint["phase"] == "approval_guardrail"
                            else (
                                "finance-telemetry-lead"
                                if checkpoint["phase"] == "monitoring" and index == len(treasury["checkpoints"]) - 1
                                else None
                            )
                        )
                    ),
                    "status": (
                        "completed"
                        if checkpoint["phase"] in {"immediate_action", "approval_guardrail"}
                        else ("in_progress" if index == len(treasury["checkpoints"]) - 1 else "pending")
                    ),
                }
                for index, checkpoint in enumerate(treasury["checkpoints"], start=1)
            ],
        }

        with tempfile.TemporaryDirectory() as tmpdir:
            status_path = Path(tmpdir) / "status.json"
            status_path.write_text(json.dumps(status_payload), encoding="utf-8")
            result = self.run_cli(
                "governance-remediation",
                "--registry",
                "examples/proposals/registry.json",
                "--history",
                "examples/proposals/history.json",
                "--status",
                str(status_path),
                "--json",
            )

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        treasury_updated = next(item for item in payload if item["trend_cluster"] == "treasury_grant:0ai-core")
        self.assertEqual(treasury_updated["release_readiness"], "guarded")
        self.assertEqual(treasury_updated["current_release_readiness"], "monitoring")
        self.assertEqual(treasury_updated["checkpoint_status_counts"]["completed"], 10)
        self.assertEqual(treasury_updated["checkpoint_status_counts"]["in_progress"], 1)
        self.assertEqual(treasury_updated["checkpoint_status_counts"]["pending"], 1)
        self.assertEqual(treasury_updated["invalid_transition_count"], 0)
        self.assertEqual(treasury_updated["invalid_audit_count"], 0)
        self.assertTrue(
            all(
                checkpoint["ready_to_start"]
                for checkpoint in treasury_updated["checkpoints"]
                if checkpoint["phase"] == "monitoring"
            )
        )

    def test_governance_remediation_invalid_transition_marks_plan_invalid(self) -> None:
        baseline = self.run_cli(
            "governance-remediation",
            "--registry",
            "examples/proposals/registry.json",
            "--history",
            "examples/proposals/history.json",
            "--json",
        )
        self.assertEqual(baseline.returncode, 2, baseline.stdout + baseline.stderr)
        baseline_payload = json.loads(baseline.stdout)
        treasury = next(item for item in baseline_payload if item["trend_cluster"] == "treasury_grant:0ai-core")
        invalid_status_payload = {
            "version": "checkpoint-status-invalid-test",
            "checkpoints": [
                {
                    "checkpoint_id": treasury["checkpoints"][0]["checkpoint_id"],
                    "previous_status": "pending",
                    "updated_at": "2026-04-16T12:00:00Z",
                    "recorded_by": "treasury-program-manager",
                    "status": "completed",
                }
            ],
        }

        with tempfile.TemporaryDirectory() as tmpdir:
            status_path = Path(tmpdir) / "status-invalid.json"
            status_path.write_text(json.dumps(invalid_status_payload), encoding="utf-8")
            result = self.run_cli(
                "governance-remediation",
                "--registry",
                "examples/proposals/registry.json",
                "--history",
                "examples/proposals/history.json",
                "--status",
                str(status_path),
                "--json",
            )

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        treasury_invalid = next(item for item in payload if item["trend_cluster"] == "treasury_grant:0ai-core")
        self.assertEqual(treasury_invalid["current_release_readiness"], "invalid")
        self.assertEqual(treasury_invalid["invalid_transition_count"], 1)
        self.assertTrue(treasury_invalid["transition_alerts"])
        self.assertTrue(
            any(
                not checkpoint["transition_valid"] and checkpoint["transition_note"]
                for checkpoint in treasury_invalid["checkpoints"]
                if checkpoint["checkpoint_id"] == treasury["checkpoints"][0]["checkpoint_id"]
            )
        )

    def test_governance_remediation_dependency_timestamp_order_marks_plan_invalid(self) -> None:
        baseline = self.run_cli(
            "governance-remediation",
            "--registry",
            "examples/proposals/registry.json",
            "--history",
            "examples/proposals/history.json",
            "--json",
        )
        self.assertEqual(baseline.returncode, 2, baseline.stdout + baseline.stderr)
        baseline_payload = json.loads(baseline.stdout)
        treasury = next(item for item in baseline_payload if item["trend_cluster"] == "treasury_grant:0ai-core")

        status_payload = {"version": "checkpoint-status-audit-order-test", "checkpoints": []}
        approval_checkpoint_id = None
        for index, checkpoint in enumerate(treasury["checkpoints"], start=1):
            if checkpoint["phase"] == "immediate_action":
                status_payload["checkpoints"].append(
                    {
                        "checkpoint_id": checkpoint["checkpoint_id"],
                        "previous_status": "in_progress",
                        "updated_at": f"2026-04-16T12:{index:02d}:00Z",
                        "recorded_by": "treasury-program-manager",
                        "status": "completed",
                    }
                )
            elif checkpoint["phase"] == "approval_guardrail" and approval_checkpoint_id is None:
                approval_checkpoint_id = checkpoint["checkpoint_id"]
                status_payload["checkpoints"].append(
                    {
                        "checkpoint_id": checkpoint["checkpoint_id"],
                        "previous_status": "in_progress",
                        "updated_at": "2026-04-16T12:01:00Z",
                        "recorded_by": "treasury-review-chair",
                        "status": "completed",
                    }
                )

        self.assertIsNotNone(approval_checkpoint_id)

        with tempfile.TemporaryDirectory() as tmpdir:
            status_path = Path(tmpdir) / "status-audit-order.json"
            status_path.write_text(json.dumps(status_payload), encoding="utf-8")
            result = self.run_cli(
                "governance-remediation",
                "--registry",
                "examples/proposals/registry.json",
                "--history",
                "examples/proposals/history.json",
                "--status",
                str(status_path),
                "--json",
            )

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        treasury_invalid = next(item for item in payload if item["trend_cluster"] == "treasury_grant:0ai-core")
        self.assertEqual(treasury_invalid["current_release_readiness"], "invalid")
        self.assertGreaterEqual(treasury_invalid["invalid_audit_count"], 1)
        self.assertTrue(treasury_invalid["audit_alerts"])
        self.assertTrue(
            any(
                not checkpoint["audit_valid"]
                and checkpoint["audit_note"]
                and "predates dependency completion" in checkpoint["audit_note"]
                for checkpoint in treasury_invalid["checkpoints"]
                if checkpoint["checkpoint_id"] == approval_checkpoint_id
            )
        )

    def test_governance_drift_suppresses_stable_pattern_signals(self) -> None:
        proposal = {
            "proposal_id": "draft-validator-stable-001",
            "title": "Routine validator weight maintenance",
            "summary": "Rebalance validator weights as part of the regular quarterly maintenance window.",
            "kind": "validator_set_change",
            "requested_by": "token-house",
            "treasury_amount_usd": 0,
            "affects_validators": True,
            "changes_governance_rules": False,
            "touches_ai_systems": False,
            "is_emergency": False,
            "external_dependencies": [],
            "actions": [
                "rebalance voting power",
                "publish maintenance summary",
            ],
        }
        history = {
            "version": "proposal-history-test",
            "entries": [
                {
                    "recorded_at": "2026-01-01T00:00:00Z",
                    "status": "executed",
                    "outcome": "completed",
                    "proposal": {
                        **proposal,
                        "proposal_id": "hist-stable-validator-001",
                        "title": "Routine validator maintenance January",
                    },
                },
                {
                    "recorded_at": "2026-02-01T00:00:00Z",
                    "status": "executed",
                    "outcome": "completed",
                    "proposal": {
                        **proposal,
                        "proposal_id": "hist-stable-validator-002",
                        "title": "Routine validator maintenance February",
                    },
                },
            ],
        }

        with tempfile.TemporaryDirectory() as tmpdir:
            root = Path(tmpdir)
            proposal_path = root / "proposal.json"
            history_path = root / "history.json"
            proposal_path.write_text(json.dumps(proposal), encoding="utf-8")
            history_path.write_text(json.dumps(history), encoding="utf-8")

            env = {"PYTHONPATH": PYTHONPATH}
            result = subprocess.run(
                [
                    sys.executable,
                    "-m",
                    "assurancectl.cli",
                    "governance-drift",
                    "--proposal",
                    str(proposal_path),
                    "--history",
                    str(history_path),
                    "--json",
                ],
                cwd=PROJECT_ROOT,
                env=env,
                text=True,
                capture_output=True,
                check=False,
            )

        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["drift_attention"], "normal")
        self.assertTrue(payload["stable_pattern"])
        self.assertIn("requester_concentration", payload["suppressed_signals"])
        self.assertIn("repeated_validator_change", payload["suppressed_signals"])


if __name__ == "__main__":
    unittest.main()
