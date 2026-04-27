from __future__ import annotations

import json
import tempfile
import unittest
from pathlib import Path

from scripts.validate_peachtree_recursive_learning import main, validate_record


class PeachtreeRecursiveLearningValidatorTests(unittest.TestCase):
    def test_fixture_record_is_valid(self) -> None:
        fixture = Path("tests/fixtures/peachtree-recursive-learning/governance-proposal-cycle.json")
        errors = validate_record(fixture)
        self.assertEqual(errors, [])

    def test_secret_marker_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as tmp:
            record_path = Path(tmp) / "record.json"
            record = {
                "schema_version": "peachtree.recursive_learning.v1",
                "cycle_id": "cycle-001",
                "artifact_type": "governance_proposal",
                "artifact_path": "examples/proposals/emergency-pause.json",
                "artifact_sha256": "sha256:abc",
                "validation_commands": ["make validate"],
                "sandbox_required": True,
                "policy_decision": "human_review_required",
                "findings": [],
                "dataset_candidate": {
                    "dataset_action": "human_review",
                    "notes": "contains token=should-not-be-here",
                },
                "self_improvement": {
                    "recommendation": "review",
                    "next_cycle_seed_candidate": True,
                    "requires_human_approval": True,
                },
            }
            record_path.write_text(json.dumps(record), encoding="utf-8")
            errors = validate_record(record_path)

        self.assertTrue(errors)
        self.assertIn("blocked secret/token marker", errors[0])

    def test_directory_main_reports_success(self) -> None:
        fixture_dir = "tests/fixtures/peachtree-recursive-learning"
        exit_code = main([fixture_dir])
        self.assertEqual(exit_code, 0)


if __name__ == "__main__":
    unittest.main()
