"""GitHub MCP connector for pull-request review lifecycle operations.

This module wraps the GitHub REST and GraphQL APIs used by MCP tools so that
review IDs returned by ``add_review_to_pr`` are directly usable by
``dismiss_pull_request_review``.

Root cause addressed (Defect #32)
----------------------------------
The GitHub REST API ``POST /repos/{owner}/{repo}/pulls/{pull_number}/reviews``
returns a review object that contains two distinct identifiers:

* ``id``      – an integer *database* ID (e.g. ``4125974959``)
* ``node_id`` – a base-64 *GraphQL node* ID (e.g. ``PRR_kwDOA...``)

The ``dismissPullRequestReview`` GraphQL mutation only accepts the node ID.
Previously ``add_review_to_pr`` surfaced only the numeric ``id``, making the
dismiss path non-functional for callers who relied solely on the create
response.

Fix applied
-----------
``add_review_to_pr`` now returns both identifiers::

    {
        "review_id": 4125974959,
        "review_node_id": "PRR_kwDOA...",
        ...
    }

``dismiss_pull_request_review`` accepts either format.  When a caller passes
a numeric ID (or its string representation), the connector fetches the
corresponding node ID from the REST API before issuing the GraphQL mutation.
"""

from __future__ import annotations

import re
import urllib.error
import urllib.request
from typing import Any

import json

# ---------------------------------------------------------------------------
# Internal helpers
# ---------------------------------------------------------------------------

# Modern compact node IDs: e.g. "PRR_kwDORtzkIs717WGv"
_NODE_ID_COMPACT_RE = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*_[A-Za-z0-9+/=_-]+$")
# Legacy base-64 node IDs: e.g. "MDExOlB1bGxSZXF1ZXN0UmV2aWV3NDEyNTk3NDk1OQ=="
_NODE_ID_BASE64_RE = re.compile(r"^[A-Za-z0-9+/]{20,}={0,2}$")


def _is_node_id(value: str) -> bool:
    """Return True when *value* looks like a GitHub GraphQL node ID.

    Handles both the modern compact format (e.g. ``PRR_kwDOA...``) and the
    legacy base-64 format (e.g. ``MDExOlB1bGxSZXF1ZXN0UmV2aWV3...``).
    """
    if not value:
        return False
    return bool(_NODE_ID_COMPACT_RE.match(value) or _NODE_ID_BASE64_RE.match(value))


def _github_request(
    method: str,
    url: str,
    *,
    token: str,
    payload: dict[str, Any] | None = None,
) -> dict[str, Any]:
    """Perform a single GitHub API request and return the parsed JSON body.

    Parameters
    ----------
    method:
        HTTP verb (``GET``, ``POST``, ``PUT``, …).
    url:
        Fully-qualified API URL.
    token:
        GitHub personal-access token or installation token.
    payload:
        Optional request body to JSON-encode.

    Raises
    ------
    GitHubAPIError
        When GitHub returns a non-2xx status code.
    """
    data: bytes | None = None
    headers: dict[str, str] = {
        "Authorization": f"Bearer {token}",
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
    }
    if payload is not None:
        data = json.dumps(payload).encode("utf-8")
        headers["Content-Type"] = "application/json"

    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req) as resp:  # noqa: S310
            return json.loads(resp.read().decode("utf-8"))
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8", errors="replace")
        raise GitHubAPIError(exc.code, body) from exc


def _graphql_request(
    query: str,
    variables: dict[str, Any],
    *,
    token: str,
) -> dict[str, Any]:
    """Execute a GitHub GraphQL query/mutation.

    Parameters
    ----------
    query:
        The GraphQL document string.
    variables:
        Variables to accompany the document.
    token:
        GitHub personal-access token or installation token.

    Raises
    ------
    GitHubAPIError
        When the HTTP layer returns a non-2xx status.
    GitHubGraphQLError
        When the GraphQL response contains ``errors``.
    """
    result = _github_request(
        "POST",
        "https://api.github.com/graphql",
        token=token,
        payload={"query": query, "variables": variables},
    )
    if result.get("errors"):
        raise GitHubGraphQLError(result["errors"])
    return result.get("data", {})


# ---------------------------------------------------------------------------
# Exceptions
# ---------------------------------------------------------------------------


class GitHubAPIError(Exception):
    """Raised when GitHub returns a non-2xx HTTP response."""

    def __init__(self, status_code: int, body: str) -> None:
        self.status_code = status_code
        self.body = body
        super().__init__(f"GitHub API error {status_code}: {body}")


class GitHubGraphQLError(Exception):
    """Raised when a GraphQL response contains errors."""

    def __init__(self, errors: list[dict[str, Any]]) -> None:
        self.errors = errors
        messages = "; ".join(e.get("message", str(e)) for e in errors)
        super().__init__(f"GitHub GraphQL errors: {messages}")


# ---------------------------------------------------------------------------
# Public connector
# ---------------------------------------------------------------------------

_DISMISS_MUTATION = """
mutation DismissPullRequestReview($reviewId: ID!, $message: String!) {
  dismissPullRequestReview(input: {pullRequestReviewId: $reviewId, message: $message}) {
    pullRequestReview {
      id
      state
    }
  }
}
"""


class GitHubMCPConnector:
    """Thin connector that bridges MCP tool calls to the GitHub API.

    Parameters
    ----------
    token:
        GitHub personal-access token or installation token with the necessary
        scopes (``repo`` or ``pull_requests: write``).
    api_base:
        Override the REST API base URL (useful for GitHub Enterprise or tests).
    """

    def __init__(self, token: str, *, api_base: str = "https://api.github.com") -> None:
        self._token = token
        self._api_base = api_base.rstrip("/")

    # ------------------------------------------------------------------
    # Review creation
    # ------------------------------------------------------------------

    def add_review_to_pr(
        self,
        *,
        owner: str,
        repo: str,
        pull_number: int,
        body: str = "",
        event: str = "COMMENT",
        comments: list[dict[str, Any]] | None = None,
        commit_id: str | None = None,
    ) -> dict[str, Any]:
        """Create a pull-request review and return compatible ID fields.

        Uses ``POST /repos/{owner}/{repo}/pulls/{pull_number}/reviews``.

        The response always contains **both**:

        * ``review_id``      – integer database ID (REST-compatible)
        * ``review_node_id`` – GraphQL node ID (e.g. ``PRR_kwDOA…``)

        This makes the returned IDs directly usable by
        :meth:`dismiss_pull_request_review` without any additional lookup.

        Parameters
        ----------
        owner:
            Repository owner login.
        repo:
            Repository name.
        pull_number:
            Pull-request number.
        body:
            Review body text.
        event:
            One of ``APPROVE``, ``REQUEST_CHANGES``, or ``COMMENT``.
        comments:
            Optional list of per-line review comments (see GitHub docs).
        commit_id:
            SHA of the commit to attach the review to.  Defaults to the
            most-recent commit of the PR when omitted.

        Returns
        -------
        dict
            Normalised review data including ``review_id`` (int) and
            ``review_node_id`` (str).
        """
        url = f"{self._api_base}/repos/{owner}/{repo}/pulls/{pull_number}/reviews"
        payload: dict[str, Any] = {"body": body, "event": event}
        if comments:
            payload["comments"] = comments
        if commit_id:
            payload["commit_id"] = commit_id

        raw = _github_request("POST", url, token=self._token, payload=payload)
        return self._normalise_review(raw)

    # ------------------------------------------------------------------
    # Review dismissal
    # ------------------------------------------------------------------

    def dismiss_pull_request_review(
        self,
        *,
        owner: str,
        repo: str,
        pull_number: int,
        review_id: int | str,
        message: str,
    ) -> dict[str, Any]:
        """Dismiss a pull-request review via the GitHub GraphQL API.

        Accepts either the integer database ID **or** the GraphQL node ID
        returned by :meth:`add_review_to_pr`.  When a numeric ID is supplied
        the connector resolves the corresponding node ID automatically via the
        REST API before issuing the GraphQL mutation.

        Parameters
        ----------
        owner:
            Repository owner login.
        repo:
            Repository name.
        pull_number:
            Pull-request number.
        review_id:
            Either the integer (or string representation of the integer)
            database ID, **or** the GraphQL node ID (e.g. ``PRR_kwDOA…``).
        message:
            Dismissal message shown to the reviewer.

        Returns
        -------
        dict
            Normalised review data after dismissal.
        """
        node_id = self._resolve_review_node_id(
            owner=owner,
            repo=repo,
            pull_number=pull_number,
            review_id=review_id,
        )
        data = _graphql_request(
            _DISMISS_MUTATION,
            {"reviewId": node_id, "message": message},
            token=self._token,
        )
        dismissed = data["dismissPullRequestReview"]["pullRequestReview"]
        return dismissed

    # ------------------------------------------------------------------
    # Internal helpers
    # ------------------------------------------------------------------

    def _normalise_review(self, raw: dict[str, Any]) -> dict[str, Any]:
        """Map GitHub REST review fields to connector-canonical field names.

        Critically, both ``review_id`` (integer) and ``review_node_id``
        (GraphQL node ID string) are surfaced.
        """
        return {
            "review_id": raw["id"],
            "review_node_id": raw["node_id"],
            "pull_request_url": raw.get("pull_request_url", ""),
            "state": raw.get("state", ""),
            "body": raw.get("body", ""),
            "submitted_at": raw.get("submitted_at"),
            "user": raw.get("user", {}).get("login") if raw.get("user") else None,
            "html_url": raw.get("html_url", ""),
        }

    def _resolve_review_node_id(
        self,
        *,
        owner: str,
        repo: str,
        pull_number: int,
        review_id: int | str,
    ) -> str:
        """Return the GraphQL node ID for *review_id*.

        If *review_id* is already a node ID string (detected by pattern),
        it is returned as-is.  Otherwise the connector fetches the review
        via the REST API to obtain the ``node_id`` field.

        Parameters
        ----------
        owner:
            Repository owner login.
        repo:
            Repository name.
        pull_number:
            Pull-request number.
        review_id:
            Numeric database ID or GraphQL node ID.
        """
        str_id = str(review_id)
        if _is_node_id(str_id):
            return str_id

        # Numeric ID supplied – look up the node_id via REST.
        url = (
            f"{self._api_base}/repos/{owner}/{repo}"
            f"/pulls/{pull_number}/reviews/{str_id}"
        )
        raw = _github_request("GET", url, token=self._token)
        node_id: str = raw["node_id"]
        return node_id
