"""Unit tests for the GitHub MCP connector – Defect #32: Review ID Compatibility.

These tests verify that:

1. ``add_review_to_pr`` surfaces **both** ``review_id`` (integer database ID)
   and ``review_node_id`` (GraphQL node ID) so callers can use either.
2. ``dismiss_pull_request_review`` accepts a GraphQL node ID directly.
3. ``dismiss_pull_request_review`` accepts a numeric (database) ID and
   resolves the node ID internally before calling GraphQL.
4. The ``_is_node_id`` helper correctly distinguishes the two formats.

All network I/O is intercepted via ``unittest.mock.patch`` so the tests run
without real GitHub credentials.
"""

from __future__ import annotations

import json
import unittest
from unittest.mock import MagicMock, patch

from mcp_connectors.github import (
    GitHubAPIError,
    GitHubGraphQLError,
    GitHubMCPConnector,
    _is_node_id,
)


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


# ---------------------------------------------------------------------------
# Helper to build a connector with a dummy token
# ---------------------------------------------------------------------------

def _make_connector() -> GitHubMCPConnector:
    return GitHubMCPConnector("ghp_test_token_placeholder")


# ---------------------------------------------------------------------------
# Tests for _is_node_id helper
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


# ---------------------------------------------------------------------------
# Tests for add_review_to_pr
# ---------------------------------------------------------------------------

class TestAddReviewToPr(unittest.TestCase):
    """Defect #32 – create path must return both review_id and review_node_id."""

    def _mock_rest_post(self, return_value: dict) -> MagicMock:
        """Return a context manager that patches _github_request for POST."""
        return patch(
            "mcp_connectors.github._github_request",
            return_value=return_value,
        )

    def test_returns_review_id_integer(self) -> None:
        connector = _make_connector()
        with self._mock_rest_post(_SAMPLE_REST_REVIEW):
            result = connector.add_review_to_pr(
                owner="owner",
                repo="repo",
                pull_number=1,
                body="Looks good!",
                event="APPROVE",
            )
        self.assertEqual(result["review_id"], NUMERIC_REVIEW_ID)

    def test_returns_review_node_id_string(self) -> None:
        """The critical fix: node ID must be present in the response."""
        connector = _make_connector()
        with self._mock_rest_post(_SAMPLE_REST_REVIEW):
            result = connector.add_review_to_pr(
                owner="owner",
                repo="repo",
                pull_number=1,
                body="Looks good!",
                event="APPROVE",
            )
        self.assertEqual(result["review_node_id"], NODE_REVIEW_ID)
        self.assertTrue(_is_node_id(result["review_node_id"]))

    def test_returns_both_ids_simultaneously(self) -> None:
        """Both IDs must co-exist in the same response dict."""
        connector = _make_connector()
        with self._mock_rest_post(_SAMPLE_REST_REVIEW):
            result = connector.add_review_to_pr(
                owner="owner",
                repo="repo",
                pull_number=1,
            )
        self.assertIn("review_id", result)
        self.assertIn("review_node_id", result)

    def test_normalises_other_fields(self) -> None:
        connector = _make_connector()
        with self._mock_rest_post(_SAMPLE_REST_REVIEW):
            result = connector.add_review_to_pr(
                owner="owner",
                repo="repo",
                pull_number=1,
            )
        self.assertEqual(result["state"], "APPROVED")
        self.assertEqual(result["user"], "octocat")
        self.assertEqual(result["body"], "Looks good!")
        self.assertIsNotNone(result["submitted_at"])

    def test_api_error_propagates(self) -> None:
        connector = _make_connector()
        with patch(
            "mcp_connectors.github._github_request",
            side_effect=GitHubAPIError(422, '{"message":"Validation Failed"}'),
        ):
            with self.assertRaises(GitHubAPIError) as ctx:
                connector.add_review_to_pr(
                    owner="owner",
                    repo="repo",
                    pull_number=1,
                )
        self.assertEqual(ctx.exception.status_code, 422)


# ---------------------------------------------------------------------------
# Tests for dismiss_pull_request_review
# ---------------------------------------------------------------------------

class TestDismissPullRequestReview(unittest.TestCase):
    """Defect #32 – dismiss path must accept both node ID and numeric ID."""

    def _mock_graphql(self, return_value: dict | None = None) -> MagicMock:
        data = return_value if return_value is not None else _SAMPLE_GRAPHQL_DISMISS_RESPONSE["data"]
        return patch("mcp_connectors.github._graphql_request", return_value=data)

    def test_dismiss_with_node_id_succeeds(self) -> None:
        """Passing the GraphQL node ID directly must work without REST lookup."""
        connector = _make_connector()
        with self._mock_graphql() as mock_gql, \
             patch("mcp_connectors.github._github_request") as mock_rest:
            result = connector.dismiss_pull_request_review(
                owner="owner",
                repo="repo",
                pull_number=1,
                review_id=NODE_REVIEW_ID,
                message="Dismissed via node ID.",
            )
        # REST should NOT be called when a node ID is supplied directly.
        mock_rest.assert_not_called()
        mock_gql.assert_called_once()
        args, kwargs = mock_gql.call_args
        # variables is the second positional argument to _graphql_request
        self.assertEqual(args[1]["reviewId"], NODE_REVIEW_ID)
        self.assertEqual(result["state"], "DISMISSED")

    def test_dismiss_with_numeric_id_resolves_node_id(self) -> None:
        """Passing a numeric ID triggers a REST lookup then calls GraphQL."""
        connector = _make_connector()
        with self._mock_graphql() as mock_gql, \
             patch(
                 "mcp_connectors.github._github_request",
                 return_value=_SAMPLE_REST_REVIEW,
             ) as mock_rest:
            result = connector.dismiss_pull_request_review(
                owner="owner",
                repo="repo",
                pull_number=1,
                review_id=NUMERIC_REVIEW_ID,
                message="Dismissed via numeric ID.",
            )
        # REST GET must be called to resolve the node ID.
        mock_rest.assert_called_once()
        rest_args, rest_kwargs = mock_rest.call_args
        self.assertEqual(rest_args[0], "GET")
        self.assertIn(str(NUMERIC_REVIEW_ID), rest_args[1])

        # GraphQL must be called with the resolved node ID.
        mock_gql.assert_called_once()
        gql_args, gql_kwargs = mock_gql.call_args
        # variables is the second positional argument to _graphql_request
        self.assertEqual(gql_args[1]["reviewId"], NODE_REVIEW_ID)
        self.assertEqual(result["state"], "DISMISSED")

    def test_dismiss_with_string_numeric_id_resolves_node_id(self) -> None:
        """A numeric ID passed as a string must also trigger the REST lookup."""
        connector = _make_connector()
        with self._mock_graphql(), \
             patch(
                 "mcp_connectors.github._github_request",
                 return_value=_SAMPLE_REST_REVIEW,
             ) as mock_rest:
            connector.dismiss_pull_request_review(
                owner="owner",
                repo="repo",
                pull_number=1,
                review_id=str(NUMERIC_REVIEW_ID),
                message="Dismissed via string numeric ID.",
            )
        mock_rest.assert_called_once()

    def test_graphql_error_propagates(self) -> None:
        connector = _make_connector()
        with patch(
            "mcp_connectors.github._graphql_request",
            side_effect=GitHubGraphQLError([{"message": "Could not resolve to a node"}]),
        ):
            with self.assertRaises(GitHubGraphQLError) as ctx:
                connector.dismiss_pull_request_review(
                    owner="owner",
                    repo="repo",
                    pull_number=1,
                    review_id=NODE_REVIEW_ID,
                    message="Should fail.",
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
                    owner="owner",
                    repo="repo",
                    pull_number=1,
                    review_id=NUMERIC_REVIEW_ID,
                    message="Should fail.",
                )
        self.assertEqual(ctx.exception.status_code, 404)


# ---------------------------------------------------------------------------
# Integration-style round-trip test (fully mocked)
# ---------------------------------------------------------------------------

class TestReviewLifecycleRoundTrip(unittest.TestCase):
    """End-to-end mocked flow: create a review then dismiss it using the
    returned ``review_node_id``."""

    def test_create_then_dismiss_using_node_id(self) -> None:
        connector = _make_connector()

        # Step 1 – create review
        with patch(
            "mcp_connectors.github._github_request",
            return_value=_SAMPLE_REST_REVIEW,
        ):
            created = connector.add_review_to_pr(
                owner="owner",
                repo="repo",
                pull_number=1,
                body="LGTM",
                event="APPROVE",
            )

        self.assertIn("review_node_id", created)
        node_id = created["review_node_id"]

        # Step 2 – dismiss using the node ID from step 1
        dismiss_data = {
            "dismissPullRequestReview": {
                "pullRequestReview": {
                    "id": node_id,
                    "state": "DISMISSED",
                }
            }
        }
        with patch(
            "mcp_connectors.github._graphql_request",
            return_value=dismiss_data,
        ) as mock_gql, patch("mcp_connectors.github._github_request") as mock_rest:
            result = connector.dismiss_pull_request_review(
                owner="owner",
                repo="repo",
                pull_number=1,
                review_id=node_id,
                message="No longer valid.",
            )

        # REST must NOT be invoked – node ID was used directly.
        mock_rest.assert_not_called()
        mock_gql.assert_called_once()
        self.assertEqual(result["state"], "DISMISSED")


if __name__ == "__main__":
    unittest.main()
