"""GitHub MCP connector – pull request operations."""

from __future__ import annotations

from typing import Any


def is_cross_repo_pr(pr: dict[str, Any]) -> bool:
    """Return True when the PR head and base come from different repositories.

    GitHub only allows ``maintainer_can_modify`` on cross-repo (fork) pull
    requests.  Sending the field for a same-repo PR causes GitHub to return
    HTTP 422 with "Fork collab can only be enabled on cross-repo pull
    requests".

    The comparison uses the repository ``id`` field (integer) which is stable
    and unambiguous.  Both ``pr["head"]["repo"]`` and ``pr["base"]["repo"]``
    must be present; if either is missing the PR is treated as same-repo so
    the field is safely omitted.
    """
    try:
        head_repo_id = pr["head"]["repo"]["id"]
        base_repo_id = pr["base"]["repo"]["id"]
    except (KeyError, TypeError):
        return False
    return head_repo_id != base_repo_id


def build_update_payload(pr: dict[str, Any], updates: dict[str, Any]) -> dict[str, Any]:
    """Build the PATCH payload for a pull request update.

    All fields in *updates* are forwarded to the payload **except**
    ``maintainer_can_modify``, which is only included when the PR is
    cross-repo.  This prevents HTTP 422 errors from GitHub when the caller
    passes ``maintainer_can_modify`` for a same-repo PR.

    Args:
        pr: The current pull request object as returned by the GitHub API
            (must include ``head.repo.id`` and ``base.repo.id``).
        updates: Mapping of fields the caller wants to change.  Recognised
            fields include ``title``, ``body``, ``state``, ``base``, and
            ``maintainer_can_modify``.

    Returns:
        A dict suitable for use as the JSON body of a ``PATCH /repos/{owner}/{repo}/pulls/{pull_number}`` request.
    """
    payload: dict[str, Any] = {}

    for key, value in updates.items():
        if key == "maintainer_can_modify":
            # Only forward this field for cross-repo (fork) PRs.
            # GitHub returns HTTP 422 if it is sent for same-repo PRs.
            if is_cross_repo_pr(pr):
                payload[key] = value
        else:
            payload[key] = value

    return payload
