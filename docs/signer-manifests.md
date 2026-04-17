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

## Apply results

`0aid signer-rotation-apply` validates the activation plan against the current
policy lineage and emits a machine-readable apply result plus the exact
checkpoint signer policy that should become active.

Example:

```bash
./0aid signer-rotation-apply --root . \
  --plan ./build/rotation/governance-chair-activation-plan.json \
  --policy-out ./build/rotation/governance-chair-applied-policy.json \
  --out ./build/rotation/governance-chair-apply-result.json
```

The apply result includes:

- `activation_plan_digest`
- `applied_policy_digest`
- the exact `applied_policy`
- `target_policy_version`
- the deterministic effective cutover time

Apply fails closed when:

- the activation plan no longer matches the current policy lineage
- the resulting policy payload differs from the plan output
- the resulting policy would fail signer-manifest validation

## Post-activation verification receipts

`0aid signer-rotation-verify` signs a post-activation verification receipt
against the applied checkpoint signer policy using the newly activated signer.

Example:

```bash
./0aid signer-rotation-verify --root . \
  --plan ./build/rotation/governance-chair-activation-plan.json \
  --policy ./build/rotation/governance-chair-applied-policy.json \
  --verified-at 2026-04-24T00:15:00Z \
  --out ./build/rotation/governance-chair-verification.json
```

Verification receipts include:

- the activation plan digest
- the applied policy digest
- the signer/key that verified the activation
- actor and organization ownership context
- a signed verification receipt with validity bounds

Verification fails closed when:

- the applied policy drifts from the activation plan
- the outgoing signer is still present
- the incoming signer is missing
- the incoming signer lacks a shared secret
- `verified_at` falls before `effective_at`, before `provisioned_at`, or after
  the incoming signer `rotate_by`

## Activation audit ledger

`0aid signer-rotation-ledger-append` appends a verified activation record into
an append-only audit ledger.

Example:

```bash
./0aid signer-rotation-ledger-append \
  --apply ./build/rotation/governance-chair-apply-result.json \
  --verification ./build/rotation/governance-chair-verification.json \
  --ledger-out ./build/rotation/activation-audit-ledger.json \
  --out ./build/rotation/governance-chair-ledger-append.json
```

The append result includes:

- the appended audit entry
- the updated ledger
- a stable append index

Each audit entry binds:

- `receipt_id`
- `target_policy_version`
- `activation_plan_digest`
- `applied_policy_digest`
- `effective_at`
- `verified_at`
- the verifying signer/key and actor ownership context
- the verification signature envelope

Ledger append fails closed when:

- the apply result and verification receipt disagree on receipt or policy digests
- the verification signature metadata is missing
- the same `receipt_id`, `target_policy_version`, or `signature_id` already exists
- the new record would move `effective_at` or `verified_at` backward
- an existing ledger entry is already malformed

## Ledger reconciliation

`0aid signer-rotation-ledger-reconcile` compares the activation audit ledger
against a current checkpoint signer policy and reports whether the active policy
is fully explained by the recorded ledger lineage.

Example:

```bash
./0aid signer-rotation-ledger-reconcile \
  --ledger ./build/rotation/activation-audit-ledger.json \
  --policy ./build/rotation/governance-chair-applied-policy.json \
  --out ./build/rotation/governance-chair-ledger-reconcile.json
```

The reconciliation report includes:

- the current policy version and digest
- the latest recorded receipt id, target policy version, and applied policy digest
- whether the current policy is explained by the ledger
- continuity issues such as missing coverage, duplicates, or lineage mismatch

Reconciliation surfaces gaps when:

- the ledger is empty but the current policy already appears rotated
- the latest recorded target policy version does not match the current policy
- the latest recorded applied policy digest does not match the current policy
- duplicate receipt ids, target policy versions, or signature ids exist in the ledger
- `effective_at` or `verified_at` ordering is not strictly increasing across entries

## Ledger export packages

`0aid signer-rotation-ledger-export` packages the current checkpoint signer
policy, the activation audit ledger, and the reconciliation report into one
stable export artifact for offline review or retention.

Example:

```bash
./0aid signer-rotation-ledger-export \
  --ledger ./build/rotation/activation-audit-ledger.json \
  --policy ./build/rotation/governance-chair-applied-policy.json \
  --out ./build/rotation/governance-chair-audit-export.json
```

The export package includes:

- the current checkpoint signer policy and digest
- the append-only activation audit ledger
- the reconciliation report
- a baseline snapshot with the latest ledger lineage, current policy digest,
  and continuity status

Operators may also pass `--reconcile` to reuse a saved reconciliation report.
Export still fails closed when that report contradicts the supplied ledger or
current policy digest.

## Export verification

`0aid signer-rotation-ledger-verify-export` validates one exported audit
package for archive retention.

Example:

```bash
./0aid signer-rotation-ledger-verify-export \
  --export ./build/rotation/governance-chair-audit-export.json \
  --out ./build/rotation/governance-chair-audit-export-verify.json
```

Verification reruns reconciliation from the embedded ledger and policy,
rebuilds the expected export package, and compares the regenerated payload with
the archived artifact. The verification report marks a bundle as
`archive_ready` only when the recomputed export matches exactly and the
embedded reconciliation status is still `consistent`.

## Archive index manifests

`0aid signer-rotation-ledger-archive-index` builds a compact archive manifest
over a retained set of export packages.

Example:

```bash
./0aid signer-rotation-ledger-archive-index \
  --exports ./build/rotation/current-audit-export.json,./build/rotation/governance-chair-audit-export.json \
  --out ./build/rotation/activation-audit-archive-index.json
```

The archive index includes:

- package path, policy version, digest, and latest receipt lineage for each retained export
- per-package archive readiness and verification issues
- aggregate chain and policy path continuity
- latest retained baseline summary

Index generation fails closed when retained packages overlap on the same
current policy version, reuse the same latest receipt id, or disagree on chain
or policy path metadata.

## Archive promotion receipts

`0aid signer-rotation-ledger-archive-promote` emits a deterministic receipt
proving that one verified export package was formally promoted into retained
archive state.

Example:

```bash
./0aid signer-rotation-ledger-archive-promote \
  --export ./build/rotation/governance-chair-audit-export.json \
  --verify ./build/rotation/governance-chair-audit-export-verify.json \
  --archive-index ./build/rotation/activation-audit-archive-index.json \
  --promoted-at 2026-04-24T01:00:00Z \
  --out ./build/rotation/governance-chair-archive-promotion.json
```

Promotion fails closed unless:

- the supplied verification report exactly matches a recomputed verification of
  the export package
- the export package is still `archive_ready`
- the archive index is internally consistent and in `consistent` status
- the archive index contains an entry for the package path whose metadata
  exactly matches the verified export lineage

The promotion receipt binds:

- the promoted package path
- current policy version and digest
- baseline lineage (`latest_receipt_id`, `latest_target_policy_version`)
- digests for the export package, baseline snapshot, verification report,
  archive index, and matched archive entry

## Retained baseline attestations

`0aid signer-rotation-ledger-attest-retained-baseline` emits a machine-readable
attestation tied to one retained archive entry and its promotion receipt.

Example:

```bash
./0aid signer-rotation-ledger-attest-retained-baseline \
  --promotion-receipt ./build/rotation/governance-chair-archive-promotion.json \
  --attested-at 2026-04-24T01:05:00Z \
  --out ./build/rotation/governance-chair-retained-baseline-attestation.json
```

The attestation records:

- the promotion receipt id and digest
- the retained package path and policy lineage
- digests for the promoted export package, baseline snapshot, verification
  report, archive index, and archive entry
- promotion and attestation timestamps

Attestation fails closed when the promotion receipt is not already in retained
state or when the attestation timestamp predates the promotion timestamp.

## Operator workflow

1. Render the current signer manifest and identify any `expiring` signers.
2. Generate a signer rotation receipt stub for the outgoing signer.
3. Review the approval actor set and replacement manifest preview.
4. Collect signed approval artifacts for every required approval role.
5. Finalize the receipt bundle and verify it remains bound to the replacement
   manifest preview.
6. Render the activation plan and resulting checkpoint signer policy payload.
7. Apply the activation plan and emit the exact replacement
   `checkpoint-signers.json` payload.
8. Publish the approved bundle together with the replacement signer manifest.
9. Update `config/governance/checkpoint-signers.json` only after the approved
   replacement manifest is ready to become the new active state.
10. Sign and retain a post-activation verification receipt proving the new
    signer set became active exactly as planned.
11. Append the apply + verification outputs into the activation audit ledger so
    the signer lineage stays append-only and reviewable over time.
12. Reconcile the current checkpoint signer policy against that ledger before
    treating the rotation lineage as continuous and complete.
13. Export the policy, ledger, and reconciliation state into a stable audit
    package before archiving the rotation baseline.
14. Verify that archived export package before accepting it as a retained baseline.
15. Rebuild the archive index manifest so the candidate retained baseline is
    represented as a concrete archive entry.
16. Promote the verified export package into retained archive state using that
    archive index entry.
17. Emit a retained-baseline attestation tied to the promotion receipt and
    archive entry lineage.
