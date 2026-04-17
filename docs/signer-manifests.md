# Signer Manifests

`0aid signer-manifest` joins the checkpoint signer policy in
`config/governance/checkpoint-signers.json` with the active identity bootstrap
in `config/identity/bootstrap.json`.

The rendered manifest is meant to be the operator-facing source of truth for:

- which actor owns each active signer
- which governance execution roles each signer covers
- which governance phases depend on that role coverage
- when each signer must be rotated
- which signer rotations are current vs expiring

## Required signer metadata

Each signer entry must declare:

- `actor_id`
- `signer_id`
- `key_id`
- `status`
- `provisioned_at`
- `rotate_by`
- `roles`
- `shared_secret`

The current repository uses development-only HMAC signers for authenticated
pre-launch governance replay. This is intentionally not a production custody
design.

## Rotation policy

`rotation_policy.reference_time` is the deterministic point-in-time used for
validation and manifest rendering.

`rotation_policy.warning_window_days` controls when a signer is marked
`expiring`.

Active signers are rejected when:

- `rotate_by` is not after `provisioned_at`
- `rotate_by` is not after `rotation_policy.reference_time`
- two active signers claim the same `actor_id`
- two active signers cover the same governance execution role
- a required governance execution role has no active signer coverage

## Output shape

The manifest contains:

- signer ownership and organization context
- per-role governance references such as `governance:phase_owner:*`
- rotation status (`current`, `expiring`, `inactive`)
- `rotation_plan`, sorted by earliest `rotate_by`

Example:

```bash
./0aid signer-manifest --root . --out ./build/signer-manifest.json
```

This command should be treated as a release-gating check for governance
bootstrap changes. If it fails, the signer policy is no longer operationally
sound enough to exercise checkpoint replay or remediation signing.
