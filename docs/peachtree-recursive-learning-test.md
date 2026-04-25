# Peachtree Recursive Learning Validation Plan

This document defines a safe, deterministic test plan for the Peachtree recursive learning loop across 0AI Assurance Network governance artifacts.

The plan is intentionally **generate-only** and **human-in-the-loop**. It does not perform autonomous on-chain actions, token issuance, production key usage, public deployment, or unsandboxed fuzzing.

## Objective

Validate that a governance artifact can pass through a controlled recursive learning loop:

```text
Governance artifact
  -> static validation
  -> governance simulation or threat scan
  -> fuzz/edge-case classification
  -> sanitized dataset row
  -> self-improvement recommendation
  -> next-cycle seed candidate
```

## Test scope

Supported first-wave artifacts:

- governance proposal JSON
- signer-rotation receipt JSON
- signer approval bundle JSON
- activation audit ledger JSON
- localnet render summary JSON
- funding deployment dry-run artifact

Out of scope:

- real signing
- real token issuance
- live chain deployment
- production keys
- unsandboxed fuzz execution
- non-owned or third-party targets

## Recursive learning invariants

A valid recursive learning cycle must satisfy all invariants:

1. **Deterministic input identity**: every input artifact has a stable SHA-256 digest.
2. **Sandbox boundary**: any execution command is represented as a dry-run or sandbox-required command.
3. **Policy gate**: every cycle has a `policy_decision` of `allow_dry_run`, `human_review_required`, or `blocked`.
4. **Dataset hygiene**: generated dataset rows do not include secrets, production keys, live endpoints, or raw private telemetry.
5. **Improvement traceability**: every self-improvement recommendation links to the input digest, finding ID, and next-cycle seed candidate.
6. **No autonomous promotion**: no generated seed, dataset, or patch is promoted without human review.

## Canonical test record

```json
{
  "schema_version": "peachtree.recursive_learning.v1",
  "cycle_id": "cycle-0001",
  "artifact_type": "governance_proposal",
  "artifact_path": "examples/proposals/emergency-pause.json",
  "artifact_sha256": "sha256:REPLACE_WITH_DIGEST",
  "validation_commands": [
    "make validate",
    "make governance-sim PROPOSAL=examples/proposals/emergency-pause.json",
    "make governance-threat-scan PROPOSAL=examples/proposals/emergency-pause.json"
  ],
  "sandbox_required": true,
  "policy_decision": "human_review_required",
  "findings": [
    {
      "finding_id": "gov-edge-0001",
      "severity": "medium",
      "class": "governance_edge_case",
      "summary": "Proposal requires deterministic simulation and threat-scan review before dataset promotion."
    }
  ],
  "dataset_candidate": {
    "dataset_name": "governance_fuzz_v1.jsonl",
    "privacy_tier": "restricted",
    "dataset_action": "human_review",
    "labels": ["governance", "fuzzing", "peachtree", "recursive-learning"]
  },
  "self_improvement": {
    "recommendation": "Add this artifact shape to the next governance fuzz seed set after maintainer review.",
    "next_cycle_seed_candidate": true,
    "requires_human_approval": true
  }
}
```

## Suggested local dry-run commands

```bash
make validate
make readiness
make governance-sim PROPOSAL=examples/proposals/emergency-pause.json
make governance-threat-scan PROPOSAL=examples/proposals/emergency-pause.json
make governance-queue REGISTRY=examples/proposals/registry.json
make governance-trends REGISTRY=examples/proposals/registry.json HISTORY=examples/proposals/history.json
```

## Acceptance criteria

- The recursive learning record has a stable schema version.
- Every artifact includes digest/provenance metadata.
- Every execution command is dry-run or sandbox-required.
- Every high-risk promotion is marked `human_review_required` or `blocked`.
- Dataset candidates include privacy, label, and promotion status fields.
- The final output includes a next-cycle recommendation but does not self-apply it.

## GitHub CI extension

A future CI job may validate fixture records with a command like:

```bash
python scripts/validate_peachtree_recursive_learning.py tests/fixtures/peachtree-recursive-learning/*.json
```

That validator should only check schema, policy posture, and redaction invariants. It should not run long fuzzing or training in pull-request CI.
