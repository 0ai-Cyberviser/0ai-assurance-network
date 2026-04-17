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
- deterministic `next_rotation_receipt_id` values
- deterministic `replacement_manifest_ref` paths

Example:

```bash
./0aid signer-manifest --root . --out ./build/signer-manifest.json
```

This command should be treated as a release-gating check for governance
bootstrap changes. If it fails, the signer policy is no longer operationally
sound enough to exercise checkpoint replay or remediation signing.

## Rotation receipts

`0aid signer-rotation-receipt` creates a machine-readable receipt stub for a
planned signer replacement.

The receipt includes:

- outgoing signer metadata
- incoming signer metadata
- approval actor requirements, resolved from `rotation_policy.approval_roles`
- `effective_at`
- a deterministic `replacement_manifest_ref`
- a full replacement-ready signer manifest preview

Example:

```bash
./0aid signer-rotation-receipt --root . \
  --outgoing-signer-id governance-chair-bot \
  --incoming-signer-id governance-chair-bot-v2 \
  --incoming-key-id governance-chair-dev-v2 \
  --effective-at 2026-04-24T00:00:00Z \
  --out ./build/rotation/governance-chair-receipt.json
```

The replacement preview only renders when:

- the incoming actor is active
- the incoming actor is actively bound to every replacement role
- the replacement does not reuse another active signer owner
- the replacement preserves all required governance execution role coverage
- `effective_at` is after the current reference time and on or before the
  outgoing signer `rotate_by`

## Operator workflow

1. Render the current signer manifest and identify any `expiring` signers.
2. Generate a signer rotation receipt stub for the outgoing signer.
3. Review the approval actor set and replacement manifest preview.
4. Collect governance approval against the receipt stub.
5. Publish the approved receipt together with the replacement signer manifest.
6. Update `config/governance/checkpoint-signers.json` only after the approved
   replacement manifest is ready to become the new active state.
