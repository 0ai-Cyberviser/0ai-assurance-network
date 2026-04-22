"""GitHub MCP connector."""
"""GitHub MCP connector — REST and GraphQL helpers for pull-request operations.

Provides:
- :class:`GitHubAPIError` — raised on non-2xx REST responses.
- :class:`GitHubGraphQLError` — raised when a GraphQL response contains errors.
- :func:`_github_request` — low-level REST helper (patchable in tests).
- :func:`_graphql_request` — low-level GraphQL helper (patchable in tests).
- :func:`_is_node_id` — distinguish GraphQL node IDs from numeric database IDs.
- :class:`GitHubMCPConnector` — high-level connector class.

No third-party dependencies — uses only stdlib ``urllib`` and ``json``.
"""

from __future__ import annotations

import json
import urllib.error
import urllib.error
import urllib.request
from typing import Any

_DISMISS_MUTATION = """
mutation DismissReview($reviewId: ID!, $message: String!) {
  dismissPullRequestReview(input: {pullRequestReviewId: $reviewId, message: $message}) {
    pullRequestReview {
      id
      state
    }
  }
}
"""

class GitHubAPIError(Exception):
import urllib.request
from typing import Any


# ---------------------------------------------------------------------------
# Exceptions
# ---------------------------------------------------------------------------


class GitHubAPIError(Exception):
    """Raised when the GitHub REST API returns a non-2xx status code."""

    def __init__(self, status_code: int, message: str) -> None:
        super().__init__(message)
        self.status_code = status_code

class GitHubGraphQLError(Exception):
    def __init__(self, errors: list[dict]) -> None:
        super().__init__(str(errors))
        self.errors = errors

def _is_node_id(s: str) -> bool:
    """Return True if *s* looks like a GitHub GraphQL node ID (non-empty and not purely numeric)."""
    return bool(s) and not s.isdigit()

def _github_request(method: str, url: str, *, token: str, **kwargs: Any) -> dict:
    """Make a GitHub REST API call and return the parsed JSON response."""
    data = kwargs.get("json")
    body = json.dumps(data).encode() if data is not None else None

class GitHubGraphQLError(Exception):
    """Raised when a GitHub GraphQL response contains one or more errors."""

    def __init__(self, errors: list[dict]) -> None:
        messages = "; ".join(e.get("message", "") for e in errors)
        super().__init__(messages)
        self.errors = errors


# ---------------------------------------------------------------------------
# Module-level helpers (patchable at mcp_connectors.github._github_request
# and mcp_connectors.github._graphql_request)
# ---------------------------------------------------------------------------

_GITHUB_API_BASE = "https://api.github.com"
_GRAPHQL_URL = "https://api.github.com/graphql"


def _github_request(
    method: str,
    url: str,
    *,
    token: str,
    json_data: Any = None,
) -> dict:
    """Make a REST API call to GitHub.

    Parameters
    ----------
    method:
        HTTP verb (e.g. ``"GET"``, ``"POST"``).
    url:
        Full URL to request.
    token:
        GitHub personal access token (used in ``Authorization`` header).
    json_data:
        Optional request body; serialised to JSON and sent with
        ``Content-Type: application/json``.

    Returns
    -------
    dict
        Parsed JSON response body.

    Raises
    ------
    GitHubAPIError
        If the response status code is not in the 2xx range.
    """
    body: bytes | None = None
    if json_data is not None:
        body = json.dumps(json_data).encode()

    req = urllib.request.Request(
        url,
        data=body,
        method=method,
        headers={
            "Authorization": f"Bearer {token}",
            "Accept": "application/vnd.github+json",
            "Content-Type": "application/json",
            "X-GitHub-Api-Version": "2022-11-28",
            "X-GitHub-Api-Version": "2022-11-28",
            "Content-Type": "application/json",
        },
    )
    try:
        with urllib.request.urlopen(req) as resp:
            return json.loads(resp.read().decode())
    except urllib.error.HTTPError as exc:
        raise GitHubAPIError(exc.code, exc.read().decode()) from exc

def _graphql_request(query: str, variables: dict, *, token: str, **kwargs: Any) -> dict:
    """Make a GitHub GraphQL API call and return the ``data`` dict from the response."""
    payload = json.dumps({"query": query, "variables": variables}).encode()
    req = urllib.request.Request(
        "https://api.github.com/graphql",
        try:
            payload = json.loads(exc.read().decode())
            message = payload.get("message", str(exc))
        except Exception:
            message = str(exc)
        raise GitHubAPIError(exc.code, message) from exc


def _graphql_request(query: str, variables: dict, *, token: str) -> dict:
    """Make a GraphQL request to GitHub.

    Parameters
    ----------
    query:
        GraphQL query or mutation string.
    variables:
        Variables dict for the query.
    token:
        GitHub personal access token.

    Returns
    -------
    dict
        The ``data`` field of the GraphQL response.

    Raises
    ------
    GitHubGraphQLError
        If the response JSON contains a top-level ``errors`` key.
    """
    payload = json.dumps({"query": query, "variables": variables}).encode()
    req = urllib.request.Request(
        _GRAPHQL_URL,
        data=payload,
        method="POST",
        headers={
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json",
        },
    )
    try:
        with urllib.request.urlopen(req) as resp:
            result = json.loads(resp.read().decode())
    except urllib.error.HTTPError as exc:
        raise GitHubAPIError(exc.code, exc.read().decode()) from exc
    if "errors" in result:
        raise GitHubGraphQLError(result["errors"])
    return result.get("data", {})

class GitHubMCPConnector:
    """Connector for GitHub MCP operations."""
    with urllib.request.urlopen(req) as resp:
        result = json.loads(resp.read().decode())

    if "errors" in result:
        raise GitHubGraphQLError(result["errors"])

    return result.get("data", {})


# ---------------------------------------------------------------------------
# Helper
# ---------------------------------------------------------------------------


def _is_node_id(value: str) -> bool:
    """Return True if *value* looks like a GraphQL node ID.

    A node ID is a non-empty string that is **not** purely numeric (numeric
    strings are treated as REST database IDs).
    """
    return bool(value) and not value.isdigit()


# ---------------------------------------------------------------------------
# Main connector class
# ---------------------------------------------------------------------------


class GitHubMCPConnector:
    """High-level connector for GitHub pull-request operations."""

    def __init__(self, token: str) -> None:
        self._token = token

    # ------------------------------------------------------------------
    # Reviews
    # ------------------------------------------------------------------

    def add_review_to_pr(
        self,
        owner: str,
        repo: str,
        pull_number: int,
        body: str = "",
        event: str = "COMMENT",
    ) -> dict:
        """Submit a review on a pull request and return normalised review data."""
        url = f"https://api.github.com/repos/{owner}/{repo}/pulls/{pull_number}/reviews"
        response = _github_request(
            "POST",
            url,
            token=self._token,
            json={"body": body, "event": event},
        )
        return {
            "review_id": response["id"],
            "review_node_id": response["node_id"],
            "state": response["state"],
            "user": response["user"]["login"],
            "body": response["body"],
            "submitted_at": response["submitted_at"],
        """Submit a review on a pull request.

        POST /repos/{owner}/{repo}/pulls/{pull_number}/reviews

        Parameters
        ----------
        owner:
            Repository owner (user or organisation login).
        repo:
            Repository name.
        pull_number:
            Pull request number.
        body:
            Optional review comment body text.
        event:
            Review event type — ``"COMMENT"``, ``"APPROVE"``, or
            ``"REQUEST_CHANGES"``.

        Returns
        -------
        dict
            A normalised response with keys:

            - ``review_id`` — integer database ID
            - ``review_node_id`` — GraphQL node ID string
            - ``state`` — review state string
            - ``user`` — login string of the reviewer
            - ``body`` — review body
            - ``submitted_at`` — ISO 8601 timestamp string
        """
        url = (
            f"{_GITHUB_API_BASE}/repos/{owner}/{repo}"
            f"/pulls/{pull_number}/reviews"
        )
        payload: dict[str, Any] = {}
        if body:
            payload["body"] = body
        if event:
            payload["event"] = event
        raw = _github_request("POST", url, token=self._token, json_data=payload)
        return {
            "review_id": raw["id"],
            "review_node_id": raw["node_id"],
            "state": raw["state"],
            "user": raw["user"]["login"],
            "body": raw.get("body", ""),
            "submitted_at": raw.get("submitted_at"),
        }

    def dismiss_pull_request_review(
        self,
        owner: str,
        repo: str,
        pull_number: int,
        review_id: int | str,
        message: str,
    ) -> dict:
        """Dismiss a pull request review.

        *review_id* may be either a GraphQL node ID string or a numeric
        (database) ID.  When a numeric ID is supplied the node ID is resolved
        via a REST call before invoking the GraphQL mutation.
        """
        if _is_node_id(str(review_id)):
            node_id = str(review_id)
        else:
            rest_response = _github_request(
                "GET",
                f"https://api.github.com/repos/{owner}/{repo}/pulls/{pull_number}/reviews/{review_id}",
                token=self._token,
            )
            node_id = rest_response["node_id"]

        data = _graphql_request(
            _DISMISS_MUTATION,
            {"reviewId": node_id, "message": message},
            token=self._token,
        )
        return data["dismissPullRequestReview"]["pullRequestReview"]
        review_id: Any,
        message: str = "",
    ) -> dict:
        """Dismiss a pull-request review using the GraphQL API.

        If *review_id* is a numeric (database) ID the method first resolves
        the GraphQL node ID via a REST GET call, then calls the GraphQL
        ``dismissPullRequestReview`` mutation.  If *review_id* is already a
        node ID it is used directly.

        Parameters
        ----------
        owner:
            Repository owner.
        repo:
            Repository name.
        pull_number:
            Pull request number.
        review_id:
            Either a numeric database ID (``int`` or numeric ``str``) or a
            GraphQL node ID string.
        message:
            Dismissal message (required by the GitHub GraphQL API).

        Returns
        -------
        dict
            A dict with at least the key ``state``.
        """
        node_id: str
        if _is_node_id(str(review_id)):
            node_id = str(review_id)
        else:
            url = (
                f"{_GITHUB_API_BASE}/repos/{owner}/{repo}"
                f"/pulls/{pull_number}/reviews/{review_id}"
            )
            raw = _github_request("GET", url, token=self._token)
            node_id = raw["node_id"]

        mutation = """
        mutation DismissReview($reviewId: ID!, $message: String!) {
          dismissPullRequestReview(input: {pullRequestReviewId: $reviewId, message: $message}) {
            pullRequestReview {
              id
              state
            }
          }
        }
        """
        variables = {"reviewId": node_id, "message": message}
        data = _graphql_request(mutation, variables, token=self._token)
        review = data["dismissPullRequestReview"]["pullRequestReview"]
        return {"state": review["state"], "id": review.get("id", node_id)}
