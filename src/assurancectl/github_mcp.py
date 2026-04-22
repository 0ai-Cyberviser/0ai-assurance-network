"""GitHub MCP connector reaction handlers.

Provides read operations for GitHub reactions on issues, issue comments,
and pull requests using the GitHub REST API.

Defect #34 fix
--------------
``get_issue_comment_reactions`` previously returned an empty list immediately
after a reaction was created because it was hitting the wrong endpoint.

Correct endpoint mapping
~~~~~~~~~~~~~~~~~~~~~~~~
* Issue comment reactions:
    GET /repos/{owner}/{repo}/issues/comments/{comment_id}/reactions
* Issue reactions (and PR reactions, since PRs are issues):
    GET /repos/{owner}/{repo}/issues/{issue_number}/reactions

Using ``/issues/{number}/reactions`` for a comment-level lookup silently
scopes the query to issue-level reactions, which returns an empty list when
only comment reactions exist.  This module uses the correct endpoint for each
reaction type and shares a single normalizer so the output shape is identical
regardless of the reaction target.
"""

from __future__ import annotations

from typing import Any


# ---------------------------------------------------------------------------
# Endpoint path templates
# ---------------------------------------------------------------------------

# Issue comment reactions – the corrected path for Defect #34
_ISSUE_COMMENT_REACTIONS_PATH = (
    "/repos/{owner}/{repo}/issues/comments/{comment_id}/reactions"
)

# Issue-level reactions (also used for PR reactions, since GitHub models PRs
# as issues for the purpose of the reactions API)
_ISSUE_REACTIONS_PATH = "/repos/{owner}/{repo}/issues/{issue_number}/reactions"

# PR reactions (alias kept explicit for readability; same URL shape)
_PR_REACTIONS_PATH = "/repos/{owner}/{repo}/issues/{pull_number}/reactions"


# ---------------------------------------------------------------------------
# Response normalizer
# ---------------------------------------------------------------------------


def _normalize_reaction(raw: dict[str, Any]) -> dict[str, Any]:
    """Normalise a raw GitHub reaction payload to a consistent shape.

    Both the issue-comment and the issue/PR reaction endpoints return the same
    JSON structure, so a single normaliser is shared across all callers.  This
    guarantees that the output of ``get_issue_comment_reactions`` and
    ``get_pr_reactions`` are byte-for-byte identical for the same raw payload.
    """
    user = raw.get("user") or {}
    return {
        "id": raw["id"],
        "content": raw["content"],
        "created_at": raw.get("created_at", ""),
        "user": {
            "login": user.get("login", ""),
            "id": user.get("id", 0),
        },
    }


# ---------------------------------------------------------------------------
# Pagination helper
# ---------------------------------------------------------------------------


def _paginate_reactions(
    client: Any,
    path: str,
    *,
    per_page: int = 100,
) -> list[dict[str, Any]]:
    """Collect all pages from a GitHub reactions endpoint.

    Args:
        client: Object with a ``get(path, *, params)`` method that returns a
            list of raw reaction dicts for the requested page.
        path: Absolute path component of the GitHub API URL.
        per_page: Number of results per page (GitHub maximum is 100).

    Returns:
        Flat list of raw reaction dicts across all pages.
    """
    results: list[dict[str, Any]] = []
    page = 1
    while True:
        page_results: list[dict[str, Any]] = client.get(
            path, params={"per_page": per_page, "page": page}
        )
        results.extend(page_results)
        if len(page_results) < per_page:
            break
        page += 1
    return results


# ---------------------------------------------------------------------------
# Public API
# ---------------------------------------------------------------------------


def get_issue_comment_reactions(
    client: Any,
    *,
    owner: str,
    repo: str,
    comment_id: int,
) -> list[dict[str, Any]]:
    """Fetch all reactions for a specific issue comment.

    Uses the correct endpoint::

        GET /repos/{owner}/{repo}/issues/comments/{comment_id}/reactions

    This is *distinct* from ``/repos/{owner}/{repo}/issues/{number}/reactions``
    which returns reactions on the issue itself.  Using the issue-level endpoint
    for a comment lookup returns an empty list (Defect #34).

    Args:
        client: HTTP client with a ``get(path, *, params)`` method.
        owner: Repository owner (user or organisation).
        repo: Repository name.
        comment_id: Numeric ID of the issue comment.

    Returns:
        List of normalised reaction dicts.
    """
    path = _ISSUE_COMMENT_REACTIONS_PATH.format(
        owner=owner, repo=repo, comment_id=comment_id
    )
    raw_reactions = _paginate_reactions(client, path)
    return [_normalize_reaction(r) for r in raw_reactions]


def get_pr_reactions(
    client: Any,
    *,
    owner: str,
    repo: str,
    pull_number: int,
) -> list[dict[str, Any]]:
    """Fetch all reactions for a pull request.

    Pull requests are modelled as issues in GitHub's data model, so this uses::

        GET /repos/{owner}/{repo}/issues/{pull_number}/reactions

    Args:
        client: HTTP client with a ``get(path, *, params)`` method.
        owner: Repository owner (user or organisation).
        repo: Repository name.
        pull_number: PR number.

    Returns:
        List of normalised reaction dicts.
    """
    path = _PR_REACTIONS_PATH.format(
        owner=owner, repo=repo, pull_number=pull_number
    )
    raw_reactions = _paginate_reactions(client, path)
    return [_normalize_reaction(r) for r in raw_reactions]


def get_issue_reactions(
    client: Any,
    *,
    owner: str,
    repo: str,
    issue_number: int,
) -> list[dict[str, Any]]:
    """Fetch all reactions for an issue.

    Args:
        client: HTTP client with a ``get(path, *, params)`` method.
        owner: Repository owner (user or organisation).
        repo: Repository name.
        issue_number: Issue number.

    Returns:
        List of normalised reaction dicts.
    """
    path = _ISSUE_REACTIONS_PATH.format(
        owner=owner, repo=repo, issue_number=issue_number
    )
    raw_reactions = _paginate_reactions(client, path)
    return [_normalize_reaction(r) for r in raw_reactions]
