#!/usr/bin/env python3
"""Deploy funding mechanism for 0AI Assurance Network.

This script validates and deploys the funding configuration for the blockchain,
including treasury allocation, validator funding, and grant distribution mechanisms.
"""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path
from typing import Any


def validate_funding_config(config: dict[str, Any]) -> tuple[bool, list[str]]:
    """Validate funding configuration structure and values.

    Args:
        config: Funding configuration dictionary

    Returns:
        Tuple of (is_valid, error_messages)
    """
    errors = []

    # Check required fields
    required_fields = ["funding_pools", "allocation_strategy", "treasury_address"]
    for field in required_fields:
        if field not in config:
            errors.append(f"Missing required field: {field}")

    # Validate funding pools
    if "funding_pools" in config:
        pools = config["funding_pools"]
        if not isinstance(pools, dict):
            errors.append("funding_pools must be a dictionary")
        else:
            total_allocation = 0
            for pool_name, pool_config in pools.items():
                if "allocation_percent" not in pool_config:
                    errors.append(f"Pool {pool_name} missing allocation_percent")
                else:
                    total_allocation += pool_config["allocation_percent"]

                if "min_balance" not in pool_config:
                    errors.append(f"Pool {pool_name} missing min_balance")

            if abs(total_allocation - 100.0) > 0.01:
                errors.append(f"Total allocation must equal 100%, got {total_allocation}%")

    # Validate allocation strategy
    if "allocation_strategy" in config:
        valid_strategies = ["proportional", "fixed", "dynamic"]
        if config["allocation_strategy"] not in valid_strategies:
            errors.append(
                f"Invalid allocation_strategy: {config['allocation_strategy']}. "
                f"Must be one of: {', '.join(valid_strategies)}"
            )

    return len(errors) == 0, errors


def load_network_topology(root: Path) -> dict[str, Any]:
    """Load network topology configuration."""
    topology_path = root / "config" / "network-topology.json"
    if not topology_path.exists():
        raise FileNotFoundError(f"Network topology not found: {topology_path}")

    with open(topology_path, encoding="utf-8") as f:
        return json.load(f)


def load_genesis_config(root: Path) -> dict[str, Any]:
    """Load genesis configuration."""
    genesis_path = root / "config" / "genesis" / "base-genesis.json"
    if not genesis_path.exists():
        raise FileNotFoundError(f"Genesis config not found: {genesis_path}")

    with open(genesis_path, encoding="utf-8") as f:
        return json.load(f)


def generate_funding_deployment(
    root: Path,
    funding_config: dict[str, Any],
    output_path: Path | None = None,
) -> dict[str, Any]:
    """Generate funding deployment configuration.

    Args:
        root: Project root directory
        funding_config: Funding configuration
        output_path: Optional path to write deployment config

    Returns:
        Deployment configuration dictionary
    """
    # Load dependent configs
    topology = load_network_topology(root)
    genesis = load_genesis_config(root)

    # Build deployment configuration
    deployment = {
        "deployment_version": "1.0.0",
        "network_id": topology.get("network_id", "0ai-testnet"),
        "genesis_time": genesis.get("genesis_time"),
        "funding_config": funding_config,
        "validators": [],
    }

    # Add validator funding allocations
    for validator in topology.get("validators", []):
        validator_funding = {
            "validator_id": validator["id"],
            "address": validator.get("address", f"placeholder-{validator['id']}"),
            "initial_stake": funding_config.get("initial_validator_stake", 0),
            "funding_pool": "validator_rewards",
        }
        deployment["validators"].append(validator_funding)

    # Write output if requested
    if output_path:
        output_path.parent.mkdir(parents=True, exist_ok=True)
        with open(output_path, "w", encoding="utf-8") as f:
            json.dump(deployment, f, indent=2)
        print(f"Deployment configuration written to: {output_path}")

    return deployment


def deploy_funding(
    deployment_config: dict[str, Any],
    dry_run: bool = False,
) -> bool:
    """Deploy funding configuration to the blockchain.

    Args:
        deployment_config: Deployment configuration
        dry_run: If True, only validate without deploying

    Returns:
        True if deployment successful, False otherwise
    """
    print(f"{'[DRY RUN] ' if dry_run else ''}Deploying funding configuration...")
    print(f"Network: {deployment_config['network_id']}")
    print(f"Validators: {len(deployment_config['validators'])}")

    # Validate deployment config
    if "funding_config" not in deployment_config:
        print("ERROR: Missing funding_config in deployment")
        return False

    funding_config = deployment_config["funding_config"]
    is_valid, errors = validate_funding_config(funding_config)

    if not is_valid:
        print("ERROR: Funding configuration validation failed:")
        for error in errors:
            print(f"  - {error}")
        return False

    if dry_run:
        print("Validation successful. Deployment skipped (dry run).")
        return True

    # In a real deployment, this would interact with the blockchain
    # For now, we simulate successful deployment
    print("Funding configuration validated and ready for deployment.")
    print("Note: Actual blockchain deployment requires chain runtime.")

    return True


def main() -> int:
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Deploy funding mechanism for 0AI Assurance Network"
    )
    parser.add_argument(
        "--root",
        type=Path,
        default=Path("."),
        help="Project root directory",
    )
    parser.add_argument(
        "--funding-config",
        type=Path,
        required=True,
        help="Path to funding configuration JSON",
    )
    parser.add_argument(
        "--output",
        type=Path,
        help="Path to write deployment configuration",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Validate without deploying",
    )

    args = parser.parse_args()

    # Load funding configuration
    try:
        with open(args.funding_config, encoding="utf-8") as f:
            funding_config = json.load(f)
    except FileNotFoundError:
        print(f"ERROR: Funding config not found: {args.funding_config}")
        return 1
    except json.JSONDecodeError as e:
        print(f"ERROR: Invalid JSON in funding config: {e}")
        return 1

    # Generate deployment configuration
    try:
        deployment_config = generate_funding_deployment(
            args.root,
            funding_config,
            args.output,
        )
    except Exception as e:
        print(f"ERROR: Failed to generate deployment: {e}")
        return 1

    # Deploy
    success = deploy_funding(deployment_config, args.dry_run)

    return 0 if success else 1


if __name__ == "__main__":
    sys.exit(main())
