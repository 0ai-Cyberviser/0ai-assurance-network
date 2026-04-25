"""Unit tests for github_mcp.pull_requests (Fix #36).

Verifies that:
- ``is_cross_repo_pr`` correctly identifies fork PRs vs same-repo PRs.
- ``build_update_payload`` suppresses ``maintainer_can_modify`` for same-repo
  PRs and forwards it unchanged for cross-repo (fork) PRs.
"""

from __future__ import annotations

import sys
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from github_mcp.pull_requests import build_update_payload, is_cross_repo_pr  # noqa: E402


def _make_pr(*, head_repo_id: int, base_repo_id: int) -> dict:
    """Return a minimal PR dict with the given repo IDs."""
    return {
        "number": 1,
        "head": {"repo": {"id": head_repo_id}},
        "base": {"repo": {"id": base_repo_id}},
    }


class TestIsCrossRepoPr(unittest.TestCase):
    """Unit tests for is_cross_repo_pr."""

    def test_same_repo_is_not_cross_repo(self) -> None:
        pr = _make_pr(head_repo_id=100, base_repo_id=100)
        self.assertFalse(is_cross_repo_pr(pr))

    def test_different_repo_ids_are_cross_repo(self) -> None:
        pr = _make_pr(head_repo_id=100, base_repo_id=200)
        self.assertTrue(is_cross_repo_pr(pr))

    def test_missing_head_repo_treated_as_same_repo(self) -> None:
        pr = {"number": 1, "head": {}, "base": {"repo": {"id": 100}}}
        self.assertFalse(is_cross_repo_pr(pr))

    def test_missing_base_repo_treated_as_same_repo(self) -> None:
        pr = {"number": 1, "head": {"repo": {"id": 100}}, "base": {}}
        self.assertFalse(is_cross_repo_pr(pr))

    def test_none_head_repo_treated_as_same_repo(self) -> None:
        pr = {"number": 1, "head": {"repo": None}, "base": {"repo": {"id": 100}}}
        self.assertFalse(is_cross_repo_pr(pr))

    def test_missing_keys_entirely_treated_as_same_repo(self) -> None:
        self.assertFalse(is_cross_repo_pr({}))


class TestBuildUpdatePayload(unittest.TestCase):
    """Unit tests for build_update_payload -- Fix #36: same-repo PR update guard."""

    # ------------------------------------------------------------------
    # Same-repo PRs: maintainer_can_modify must be suppressed
    # ------------------------------------------------------------------

    def test_same_repo_pr_omits_maintainer_can_modify_true(self) -> None:
        """maintainer_can_modify=True must be dropped for same-repo PRs."""
        pr = _make_pr(head_repo_id=100, base_repo_id=100)
        payload = build_update_payload(pr, {"title": "New title", "maintainer_can_modify": True})
        self.assertNotIn("maintainer_can_modify", payload)
        self.assertEqual(payload["title"], "New title")

    def test_same_repo_pr_omits_maintainer_can_modify_false(self) -> None:
        """maintainer_can_modify=False must also be dropped for same-repo PRs."""
        pr = _make_pr(head_repo_id=100, base_repo_id=100)
        payload = build_update_payload(pr, {"maintainer_can_modify": False})
        self.assertNotIn("maintainer_can_modify", payload)

    def test_same_repo_pr_forwards_other_fields(self) -> None:
        """Non-maintainer_can_modify fields are always forwarded."""
        pr = _make_pr(head_repo_id=100, base_repo_id=100)
        updates = {"title": "Updated title", "body": "New body", "state": "closed"}
        payload = build_update_payload(pr, updates)
        self.assertEqual(payload["title"], "Updated title")
        self.assertEqual(payload["body"], "New body")
        self.assertEqual(payload["state"], "closed")
        self.assertNotIn("maintainer_can_modify", payload)

    def test_same_repo_pr_no_maintainer_can_modify_in_updates(self) -> None:
        """When maintainer_can_modify is absent from updates, payload is unaffected."""
        pr = _make_pr(head_repo_id=100, base_repo_id=100)
        payload = build_update_payload(pr, {"title": "Title only"})
        self.assertEqual(payload, {"title": "Title only"})

    # ------------------------------------------------------------------
    # Cross-repo (fork) PRs: maintainer_can_modify must be forwarded
    # ------------------------------------------------------------------

    def test_cross_repo_pr_includes_maintainer_can_modify_true(self) -> None:
        """maintainer_can_modify=True must be forwarded for fork PRs."""
        pr = _make_pr(head_repo_id=100, base_repo_id=200)
        payload = build_update_payload(pr, {"title": "Fork PR", "maintainer_can_modify": True})
        self.assertIn("maintainer_can_modify", payload)
        self.assertTrue(payload["maintainer_can_modify"])

    def test_cross_repo_pr_includes_maintainer_can_modify_false(self) -> None:
        """maintainer_can_modify=False must be forwarded for fork PRs."""
        pr = _make_pr(head_repo_id=100, base_repo_id=200)
        payload = build_update_payload(pr, {"maintainer_can_modify": False})
        self.assertIn("maintainer_can_modify", payload)
        self.assertFalse(payload["maintainer_can_modify"])

    def test_cross_repo_pr_forwards_all_fields(self) -> None:
        """All fields including maintainer_can_modify are forwarded for fork PRs."""
        pr = _make_pr(head_repo_id=100, base_repo_id=200)
        updates = {"title": "Fork update", "body": "body", "maintainer_can_modify": True}
        payload = build_update_payload(pr, updates)
        self.assertEqual(payload["title"], "Fork update")
        self.assertEqual(payload["body"], "body")
        self.assertTrue(payload["maintainer_can_modify"])

    # ------------------------------------------------------------------
    # Edge cases: malformed / incomplete PR objects
    # ------------------------------------------------------------------

    def test_missing_repo_info_omits_maintainer_can_modify(self) -> None:
        """PRs with missing repo info are treated as same-repo: field suppressed."""
        pr = {"number": 1, "head": {}, "base": {}}
        payload = build_update_payload(pr, {"title": "Safe update", "maintainer_can_modify": True})
        self.assertNotIn("maintainer_can_modify", payload)
        self.assertEqual(payload["title"], "Safe update")

    def test_empty_updates_returns_empty_payload(self) -> None:
        pr = _make_pr(head_repo_id=100, base_repo_id=100)
        self.assertEqual(build_update_payload(pr, {}), {})


if __name__ == "__main__":
    unittest.main()
