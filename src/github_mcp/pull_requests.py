"""GitHub MCP connector – pull request operations."""

from __future__ import annotations

import json
import urllib.error
import urllib.request
from typing import Any


def is_cross_repo_pr(pr: dict[str, Any]) -> bool:
    """Return True when the PR head and base come from different repositories.

    GitHub only allows ``maintainer_can_modify`` on cross-repo (fork) pull
    requests.  Sending the field for a same-repo PR causes GitHub to return
    HTTP 422 with "Fork collab can only be enabled on cross-repo pull
    requests".

    The comparison uses the repository ``id`` field (integer) which is stable
    and unambiguous.  Both ``pr["head"]["repo"]`` and ``pr["base"]["repo"]``
    must be present; if either is missing or either ``id`` is not a valid
    integer, the PR is treated as same-repo so the field is safely omitted.
    """
    try:
        head_repo_id = pr["head"]["repo"]["id"]
        base_repo_id = pr["base"]["repo"]["id"]
    except (KeyError, TypeError):
        return False

    if type(head_repo_id) is not int or type(base_repo_id) is not int:
        return False
    return head_repo_id != base_repo_id


def build_update_payload(pr: dict[str, Any], updates: dict[str, Any]) -> dict[str, Any]:
    """Build the PATCH payload for a pull request update.

    All fields in *updates* are forwarded to the payload **except**
    ``maintainer_can_modify``, which is only included when the PR is
    cross-repo.  This prevents HTTP 422 errors from GitHub when the caller
    passes ``maintainer_can_modify`` for a same-repo PR.

    Args:
        pr: The current pull request object as returned by the GitHub API.
            When available, ``head.repo.id`` and ``base.repo.id`` are used
            to detect cross-repo PRs; if that repo information is missing or
            invalid, the PR is treated as same-repo and
            ``maintainer_can_modify`` is safely omitted.
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


def update_pull_request(
    owner: str,
    repo: str,
    pull_number: int,
    pr: dict[str, Any],
    updates: dict[str, Any],
    *,
    token: str,
) -> dict[str, Any]:
    """PATCH a pull request on GitHub, safely suppressing ``maintainer_can_modify``
    for same-repo PRs to avoid HTTP 422.

    Args:
        owner: Repository owner.
        repo: Repository name.
        pull_number: Pull request number.
        pr: The current pull request object (used to detect cross-repo).
        updates: Fields to update (see :func:`build_update_payload`).
        token: GitHub API token.

    Returns:
        The updated pull request object as returned by the GitHub API.

    Raises:
        urllib.error.HTTPError: On non-2xx responses.
    """
    payload = build_update_payload(pr, updates)
    url = f"https://api.github.com/repos/{owner}/{repo}/pulls/{pull_number}"
    body = json.dumps(payload).encode()
    req = urllib.request.Request(
        url,
        data=body,
        method="PATCH",
        headers={
            "Authorization": f"Bearer {token}",
            "Accept": "application/vnd.github+json",
            "Content-Type": "application/json",
            "X-GitHub-Api-Version": "2022-11-28",
        },
    )
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read().decode())
