from __future__ import annotations

import csv
import json
import re
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


def test_launch_control_manifest_references_existing_files() -> None:
    manifest_path = PACK_ROOT / "manifest.json"
    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))

    assert manifest["pack"] == "launch-control-center"
    assert manifest["version"]
    assert manifest["repository_scope"]
    assert "artifacts" in manifest and manifest["artifacts"]

    for artifact in manifest["artifacts"]:
        artifact_path = REPO_ROOT / artifact["path"]
        assert artifact_path.exists(), f"Missing artifact: {artifact['path']}"
        assert artifact.get("owner_role"), f"Missing owner_role: {artifact['path']}"


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


def test_allowlist_rows_have_valid_shape() -> None:
    rows = _read_csv(PACK_ROOT / "templates" / "allowlist.csv")
    assert rows, "allowlist.csv should include at least one sample row"
    for row in rows:
        assert ADDRESS_RE.fullmatch(row["wallet_address"]) is not None
        assert int(row["max_mint"]) > 0
        assert row["tier"]


def test_vesting_rows_have_valid_shape() -> None:
    rows = _read_csv(PACK_ROOT / "templates" / "vesting_beneficiaries.csv")
    assert rows, "vesting_beneficiaries.csv should include at least one sample row"
    for row in rows:
        assert ADDRESS_RE.fullmatch(row["beneficiary_address"]) is not None
        assert int(row["allocation_tokens"]) > 0
        assert ISO_UTC_RE.fullmatch(row["start_utc"]) is not None
        assert int(row["cliff_days"]) >= 0
        assert int(row["duration_days"]) > 0
        assert int(row["slice_days"]) > 0
        assert row["revocable"] in {"true", "false"}
        assert row["label"]


def test_incident_rows_have_valid_shape() -> None:
    rows = _read_csv(PACK_ROOT / "templates" / "incident_log.csv")
    assert rows, "incident_log.csv should include at least one sample row"

    allowed_severity = {"critical", "high", "medium", "low"}
    allowed_status = {"open", "monitoring", "resolved"}

    for row in rows:
        assert row["incident_id"]
        assert ISO_UTC_RE.fullmatch(row["opened_utc"]) is not None
        assert row["severity"] in allowed_severity
        assert row["status"] in allowed_status
        if row["next_update_utc"]:
            assert ISO_UTC_RE.fullmatch(row["next_update_utc"]) is not None
        if row["closed_utc"]:
            assert ISO_UTC_RE.fullmatch(row["closed_utc"]) is not None


def test_support_macros_have_required_fields() -> None:
    rows = _read_csv(PACK_ROOT / "templates" / "support_macros.csv")
    assert rows, "support_macros.csv should include at least one sample row"
    for row in rows:
        assert row["macro_id"].startswith("M")
        assert row["category"]
        assert row["title"]
        assert row["message_template"]
        assert row["escalation_path"]
