"""Artifact rendering for the 0AI Assurance Network repo skeleton."""

from __future__ import annotations

import json
from pathlib import Path
from typing import Any

from .config import LoadedConfig


def build_dir(config: LoadedConfig) -> Path:
    return config.root / "build" / "localnet"


def compose_text(config: LoadedConfig) -> str:
    topology = config.topology
    image = topology["container_image"]
    binary = topology["binary"]

    services = [
        "services:",
        "  seed-1:",
        f"    image: {image}",
        "    container_name: 0ai-seed-1",
        "    command:",
        f"      - {binary}",
        "      - start",
        "      - --home",
        "      - /chain/seed-1",
        "      - --p2p.laddr=tcp://0.0.0.0:26656",
        "    ports:",
        "      - \"26656:26656\"",
        "    volumes:",
        "      - seed-1-data:/chain/seed-1",
    ]

    peers = ",".join(
        f"{seed['id']}@{seed['p2p_host']}:{seed['p2p_port']}" for seed in topology["seed_nodes"]
    )

    for validator in topology["validators"]:
        moniker = validator["moniker"]
        services.extend(
            [
                f"  {moniker}:",
                f"    image: {image}",
                f"    container_name: 0ai-{moniker}",
                "    command:",
                f"      - {binary}",
                "      - start",
                "      - --home",
                f"      - /chain/{moniker}",
                "      - --rpc.laddr=tcp://0.0.0.0:26657",
                f"      - --p2p.laddr=tcp://0.0.0.0:{validator['p2p_port']}",
                f"      - --p2p.persistent_peers={peers}",
                "    ports:",
                f"      - \"{validator['rpc_port']}:26657\"",
                f"      - \"{validator['p2p_port']}:{validator['p2p_port']}\"",
                f"      - \"{validator['app_port']}:1317\"",
                f"      - \"{validator['prometheus_port']}:26660\"",
                "    volumes:",
                f"      - {moniker}-data:/chain/{moniker}",
                "    depends_on:",
                "      - seed-1",
            ]
        )

    services.append("volumes:")
    services.append("  seed-1-data:")
    for validator in topology["validators"]:
        services.append(f"  {validator['moniker']}-data:")

    return "\n".join(services) + "\n"


def network_summary(config: LoadedConfig) -> dict[str, Any]:
    topology = config.topology
    genesis = config.genesis
    return {
        "network_name": topology["network_name"],
        "chain_id": topology["chain_id"],
        "mode": topology["mode"],
        "validator_count": len(topology["validators"]),
        "seed_count": len(topology["seed_nodes"]),
        "governance_houses": topology["governance"]["houses"],
        "base_denom": genesis["denoms"]["base"],
        "validators": [
            {
                "id": validator["id"],
                "moniker": validator["moniker"],
                "rpc_port": validator["rpc_port"],
                "p2p_port": validator["p2p_port"],
                "voting_power": validator["voting_power"],
            }
            for validator in topology["validators"]
        ],
    }


def rendered_genesis(config: LoadedConfig) -> dict[str, Any]:
    rendered = dict(config.genesis)
    rendered["localnet"] = {
        "rendered_from": "config/genesis/base-genesis.json",
        "validators": [validator["moniker"] for validator in config.topology["validators"]],
        "seed_nodes": [seed["moniker"] for seed in config.topology["seed_nodes"]],
    }
    return rendered


def write_localnet_artifacts(config: LoadedConfig) -> Path:
    output = build_dir(config)
    output.mkdir(parents=True, exist_ok=True)

    (output / "docker-compose.yml").write_text(compose_text(config), encoding="utf-8")
    (output / "network-summary.json").write_text(
        json.dumps(network_summary(config), indent=2) + "\n",
        encoding="utf-8",
    )
    (output / "genesis.rendered.json").write_text(
        json.dumps(rendered_genesis(config), indent=2) + "\n",
        encoding="utf-8",
    )
    return output
