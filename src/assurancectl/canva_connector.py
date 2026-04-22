"""Canva MCP connector utilities for asset upload and folder listing.

Addresses Defect #35: Asset Upload Not Reflected in Uploads Folder Listing.

Canva's backend exhibits eventual consistency for newly uploaded assets:
an upload that returns a successful job response and asset ID may not be
immediately visible in the Uploads folder listing.  The ``verify_upload``
helper implements exponential-backoff retry logic so that callers can
confirm visibility before trusting the folder listing.
"""

from __future__ import annotations

import time
from typing import Any


# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------

UPLOADS_FOLDER_ID = "uploads"
ROOT_FOLDER_ID = "root"

# Retry defaults for verify_upload (seconds).
_DEFAULT_MAX_RETRIES = 3
_DEFAULT_BASE_DELAY = 2.0


# ---------------------------------------------------------------------------
# Simulated Canva API layer (replaceable with real HTTP calls)
# ---------------------------------------------------------------------------


class CanvaAPIError(Exception):
    """Raised when a Canva API operation fails."""


class CanvaClient:
    """Thin wrapper around Canva Connect API calls.

    In production, replace the ``_assets`` store and method stubs with real
    HTTP requests to the Canva Connect API.  The interface is kept simple so
    that the retry / verification logic in this module can be unit-tested
    without a live Canva account.
    """

    def __init__(self) -> None:
        # In-memory store used for unit tests; not used in live mode.
        self._assets: dict[str, dict[str, Any]] = {}

    # ------------------------------------------------------------------
    # Upload
    # ------------------------------------------------------------------

    def upload_asset_from_url(
        self,
        *,
        url: str,
        name: str,
        folder_id: str = UPLOADS_FOLDER_ID,
    ) -> dict[str, Any]:
        """Upload an asset from a URL into *folder_id*.

        Returns a dict with at least ``asset_id`` and ``status`` keys.
        Raises :class:`CanvaAPIError` on failure.
        """
        raise NotImplementedError(  # pragma: no cover
            "Replace with a real Canva Connect API call in production."
        )

    # ------------------------------------------------------------------
    # Folder listing
    # ------------------------------------------------------------------

    def list_folder_items(
        self,
        folder_id: str,
        *,
        item_types: list[str] | None = None,
        continuation_token: str | None = None,
    ) -> dict[str, Any]:
        """Return items in *folder_id*.

        Returns a dict with at least an ``items`` list.  Supports optional
        ``item_types`` filter and pagination via *continuation_token*.
        Raises :class:`CanvaAPIError` on failure.
        """
        raise NotImplementedError(  # pragma: no cover
            "Replace with a real Canva Connect API call in production."
        )


# ---------------------------------------------------------------------------
# Upload verification with retry
# ---------------------------------------------------------------------------


def verify_upload(
    asset_id: str,
    *,
    client: CanvaClient,
    folder_id: str = UPLOADS_FOLDER_ID,
    item_types: list[str] | None = None,
    max_retries: int = _DEFAULT_MAX_RETRIES,
    base_delay: float = _DEFAULT_BASE_DELAY,
    _sleep: Any = time.sleep,
) -> bool:
    """Verify that *asset_id* appears in *folder_id* after upload.

    Canva's backend may take a short time to index a newly uploaded asset in
    the Uploads folder even though the upload job itself has already
    succeeded.  This function retries the folder listing up to *max_retries*
    times with exponential backoff (``base_delay * 2^attempt`` seconds) to
    accommodate that eventual-consistency window.

    Parameters
    ----------
    asset_id:
        The asset ID returned by the upload job.
    client:
        A :class:`CanvaClient` instance (or compatible object) used for API
        calls.
    folder_id:
        The folder to check.  Defaults to :data:`UPLOADS_FOLDER_ID`.
    item_types:
        Optional list of item-type filters passed to ``list_folder_items``.
    max_retries:
        Maximum number of listing attempts before giving up.
    base_delay:
        Base sleep duration in seconds between retries (doubled each attempt).
    _sleep:
        Injected sleep callable; use ``lambda _: None`` in tests to avoid
        real delays.

    Returns
    -------
    bool
        ``True`` if the asset is found within the retry window, ``False``
        otherwise.
    """
    for attempt in range(max_retries):
        items = list_all_folder_items(
            folder_id,
            client=client,
            item_types=item_types,
        )
        if any(item.get("asset_id") == asset_id for item in items):
            return True
        if attempt < max_retries - 1:
            _sleep(base_delay * (2**attempt))
    return False


# ---------------------------------------------------------------------------
# Paginated folder listing helper
# ---------------------------------------------------------------------------


def list_all_folder_items(
    folder_id: str,
    *,
    client: CanvaClient,
    item_types: list[str] | None = None,
) -> list[dict[str, Any]]:
    """Return all items in *folder_id*, following pagination tokens.

    Partial pagination was one of the suspected root causes of Defect #35.
    This helper exhausts all pages so callers receive the complete listing.
    """
    items: list[dict[str, Any]] = []
    continuation_token: str | None = None
    while True:
        result = client.list_folder_items(
            folder_id,
            item_types=item_types,
            continuation_token=continuation_token,
        )
        page_items: list[dict[str, Any]] = result.get("items", [])
        items.extend(page_items)
        continuation_token = result.get("continuation_token")
        if not continuation_token:
            break
    return items
