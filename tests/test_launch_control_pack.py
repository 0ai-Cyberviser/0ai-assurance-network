from __future__ import annotations

import csv
import json
import re
import unittest
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
PACK_ROOT = REPO_ROOT / "docs" / "launch-control"
ISO_UTC_RE = re.compile(r"^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$")
ADDRESS_RE = re.compile(r"^0x[a-fA-F0-9]{40}$")


def _read_csv(path: Path) -> list[dict[str, str]]:
    with path.open(newline="", encoding="utf-8") as fh:
        reader = csv.DictReader(fh)
        return list(reader)


def _read_csv_header(path: Path) -> list[str]:
    with path.open(newline="", encoding="utf-8") as fh:
        reader = csv.reader(fh)
        return next(reader)


class LaunchControlPackTests(unittest.TestCase):
    def test_manifest_references_existing_files(self) -> None:
        manifest_path = PACK_ROOT / "manifest.json"
        manifest = json.loads(manifest_path.read_text(encoding="utf-8"))

        self.assertEqual(manifest["pack"], "launch-control-center")
        self.assertTrue(manifest["version"])
        self.assertTrue(manifest["repository_scope"])
        self.assertIn("artifacts", manifest)
        self.assertTrue(manifest["artifacts"])

        for artifact in manifest["artifacts"]:
            artifact_path = REPO_ROOT / artifact["path"]
            self.assertTrue(artifact_path.exists(), f"Missing artifact: {artifact['path']}")
            self.assertTrue(artifact.get("owner_role"), f"Missing owner_role: {artifact['path']}")

    def test_csv_headers_match_contract(self) -> None:
        expected_headers = {
            "allowlist.csv": ["wallet_address", "max_mint", "tier", "notes"],
            "vesting_beneficiaries.csv": [
                "beneficiary_address",
                "allocation_tokens",
                "start_utc",
                "cliff_days",
                "duration_days",
                "slice_days",
                "revocable",
                "label",
            ],
            "support_macros.csv": [
                "macro_id",
                "category",
                "title",
                "message_template",
                "escalation_path",
            ],
            "incident_log.csv": [
                "incident_id",
                "opened_utc",
                "severity",
                "status",
                "detected_by",
                "component",
                "symptom",
                "impact",
                "mitigation",
                "owner",
                "next_update_utc",
                "closed_utc",
                "rca_summary",
                "followup_ticket",
            ],
        }

        for filename, header in expected_headers.items():
            csv_path = PACK_ROOT / "templates" / filename
            self.assertTrue(csv_path.exists(), f"Missing CSV template: {filename}")
            self.assertEqual(_read_csv_header(csv_path), header)

    def test_allowlist_rows_have_valid_shape(self) -> None:
        rows = _read_csv(PACK_ROOT / "templates" / "allowlist.csv")
        self.assertTrue(rows, "allowlist.csv should include at least one sample row")
        for row in rows:
            self.assertIsNotNone(ADDRESS_RE.fullmatch(row["wallet_address"]))
            self.assertGreater(int(row["max_mint"]), 0)
            self.assertTrue(row["tier"])

    def test_vesting_rows_have_valid_shape(self) -> None:
        rows = _read_csv(PACK_ROOT / "templates" / "vesting_beneficiaries.csv")
        self.assertTrue(rows, "vesting_beneficiaries.csv should include at least one sample row")
        for row in rows:
            self.assertIsNotNone(ADDRESS_RE.fullmatch(row["beneficiary_address"]))
            self.assertGreater(int(row["allocation_tokens"]), 0)
            self.assertIsNotNone(ISO_UTC_RE.fullmatch(row["start_utc"]))
            self.assertGreaterEqual(int(row["cliff_days"]), 0)
            self.assertGreater(int(row["duration_days"]), 0)
            self.assertGreater(int(row["slice_days"]), 0)
            self.assertIn(row["revocable"], {"true", "false"})
            self.assertTrue(row["label"])

    def test_incident_rows_have_valid_shape(self) -> None:
        rows = _read_csv(PACK_ROOT / "templates" / "incident_log.csv")
        self.assertTrue(rows, "incident_log.csv should include at least one sample row")

        allowed_severity = {"critical", "high", "medium", "low"}
        allowed_status = {"open", "monitoring", "resolved"}

        for row in rows:
            self.assertTrue(row["incident_id"])
            self.assertIsNotNone(ISO_UTC_RE.fullmatch(row["opened_utc"]))
            self.assertIn(row["severity"], allowed_severity)
            self.assertIn(row["status"], allowed_status)
            if row["next_update_utc"]:
                self.assertIsNotNone(ISO_UTC_RE.fullmatch(row["next_update_utc"]))
            if row["closed_utc"]:
                self.assertIsNotNone(ISO_UTC_RE.fullmatch(row["closed_utc"]))

    def test_support_macros_have_required_fields(self) -> None:
        rows = _read_csv(PACK_ROOT / "templates" / "support_macros.csv")
        self.assertTrue(rows, "support_macros.csv should include at least one sample row")
        for row in rows:
            self.assertTrue(row["macro_id"].startswith("M"))
            self.assertTrue(row["category"])
            self.assertTrue(row["title"])
            self.assertTrue(row["message_template"])
            self.assertTrue(row["escalation_path"])


if __name__ == "__main__":
    unittest.main()
