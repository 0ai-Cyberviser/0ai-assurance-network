"""Tests for the GitHub MCP connector helpers (src/assurancectl/github_connector.py).

Covers Defect #33: list_pull_request_reviews must return COMMENTED reviews.
"""

from __future__ import annotations

import sys
import unittest
from pathlib import Path

# Ensure the src tree is importable when tests are run directly.
sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from assurancectl.github_connector import (
    ALLOWED_REVIEW_STATES,
    list_pull_request_reviews,
    normalize_review,
)


def _make_review(
    *,
    review_id: int = 1,
    node_id: str = "PRR_node1",
    state: str = "COMMENTED",
    body: str = "lgtm",
    submitted_at: str = "2026-04-17T12:00:00Z",
    commit_id: str = "abc123",
    html_url: str = "https://github.com/owner/repo/pull/1#pullrequestreview-1",
    user: dict | None = None,
) -> dict:
    return {
        "id": review_id,
        "node_id": node_id,
        "state": state,
        "body": body,
        "submitted_at": submitted_at,
        "commit_id": commit_id,
        "html_url": html_url,
        "user": user or {"login": "reviewer"},
    }


class TestAllowedReviewStates(unittest.TestCase):
    """ALLOWED_REVIEW_STATES must cover every documented GitHub review state."""

    def test_commented_is_included(self) -> None:
        """Defect #33: COMMENTED must be in ALLOWED_REVIEW_STATES."""
        self.assertIn("COMMENTED", ALLOWED_REVIEW_STATES)

    def test_approved_is_included(self) -> None:
        self.assertIn("APPROVED", ALLOWED_REVIEW_STATES)

    def test_changes_requested_is_included(self) -> None:
        self.assertIn("CHANGES_REQUESTED", ALLOWED_REVIEW_STATES)

    def test_dismissed_is_included(self) -> None:
        self.assertIn("DISMISSED", ALLOWED_REVIEW_STATES)

    def test_pending_is_included(self) -> None:
        self.assertIn("PENDING", ALLOWED_REVIEW_STATES)


class TestNormalizeReview(unittest.TestCase):
    """normalize_review should preserve all allowed states."""

    def test_commented_review_is_preserved(self) -> None:
        """Defect #33: COMMENTED reviews must not be dropped by the normalizer."""
        raw = _make_review(state="COMMENTED")
        result = normalize_review(raw)
        self.assertIsNotNone(result)
        assert result is not None
        self.assertEqual(result["state"], "COMMENTED")
        self.assertEqual(result["review_id"], 1)

    def test_approved_review_is_preserved(self) -> None:
        raw = _make_review(state="APPROVED")
        result = normalize_review(raw)
        self.assertIsNotNone(result)
        assert result is not None
        self.assertEqual(result["state"], "APPROVED")

    def test_changes_requested_review_is_preserved(self) -> None:
        raw = _make_review(state="CHANGES_REQUESTED")
        result = normalize_review(raw)
        self.assertIsNotNone(result)
        assert result is not None
        self.assertEqual(result["state"], "CHANGES_REQUESTED")

    def test_dismissed_review_is_preserved(self) -> None:
        raw = _make_review(state="DISMISSED")
        result = normalize_review(raw)
        self.assertIsNotNone(result)
        assert result is not None
        self.assertEqual(result["state"], "DISMISSED")

    def test_pending_review_is_preserved(self) -> None:
        raw = _make_review(state="PENDING")
        result = normalize_review(raw)
        self.assertIsNotNone(result)
        assert result is not None
        self.assertEqual(result["state"], "PENDING")

    def test_unknown_state_is_dropped(self) -> None:
        raw = _make_review(state="UNKNOWN_STATE")
        self.assertIsNone(normalize_review(raw))

    def test_missing_id_is_dropped(self) -> None:
        raw = _make_review()
        del raw["id"]
        self.assertIsNone(normalize_review(raw))

    def test_non_dict_input_is_dropped(self) -> None:
        self.assertIsNone(normalize_review("not-a-dict"))  # type: ignore[arg-type]

    def test_state_is_normalised_to_upper(self) -> None:
        raw = _make_review(state="commented")
        result = normalize_review(raw)
        self.assertIsNotNone(result)
        assert result is not None
        self.assertEqual(result["state"], "COMMENTED")

    def test_fields_are_mapped_correctly(self) -> None:
        raw = _make_review(
            review_id=99,
            node_id="PRR_xyz",
            state="APPROVED",
            body="Ship it",
            submitted_at="2026-04-18T08:00:00Z",
            commit_id="deadbeef",
            html_url="https://github.com/owner/repo/pull/2#pullrequestreview-99",
            user={"login": "alice"},
        )
        result = normalize_review(raw)
        self.assertIsNotNone(result)
        assert result is not None
        self.assertEqual(result["review_id"], 99)
        self.assertEqual(result["review_node_id"], "PRR_xyz")
        self.assertEqual(result["state"], "APPROVED")
        self.assertEqual(result["body"], "Ship it")
        self.assertEqual(result["submitted_at"], "2026-04-18T08:00:00Z")
        self.assertEqual(result["commit_id"], "deadbeef")
        self.assertEqual(result["html_url"], "https://github.com/owner/repo/pull/2#pullrequestreview-99")
        self.assertEqual(result["user"], {"login": "alice"})


class TestListPullRequestReviews(unittest.TestCase):
    """list_pull_request_reviews must return every submitted review, including COMMENTED."""

    def test_commented_review_appears_in_listing(self) -> None:
        """Defect #33 regression: COMMENTED review must not be omitted from results."""
        raw_reviews = [_make_review(review_id=1, state="COMMENTED")]
        results = list_pull_request_reviews(raw_reviews)
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0]["state"], "COMMENTED")
        self.assertEqual(results[0]["review_id"], 1)

    def test_mixed_states_all_appear(self) -> None:
        raw_reviews = [
            _make_review(review_id=1, state="COMMENTED"),
            _make_review(review_id=2, state="APPROVED"),
            _make_review(review_id=3, state="CHANGES_REQUESTED"),
            _make_review(review_id=4, state="DISMISSED"),
            _make_review(review_id=5, state="PENDING"),
        ]
        results = list_pull_request_reviews(raw_reviews)
        self.assertEqual(len(results), 5)
        states = [r["state"] for r in results]
        self.assertIn("COMMENTED", states)
        self.assertIn("APPROVED", states)
        self.assertIn("CHANGES_REQUESTED", states)
        self.assertIn("DISMISSED", states)
        self.assertIn("PENDING", states)

    def test_empty_list_returns_empty(self) -> None:
        self.assertEqual(list_pull_request_reviews([]), [])

    def test_non_list_input_returns_empty(self) -> None:
        self.assertEqual(list_pull_request_reviews(None), [])  # type: ignore[arg-type]

    def test_invalid_entries_are_skipped(self) -> None:
        """Reviews with unknown states or missing IDs are dropped; valid ones kept."""
        raw_reviews = [
            _make_review(review_id=1, state="COMMENTED"),
            _make_review(review_id=2, state="UNKNOWN"),
            {"state": "APPROVED"},  # missing id
        ]
        results = list_pull_request_reviews(raw_reviews)
        self.assertEqual(len(results), 1)
        self.assertEqual(results[0]["review_id"], 1)

    def test_ordering_is_preserved(self) -> None:
        raw_reviews = [
            _make_review(review_id=10, state="APPROVED"),
            _make_review(review_id=20, state="COMMENTED"),
            _make_review(review_id=30, state="DISMISSED"),
        ]
        results = list_pull_request_reviews(raw_reviews)
        self.assertEqual([r["review_id"] for r in results], [10, 20, 30])

    def test_only_commented_review_on_pr_is_not_lost(self) -> None:
        """Regression: a PR with only a COMMENTED review must return that review."""
        raw_reviews = [
            _make_review(review_id=42, state="COMMENTED", body="Looks good, just a note."),
        ]
        results = list_pull_request_reviews(raw_reviews)
        self.assertNotEqual(results, [], "Expected COMMENTED review to be returned, got empty list")
        self.assertEqual(results[0]["review_id"], 42)
        self.assertEqual(results[0]["body"], "Looks good, just a note.")


if __name__ == "__main__":
    unittest.main()
