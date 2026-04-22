"""Tests for the Canva MCP connector (Defect #35 regression coverage).

Verifies that:
- ``verify_upload`` returns True when the asset appears on the first attempt.
- ``verify_upload`` retries and returns True when the asset appears on a later attempt.
- ``verify_upload`` returns False after exhausting all retries without finding the asset.
- ``list_all_folder_items`` follows pagination tokens until all pages are consumed.
- ``verify_upload`` passes ``item_types`` through to ``list_folder_items``.
"""

from __future__ import annotations

import sys
import unittest
from pathlib import Path
from typing import Any
from unittest.mock import MagicMock, call

# Allow direct imports from the src layout without installing the package.
sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "src"))

from assurancectl.canva_connector import (  # noqa: E402
    UPLOADS_FOLDER_ID,
    CanvaClient,
    list_all_folder_items,
    verify_upload,
)


class _StubClient(CanvaClient):
    """CanvaClient stub that replays pre-configured ``list_folder_items`` responses."""

    def __init__(self, responses: list[dict[str, Any]]) -> None:
        super().__init__()
        self._responses = iter(responses)

    def list_folder_items(
        self,
        folder_id: str,
        *,
        item_types: list[str] | None = None,
        continuation_token: str | None = None,
    ) -> dict[str, Any]:
        return next(self._responses)


class TestVerifyUpload(unittest.TestCase):
    """Unit tests for verify_upload."""

    def _no_sleep(self, _: float) -> None:
        pass

    def test_found_on_first_attempt(self) -> None:
        """Asset visible immediately: verify_upload returns True after one call."""
        client = _StubClient(
            [{"items": [{"asset_id": "abc123", "name": "photo.jpg"}]}]
        )
        result = verify_upload("abc123", client=client, _sleep=self._no_sleep)
        self.assertTrue(result)

    def test_found_on_second_attempt(self) -> None:
        """Asset not visible initially but appears on the second attempt (eventual consistency)."""
        client = _StubClient(
            [
                {"items": []},  # attempt 1 – not yet indexed
                {"items": [{"asset_id": "abc123", "name": "photo.jpg"}]},  # attempt 2
            ]
        )
        result = verify_upload(
            "abc123", client=client, max_retries=3, _sleep=self._no_sleep
        )
        self.assertTrue(result)

    def test_found_on_last_attempt(self) -> None:
        """Asset appears only on the final retry."""
        client = _StubClient(
            [
                {"items": []},
                {"items": []},
                {"items": [{"asset_id": "abc123", "name": "photo.jpg"}]},
            ]
        )
        result = verify_upload(
            "abc123", client=client, max_retries=3, _sleep=self._no_sleep
        )
        self.assertTrue(result)

    def test_not_found_exhausts_retries(self) -> None:
        """Asset never appears: verify_upload returns False after max_retries attempts."""
        client = _StubClient(
            [{"items": []}, {"items": []}, {"items": []}]
        )
        result = verify_upload(
            "abc123", client=client, max_retries=3, _sleep=self._no_sleep
        )
        self.assertFalse(result)

    def test_sleep_called_between_retries(self) -> None:
        """Exponential backoff delays are applied between failed attempts."""
        sleep_mock = MagicMock()
        # Asset found on 3rd attempt (2 sleeps expected).
        client = _StubClient(
            [
                {"items": []},
                {"items": []},
                {"items": [{"asset_id": "xyz", "name": "img.png"}]},
            ]
        )
        verify_upload(
            "xyz",
            client=client,
            max_retries=3,
            base_delay=1.0,
            _sleep=sleep_mock,
        )
        # Delays should be 1 * 2^0 = 1.0 and 1 * 2^1 = 2.0.
        sleep_mock.assert_has_calls([call(1.0), call(2.0)])

    def test_no_sleep_on_last_attempt(self) -> None:
        """No sleep occurs after the final attempt (whether or not it succeeds)."""
        sleep_mock = MagicMock()
        client = _StubClient(
            [{"items": []}, {"items": []}, {"items": []}]  # never found
        )
        verify_upload(
            "missing",
            client=client,
            max_retries=3,
            base_delay=1.0,
            _sleep=sleep_mock,
        )
        # Only 2 sleeps for a 3-attempt run (no sleep after the last attempt).
        self.assertEqual(sleep_mock.call_count, 2)

    def test_item_types_forwarded(self) -> None:
        """item_types filter is passed through to list_folder_items."""
        mock_client = MagicMock(spec=CanvaClient)
        mock_client.list_folder_items.return_value = {
            "items": [{"asset_id": "q1", "name": "banner.png"}]
        }
        verify_upload(
            "q1",
            client=mock_client,
            item_types=["image"],
            _sleep=self._no_sleep,
        )
        mock_client.list_folder_items.assert_called_with(
            UPLOADS_FOLDER_ID,
            item_types=["image"],
            continuation_token=None,
        )

    def test_custom_folder_id(self) -> None:
        """verify_upload checks the specified folder, not the default uploads folder."""
        mock_client = MagicMock(spec=CanvaClient)
        mock_client.list_folder_items.return_value = {
            "items": [{"asset_id": "a1"}]
        }
        result = verify_upload(
            "a1",
            client=mock_client,
            folder_id="custom-folder",
            _sleep=self._no_sleep,
        )
        self.assertTrue(result)
        mock_client.list_folder_items.assert_called_with(
            "custom-folder",
            item_types=None,
            continuation_token=None,
        )

    def test_asset_not_confused_with_other_ids(self) -> None:
        """Only an exact asset_id match is accepted."""
        client = _StubClient(
            [{"items": [{"asset_id": "other-id", "name": "other.jpg"}]}]
        )
        result = verify_upload(
            "target-id", client=client, max_retries=1, _sleep=self._no_sleep
        )
        self.assertFalse(result)


class TestListAllFolderItems(unittest.TestCase):
    """Unit tests for list_all_folder_items."""

    def test_single_page(self) -> None:
        """Single-page responses are returned directly."""
        client = _StubClient(
            [{"items": [{"asset_id": "a1"}, {"asset_id": "a2"}]}]
        )
        items = list_all_folder_items(UPLOADS_FOLDER_ID, client=client)
        self.assertEqual(len(items), 2)
        self.assertEqual(items[0]["asset_id"], "a1")

    def test_multiple_pages(self) -> None:
        """Pagination tokens are followed until all pages are consumed."""
        client = _StubClient(
            [
                {"items": [{"asset_id": "a1"}], "continuation_token": "tok1"},
                {"items": [{"asset_id": "a2"}], "continuation_token": "tok2"},
                {"items": [{"asset_id": "a3"}]},  # last page – no token
            ]
        )
        items = list_all_folder_items(UPLOADS_FOLDER_ID, client=client)
        self.assertEqual([i["asset_id"] for i in items], ["a1", "a2", "a3"])

    def test_empty_folder(self) -> None:
        """An empty folder returns an empty list."""
        client = _StubClient([{"items": []}])
        items = list_all_folder_items(UPLOADS_FOLDER_ID, client=client)
        self.assertEqual(items, [])

    def test_item_types_forwarded(self) -> None:
        """item_types is passed to list_folder_items on every page request."""
        mock_client = MagicMock(spec=CanvaClient)
        mock_client.list_folder_items.side_effect = [
            {"items": [{"asset_id": "img1"}], "continuation_token": "tok"},
            {"items": [{"asset_id": "img2"}]},
        ]
        items = list_all_folder_items(
            UPLOADS_FOLDER_ID, client=mock_client, item_types=["image"]
        )
        self.assertEqual(len(items), 2)
        for c in mock_client.list_folder_items.call_args_list:
            self.assertEqual(c.kwargs["item_types"], ["image"])


if __name__ == "__main__":
    unittest.main()
