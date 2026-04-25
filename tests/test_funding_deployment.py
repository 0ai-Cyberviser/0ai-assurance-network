"""Tests for blockchain funding deployment."""

from __future__ import annotations

import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


PROJECT_ROOT = Path(__file__).resolve().parents[1]


class FundingDeploymentTests(unittest.TestCase):
    """Test cases for funding deployment functionality."""

    def test_funding_config_validation_success(self) -> None:
        """Test that valid funding configuration passes validation."""
        result = subprocess.run(
            [
                sys.executable,
                str(PROJECT_ROOT / "scripts" / "deploy_funding.py"),
                "--root",
                str(PROJECT_ROOT),
                "--funding-config",
                str(PROJECT_ROOT / "config" / "governance" / "funding-config.json"),
                "--dry-run",
            ],
            capture_output=True,
            text=True,
            check=False,
        )

        self.assertEqual(result.returncode, 0, f"Deployment failed: {result.stderr}")
        self.assertIn("Validation successful", result.stdout)

    def test_funding_config_invalid_allocation(self) -> None:
        """Test that invalid allocation percentages are rejected."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            invalid_config = {
                "funding_pools": {
                    "pool1": {"allocation_percent": 60.0, "min_balance": 1000},
                    "pool2": {"allocation_percent": 60.0, "min_balance": 1000},
                },
                "allocation_strategy": "proportional",
                "treasury_address": "test-treasury",
            }
            json.dump(invalid_config, f)
            config_path = Path(f.name)

        try:
            result = subprocess.run(
                [
                    sys.executable,
                    str(PROJECT_ROOT / "scripts" / "deploy_funding.py"),
                    "--root",
                    str(PROJECT_ROOT),
                    "--funding-config",
                    str(config_path),
                    "--dry-run",
                ],
                capture_output=True,
                text=True,
                check=False,
            )

            self.assertNotEqual(result.returncode, 0)
            self.assertIn("Total allocation must equal 100%", result.stdout)
        finally:
            config_path.unlink()

    def test_funding_config_missing_required_field(self) -> None:
        """Test that missing required fields are detected."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            invalid_config = {
                "funding_pools": {
                    "pool1": {"allocation_percent": 100.0, "min_balance": 1000},
                },
                # Missing allocation_strategy and treasury_address
            }
            json.dump(invalid_config, f)
            config_path = Path(f.name)

        try:
            result = subprocess.run(
                [
                    sys.executable,
                    str(PROJECT_ROOT / "scripts" / "deploy_funding.py"),
                    "--root",
                    str(PROJECT_ROOT),
                    "--funding-config",
                    str(config_path),
                    "--dry-run",
                ],
                capture_output=True,
                text=True,
                check=False,
            )

            self.assertNotEqual(result.returncode, 0)
            self.assertIn("Missing required field", result.stdout)
        finally:
            config_path.unlink()

    def test_funding_deployment_output_generation(self) -> None:
        """Test that deployment configuration is correctly generated."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as output_file:
            output_path = Path(output_file.name)

        try:
            result = subprocess.run(
                [
                    sys.executable,
                    str(PROJECT_ROOT / "scripts" / "deploy_funding.py"),
                    "--root",
                    str(PROJECT_ROOT),
                    "--funding-config",
                    str(PROJECT_ROOT / "config" / "governance" / "funding-config.json"),
                    "--output",
                    str(output_path),
                    "--dry-run",
                ],
                capture_output=True,
                text=True,
                check=False,
            )

            self.assertEqual(result.returncode, 0, f"Deployment failed: {result.stderr}")

            # Verify output file was created
            self.assertTrue(output_path.exists(), "Output file was not created")

            # Load and validate output
            with open(output_path, encoding="utf-8") as f:
                deployment_config = json.load(f)

            self.assertIn("deployment_version", deployment_config)
            self.assertIn("network_id", deployment_config)
            self.assertIn("funding_config", deployment_config)
            self.assertIn("validators", deployment_config)
            self.assertIsInstance(deployment_config["validators"], list)

        finally:
            if output_path.exists():
                output_path.unlink()

    def test_funding_config_invalid_strategy(self) -> None:
        """Test that invalid allocation strategy is rejected."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as f:
            invalid_config = {
                "funding_pools": {
                    "pool1": {"allocation_percent": 100.0, "min_balance": 1000},
                },
                "allocation_strategy": "invalid_strategy",
                "treasury_address": "test-treasury",
            }
            json.dump(invalid_config, f)
            config_path = Path(f.name)

        try:
            result = subprocess.run(
                [
                    sys.executable,
                    str(PROJECT_ROOT / "scripts" / "deploy_funding.py"),
                    "--root",
                    str(PROJECT_ROOT),
                    "--funding-config",
                    str(config_path),
                    "--dry-run",
                ],
                capture_output=True,
                text=True,
                check=False,
            )

            self.assertNotEqual(result.returncode, 0)
            self.assertIn("Invalid allocation_strategy", result.stdout)
        finally:
            config_path.unlink()

    def test_validator_funding_allocation(self) -> None:
        """Test that validators receive correct funding allocation in deployment."""
        with tempfile.NamedTemporaryFile(mode="w", suffix=".json", delete=False) as output_file:
            output_path = Path(output_file.name)

        try:
            result = subprocess.run(
                [
                    sys.executable,
                    str(PROJECT_ROOT / "scripts" / "deploy_funding.py"),
                    "--root",
                    str(PROJECT_ROOT),
                    "--funding-config",
                    str(PROJECT_ROOT / "config" / "governance" / "funding-config.json"),
                    "--output",
                    str(output_path),
                    "--dry-run",
                ],
                capture_output=True,
                text=True,
                check=False,
            )

            self.assertEqual(result.returncode, 0)

            with open(output_path, encoding="utf-8") as f:
                deployment_config = json.load(f)

            # Load network topology to get expected validator count
            with open(PROJECT_ROOT / "config" / "network-topology.json", encoding="utf-8") as f:
                topology = json.load(f)

            expected_validator_count = len(topology.get("validators", []))
            actual_validator_count = len(deployment_config["validators"])

            self.assertEqual(
                actual_validator_count,
                expected_validator_count,
                f"Expected {expected_validator_count} validators, got {actual_validator_count}",
            )

            # Verify each validator has required fields
            for validator in deployment_config["validators"]:
                self.assertIn("validator_id", validator)
                self.assertIn("address", validator)
                self.assertIn("initial_stake", validator)
                self.assertIn("funding_pool", validator)

        finally:
            if output_path.exists():
                output_path.unlink()


if __name__ == "__main__":
    unittest.main()
