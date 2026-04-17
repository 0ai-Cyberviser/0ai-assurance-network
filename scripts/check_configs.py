#!/usr/bin/env python3
"""Validate the 0AI Assurance Network repo skeleton config."""

from __future__ import annotations

import sys
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
SRC = ROOT / "src"
if str(SRC) not in sys.path:
    sys.path.insert(0, str(SRC))

from assurancectl.cli import main  # noqa: E402


if __name__ == "__main__":
    raise SystemExit(main(["validate"]))
