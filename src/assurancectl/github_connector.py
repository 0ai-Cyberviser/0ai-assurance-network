"""GitHub MCP connector helpers for the 0AI Assurance Network.

Provides normalization and filtering utilities used when consuming GitHub
MCP tool responses, ensuring that all review states (including COMMENTED)
are preserved and returned correctly.

Defect #33 fix: COMMENTED reviews were previously excluded by an incomplete
ALLOWED_REVIEW_STATES list.  The list now includes every state that the
GitHub REST API can return for a pull-request review.
"""

from __future__ import annotations

from typing import Any


# All valid GitHub pull-request review states.
# "COMMENTED" must be included; omitting it caused list_pull_request_reviews
# to silently drop reviews submitted with action=COMMENT (Defect #33).
ALLOWED_REVIEW_STATES: frozenset[str] = frozenset(
    [
        "COMMENTED",
        "APPROVED",
        "CHANGES_REQUESTED",
        "DISMISSED",
        "PENDING",
    ]
)


def normalize_review(raw: dict[str, Any]) -> dict[str, Any] | None:
    """Normalize a single raw pull-request review object from the GitHub API.

    Returns the normalized review dict, or ``None`` when the raw object is
    missing required fields or has an unrecognised state.

    All states listed in :data:`ALLOWED_REVIEW_STATES` are preserved.
    In particular, ``COMMENTED`` reviews are **not** filtered out.
    """
    if not isinstance(raw, dict):
        return None

    state = str(raw.get("state", "")).upper()
    if state not in ALLOWED_REVIEW_STATES:
        return None

    review_id = raw.get("id")
    if review_id is None:
        return None

    return {
        "review_id": review_id,
        "review_node_id": raw.get("node_id"),
        "state": state,
        "user": raw.get("user") or raw.get("author"),
        "body": raw.get("body", ""),
        "submitted_at": raw.get("submitted_at"),
        "html_url": raw.get("html_url"),
        "commit_id": raw.get("commit_id"),
    }


def list_pull_request_reviews(
    raw_reviews: list[dict[str, Any]],
) -> list[dict[str, Any]]:
    """Normalize a list of raw pull-request review objects from the GitHub API.

    Accepts the raw array returned by the GitHub REST endpoint
    ``GET /repos/{owner}/{repo}/pulls/{pull_number}/reviews`` and returns a
    list of normalized review dicts.

    Reviews with any state in :data:`ALLOWED_REVIEW_STATES` are included.
    Reviews with an unrecognised state or missing ``id`` are silently dropped.

    Args:
        raw_reviews: The raw list of review objects as returned by the GitHub
            API (or GitHub MCP tool response).

    Returns:
        A list of normalized review dicts.  The list preserves the original
        ordering from the API response.
    """
    if not isinstance(raw_reviews, list):
        return []

    normalized: list[dict[str, Any]] = []
    for raw in raw_reviews:
        review = normalize_review(raw)
        if review is not None:
            normalized.append(review)
    return normalized
