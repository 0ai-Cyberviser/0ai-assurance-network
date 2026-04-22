"""Tests for github_mcp.pull_requests – same-repo PR update guard (Fix #36).

Unit tests for GitHub MCP connector reaction handlers (Defect #34).

These tests verify that:

1. ``get_issue_comment_reactions`` calls the correct endpoint
   ``/repos/{owner}/{repo}/issues/comments/{comment_id}/reactions``, not the
   wrong issue-level endpoint ``/repos/{owner}/{repo}/issues/{number}/reactions``
   which silently returns an empty list.

2. The normaliser produces the same output shape for issue-comment reactions
   and PR reactions, confirming both paths are consistent.

3. Pagination works correctly, collecting all pages.

Tests for the GitHub MCP connector.

Verifies that:
- ``GitHubAPIError`` and ``GitHubGraphQLError`` carry the expected attributes.
- ``_is_node_id`` correctly distinguishes numeric DB IDs from GraphQL node IDs.
- ``_github_request`` serialises the request correctly and raises
  ``GitHubAPIError`` on non-2xx responses.
- ``_graphql_request`` raises ``GitHubGraphQLError`` when the response
  contains errors.
- ``GitHubMCPConnector.add_review_to_pr`` maps the raw API response to the
  normalised dict shape.
- ``GitHubMCPConnector.dismiss_pull_request_review`` resolves numeric IDs
  via a REST GET before calling the GraphQL mutation.
- ``GitHubMCPConnector.dismiss_pull_request_review`` uses a node ID directly
  without an extra REST GET when given one.
"""

from __future__ import annotations

import sys
import unittest
from pathlib import Path
from typing import Any
from unittest.mock import MagicMock, patch

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from github_mcp.pull_requests import build_update_payload, is_cross_repo_pr  # noqa: E402
from assurancectl.github_mcp import (  # noqa: E402
    _ISSUE_COMMENT_REACTIONS_PATH,
    _ISSUE_REACTIONS_PATH,
    _PR_REACTIONS_PATH,
    _normalize_reaction,
    get_issue_comment_reactions,
    get_issue_reactions,
    get_pr_reactions,
)
import mcp_connectors.github as gh_mod  # noqa: E402
from mcp_connectors.github import (  # noqa: E402
    GitHubAPIError,
    GitHubGraphQLError,
    GitHubMCPConnector,
    _is_node_id,
)


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
    """Unit tests for build_update_payload – Fix #36: same-repo PR update guard."""

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
# Sample fixtures
# ---------------------------------------------------------------------------

NUMERIC_REVIEW_ID = 4125974959
NODE_REVIEW_ID = "PRR_kwDORtzkIs717WGv"

_SAMPLE_REST_REVIEW = {
    "id": NUMERIC_REVIEW_ID,
    "node_id": NODE_REVIEW_ID,
    "user": {"login": "octocat"},
    "body": "Looks good!",
    "state": "APPROVED",
    "html_url": "https://github.com/owner/repo/pull/1#pullrequestreview-4125974959",
    "pull_request_url": "https://api.github.com/repos/owner/repo/pulls/1",
    "submitted_at": "2026-04-17T10:00:00Z",
}

_SAMPLE_GRAPHQL_DISMISS_RESPONSE = {
    "data": {
        "dismissPullRequestReview": {
            "pullRequestReview": {
                "id": NODE_REVIEW_ID,
                "state": "DISMISSED",
            }
        }
    }
}

_GRAPHQL_DISMISS_DATA = {
    "dismissPullRequestReview": {
        "pullRequestReview": {
            "id": "PRR_kwDOAbc123",
            "state": "DISMISSED",
        }
    }
}


def _make_connector() -> GitHubMCPConnector:
    return GitHubMCPConnector("ghp_test_token_placeholder")


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


# ---------------------------------------------------------------------------
# Exception tests
# ---------------------------------------------------------------------------


class TestGitHubAPIError(unittest.TestCase):
    def test_message_in_args(self) -> None:
        err = GitHubAPIError(404, "Not Found")
        self.assertEqual(str(err), "Not Found")

    def test_status_code_attribute(self) -> None:
        err = GitHubAPIError(422, "Unprocessable Entity")
        self.assertEqual(err.status_code, 422)

    def test_is_exception(self) -> None:
        self.assertIsInstance(GitHubAPIError(500, "oops"), Exception)


class TestGitHubGraphQLError(unittest.TestCase):
    def test_single_error_message(self) -> None:
        err = GitHubGraphQLError([{"message": "field not found"}])
        self.assertIn("field not found", str(err))

    def test_multiple_errors_joined(self) -> None:
        err = GitHubGraphQLError([
            {"message": "first error"}, {"message": "second error"}
        ])
        self.assertIn("first error", str(err))
        self.assertIn("second error", str(err))

    def test_errors_attribute(self) -> None:
        errors = [{"message": "boom"}]
        err = GitHubGraphQLError(errors)
        self.assertIs(err.errors, errors)

    def test_missing_message_key(self) -> None:
        err = GitHubGraphQLError([{"type": "NOT_FOUND"}])
        self.assertEqual(str(err), "")


# ---------------------------------------------------------------------------
# _is_node_id tests
# ---------------------------------------------------------------------------


class TestIsNodeId(unittest.TestCase):
    def test_node_id_recognised(self) -> None:
        self.assertTrue(_is_node_id("PRR_kwDORtzkIs717WGv"))
        self.assertTrue(_is_node_id("PR_kwDOABC123"))
        self.assertTrue(_is_node_id("MDExOlB1bGxSZXF1ZXN0UmV2aWV3NDEyNTk3NDk1OQ=="))

    def test_numeric_id_not_recognised(self) -> None:
        self.assertFalse(_is_node_id("4125974959"))
        self.assertFalse(_is_node_id("123"))

    def test_empty_string_not_recognised(self) -> None:
        self.assertFalse(_is_node_id(""))

    def test_numeric_string_is_not_node_id(self) -> None:
        self.assertFalse(_is_node_id("12345678"))

    def test_alphanumeric_is_node_id(self) -> None:
        self.assertTrue(_is_node_id("PRR_kwDOA12345"))

    def test_plain_string_is_node_id(self) -> None:
        self.assertTrue(_is_node_id("abc"))

    def test_zero_is_not_node_id(self) -> None:
        self.assertFalse(_is_node_id("0"))


# ---------------------------------------------------------------------------
# _github_request tests
# ---------------------------------------------------------------------------


class TestGithubRequest(unittest.TestCase):
    def _make_mock_response(self, body: dict, status: int = 200):
        import json
        mock_resp = MagicMock()
        mock_resp.read.return_value = json.dumps(body).encode()
        mock_resp.__enter__ = lambda s: s
        mock_resp.__exit__ = MagicMock(return_value=False)
        return mock_resp

    def test_get_request_returns_parsed_json(self) -> None:
        response_body = {"id": 1, "state": "APPROVED"}
        mock_resp = self._make_mock_response(response_body)
        with patch("urllib.request.urlopen", return_value=mock_resp):
            result = gh_mod._github_request(
                "GET", "https://api.github.com/test", token="tok"
            )
        self.assertEqual(result, response_body)

    def test_post_with_json_data(self) -> None:
        response_body = {"ok": True}
        mock_resp = self._make_mock_response(response_body)
        with patch("urllib.request.urlopen", return_value=mock_resp) as mock_open:
            gh_mod._github_request(
                "POST",
                "https://api.github.com/test",
                token="tok",
                json_data={"event": "APPROVE"},
            )
        import json
        req_arg = mock_open.call_args[0][0]
        self.assertEqual(json.loads(req_arg.data.decode()), {"event": "APPROVE"})

    def test_raises_github_api_error_on_non_2xx(self) -> None:
        import urllib.error
        http_err = urllib.error.HTTPError(
            url="https://api.github.com/test",
            code=404,
            msg="Not Found",
            hdrs=None,  # type: ignore[arg-type]
            fp=None,
        )
        http_err.read = lambda: b'{"message": "Not Found"}'
        with patch("urllib.request.urlopen", side_effect=http_err):
            with self.assertRaises(GitHubAPIError) as ctx:
                gh_mod._github_request(
                    "GET", "https://api.github.com/test", token="tok"
                )
        self.assertEqual(ctx.exception.status_code, 404)

    def test_authorization_header_set(self) -> None:
        response_body = {}
        mock_resp = self._make_mock_response(response_body)
        with patch("urllib.request.urlopen", return_value=mock_resp) as mock_open:
            gh_mod._github_request(
                "GET", "https://api.github.com/test", token="mytoken"
            )
        req_arg = mock_open.call_args[0][0]
        self.assertIn("mytoken", req_arg.get_header("Authorization"))


# ---------------------------------------------------------------------------
# _graphql_request tests
# ---------------------------------------------------------------------------


class TestGraphqlRequest(unittest.TestCase):
    def _make_mock_response(self, body: dict):
        import json
        mock_resp = MagicMock()
        mock_resp.read.return_value = json.dumps(body).encode()
        mock_resp.__enter__ = lambda s: s
        mock_resp.__exit__ = MagicMock(return_value=False)
        return mock_resp

    def test_returns_data_field(self) -> None:
        mock_resp = self._make_mock_response({"data": {"someField": "value"}})
        with patch("urllib.request.urlopen", return_value=mock_resp):
            result = gh_mod._graphql_request("{query}", {}, token="tok")
        self.assertEqual(result, {"someField": "value"})

    def test_raises_graphql_error_on_errors_key(self) -> None:
        mock_resp = self._make_mock_response(
            {"errors": [{"message": "some graphql error"}]}
        )
        with patch("urllib.request.urlopen", return_value=mock_resp):
            with self.assertRaises(GitHubGraphQLError) as ctx:
                gh_mod._graphql_request("{query}", {}, token="tok")
        self.assertIn("some graphql error", str(ctx.exception))


# ---------------------------------------------------------------------------
# add_review_to_pr tests
# ---------------------------------------------------------------------------


class TestAddReviewToPr(unittest.TestCase):
    _RAW_REVIEW = {
        "id": 4125974959,
        "node_id": "PRR_kwDOAbc123",
        "state": "APPROVED",
        "user": {"login": "octocat"},
        "body": "LGTM",
        "submitted_at": "2026-04-17T10:00:00Z",
    }

    def _connector(self) -> GitHubMCPConnector:
        return GitHubMCPConnector(token="test-token")

    def test_returns_normalised_dict(self) -> None:
        with patch.object(gh_mod, "_github_request", return_value=self._RAW_REVIEW):
            result = self._connector().add_review_to_pr("owner", "repo", 42)
        self.assertEqual(result["review_id"], 4125974959)
        self.assertEqual(result["review_node_id"], "PRR_kwDOAbc123")
        self.assertEqual(result["state"], "APPROVED")
        self.assertEqual(result["user"], "octocat")
        self.assertEqual(result["body"], "LGTM")
        self.assertEqual(result["submitted_at"], "2026-04-17T10:00:00Z")

    def test_posts_to_correct_url(self) -> None:
        mock_fn = MagicMock(return_value=self._RAW_REVIEW)
        with patch.object(gh_mod, "_github_request", mock_fn):
            self._connector().add_review_to_pr("myorg", "myrepo", 7)
        url_arg = mock_fn.call_args[0][1]
        self.assertIn("/repos/myorg/myrepo/pulls/7/reviews", url_arg)

    def test_body_and_event_in_payload(self) -> None:
        mock_fn = MagicMock(return_value=self._RAW_REVIEW)
        with patch.object(gh_mod, "_github_request", mock_fn):
            self._connector().add_review_to_pr(
                "o", "r", 1, body="nice work", event="APPROVE"
            )
        payload = mock_fn.call_args[1]["json_data"]
        self.assertEqual(payload["body"], "nice work")
        self.assertEqual(payload["event"], "APPROVE")

    def test_empty_body_omitted_from_payload(self) -> None:
        mock_fn = MagicMock(return_value=self._RAW_REVIEW)
        with patch.object(gh_mod, "_github_request", mock_fn):
            self._connector().add_review_to_pr("o", "r", 1, body="")
        payload = mock_fn.call_args[1]["json_data"]
        self.assertNotIn("body", payload)

    def test_returns_review_id_integer(self) -> None:
        connector = _make_connector()
        with patch("mcp_connectors.github._github_request", return_value=_SAMPLE_REST_REVIEW):
            result = connector.add_review_to_pr(
                owner="owner", repo="repo", pull_number=1, body="Looks good!", event="APPROVE"
            )
        self.assertEqual(result["review_id"], NUMERIC_REVIEW_ID)

    def test_returns_review_node_id_string(self) -> None:
        connector = _make_connector()
        with patch("mcp_connectors.github._github_request", return_value=_SAMPLE_REST_REVIEW):
            result = connector.add_review_to_pr(
                owner="owner", repo="repo", pull_number=1, body="Looks good!", event="APPROVE"
            )
        self.assertEqual(result["review_node_id"], NODE_REVIEW_ID)
        self.assertTrue(_is_node_id(result["review_node_id"]))

    def test_returns_both_ids_simultaneously(self) -> None:
        connector = _make_connector()
        with patch("mcp_connectors.github._github_request", return_value=_SAMPLE_REST_REVIEW):
            result = connector.add_review_to_pr(owner="owner", repo="repo", pull_number=1)
        self.assertIn("review_id", result)
        self.assertIn("review_node_id", result)

    def test_normalises_other_fields(self) -> None:
        connector = _make_connector()
        with patch("mcp_connectors.github._github_request", return_value=_SAMPLE_REST_REVIEW):
            result = connector.add_review_to_pr(
                owner="owner", repo="repo", pull_number=1, body="Looks good!", event="APPROVE"
            )
        self.assertEqual(result["state"], _SAMPLE_REST_REVIEW["state"])
        self.assertEqual(result["user"], _SAMPLE_REST_REVIEW["user"]["login"])
        self.assertEqual(result["body"], _SAMPLE_REST_REVIEW["body"])
        self.assertEqual(result["submitted_at"], _SAMPLE_REST_REVIEW["submitted_at"])

    def test_api_error_propagates(self) -> None:
        connector = _make_connector()
        with patch(
            "mcp_connectors.github._github_request",
            side_effect=GitHubAPIError(422, '{"message":"Validation Failed"}'),
        ): 
            with self.assertRaises(GitHubAPIError) as ctx:
                connector.add_review_to_pr(owner="owner", repo="repo", pull_number=1)
        self.assertEqual(ctx.exception.status_code, 422)


# ---------------------------------------------------------------------------
# dismiss_pull_request_review tests
# ---------------------------------------------------------------------------


class TestDismissPullRequestReview(unittest.TestCase):
    def _connector(self) -> GitHubMCPConnector:
        return GitHubMCPConnector(token="test-token")

    def test_uses_node_id_directly(self) -> None:
        mock_rest = MagicMock()
        mock_gql = MagicMock(return_value=_GRAPHQL_DISMISS_DATA)
        with patch.object(gh_mod, "_github_request", mock_rest), \
             patch.object(gh_mod, "_graphql_request", mock_gql):
            result = self._connector().dismiss_pull_request_review(
                "owner", "repo", 1, review_id="PRR_kwDOAbc123", message="oops"
            )
        mock_rest.assert_not_called()
        self.assertEqual(result["state"], "DISMISSED")

    def test_resolves_numeric_id_via_rest(self) -> None:
        raw_review = {"node_id": "PRR_kwDOAbc123", "id": 99}
        mock_rest = MagicMock(return_value=raw_review)
        mock_gql = MagicMock(return_value=_GRAPHQL_DISMISS_DATA)
        with patch.object(gh_mod, "_github_request", mock_rest), \
             patch.object(gh_mod, "_graphql_request", mock_gql):
            result = self._connector().dismiss_pull_request_review(
                "owner", "repo", 42, review_id=99, message="dismiss"
            )
        mock_rest.assert_called_once()
        get_url = mock_rest.call_args[0][1]
        self.assertIn("/pulls/42/reviews/99", get_url)
        self.assertEqual(result["state"], "DISMISSED")

    def test_graphql_mutation_called_with_node_id(self) -> None:
        raw_review = {"node_id": "PRR_node", "id": 7}
        mock_rest = MagicMock(return_value=raw_review)
        mock_gql = MagicMock(return_value=_GRAPHQL_DISMISS_DATA)
        with patch.object(gh_mod, "_github_request", mock_rest), \
             patch.object(gh_mod, "_graphql_request", mock_gql):
            self._connector().dismiss_pull_request_review(
                "o", "r", 1, review_id=7, message="bye"
            )
        variables = mock_gql.call_args[0][1]
        self.assertEqual(variables["reviewId"], "PRR_node")
        self.assertEqual(variables["message"], "bye")

    def test_returns_state_key(self) -> None:
        mock_rest = MagicMock()
        mock_gql = MagicMock(return_value=_GRAPHQL_DISMISS_DATA)
        with patch.object(gh_mod, "_github_request", mock_rest), \
             patch.object(gh_mod, "_graphql_request", mock_gql):
            result = self._connector().dismiss_pull_request_review(
                "o", "r", 1, review_id="PRR_kwDOAbc123"
            )
        self.assertIn("state", result)

    def test_dismiss_with_node_id_succeeds(self) -> None:
        connector = _make_connector()
        dismiss_data = _SAMPLE_GRAPHQL_DISMISS_RESPONSE["data"]
        with patch("mcp_connectors.github._graphql_request", return_value=dismiss_data) as mock_gql, \
             patch("mcp_connectors.github._github_request") as mock_rest:
            result = connector.dismiss_pull_request_review(
                owner="owner", repo="repo", pull_number=1,
                review_id=NODE_REVIEW_ID, message="Dismissed via node ID."
            )
        mock_rest.assert_not_called()
        mock_gql.assert_called_once()
        args, kwargs = mock_gql.call_args
        self.assertEqual(args[1]["reviewId"], NODE_REVIEW_ID)
        self.assertEqual(result["state"], "DISMISSED")

    def test_dismiss_with_numeric_id_resolves_node_id(self) -> None:
        connector = _make_connector()
        dismiss_data = _SAMPLE_GRAPHQL_DISMISS_RESPONSE["data"]
        with patch("mcp_connectors.github._graphql_request", return_value=dismiss_data) as mock_gql, \
             patch("mcp_connectors.github._github_request", return_value=_SAMPLE_REST_REVIEW) as mock_rest:
            result = connector.dismiss_pull_request_review(
                owner="owner", repo="repo", pull_number=1,
                review_id=NUMERIC_REVIEW_ID, message="Dismissed via numeric ID."
            )
        mock_rest.assert_called_once()
        rest_args, _ = mock_rest.call_args
        self.assertEqual(rest_args[0], "GET")
        self.assertIn(str(NUMERIC_REVIEW_ID), rest_args[1])
        mock_gql.assert_called_once()
        gql_args, _ = mock_gql.call_args
        self.assertEqual(gql_args[1]["reviewId"], NODE_REVIEW_ID)
        self.assertEqual(result["state"], "DISMISSED")

    def test_dismiss_with_string_numeric_id_resolves_node_id(self) -> None:
        connector = _make_connector()
        dismiss_data = _SAMPLE_GRAPHQL_DISMISS_RESPONSE["data"]
        with patch("mcp_connectors.github._graphql_request", return_value=dismiss_data) as mock_gql, \
             patch("mcp_connectors.github._github_request", return_value=_SAMPLE_REST_REVIEW) as mock_rest:
            result = connector.dismiss_pull_request_review(
                owner="owner", repo="repo", pull_number=1,
                review_id=str(NUMERIC_REVIEW_ID), message="Dismissed via string numeric ID."
            )
        mock_rest.assert_called_once()
        rest_args, _ = mock_rest.call_args
        self.assertEqual(rest_args[0], "GET")
        self.assertIn(str(NUMERIC_REVIEW_ID), rest_args[1])
        mock_gql.assert_called_once()
        gql_args, _ = mock_gql.call_args
        self.assertEqual(gql_args[1]["reviewId"], NODE_REVIEW_ID)
        self.assertEqual(result["state"], "DISMISSED")

    def test_graphql_error_propagates(self) -> None:
        connector = _make_connector()
        with patch(
            "mcp_connectors.github._graphql_request",
            side_effect=GitHubGraphQLError([{"message": "Could not resolve to a node"}]),
        ):
            with self.assertRaises(GitHubGraphQLError) as ctx:
                connector.dismiss_pull_request_review(
                    owner="owner", repo="repo", pull_number=1,
                    review_id=NODE_REVIEW_ID, message="Should fail."
                )
        self.assertIn("Could not resolve to a node", str(ctx.exception))

    def test_rest_lookup_failure_propagates_on_numeric_id(self) -> None:
        connector = _make_connector()
        with patch(
            "mcp_connectors.github._github_request",
            side_effect=GitHubAPIError(404, '{"message":"Not Found"}'),
        ):
            with self.assertRaises(GitHubAPIError) as ctx:
                connector.dismiss_pull_request_review(
                    owner="owner", repo="repo", pull_number=1,
                    review_id=NUMERIC_REVIEW_ID, message="Should fail."
                )
        self.assertEqual(ctx.exception.status_code, 404)


# ---------------------------------------------------------------------------
# Integration-style round-trip test (fully mocked)
# ---------------------------------------------------------------------------


class TestReviewLifecycleRoundTrip(unittest.TestCase):
    def test_create_then_dismiss_using_node_id(self) -> None:
        connector = _make_connector()
        with patch("mcp_connectors.github._github_request", return_value=_SAMPLE_REST_REVIEW):
            created = connector.add_review_to_pr(
                owner="owner", repo="repo", pull_number=1, body="LGTM", event="APPROVE",
            )
        self.assertIn("review_node_id", created)
        node_id = created["review_node_id"]
        dismiss_data = {
            "dismissPullRequestReview": {
                "pullRequestReview": {"id": node_id, "state": "DISMISSED"}
            }
        }
        with patch("mcp_connectors.github._graphql_request", return_value=dismiss_data) as mock_gql, \
             patch("mcp_connectors.github._github_request") as mock_rest:
            result = connector.dismiss_pull_request_review(
                owner="owner", repo="repo", pull_number=1,
                review_id=node_id, message="No longer valid.",
            )
        mock_rest.assert_not_called()
        mock_gql.assert_called_once()
        self.assertEqual(result["state"], "DISMISSED")


if __name__ == "__main__":
    unittest.main()
