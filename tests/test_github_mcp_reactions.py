"""Unit tests for GitHub MCP connector reaction handlers (Defect #34).

These tests verify that:

1. ``get_issue_comment_reactions`` calls the correct endpoint
   ``/repos/{owner}/{repo}/issues/comments/{comment_id}/reactions``, not the
   wrong issue-level endpoint ``/repos/{owner}/{repo}/issues/{number}/reactions``
   which silently returns an empty list.

2. The normaliser produces the same output shape for issue-comment reactions
   and PR reactions, confirming both paths are consistent.

3. Pagination works correctly, collecting all pages.
"""

from __future__ import annotations

import sys
import unittest
from pathlib import Path
from typing import Any
from unittest.mock import MagicMock

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from assurancectl.github_mcp import (  # noqa: E402
    _ISSUE_COMMENT_REACTIONS_PATH,
    _ISSUE_REACTIONS_PATH,
    _PR_REACTIONS_PATH,
    _normalize_reaction,
    get_issue_comment_reactions,
    get_issue_reactions,
    get_pr_reactions,
)


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _raw_reaction(
    reaction_id: int = 1,
    content: str = "+1",
    login: str = "octocat",
    user_id: int = 1,
    created_at: str = "2026-04-17T12:00:00Z",
) -> dict[str, Any]:
    return {
        "id": reaction_id,
        "content": content,
        "created_at": created_at,
        "user": {"login": login, "id": user_id},
    }


def _mock_client(pages: list[list[dict[str, Any]]]) -> MagicMock:
    mock = MagicMock()
    mock.get.side_effect = pages
    return mock


# ---------------------------------------------------------------------------
# _normalize_reaction
# ---------------------------------------------------------------------------


class TestNormalizeReaction(unittest.TestCase):
    def test_basic_fields(self) -> None:
        raw = _raw_reaction(reaction_id=42, content="heart", login="alice", user_id=99)
        norm = _normalize_reaction(raw)
        self.assertEqual(norm["id"], 42)
        self.assertEqual(norm["content"], "heart")
        self.assertEqual(norm["created_at"], "2026-04-17T12:00:00Z")
        self.assertEqual(norm["user"]["login"], "alice")
        self.assertEqual(norm["user"]["id"], 99)

    def test_null_user_defaults_to_empty(self) -> None:
        raw = {"id": 1, "content": "+1", "user": None}
        norm = _normalize_reaction(raw)
        self.assertEqual(norm["user"]["login"], "")
        self.assertEqual(norm["user"]["id"], 0)

    def test_missing_user_defaults_to_empty(self) -> None:
        raw = {"id": 1, "content": "+1"}
        norm = _normalize_reaction(raw)
        self.assertEqual(norm["user"]["login"], "")
        self.assertEqual(norm["user"]["id"], 0)

    def test_missing_created_at_defaults_to_empty_string(self) -> None:
        raw = {"id": 1, "content": "+1", "user": {"login": "u", "id": 1}}
        norm = _normalize_reaction(raw)
        self.assertEqual(norm["created_at"], "")


# ---------------------------------------------------------------------------
# get_issue_comment_reactions  (Defect #34)
# ---------------------------------------------------------------------------


class TestGetIssueCommentReactions(unittest.TestCase):
    def test_calls_issue_comment_reactions_endpoint(self) -> None:
        client = _mock_client([[]])
        get_issue_comment_reactions(client, owner="acme", repo="widgets", comment_id=99)
        actual_path: str = client.get.call_args[0][0]
        expected_path = _ISSUE_COMMENT_REACTIONS_PATH.format(
            owner="acme", repo="widgets", comment_id=99
        )
        self.assertEqual(actual_path, expected_path)

    def test_does_not_call_issue_level_endpoint(self) -> None:
        client = _mock_client([[]])
        get_issue_comment_reactions(client, owner="acme", repo="widgets", comment_id=99)
        actual_path: str = client.get.call_args[0][0]
        wrong_path = _ISSUE_REACTIONS_PATH.format(
            owner="acme", repo="widgets", issue_number=99
        )
        self.assertNotEqual(actual_path, wrong_path)

    def test_returns_normalized_reactions(self) -> None:
        reaction = _raw_reaction(reaction_id=7, content="+1")
        client = _mock_client([[reaction], []])
        reactions = get_issue_comment_reactions(
            client, owner="acme", repo="widgets", comment_id=99
        )
        self.assertEqual(len(reactions), 1)
        self.assertEqual(reactions[0]["id"], 7)
        self.assertEqual(reactions[0]["content"], "+1")

    def test_empty_when_no_reactions(self) -> None:
        client = _mock_client([[]])
        reactions = get_issue_comment_reactions(
            client, owner="acme", repo="widgets", comment_id=1
        )
        self.assertEqual(reactions, [])

    def test_paginates_across_multiple_pages(self) -> None:
        page1 = [_raw_reaction(i, "+1") for i in range(1, 101)]
        page2 = [_raw_reaction(101, "heart")]
        client = _mock_client([page1, page2])
        reactions = get_issue_comment_reactions(
            client, owner="acme", repo="widgets", comment_id=5
        )
        self.assertEqual(len(reactions), 101)
        self.assertEqual(client.get.call_count, 2)

    def test_immediately_readable_after_creation(self) -> None:
        created_reaction = _raw_reaction(reaction_id=555, content="+1")
        client = _mock_client([[created_reaction], []])
        reactions = get_issue_comment_reactions(
            client, owner="owner", repo="repo", comment_id=42
        )
        self.assertGreater(len(reactions), 0)
        self.assertEqual(reactions[0]["id"], 555)
        self.assertEqual(reactions[0]["content"], "+1")


# ---------------------------------------------------------------------------
# get_pr_reactions
# ---------------------------------------------------------------------------


class TestGetPrReactions(unittest.TestCase):
    def test_calls_pr_reactions_endpoint(self) -> None:
        client = _mock_client([[]])
        get_pr_reactions(client, owner="acme", repo="widgets", pull_number=7)
        actual_path: str = client.get.call_args[0][0]
        expected_path = _PR_REACTIONS_PATH.format(
            owner="acme", repo="widgets", pull_number=7
        )
        self.assertEqual(actual_path, expected_path)

    def test_returns_normalized_reactions(self) -> None:
        reaction = _raw_reaction(reaction_id=3, content="hooray")
        client = _mock_client([[reaction], []])
        reactions = get_pr_reactions(client, owner="acme", repo="widgets", pull_number=7)
        self.assertEqual(len(reactions), 1)
        self.assertEqual(reactions[0]["content"], "hooray")

    def test_issue_comment_and_pr_reactions_share_normalizer(self) -> None:
        raw = _raw_reaction(reaction_id=1, content="+1")
        ic_client = _mock_client([[raw], []])
        pr_client = _mock_client([[raw], []])
        ic_reactions = get_issue_comment_reactions(
            ic_client, owner="o", repo="r", comment_id=10
        )
        pr_reactions = get_pr_reactions(pr_client, owner="o", repo="r", pull_number=10)
        self.assertEqual(ic_reactions, pr_reactions)


# ---------------------------------------------------------------------------
# get_issue_reactions
# ---------------------------------------------------------------------------


class TestGetIssueReactions(unittest.TestCase):
    def test_calls_issue_reactions_endpoint(self) -> None:
        client = _mock_client([[]])
        get_issue_reactions(client, owner="acme", repo="widgets", issue_number=3)
        actual_path: str = client.get.call_args[0][0]
        expected_path = _ISSUE_REACTIONS_PATH.format(
            owner="acme", repo="widgets", issue_number=3
        )
        self.assertEqual(actual_path, expected_path)

    def test_issue_and_comment_endpoints_are_distinct(self) -> None:
        issue_client = _mock_client([[]])
        comment_client = _mock_client([[]])
        get_issue_reactions(issue_client, owner="o", repo="r", issue_number=5)
        get_issue_comment_reactions(comment_client, owner="o", repo="r", comment_id=5)
        issue_path: str = issue_client.get.call_args[0][0]
        comment_path: str = comment_client.get.call_args[0][0]
        self.assertNotEqual(issue_path, comment_path)
        self.assertIn("/issues/5/reactions", issue_path)
        self.assertIn("/issues/comments/5/reactions", comment_path)


if __name__ == "__main__":
    unittest.main()
