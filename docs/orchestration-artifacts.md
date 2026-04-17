# Governance Orchestration Artifacts

`assurancectl` governance commands can emit a versioned artifact file for
external automation:

- `governance-sim`
- `governance-queue`
- `governance-trends`
- `governance-remediation`
- `governance-replay`
- `governance-drift`

Use `--artifact-out <path>` alongside the normal command arguments.

Example:

```bash
PYTHONPATH=src python -m assurancectl.cli governance-remediation \
  --registry examples/proposals/registry.json \
  --history examples/proposals/history.json \
  --artifact-out build/artifacts/governance-remediation.json
```

## Schema Contract

Every artifact file uses the same top-level envelope:

```json
{
  "schema": "0ai.assurance.governance.artifact",
  "schema_version": "1.0.0",
  "compatibility": {
    "breaking_change": "increment major",
    "additive_change": "increment minor",
    "clarification_change": "increment patch"
  },
  "artifact_type": "governance_remediation",
  "command": "governance-remediation",
  "sources": {
    "registry": "examples/proposals/registry.json",
    "history": "examples/proposals/history.json"
  },
  "payload": []
}
```

## Compatibility Rules

- Major version changes are required for breaking schema changes.
- Minor version changes are required for additive fields or new artifact types.
- Patch version changes are limited to non-structural clarifications.

External automation should treat `schema` + `schema_version` as the contract
key. Consumers that only support `1.x` should reject artifacts with a different
major version.

## Operator Behavior

- Human-readable CLI output remains the default.
- `--json` still prints the direct payload to stdout.
- `--artifact-out` writes the versioned artifact envelope to disk without
  changing the stdout format.

## Consumer Example

See [examples/orchestrators/consume_governance_artifact.py](../examples/orchestrators/consume_governance_artifact.py)
for a small consumer that validates the schema version and extracts
remediation-blocking clusters.
