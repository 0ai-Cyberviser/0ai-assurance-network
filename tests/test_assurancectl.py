from __future__ import annotations

import hashlib
import hmac
import json
import shutil
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

    def load_checkpoint_signer_policy(self) -> dict[str, object]:
        return json.loads(
            (PROJECT_ROOT / "config" / "governance" / "checkpoint-signers.json").read_text(encoding="utf-8")
        )

    def build_signed_event(
        self,
        *,
        checkpoint_id: str,
        previous_status: str,
        new_status: str,
        updated_at: str,
        recorded_by: str,
        rationale: str,
        signature_id: str,
        signer_id: str | None = None,
        actor_id: str | None = None,
    ) -> dict[str, object]:
        signer_policy = self.load_checkpoint_signer_policy()
        signers = signer_policy["signers"]
        if signer_id is None:
            signer_entry = next(entry for entry in signers if recorded_by in entry["roles"])
        else:
            signer_entry = next(entry for entry in signers if entry["signer_id"] == signer_id)
        signature_meta = {
            "format": signer_policy["signature_format"],
            "signer_id": signer_entry["signer_id"],
            "key_id": signer_entry["key_id"],
            "signature_id": signature_id,
            "signed_at": "2026-04-16T10:00:00Z",
            "expires_at": "2026-04-17T10:00:00Z",
        }
        resolved_actor_id = actor_id or str(signer_entry["actor_id"])
        message = json.dumps(
            {
                "checkpoint_id": checkpoint_id,
                "previous_status": previous_status,
                "new_status": new_status,
                "updated_at": updated_at,
                "recorded_by": recorded_by,
                "actor_id": resolved_actor_id,
                "rationale": rationale,
                "signature": signature_meta,
            },
            sort_keys=True,
            separators=(",", ":"),
        )
        signature_value = hmac.new(
            str(signer_entry["shared_secret"]).encode("utf-8"),
            message.encode("utf-8"),
            hashlib.sha256,
        ).hexdigest()
        return {
            "checkpoint_id": checkpoint_id,
            "previous_status": previous_status,
            "new_status": new_status,
            "updated_at": updated_at,
            "recorded_by": recorded_by,
            "actor_id": resolved_actor_id,
            "rationale": rationale,
            "signature": {
                **signature_meta,
                "value": signature_value,
            },
        }

    def build_treasury_event_log(self, checkpoints: list[dict[str, object]]) -> dict[str, object]:
        event_payload: dict[str, object] = {"version": "checkpoint-event-log-test", "events": []}
        events = event_payload["events"]
        assert isinstance(events, list)
        completed_counter = 0
        for checkpoint in checkpoints:
            checkpoint_id = str(checkpoint["checkpoint_id"])
            owner_role = str(checkpoint["owner_role"])
            phase = str(checkpoint["phase"])
            if phase in {"immediate_action", "approval_guardrail"}:
                completed_counter += 1
                events.append(
                    self.build_signed_event(
                        checkpoint_id=checkpoint_id,
                        previous_status="pending",
                        new_status="in_progress",
                        updated_at=f"2026-04-16T11:{completed_counter:02d}:00Z",
                        recorded_by=owner_role,
                        rationale=f"Started {checkpoint_id}.",
                        signature_id=f"{checkpoint_id}-start",
                    )
                )
                events.append(
                    self.build_signed_event(
                        checkpoint_id=checkpoint_id,
                        previous_status="in_progress",
                        new_status="completed",
                        updated_at=f"2026-04-16T12:{completed_counter:02d}:00Z",
                        recorded_by=owner_role,
                        rationale=f"Completed {checkpoint_id}.",
                        signature_id=f"{checkpoint_id}-complete",
                    )
                )
            elif phase == "monitoring" and checkpoint_id.endswith("-1"):
                events.append(
                    self.build_signed_event(
                        checkpoint_id=checkpoint_id,
                        previous_status="pending",
                        new_status="in_progress",
                        updated_at="2026-04-16T13:00:00Z",
                        recorded_by=owner_role,
                        rationale=f"Started monitoring for {checkpoint_id}.",
                        signature_id=f"{checkpoint_id}-monitoring",
                    )
                )
        return event_payload

    def test_validate_passes(self) -> None:
        result = self.run_cli("validate")
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("skeleton config: OK", result.stdout)

    def test_render_localnet_creates_artifacts(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            root = Path(tmpdir)
            shutil.copytree(PROJECT_ROOT / "config", root / "config")

            result = self.run_cli("--root", str(root), "render-localnet")
            self.assertEqual(result.returncode, 0, result.stderr)
            build = root / "build" / "localnet"
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
                ("config/governance/inference-policy.json", "config/governance/inference-policy.json"),
                ("config/governance/checkpoint-signers.json", "config/governance/checkpoint-signers.json"),
                ("config/modules/milestone-1.json", "config/modules/milestone-1.json"),
                ("config/identity/bootstrap.json", "config/identity/bootstrap.json"),
            ):
                target_path = root / target
                target_path.parent.mkdir(parents=True, exist_ok=True)
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

    def test_validate_fails_when_identity_binding_missing(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            root = Path(tmpdir)
            for source, target in (
                ("config/network-topology.json", "config/network-topology.json"),
                ("config/genesis/base-genesis.json", "config/genesis/base-genesis.json"),
                ("config/policy/release-guards.json", "config/policy/release-guards.json"),
                ("config/governance/inference-policy.json", "config/governance/inference-policy.json"),
                ("config/governance/checkpoint-signers.json", "config/governance/checkpoint-signers.json"),
                ("config/modules/milestone-1.json", "config/modules/milestone-1.json"),
                ("config/identity/bootstrap.json", "config/identity/bootstrap.json"),
            ):
                target_path = root / target
                target_path.parent.mkdir(parents=True, exist_ok=True)
                target_path.write_text((PROJECT_ROOT / source).read_text(encoding="utf-8"), encoding="utf-8")

            identity_path = root / "config" / "identity" / "bootstrap.json"
            identity_payload = json.loads(identity_path.read_text(encoding="utf-8"))
            identity_payload["role_bindings"] = [
                binding
                for binding in identity_payload["role_bindings"]
                if binding["role"] != "network_admin"
            ]
            identity_path.write_text(json.dumps(identity_payload), encoding="utf-8")

            env = {"PYTHONPATH": PYTHONPATH}
            result = subprocess.run(
                [sys.executable, "-m", "assurancectl.cli", "--root", str(root), "validate"],
                cwd=PROJECT_ROOT,
                env=env,
                text=True,
                capture_output=True,
                check=False,
            )

            self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
            self.assertIn("identity bootstrap missing active bindings", result.stderr)

    def test_validate_fails_when_identity_binding_is_duplicated(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            root = Path(tmpdir)
            for source, target in (
                ("config/network-topology.json", "config/network-topology.json"),
                ("config/genesis/base-genesis.json", "config/genesis/base-genesis.json"),
                ("config/policy/release-guards.json", "config/policy/release-guards.json"),
                ("config/governance/inference-policy.json", "config/governance/inference-policy.json"),
                ("config/governance/checkpoint-signers.json", "config/governance/checkpoint-signers.json"),
                ("config/modules/milestone-1.json", "config/modules/milestone-1.json"),
                ("config/identity/bootstrap.json", "config/identity/bootstrap.json"),
            ):
                target_path = root / target
                target_path.parent.mkdir(parents=True, exist_ok=True)
                target_path.write_text((PROJECT_ROOT / source).read_text(encoding="utf-8"), encoding="utf-8")

            identity_path = root / "config" / "identity" / "bootstrap.json"
            identity_payload = json.loads(identity_path.read_text(encoding="utf-8"))
            identity_payload["role_bindings"].append(identity_payload["role_bindings"][0])
            identity_path.write_text(json.dumps(identity_payload), encoding="utf-8")

            env = {"PYTHONPATH": PYTHONPATH}
            result = subprocess.run(
                [sys.executable, "-m", "assurancectl.cli", "--root", str(root), "validate"],
                cwd=PROJECT_ROOT,
                env=env,
                text=True,
                capture_output=True,
                check=False,
            )

            self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
            self.assertIn("duplicate identity role binding", result.stderr)

    def test_validate_fails_when_checkpoint_signer_actor_binding_is_missing(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            root = Path(tmpdir)
            for source, target in (
                ("config/network-topology.json", "config/network-topology.json"),
                ("config/genesis/base-genesis.json", "config/genesis/base-genesis.json"),
                ("config/policy/release-guards.json", "config/policy/release-guards.json"),
                ("config/governance/inference-policy.json", "config/governance/inference-policy.json"),
                ("config/governance/checkpoint-signers.json", "config/governance/checkpoint-signers.json"),
                ("config/modules/milestone-1.json", "config/modules/milestone-1.json"),
                ("config/identity/bootstrap.json", "config/identity/bootstrap.json"),
            ):
                target_path = root / target
                target_path.parent.mkdir(parents=True, exist_ok=True)
                target_path.write_text((PROJECT_ROOT / source).read_text(encoding="utf-8"), encoding="utf-8")

            identity_path = root / "config" / "identity" / "bootstrap.json"
            identity_payload = json.loads(identity_path.read_text(encoding="utf-8"))
            identity_payload["role_bindings"] = [
                binding
                for binding in identity_payload["role_bindings"]
                if binding["role"] != "treasury-program-manager"
            ]
            identity_path.write_text(json.dumps(identity_payload), encoding="utf-8")

            env = {"PYTHONPATH": PYTHONPATH}
            result = subprocess.run(
                [sys.executable, "-m", "assurancectl.cli", "--root", str(root), "validate"],
                cwd=PROJECT_ROOT,
                env=env,
                text=True,
                capture_output=True,
                check=False,
            )

            self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
            self.assertIn("identity bootstrap missing active bindings for roles: treasury-program-manager", result.stderr)

    def test_validate_fails_when_checkpoint_signer_actor_ownership_is_duplicated(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            root = Path(tmpdir)
            for source, target in (
                ("config/network-topology.json", "config/network-topology.json"),
                ("config/genesis/base-genesis.json", "config/genesis/base-genesis.json"),
                ("config/policy/release-guards.json", "config/policy/release-guards.json"),
                ("config/governance/inference-policy.json", "config/governance/inference-policy.json"),
                ("config/governance/checkpoint-signers.json", "config/governance/checkpoint-signers.json"),
                ("config/modules/milestone-1.json", "config/modules/milestone-1.json"),
                ("config/identity/bootstrap.json", "config/identity/bootstrap.json"),
            ):
                target_path = root / target
                target_path.parent.mkdir(parents=True, exist_ok=True)
                target_path.write_text((PROJECT_ROOT / source).read_text(encoding="utf-8"), encoding="utf-8")

            signer_path = root / "config" / "governance" / "checkpoint-signers.json"
            signer_payload = json.loads(signer_path.read_text(encoding="utf-8"))
            signer_payload["signers"][1]["actor_id"] = signer_payload["signers"][0]["actor_id"]
            signer_path.write_text(json.dumps(signer_payload), encoding="utf-8")

            env = {"PYTHONPATH": PYTHONPATH}
            result = subprocess.run(
                [sys.executable, "-m", "assurancectl.cli", "--root", str(root), "validate"],
                cwd=PROJECT_ROOT,
                env=env,
                text=True,
                capture_output=True,
                check=False,
            )

            self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
            self.assertIn("duplicate checkpoint signer actor ownership", result.stderr)

    def test_validate_fails_when_checkpoint_signer_rotation_is_stale(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            root = Path(tmpdir)
            for source, target in (
                ("config/network-topology.json", "config/network-topology.json"),
                ("config/genesis/base-genesis.json", "config/genesis/base-genesis.json"),
                ("config/policy/release-guards.json", "config/policy/release-guards.json"),
                ("config/governance/inference-policy.json", "config/governance/inference-policy.json"),
                ("config/governance/checkpoint-signers.json", "config/governance/checkpoint-signers.json"),
                ("config/modules/milestone-1.json", "config/modules/milestone-1.json"),
                ("config/identity/bootstrap.json", "config/identity/bootstrap.json"),
            ):
                target_path = root / target
                target_path.parent.mkdir(parents=True, exist_ok=True)
                target_path.write_text((PROJECT_ROOT / source).read_text(encoding="utf-8"), encoding="utf-8")

            signer_path = root / "config" / "governance" / "checkpoint-signers.json"
            signer_payload = json.loads(signer_path.read_text(encoding="utf-8"))
            signer_payload["signers"][0]["rotate_by"] = "2026-04-01T00:00:00Z"
            signer_path.write_text(json.dumps(signer_payload), encoding="utf-8")

            env = {"PYTHONPATH": PYTHONPATH}
            result = subprocess.run(
                [sys.executable, "-m", "assurancectl.cli", "--root", str(root), "validate"],
                cwd=PROJECT_ROOT,
                env=env,
                text=True,
                capture_output=True,
                check=False,
            )

            self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
            self.assertIn("stale rotation metadata", result.stderr)

    def test_validate_fails_when_checkpoint_signer_coverage_is_incomplete(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            root = Path(tmpdir)
            for source, target in (
                ("config/network-topology.json", "config/network-topology.json"),
                ("config/genesis/base-genesis.json", "config/genesis/base-genesis.json"),
                ("config/policy/release-guards.json", "config/policy/release-guards.json"),
                ("config/governance/inference-policy.json", "config/governance/inference-policy.json"),
                ("config/governance/checkpoint-signers.json", "config/governance/checkpoint-signers.json"),
                ("config/modules/milestone-1.json", "config/modules/milestone-1.json"),
                ("config/identity/bootstrap.json", "config/identity/bootstrap.json"),
            ):
                target_path = root / target
                target_path.parent.mkdir(parents=True, exist_ok=True)
                target_path.write_text((PROJECT_ROOT / source).read_text(encoding="utf-8"), encoding="utf-8")

            signer_path = root / "config" / "governance" / "checkpoint-signers.json"
            signer_payload = json.loads(signer_path.read_text(encoding="utf-8"))
            signer_payload["signers"] = [
                signer
                for signer in signer_payload["signers"]
                if signer["signer_id"] != "treasury-review-chair-bot"
            ]
            signer_path.write_text(json.dumps(signer_payload), encoding="utf-8")

            env = {"PYTHONPATH": PYTHONPATH}
            result = subprocess.run(
                [sys.executable, "-m", "assurancectl.cli", "--root", str(root), "validate"],
                cwd=PROJECT_ROOT,
                env=env,
                text=True,
                capture_output=True,
                check=False,
            )

            self.assertEqual(result.returncode, 1, result.stdout + result.stderr)
            self.assertIn("checkpoint signer coverage missing roles: treasury-review-chair", result.stderr)

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

    def test_governance_queue_writes_versioned_artifact_file(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            artifact_path = Path(tmpdir) / "queue-artifact.json"
            result = self.run_cli(
                "governance-queue",
                "--registry",
                "examples/proposals/registry.json",
                "--history",
                "examples/proposals/history.json",
                "--artifact-out",
                str(artifact_path),
                "--json",
            )
            artifact = json.loads(artifact_path.read_text(encoding="utf-8"))

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload[0]["proposal_id"], "draft-pause-001")
        self.assertEqual(artifact["schema"], "0ai.assurance.governance.artifact")
        self.assertEqual(artifact["schema_version"], "1.0.0")
        self.assertEqual(artifact["artifact_type"], "governance_queue")
        self.assertEqual(artifact["command"], "governance-queue")
        self.assertEqual(artifact["sources"]["registry"], "examples/proposals/registry.json")
        self.assertEqual(artifact["sources"]["history"], "examples/proposals/history.json")
        self.assertEqual(artifact["payload"][0]["proposal_id"], "draft-pause-001")

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

    def test_governance_trends_writes_versioned_artifact_file(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            artifact_path = Path(tmpdir) / "trends-artifact.json"
            result = self.run_cli(
                "governance-trends",
                "--registry",
                "examples/proposals/registry.json",
                "--history",
                "examples/proposals/history.json",
                "--artifact-out",
                str(artifact_path),
                "--json",
            )
            artifact = json.loads(artifact_path.read_text(encoding="utf-8"))

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload[0]["trend_cluster"], "emergency_pause:safety-council")
        self.assertEqual(artifact["artifact_type"], "governance_trends")
        self.assertEqual(artifact["payload"][0]["trend_cluster"], "emergency_pause:safety-council")

    def test_governance_sim_with_history_writes_versioned_artifact_file(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            artifact_path = Path(tmpdir) / "sim-artifact.json"
            result = self.run_cli(
                "governance-sim",
                "--proposal",
                "examples/proposals/treasury-grant.json",
                "--history",
                "examples/proposals/history.json",
                "--artifact-out",
                str(artifact_path),
                "--json",
            )
            artifact = json.loads(artifact_path.read_text(encoding="utf-8"))

        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertIn("report", payload)
        self.assertEqual(artifact["artifact_type"], "governance_simulation_with_drift")
        self.assertEqual(artifact["payload"]["report"]["proposal_id"], "draft-grant-001")
        self.assertEqual(artifact["payload"]["drift"]["drift_attention"], "review")

    def test_governance_drift_writes_versioned_artifact_file(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            artifact_path = Path(tmpdir) / "drift-artifact.json"
            result = self.run_cli(
                "governance-drift",
                "--proposal",
                "examples/proposals/emergency-pause.json",
                "--history",
                "examples/proposals/history.json",
                "--artifact-out",
                str(artifact_path),
                "--json",
            )
            artifact = json.loads(artifact_path.read_text(encoding="utf-8"))

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["drift_attention"], "escalate")
        self.assertEqual(artifact["artifact_type"], "governance_drift")
        self.assertEqual(artifact["payload"]["drift_attention"], "escalate")

    def test_governance_replay_writes_versioned_artifact_file(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            event_path = Path(tmpdir) / "events.json"
            event_path.write_text(
                (PROJECT_ROOT / "examples" / "proposals" / "checkpoint-events.json").read_text(encoding="utf-8"),
                encoding="utf-8",
            )
            artifact_path = Path(tmpdir) / "replay-artifact.json"
            result = self.run_cli(
                "governance-replay",
                "--status",
                str(event_path),
                "--artifact-out",
                str(artifact_path),
                "--json",
            )
            artifact = json.loads(artifact_path.read_text(encoding="utf-8"))

        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["source_kind"], "event_log")
        self.assertEqual(artifact["artifact_type"], "governance_replay")
        self.assertEqual(artifact["sources"]["status"], str(event_path))
        self.assertEqual(artifact["payload"]["source_kind"], "event_log")

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

    def test_governance_remediation_writes_versioned_artifact_file(self) -> None:
        with tempfile.TemporaryDirectory() as tmpdir:
            artifact_path = Path(tmpdir) / "remediation-artifact.json"
            result = self.run_cli(
                "governance-remediation",
                "--registry",
                "examples/proposals/registry.json",
                "--history",
                "examples/proposals/history.json",
                "--artifact-out",
                str(artifact_path),
                "--json",
            )
            artifact = json.loads(artifact_path.read_text(encoding="utf-8"))

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload[0]["trend_cluster"], "emergency_pause:safety-council")
        self.assertEqual(artifact["artifact_type"], "governance_remediation")
        self.assertEqual(artifact["payload"][0]["trend_cluster"], "emergency_pause:safety-council")
        self.assertEqual(artifact["compatibility"]["breaking_change"], "increment major")

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
                    "actor_id": (
                        "op-treasury-program-manager-1"
                        if checkpoint["phase"] == "immediate_action"
                        else (
                            "op-treasury-review-chair-1"
                            if checkpoint["phase"] == "approval_guardrail"
                            else (
                                "op-finance-telemetry-lead-1"
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
        immediate_action = next(
            checkpoint
            for checkpoint in treasury_updated["checkpoints"]
            if checkpoint["phase"] == "immediate_action"
        )
        self.assertTrue(immediate_action["eligible_actors"])
        self.assertEqual(immediate_action["eligible_actors"][0]["actor_id"], "op-treasury-program-manager-1")
        monitoring_checkpoint = next(
            checkpoint
            for checkpoint in treasury_updated["checkpoints"]
            if checkpoint["checkpoint_id"] == "treasury-grant-0ai-core-monitoring-1"
        )
        self.assertEqual(monitoring_checkpoint["actor_id"], "op-finance-telemetry-lead-1")
        self.assertEqual(monitoring_checkpoint["assigned_actor"]["display_name"], "Finance Telemetry Lead 1")
        self.assertTrue(
            all(
                checkpoint["ready_to_start"]
                for checkpoint in treasury_updated["checkpoints"]
                if checkpoint["phase"] == "monitoring"
            )
        )

    def test_governance_replay_reconstructs_checkpoint_state_from_event_log(self) -> None:
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
        event_payload = self.build_treasury_event_log(treasury["checkpoints"])

        with tempfile.TemporaryDirectory() as tmpdir:
            event_path = Path(tmpdir) / "events.json"
            event_path.write_text(json.dumps(event_payload), encoding="utf-8")
            result = self.run_cli("governance-replay", "--status", str(event_path), "--json")

        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["source_kind"], "event_log")
        self.assertEqual(payload["invalid_event_count"], 0)
        self.assertGreater(payload["replay_event_count"], 0)
        monitoring = next(
            checkpoint
            for checkpoint in payload["checkpoints"]
            if checkpoint["checkpoint_id"] == "treasury-grant-0ai-core-monitoring-1"
        )
        self.assertEqual(monitoring["status"], "in_progress")
        self.assertEqual(monitoring["recorded_by"], "finance-telemetry-lead")
        self.assertEqual(monitoring["actor_id"], "op-finance-telemetry-lead-1")
        self.assertEqual(monitoring["actor"]["display_name"], "Finance Telemetry Lead 1")

    def test_governance_replay_rejects_duplicate_events(self) -> None:
        duplicate_payload = {
            "version": "checkpoint-event-log-duplicate-test",
            "events": [
                self.build_signed_event(
                    checkpoint_id="treasury-grant-0ai-core-immediate_action-1",
                    previous_status="pending",
                    new_status="in_progress",
                    updated_at="2026-04-16T11:01:00Z",
                    recorded_by="treasury-program-manager",
                    rationale="Started milestone release planning.",
                    signature_id="duplicate-1",
                ),
                self.build_signed_event(
                    checkpoint_id="treasury-grant-0ai-core-immediate_action-1",
                    previous_status="pending",
                    new_status="in_progress",
                    updated_at="2026-04-16T11:02:00Z",
                    recorded_by="treasury-program-manager",
                    rationale="Duplicate write should be rejected.",
                    signature_id="duplicate-2",
                ),
            ],
        }

        with tempfile.TemporaryDirectory() as tmpdir:
            event_path = Path(tmpdir) / "events-duplicate.json"
            event_path.write_text(json.dumps(duplicate_payload), encoding="utf-8")
            result = self.run_cli("governance-replay", "--status", str(event_path), "--json")

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["source_kind"], "event_log")
        self.assertEqual(payload["invalid_event_count"], 1)
        self.assertTrue(payload["event_alerts"])
        self.assertIn("contradictory event history", payload["event_alerts"][0])

    def test_governance_replay_rejects_illegal_lifecycle_transition(self) -> None:
        invalid_transition_payload = {
            "version": "checkpoint-event-log-illegal-transition-test",
            "events": [
                self.build_signed_event(
                    checkpoint_id="treasury-grant-0ai-core-immediate_action-1",
                    previous_status="pending",
                    new_status="completed",
                    updated_at="2026-04-16T11:01:00Z",
                    recorded_by="treasury-program-manager",
                    rationale="Skipping in-progress should be rejected.",
                    signature_id="illegal-transition-1",
                )
            ],
        }

        with tempfile.TemporaryDirectory() as tmpdir:
            event_path = Path(tmpdir) / "events-illegal-transition.json"
            event_path.write_text(json.dumps(invalid_transition_payload), encoding="utf-8")
            result = self.run_cli("governance-replay", "--status", str(event_path), "--json")

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["source_kind"], "event_log")
        self.assertEqual(payload["invalid_event_count"], 1)
        self.assertIn("illegal lifecycle transition pending -> completed", payload["event_alerts"][0])

    def test_governance_replay_rejects_invalid_signature(self) -> None:
        signed_event = self.build_signed_event(
            checkpoint_id="treasury-grant-0ai-core-immediate_action-1",
            previous_status="pending",
            new_status="in_progress",
            updated_at="2026-04-16T11:01:00Z",
            recorded_by="treasury-program-manager",
            rationale="Started milestone release planning.",
            signature_id="invalid-signature-1",
        )
        signature = signed_event["signature"]
        assert isinstance(signature, dict)
        signature["value"] = "0" * 64
        invalid_signature_payload = {
            "version": "checkpoint-event-log-invalid-signature-test",
            "events": [signed_event],
        }

        with tempfile.TemporaryDirectory() as tmpdir:
            event_path = Path(tmpdir) / "events-invalid-signature.json"
            event_path.write_text(json.dumps(invalid_signature_payload), encoding="utf-8")
            result = self.run_cli("governance-replay", "--status", str(event_path), "--json")

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["source_kind"], "event_log")
        self.assertEqual(payload["invalid_event_count"], 1)
        self.assertIn("invalid signature", payload["event_alerts"][0])

    def test_governance_replay_rejects_wrong_role_signer(self) -> None:
        wrong_role_payload = {
            "version": "checkpoint-event-log-wrong-role-test",
            "events": [
                self.build_signed_event(
                    checkpoint_id="treasury-grant-0ai-core-immediate_action-1",
                    previous_status="pending",
                    new_status="in_progress",
                    updated_at="2026-04-16T11:01:00Z",
                    recorded_by="treasury-program-manager",
                    rationale="Started milestone release planning.",
                    signature_id="wrong-role-1",
                    signer_id="treasury-review-chair-bot",
                )
            ],
        }

        with tempfile.TemporaryDirectory() as tmpdir:
            event_path = Path(tmpdir) / "events-wrong-role.json"
            event_path.write_text(json.dumps(wrong_role_payload), encoding="utf-8")
            result = self.run_cli("governance-replay", "--status", str(event_path), "--json")

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["source_kind"], "event_log")
        self.assertEqual(payload["invalid_event_count"], 1)
        self.assertIn("is not authorized for role treasury-program-manager", payload["event_alerts"][0])

    def test_governance_replay_rejects_signer_actor_mismatch(self) -> None:
        mismatched_actor_payload = {
            "version": "checkpoint-event-log-actor-mismatch-test",
            "events": [
                self.build_signed_event(
                    checkpoint_id="treasury-grant-0ai-core-immediate_action-1",
                    previous_status="pending",
                    new_status="in_progress",
                    updated_at="2026-04-16T11:01:00Z",
                    recorded_by="treasury-program-manager",
                    actor_id="op-treasury-review-chair-1",
                    rationale="Actor mismatch should be rejected.",
                    signature_id="actor-mismatch-1",
                )
            ],
        }

        with tempfile.TemporaryDirectory() as tmpdir:
            event_path = Path(tmpdir) / "events-actor-mismatch.json"
            event_path.write_text(json.dumps(mismatched_actor_payload), encoding="utf-8")
            result = self.run_cli("governance-replay", "--status", str(event_path), "--json")

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["invalid_event_count"], 1)
        self.assertIn("does not match event actor_id", payload["event_alerts"][0])

    def test_governance_replay_rejects_signature_replay_attempt(self) -> None:
        replayed_signature_payload = {
            "version": "checkpoint-event-log-replay-test",
            "events": [
                self.build_signed_event(
                    checkpoint_id="treasury-grant-0ai-core-immediate_action-1",
                    previous_status="pending",
                    new_status="in_progress",
                    updated_at="2026-04-16T11:01:00Z",
                    recorded_by="treasury-program-manager",
                    rationale="Started milestone release planning.",
                    signature_id="replay-attempt-1",
                ),
                self.build_signed_event(
                    checkpoint_id="treasury-grant-0ai-core-immediate_action-1",
                    previous_status="in_progress",
                    new_status="completed",
                    updated_at="2026-04-16T12:01:00Z",
                    recorded_by="treasury-program-manager",
                    rationale="Milestone release plan approved.",
                    signature_id="replay-attempt-1",
                ),
            ],
        }

        with tempfile.TemporaryDirectory() as tmpdir:
            event_path = Path(tmpdir) / "events-replay-attempt.json"
            event_path.write_text(json.dumps(replayed_signature_payload), encoding="utf-8")
            result = self.run_cli("governance-replay", "--status", str(event_path), "--json")

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["source_kind"], "event_log")
        self.assertEqual(payload["invalid_event_count"], 1)
        self.assertIn("replay attempt detected", payload["event_alerts"][0])

    def test_governance_remediation_rejects_checkpoint_owner_role_mismatch(self) -> None:
        wrong_owner_payload = {
            "version": "checkpoint-event-log-owner-mismatch-test",
            "events": [
                self.build_signed_event(
                    checkpoint_id="treasury-grant-0ai-core-immediate_action-1",
                    previous_status="pending",
                    new_status="in_progress",
                    updated_at="2026-04-16T11:01:00Z",
                    recorded_by="treasury-review-chair",
                    rationale="Wrong phase owner should be rejected during remediation.",
                    signature_id="owner-mismatch-1",
                    signer_id="treasury-review-chair-bot",
                )
            ],
        }

        with tempfile.TemporaryDirectory() as tmpdir:
            event_path = Path(tmpdir) / "events-owner-mismatch.json"
            event_path.write_text(json.dumps(wrong_owner_payload), encoding="utf-8")
            result = self.run_cli(
                "governance-remediation",
                "--registry",
                "examples/proposals/registry.json",
                "--history",
                "examples/proposals/history.json",
                "--status",
                str(event_path),
                "--json",
            )

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        treasury_plan = next(item for item in payload if item["trend_cluster"] == "treasury_grant:0ai-core")
        self.assertEqual(treasury_plan["current_release_readiness"], "invalid")
        self.assertGreater(treasury_plan["invalid_audit_count"], 0)
        self.assertTrue(
            any("Checkpoint actor role mismatch" in alert for alert in treasury_plan["audit_alerts"])
        )

    def test_governance_replay_rejects_invalid_event_log_schema(self) -> None:
        invalid_schema_payload = {
            "version": "checkpoint-event-log-invalid-schema-test",
            "events": {"unexpected": "object"},
        }

        with tempfile.TemporaryDirectory() as tmpdir:
            event_path = Path(tmpdir) / "events-invalid-schema.json"
            event_path.write_text(json.dumps(invalid_schema_payload), encoding="utf-8")
            result = self.run_cli("governance-replay", "--status", str(event_path), "--json")

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        self.assertEqual(payload["source_kind"], "event_log")
        self.assertEqual(payload["invalid_event_count"], 1)
        self.assertEqual(payload["event_alerts"], ["Invalid event log schema: 'events' must be a list."])

    def test_governance_remediation_event_log_updates_current_readiness(self) -> None:
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
        event_payload = self.build_treasury_event_log(treasury["checkpoints"])

        with tempfile.TemporaryDirectory() as tmpdir:
            event_path = Path(tmpdir) / "events-remediation.json"
            event_path.write_text(json.dumps(event_payload), encoding="utf-8")
            result = self.run_cli(
                "governance-remediation",
                "--registry",
                "examples/proposals/registry.json",
                "--history",
                "examples/proposals/history.json",
                "--status",
                str(event_path),
                "--json",
            )

        self.assertEqual(result.returncode, 2, result.stdout + result.stderr)
        payload = json.loads(result.stdout)
        treasury_updated = next(item for item in payload if item["trend_cluster"] == "treasury_grant:0ai-core")
        self.assertEqual(treasury_updated["current_release_readiness"], "monitoring")
        self.assertEqual(treasury_updated["invalid_event_count"], 0)
        self.assertEqual(treasury_updated["invalid_transition_count"], 0)
        self.assertEqual(treasury_updated["invalid_audit_count"], 0)
        self.assertGreater(treasury_updated["replay_event_count"], 0)

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
                    "actor_id": "op-treasury-program-manager-1",
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
                        "actor_id": "op-treasury-program-manager-1",
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
                        "actor_id": "op-treasury-review-chair-1",
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
