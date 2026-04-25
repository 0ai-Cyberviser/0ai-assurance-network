from __future__ import annotations

import csv
import json
from pathlib import Path


REPO_ROOT = Path(__file__).resolve().parents[1]
PACK_ROOT = REPO_ROOT / "docs" / "launch-control"


def _read_csv_header(path: Path) -> list[str]:
    with path.open(newline="", encoding="utf-8") as fh:
        reader = csv.reader(fh)
        return next(reader)


def test_launch_control_manifest_references_existing_files() -> None:
    manifest_path = PACK_ROOT / "manifest.json"
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))

    assert manifest["pack"] == "launch-control-center"
    assert "artifacts" in manifest and manifest["artifacts"]

    for artifact in manifest["artifacts"]:
        artifact_path = REPO_ROOT / artifact["path"]
        assert artifact_path.exists(), f"Missing artifact: {artifact['path']}"


def test_launch_control_csv_headers_match_contract() -> None:
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
        assert csv_path.exists(), f"Missing CSV template: {filename}"
        assert _read_csv_header(csv_path) == header
