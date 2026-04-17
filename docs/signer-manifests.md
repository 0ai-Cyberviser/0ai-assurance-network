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

## Signed approvals

`0aid signer-rotation-approve` turns a receipt stub into a signed approval
artifact for one required approval role.

Example:

```bash
./0aid signer-rotation-approve --root . \
  --receipt ./build/rotation/governance-chair-receipt.json \
  --role governance-ops \
  --signer-id governance-ops-bot \
  --approved-at 2026-04-23T00:00:00Z \
  --out ./build/rotation/governance-chair-governance-ops.json
```

Approval artifacts are bound to the exact receipt by `receipt_id`,
`receipt_digest`, and `replacement_manifest_ref`, and they are signed with the
configured development HMAC signer for the selected approval role.

Approvals are rejected when:

- the signer is not eligible for the requested approval role
- the receipt has drifted from the current checkpoint signer configuration
- `approved_at` falls before signer `provisioned_at` or after signer `rotate_by`
- `approved_at` falls after the receipt `effective_at`

## Finalized bundles

`0aid signer-rotation-finalize` validates the full approval set and emits a
publication-ready bundle containing:

- the original receipt stub
- the verified approval artifacts
- the replacement signer manifest preview
- final bundle metadata such as `status = approved` and `finalized_at`

Example:

```bash
./0aid signer-rotation-finalize --root . \
  --receipt ./build/rotation/governance-chair-receipt.json \
  --approvals ./build/rotation/governance-chair-governance-ops.json,./build/rotation/governance-chair-token-house.json,./build/rotation/governance-chair-telemetry.json \
  --out ./build/rotation/governance-chair-approved-bundle.json
```

Finalization fails closed when:

- any required approval role is missing
- the same signer or actor tries to satisfy multiple approval roles
- approval signatures do not validate against the configured signer secret
- approval artifacts reference a different receipt digest or replacement manifest

## Activation plans

`0aid signer-rotation-activate` consumes an approved bundle and emits the next
operator artifact: a deterministic activation plan plus a full replacement
policy payload for `config/governance/checkpoint-signers.json`.

Example:

```bash
./0aid signer-rotation-activate --root . \
  --bundle ./build/rotation/governance-chair-approved-bundle.json \
  --incoming-shared-secret dev-secret-governance-chair-v2 \
  --out ./build/rotation/governance-chair-activation-plan.json
```

The activation plan includes:

- the current and target checkpoint signer policy versions
- a policy patch summary for removing the outgoing signer and adding the approved replacement
- a fully rendered resulting checkpoint signer policy
- ordered activation steps for publishing and applying the approved rotation

Activation fails closed when:

- the approved bundle no longer matches the current signer policy lineage
- the outgoing signer is already absent from the current policy
- the incoming signer already exists in the current policy
- the incoming shared secret is missing
- the resulting policy would fail signer-manifest validation

## Operator workflow

1. Render the current signer manifest and identify any `expiring` signers.
2. Generate a signer rotation receipt stub for the outgoing signer.
3. Review the approval actor set and replacement manifest preview.
4. Collect signed approval artifacts for every required approval role.
5. Finalize the receipt bundle and verify it remains bound to the replacement
   manifest preview.
6. Render the activation plan and resulting checkpoint signer policy payload.
7. Publish the approved bundle together with the replacement signer manifest.
8. Update `config/governance/checkpoint-signers.json` only after the approved
   replacement manifest is ready to become the new active state.
