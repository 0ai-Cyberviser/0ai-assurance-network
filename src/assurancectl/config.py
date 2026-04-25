"""Config loading and validation for the 0AI Assurance Network skeleton."""

from __future__ import annotations

import json
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


class ValidationError(ValueError):
    """Raised when repo skeleton config is invalid."""


@dataclass(frozen=True)
class LoadedConfig:
    """Typed container for the repo skeleton config."""

    root: Path
    topology: dict[str, Any]
    genesis: dict[str, Any]
    policy: dict[str, Any]
    inference_policy: dict[str, Any]
    checkpoint_signers: dict[str, Any]
    module_plan: dict[str, Any]
    identity_bootstrap: dict[str, Any]


def root_dir(explicit_root: str | Path | None = None) -> Path:
    if explicit_root is not None:
        return Path(explicit_root).resolve()
    return Path(__file__).resolve().parents[2]


def _load_json(path: Path) -> dict[str, Any]:
    with path.open("r", encoding="utf-8") as handle:
        return json.load(handle)


def load_config(explicit_root: str | Path | None = None) -> LoadedConfig:
    root = root_dir(explicit_root)
    config = root / "config"
    return LoadedConfig(
        root=root,
        topology=_load_json(config / "network-topology.json"),
        genesis=_load_json(config / "genesis" / "base-genesis.json"),
        policy=_load_json(config / "policy" / "release-guards.json"),
        inference_policy=_load_json(config / "governance" / "inference-policy.json"),
        checkpoint_signers=_load_json(config / "governance" / "checkpoint-signers.json"),
        module_plan=_load_json(config / "modules" / "milestone-1.json"),
        identity_bootstrap=_load_json(config / "identity" / "bootstrap.json"),
    )


def _assert(condition: bool, message: str) -> None:
    if not condition:
        raise ValidationError(message)


def _parse_utc_timestamp(value: Any, field_name: str) -> datetime:
    text = str(value or "")
    _assert(text != "", f"{field_name} must be set")
    try:
        parsed = datetime.fromisoformat(text.replace("Z", "+00:00"))
    except ValueError as exc:
        raise ValidationError(f"{field_name} must be RFC3339 timestamp") from exc
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed.astimezone(timezone.utc)


def validate_topology(topology: dict[str, Any]) -> None:
    _assert(topology["mode"] == "permissioned_testnet", "mode must be permissioned_testnet")
    validators = topology["validators"]
    seeds = topology["seed_nodes"]
    _assert(len(validators) >= 7, "at least seven validators are required in the initial topology")
    _assert(len(seeds) >= 1, "at least one seed node is required")

    ids = set()
    monikers = set()
    ports = set()
    voting_power_total = 0

    for validator in validators:
        validator_id = validator["id"]
        moniker = validator["moniker"]
        _assert(validator_id not in ids, f"duplicate validator id: {validator_id}")
        _assert(moniker not in monikers, f"duplicate validator moniker: {moniker}")
        ids.add(validator_id)
        monikers.add(moniker)

        for key in ("rpc_port", "p2p_port", "app_port", "prometheus_port"):
            port = int(validator[key])
            _assert(port not in ports, f"duplicate port detected: {port}")
            _assert(1024 <= port <= 65535, f"invalid port: {port}")
            ports.add(port)

        voting_power_total += int(validator["voting_power"])

    for seed in seeds:
        _assert(seed["id"] not in ids, f"seed id conflicts with validator id: {seed['id']}")

    _assert(voting_power_total == 100, "validator voting power must sum to 100")
    _assert(
        topology["governance"]["critical_actions_require_dual_approval"] is True,
        "critical actions must require dual approval",
    )


def validate_genesis(genesis: dict[str, Any], topology: dict[str, Any]) -> None:
    _assert(genesis["chain_id"] == topology["chain_id"], "genesis chain_id must match topology chain_id")
    _assert(genesis["launch_mode"] == "permissioned_testnet", "launch_mode must be permissioned_testnet")

    fee_split = genesis["treasury"]["fee_split_percent"]
    _assert(sum(int(value) for value in fee_split.values()) == 100, "fee split must sum to 100")
    _assert(genesis["governance"]["dual_house_enabled"] is True, "dual-house governance must be enabled")
    _assert(
        genesis["incident"]["public_reason_codes_required"] is True,
        "incident module must require public reason codes",
    )


def validate_policy(policy: dict[str, Any]) -> None:
    required = set(policy["required_before_public_launch"])
    prohibited = set(policy["prohibited_shortcuts"])

    _assert("external_legal_review" in required, "external legal review must be required")
    _assert("external_security_audit" in required, "external security audit must be required")
    _assert(
        "public_retail_sale_without_legal_review" in prohibited,
        "public retail sale shortcut must be prohibited",
    )
    _assert(
        policy["safe_mode_defaults"]["permissioned_testnet"] is True,
        "permissioned_testnet safe default must be enabled",
    )
    _assert(
        policy["safe_mode_defaults"]["public_transferability"] is False,
        "public_transferability safe default must be disabled",
    )


def _required_signer_roles(inference_policy: dict[str, Any]) -> set[str]:
    remediation = inference_policy.get("remediation", {})
    execution_defaults = remediation.get("execution_defaults", {})
    required: set[str] = set()
    required.update(str(role) for role in execution_defaults.get("phase_owners", {}).values())
    for override in execution_defaults.get("owner_overrides", {}).values():
        required.update(str(role) for role in override.values())
    return {role for role in required if role}


def validate_checkpoint_signers(checkpoint_signers: dict[str, Any], inference_policy: dict[str, Any]) -> None:
    _assert(
        checkpoint_signers["signature_format"] == "0ai-hmac-sha256-v1",
        "checkpoint signer signature_format must be 0ai-hmac-sha256-v1",
    )
    _assert(
        bool(checkpoint_signers["require_signatures_for_event_logs"]),
        "checkpoint signer policy must require signatures for event logs",
    )
    _assert(
        int(checkpoint_signers["maximum_signature_validity_seconds"]) > 0,
        "checkpoint signer validity window must be positive",
    )
    rotation_policy = dict(checkpoint_signers.get("rotation_policy", {}))
    reference_time = _parse_utc_timestamp(
        rotation_policy.get("reference_time"),
        "checkpoint signer rotation_policy.reference_time",
    )
    warning_window_days = int(rotation_policy.get("warning_window_days", 0))
    _assert(
        warning_window_days > 0,
        "checkpoint signer rotation_policy.warning_window_days must be positive",
    )
    approval_roles = [str(role) for role in rotation_policy.get("approval_roles", [])]
    _assert(
        approval_roles,
        "checkpoint signer rotation_policy.approval_roles must not be empty",
    )
    signers = list(checkpoint_signers["signers"])
    _assert(signers, "at least one checkpoint signer must be configured")

    signer_ids: set[str] = set()
    key_ids: set[str] = set()
    role_bindings: set[str] = set()
    active_actor_ids: set[str] = set()
    active_role_coverage: dict[str, str] = {}
    required_roles = _required_signer_roles(inference_policy)
    for signer in signers:
        actor_id = str(signer.get("actor_id", ""))
        signer_id = str(signer["signer_id"])
        key_id = str(signer["key_id"])
        shared_secret = str(signer["shared_secret"])
        status = str(signer.get("status", ""))
        roles = [str(role) for role in signer["roles"]]
        provisioned_at = _parse_utc_timestamp(
            signer.get("provisioned_at"),
            f"checkpoint signer {signer_id} provisioned_at",
        )
        rotate_by = _parse_utc_timestamp(
            signer.get("rotate_by"),
            f"checkpoint signer {signer_id} rotate_by",
        )
        _assert(actor_id != "", f"checkpoint signer {signer_id} must declare an actor_id")
        _assert(signer_id not in signer_ids, f"duplicate checkpoint signer_id: {signer_id}")
        _assert(key_id not in key_ids, f"duplicate checkpoint key_id: {key_id}")
        _assert(shared_secret != "", f"checkpoint signer {signer_id} must declare a shared_secret")
        _assert(status in {"active", "inactive"}, f"checkpoint signer {signer_id} has invalid status")
        _assert(
            rotate_by > provisioned_at,
            f"checkpoint signer {signer_id} rotate_by must be after provisioned_at",
        )
        _assert(roles, f"checkpoint signer {signer_id} must declare at least one role")
        for role in roles:
            binding = f"{signer_id}:{role}"
            _assert(binding not in role_bindings, f"duplicate checkpoint signer role binding: {binding}")
            role_bindings.add(binding)
        if status == "active":
            _assert(
                actor_id not in active_actor_ids,
                f"duplicate checkpoint signer actor ownership: {actor_id}",
            )
            _assert(
                rotate_by > reference_time,
                f"checkpoint signer {signer_id} has stale rotation metadata",
            )
            for role in roles:
                owner = active_role_coverage.get(role)
                _assert(
                    owner is None,
                    f"duplicate checkpoint signer role coverage: {role}",
                )
                active_role_coverage[role] = signer_id
            active_actor_ids.add(actor_id)
        signer_ids.add(signer_id)
        key_ids.add(key_id)

    missing_roles = sorted(required_roles - set(active_role_coverage))
    _assert(
        not missing_roles,
        f"checkpoint signer coverage missing roles: {', '.join(missing_roles)}",
    )
    missing_approval_roles = sorted(set(approval_roles) - set(active_role_coverage))
    _assert(
        not missing_approval_roles,
        f"checkpoint signer approval roles missing active coverage: {', '.join(missing_approval_roles)}",
    )


def validate_module_plan(module_plan: dict[str, Any]) -> None:
    _assert(module_plan["version"] != "", "module plan version must be set")
    _assert(module_plan["milestone"] != "", "module plan milestone must be set")
    _assert(module_plan["scope"] != "", "module plan scope must be set")

    module_names: set[str] = set()
    for collection_name in ("mvp_modules", "dependency_surfaces"):
        collection = list(module_plan[collection_name])
        _assert(collection, f"module plan {collection_name} must not be empty")
        for module in collection:
            name = str(module["name"])
            _assert(name not in module_names, f"duplicate module name: {name}")
            module_names.add(name)
            _assert(module["purpose"] != "", f"module {name} must declare a purpose")
            _assert(module["state"], f"module {name} must declare state")
            _assert(module["transactions"], f"module {name} must declare transactions")
            _assert(module["operator_permissions"], f"module {name} must declare operator permissions")
            for state in module["state"]:
                _assert(state["key"] != "", f"module {name} has state entry with empty key")
                _assert(state["type"] != "", f"module {name} state {state['key']} must declare a type")
            transaction_names: set[str] = set()
            for tx in module["transactions"]:
                tx_name = str(tx["name"])
                _assert(tx_name not in transaction_names, f"module {name} has duplicate transaction {tx_name}")
                _assert(tx["actor_roles"], f"module {name} transaction {tx_name} must declare actor roles")
                transaction_names.add(tx_name)

    rollout = list(module_plan["rollout"])
    _assert(rollout, "module plan rollout must not be empty")
    phase_numbers: list[int] = []
    phase_names: set[str] = set()
    for phase in rollout:
        phase_number = int(phase["phase"])
        phase_name = str(phase["name"])
        _assert(phase_name not in phase_names, f"duplicate rollout phase name: {phase_name}")
        _assert(phase["deliverables"], f"rollout phase {phase_name} must declare deliverables")
        phase_numbers.append(phase_number)
        phase_names.add(phase_name)
    _assert(phase_numbers == list(range(1, len(phase_numbers) + 1)), "rollout phases must be sequential starting at 1")
    seen_phase_names: set[str] = set()
    for phase in rollout:
        phase_name = str(phase["name"])
        for dependency in phase["depends_on"]:
            _assert(
                dependency in seen_phase_names,
                f"rollout phase {phase_name} depends on unknown or future phase {dependency}",
            )
        seen_phase_names.add(phase_name)


def _required_identity_roles(module_plan: dict[str, Any]) -> set[str]:
    required: set[str] = set()
    for collection_name in ("mvp_modules", "dependency_surfaces"):
        for module in module_plan[collection_name]:
            for tx in module["transactions"]:
                required.update(str(role) for role in tx["actor_roles"])
            for permission in module["operator_permissions"]:
                required.add(str(permission["role"]))
    return required


def _required_governance_identity_roles(
    inference_policy: dict[str, Any],
    checkpoint_signers: dict[str, Any],
) -> set[str]:
    required: set[str] = set()
    execution_defaults = (
        inference_policy.get("remediation", {})
        .get("execution_defaults", {})
    )
    required.update(str(role) for role in execution_defaults.get("phase_owners", {}).values())
    for override in execution_defaults.get("owner_overrides", {}).values():
        required.update(str(role) for role in override.values())
    for signer in checkpoint_signers.get("signers", []):
        required.update(str(role) for role in signer.get("roles", []))
    return {role for role in required if role}


def _allowed_identity_roles(
    module_plan: dict[str, Any],
    inference_policy: dict[str, Any],
    checkpoint_signers: dict[str, Any],
) -> set[str]:
    return (
        _required_identity_roles(module_plan)
        | _required_governance_identity_roles(inference_policy, checkpoint_signers)
    )


def validate_identity_bootstrap(
    identity_bootstrap: dict[str, Any],
    module_plan: dict[str, Any],
    topology: dict[str, Any],
    inference_policy: dict[str, Any],
    checkpoint_signers: dict[str, Any],
) -> None:
    _assert(identity_bootstrap["version"] != "", "identity bootstrap version must be set")
    _assert(identity_bootstrap["chain_id"] == topology["chain_id"], "identity bootstrap chain id must match topology")

    actors = list(identity_bootstrap["actors"])
    role_bindings = list(identity_bootstrap["role_bindings"])
    _assert(actors, "identity bootstrap actors must not be empty")
    _assert(role_bindings, "identity bootstrap role bindings must not be empty")

    allowed_actor_types = {"organization", "operator", "council", "service_account"}
    allowed_status = {"active", "inactive"}
    actors_by_id: dict[str, dict[str, Any]] = {}
    for actor in actors:
        actor_id = str(actor["actor_id"])
        _assert(actor_id not in actors_by_id, f"duplicate identity actor id: {actor_id}")
        _assert(actor["actor_type"] in allowed_actor_types, f"identity actor {actor_id} has invalid actor_type")
        _assert(actor["status"] in allowed_status, f"identity actor {actor_id} has invalid status")
        _assert(str(actor["display_name"]) != "", f"identity actor {actor_id} must declare a display_name")
        actors_by_id[actor_id] = actor

    for actor in actors:
        organization_id = str(actor.get("organization_id", "") or "")
        if not organization_id:
            continue
        _assert(organization_id in actors_by_id, f"identity actor {actor['actor_id']} references unknown organization")
        organization = actors_by_id[organization_id]
        _assert(
            organization["actor_type"] in {"organization", "council"},
            f"identity actor {actor['actor_id']} organization must be organization or council",
        )

    required_roles = _allowed_identity_roles(module_plan, inference_policy, checkpoint_signers)
    seen_bindings: set[str] = set()
    active_roles: set[str] = set()
    active_actor_role_bindings: set[str] = set()
    for binding in role_bindings:
        actor_id = str(binding["actor_id"])
        role = str(binding["role"])
        scope = str(binding["scope"])
        granted_by = str(binding["granted_by"])
        status = str(binding["status"])
        _assert(actor_id in actors_by_id, f"identity role binding references unknown actor {actor_id}")
        _assert(role in required_roles, f"identity role binding uses undeclared role {role}")
        _assert(scope != "", f"identity role binding {actor_id}/{role} must declare a scope")
        _assert(granted_by != "", f"identity role binding {actor_id}/{role} must declare granted_by")
        _assert(status in allowed_status, f"identity role binding {actor_id}/{role} has invalid status")
        binding_key = f"{actor_id}|{role}|{scope}"
        _assert(binding_key not in seen_bindings, f"duplicate identity role binding: {binding_key}")
        seen_bindings.add(binding_key)
        if status == "active":
            active_roles.add(role)
            active_actor_role_bindings.add(f"{actor_id}|{role}")

    missing_roles = sorted(required_roles - active_roles)
    _assert(not missing_roles, f"identity bootstrap missing active bindings for roles: {', '.join(missing_roles)}")
    for signer in checkpoint_signers["signers"]:
        actor_id = str(signer["actor_id"])
        signer_id = str(signer["signer_id"])
        _assert(actor_id in actors_by_id, f"checkpoint signer {signer_id} references unknown actor {actor_id}")
        actor = actors_by_id[actor_id]
        _assert(actor["status"] == "active", f"checkpoint signer {signer_id} references inactive actor {actor_id}")
        for role in signer["roles"]:
            binding_key = f"{actor_id}|{role}"
            _assert(
                binding_key in active_actor_role_bindings,
                f"checkpoint signer {signer_id} role {role} is not backed by an active identity binding",
            )


def validate_all(config: LoadedConfig) -> None:
    validate_topology(config.topology)
    validate_genesis(config.genesis, config.topology)
    validate_policy(config.policy)
    validate_checkpoint_signers(config.checkpoint_signers, config.inference_policy)
    validate_module_plan(config.module_plan)
    validate_identity_bootstrap(
        config.identity_bootstrap,
        config.module_plan,
        config.topology,
        config.inference_policy,
        config.checkpoint_signers,
    )
