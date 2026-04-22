"""GitHub MCP connector."""

from __future__ import annotations

import json
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
    req = urllib.request.Request(
        url,
        data=body,
        method=method,
        headers={
            "Authorization": f"Bearer {token}",
            "Accept": "application/vnd.github+json",
            "Content-Type": "application/json",
            "X-GitHub-Api-Version": "2022-11-28",
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

    def __init__(self, token: str) -> None:
        self._token = token

    def add_review_to_pr(
        self,
        owner: str,
        repo: str,
        pull_number: int,
        body: str = "",
        event: str = "COMMENT",
        **kwargs: Any,
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
        }

    def dismiss_pull_request_review(
        self,
        owner: str,
        repo: str,
        pull_number: int,
        review_id: int | str,
        message: str,
        **kwargs: Any,
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
