# 0AI Assurance Network

[![CI](https://github.com/0ai-Cyberviser/0ai-assurance-network/actions/workflows/ci.yml/badge.svg)](https://github.com/0ai-Cyberviser/0ai-assurance-network/actions/workflows/ci.yml)

This directory is the standalone seed repository for the future `0AI Assurance
Network` appchain. It is separate from the earlier concept package in
[`initiatives/0ai-assurance-network`](../../initiatives/0ai-assurance-network)
and is intended to stand on its own as the controlled codebase for network
design, governance tooling, and bootstrap operations.

The emphasis here is operational:

- permissioned testnet topology
- base genesis parameters
- config validation
- localnet render output
- operator runbook scaffolding

This is still not a launched chain. It is a controlled build target for the
next engineering phase.

## Ownership And Scope

- owner: `Johnny Watters` (`0ai-Cyberviser`)
- company: `0AI`
- project type: pre-launch network and governance tooling
- live token status: no token or public chain has been launched from this repo

The code, docs, branding, and launch design in this repository are controlled
by `Johnny Watters`. This repository does not itself create a live cryptoasset,
grant protocol equity, or authorize a public token sale.

Canonical repository:

- https://github.com/0ai-Cyberviser/0ai-assurance-network

## Layout

```text
projects/0ai-assurance-network/
├── Makefile
├── README.md
├── cmd/0aid/
├── config/
│   ├── genesis/base-genesis.json
│   ├── governance/checkpoint-signers.json
│   ├── governance/inference-policy.json
│   ├── identity/bootstrap.json
│   ├── modules/milestone-1.json
│   ├── network-topology.json
│   └── policy/release-guards.json
├── docs/
├── examples/proposals/
├── internal/project/
├── scripts/
    ├── check_configs.py
    └── render_localnet.py
├── src/assurancectl/
└── tests/
```

## Commands

```bash
make -C projects/0ai-assurance-network validate
make -C projects/0ai-assurance-network render-localnet
make -C projects/0ai-assurance-network readiness
make -C projects/0ai-assurance-network governance-sim PROPOSAL=examples/proposals/treasury-grant.json
make -C projects/0ai-assurance-network governance-queue REGISTRY=examples/proposals/registry.json
make -C projects/0ai-assurance-network governance-trends REGISTRY=examples/proposals/registry.json HISTORY=examples/proposals/history.json
make -C projects/0ai-assurance-network governance-remediation REGISTRY=examples/proposals/registry.json HISTORY=examples/proposals/history.json
make -C projects/0ai-assurance-network governance-replay STATUS=examples/proposals/checkpoint-events.json
make -C projects/0ai-assurance-network governance-drift PROPOSAL=examples/proposals/emergency-pause.json HISTORY=examples/proposals/history.json
make -C projects/0ai-assurance-network go-build
make -C projects/0ai-assurance-network go-test
make -C projects/0ai-assurance-network module-plan
make -C projects/0ai-assurance-network identity-plan
make -C projects/0ai-assurance-network signer-manifest
make -C projects/0ai-assurance-network signer-rotation-receipt OUTGOING_SIGNER_ID=governance-chair-bot INCOMING_SIGNER_ID=governance-chair-bot-v2 INCOMING_KEY_ID=governance-chair-dev-v2 EFFECTIVE_AT=2026-04-24T00:00:00Z
make -C projects/0ai-assurance-network signer-rotation-approve RECEIPT=build/rotation/governance-chair-receipt.json ROLE=governance-ops SIGNER_ID=governance-ops-bot APPROVED_AT=2026-04-23T00:00:00Z
make -C projects/0ai-assurance-network signer-rotation-finalize RECEIPT=build/rotation/governance-chair-receipt.json APPROVALS=build/rotation/governance-chair-governance-ops.json,build/rotation/governance-chair-token-house.json,build/rotation/governance-chair-telemetry.json
make -C projects/0ai-assurance-network signer-rotation-activate BUNDLE=build/rotation/governance-chair-approved-bundle.json INCOMING_SHARED_SECRET=dev-secret-governance-chair-v2
make -C projects/0ai-assurance-network signer-rotation-apply PLAN=build/rotation/governance-chair-activation-plan.json POLICY_OUT=build/rotation/governance-chair-applied-policy.json
make -C projects/0ai-assurance-network signer-rotation-verify PLAN=build/rotation/governance-chair-activation-plan.json POLICY=build/rotation/governance-chair-applied-policy.json VERIFIED_AT=2026-04-24T00:15:00Z
make -C projects/0ai-assurance-network signer-rotation-ledger-append APPLY=build/rotation/governance-chair-apply-result.json VERIFICATION=build/rotation/governance-chair-verification.json LEDGER_OUT=build/rotation/activation-audit-ledger.json
make -C projects/0ai-assurance-network signer-rotation-ledger-reconcile LEDGER=build/rotation/activation-audit-ledger.json POLICY=build/rotation/governance-chair-applied-policy.json
make -C projects/0ai-assurance-network signer-rotation-ledger-export LEDGER=build/rotation/activation-audit-ledger.json POLICY=build/rotation/governance-chair-applied-policy.json OUT=build/rotation/governance-chair-audit-export.json
make -C projects/0ai-assurance-network signer-rotation-ledger-verify-export EXPORT=build/rotation/governance-chair-audit-export.json OUT=build/rotation/governance-chair-audit-export-verify.json
make -C projects/0ai-assurance-network signer-rotation-ledger-archive-index EXPORTS=build/rotation/current-audit-export.json,build/rotation/governance-chair-audit-export.json OUT=build/rotation/activation-audit-archive-index.json
make -C projects/0ai-assurance-network init-node ID=val-3
make -C projects/0ai-assurance-network collect-validator BUNDLE=build/nodes/val-3 OUT=build/collection/val-3.json
make -C projects/0ai-assurance-network assemble-genesis COLLECTION=build/collection OUT=build/assembled/genesis-plan.json
make -C projects/0ai-assurance-network assemble-localnet COLLECTION=build/collection OUT=build/assembled
```

`validate` checks that the topology, genesis, and governance policy files are
internally consistent.

`render-localnet` generates:

- `build/localnet/docker-compose.yml`
- `build/localnet/network-summary.json`
- `build/localnet/genesis.rendered.json`

`readiness` scores the current launch posture and calls out blockers before the
project pretends it is closer to launch than it really is.

`governance-sim` runs the explainable governance inference engine against a
proposal JSON document and outputs proposal class, risk score, required houses,
and remediation.

`governance-queue` scores a registry of proposals, sorts them by urgency and
proposal class, and gives operators a queue view for governance review.

`governance-drift` compares a proposal against recorded governance history and
flags repeated emergency use, adverse precedents, treasury growth, and other
pattern-level risks that one-shot proposal scoring would miss. It also clusters
precedent by proposal kind and requester, and can suppress recurring-pattern
signals when the historical cluster is stable and clean.

`governance-trends` clusters the entire active queue into portfolio-level
governance patterns so operators can spot systemic validator churn, recurring
emergency governance, or treasury growth concentration across proposals. It now
uses time-windowed baselines so recent activity can be compared against prior
cluster history instead of treated as flat repetition, and it reports
kind-level seasonal pressure so operators can see when a proposal kind is above
its normal cadence.

`governance-remediation` turns those trend clusters into structured mitigation
bundles with severity, release blockers, immediate actions, approval
guardrails, monitoring steps, and machine-readable checkpoints with explicit
owner roles, eligible bootstrap actors, assigned actor context, completion
criteria, dependency ordering between phases, and status-aware rollups for
current execution readiness. That keeps the governance path operationally
useful once the engine detects an unstable cluster.

Each governance command also supports `--artifact-out <path>` so external
automation can consume a stable, versioned orchestration artifact without
scraping CLI text. The artifact contract is documented in
[docs/orchestration-artifacts.md](docs/orchestration-artifacts.md).

`governance-replay` reconstructs deterministic current checkpoint state from an
append-only event log or a derived snapshot. Event logs are stricter than
snapshots: each event must include `checkpoint_id`, `previous_status`,
`new_status`, `updated_at`, `recorded_by`, `actor_id`, and `rationale`, and
replay rejects duplicate, contradictory, illegal, or out-of-order history.
Signed event logs must also include a `signature` object with:

- `format`
- `signer_id`
- `key_id`
- `signature_id`
- `signed_at`
- `expires_at`
- `value`

Signer-to-role policy lives in
`config/governance/checkpoint-signers.json`. The current repository ships
development-only HMAC signers so the pre-launch testnet can exercise
authenticated updates without pretending to solve production custody. Each
signer is bound to an `actor_id`, and replay only accepts signed updates when
that actor is active in `config/identity/bootstrap.json` and actively bound to
the claimed role.

Status inputs are transition-aware. Each non-pending checkpoint update should
include `previous_status`, `updated_at`, `recorded_by`, and `actor_id`, and the
engine only accepts lifecycle moves that stay inside the deterministic path
`pending -> in_progress -> completed`. It also validates dependency timestamp
ordering so downstream checkpoints cannot appear to complete before the latest
completed prerequisite. Remediation accepts either a derived checkpoint snapshot
or a replayable append-only event log via `--status`, and surfaces resolved
actor/org context in the machine-readable payload so operators can audit who is
actually assigned to each governance checkpoint.

For event logs, non-pending updates are only accepted when the signature is
valid, the signer is authorized for the `recorded_by` role, and the update falls
inside the signature validity window. Reused `signature_id` values are rejected
as replay attempts.

Rejected or expired signatures are surfaced as event alerts during
`governance-replay`, and any remediation run consuming that event log moves the
affected cluster to `current_release_readiness = invalid` until the bad update
is replaced with a new signed record.

`go-build` and `go-test` operate on the `0aid` binary skeleton and the internal
Go project package.

`signer-manifest` renders the active checkpoint signer ownership map and
rotation plan from the signer policy and identity bootstrap. It fails closed if
rotation metadata is stale, if two active signers claim the same actor
ownership, or if a governance execution role has no active signer coverage. The
rotation contract is described in [docs/signer-manifests.md](docs/signer-manifests.md).

`signer-rotation-receipt` renders a machine-readable rotation receipt stub for
one outgoing signer plus a replacement-ready signer manifest preview. The stub
records the outgoing signer, the incoming signer/key, approval actors, the
effective cutover time, and the manifest path the operator should publish with
the receipt. Rotation receipts fail closed when the replacement would leave a
required governance role uncovered, reuse another active actor owner, or use an
invalid effective-at ordering.

`signer-rotation-approve` creates a signed approval artifact for one required
approval role on top of a receipt stub. It fails closed when the signer is not
eligible for that role, when the receipt has drifted from current config, or
when the approval timestamp falls outside signer validity or after the planned
cutover.

`signer-rotation-finalize` validates the full approval set and emits a finalized
bundle that can be published together with the replacement manifest. It rejects
missing roles, duplicate approvers, receipt-digest drift, or invalid approval
signatures.

`signer-rotation-activate` consumes an approved bundle and emits a deterministic
activation plan plus a machine-readable replacement policy for
`config/governance/checkpoint-signers.json`. It fails closed when the approved
bundle has drifted from the current policy lineage or when the incoming signer
secret is missing.

`signer-rotation-apply` validates the activation plan against the current
policy lineage and emits the exact applied checkpoint signer policy plus stable
digests for the plan and policy payload. Operators can direct the resulting
policy into a standalone file with `--policy-out` or write that file directly
to `config/governance/checkpoint-signers.json`.

`signer-rotation-verify` signs a post-activation verification receipt against
the applied policy using the newly activated signer. Verification fails closed
when the applied policy drifts from the activation plan, the outgoing signer is
still present, the incoming signer is missing, or the verification timestamp
falls outside the new signer validity window.

`signer-rotation-ledger-append` binds the apply result and verification receipt
into an append-only activation audit ledger. It rejects mismatched receipt or
policy digests, duplicate receipt IDs, duplicate target policy versions,
replayed verification signatures, and non-monotonic effective or verified
timestamps.

`signer-rotation-ledger-reconcile` reads the activation audit ledger against a
current checkpoint signer policy and reports whether the active policy can be
explained by the latest recorded activation. It surfaces missing ledger
coverage, duplicate continuity records, and any current policy version or
digest that no longer matches the recorded ledger lineage.

`signer-rotation-ledger-export` packages that same ledger lineage into a stable
offline review artifact. The export bundles:

- the current checkpoint signer policy and digest
- the append-only activation audit ledger
- the reconciliation report
- a baseline snapshot covering the latest ledger lineage and current continuity state

Exports fail closed when the reconciliation report contradicts the ledger or
current policy digests, so operators cannot archive a self-inconsistent bundle
as if it were a valid baseline.

`signer-rotation-ledger-verify-export` replays that package deterministically
and emits a machine-readable archive readiness report. Verification stays local
to the exported payload: it reruns reconciliation, regenerates the expected
baseline snapshot, and rejects any bundle whose embedded digests or lineage
metadata drift from the recomputed state.

`signer-rotation-ledger-archive-index` builds a compact archive manifest over a
set of retained export packages. The index summarizes current policy versions,
latest receipt lineage, archive readiness, and duplicate or contradictory
baseline metadata across the retained package set.

The generated compose file assumes a future `0aid` chain binary packaged in a
container image. It is intentionally parameterized so the image and binary path
can change without rewriting topology data.

The `0aid` binary now exists as a narrow operator-facing Go entrypoint. It
currently supports:

- `version`
- `module-map`
- `module-plan`
- `identity-plan`
- `signer-manifest`
- `signer-rotation-receipt`
- `signer-rotation-approve`
- `signer-rotation-finalize`
- `signer-rotation-activate`
- `signer-rotation-apply`
- `signer-rotation-verify`
- `signer-rotation-ledger-append`
- `signer-rotation-ledger-reconcile`
- `signer-rotation-ledger-export`
- `signer-rotation-ledger-verify-export`
- `signer-rotation-ledger-archive-index`
- `show-plan`
- `init-genesis`
- `render-validator`
- `render-identity`
- `init-node`
- `collect-validator`
- `assemble-genesis`
- `assemble-localnet`

Bootstrap examples:

```bash
./0aid render-identity --root . --id val-3
./0aid module-plan --root . --out ./build/module-plan.json
./0aid identity-plan --root . --out ./build/identity-plan.json
./0aid signer-manifest --root . --out ./build/signer-manifest.json
./0aid signer-rotation-receipt --root . \
  --outgoing-signer-id governance-chair-bot \
  --incoming-signer-id governance-chair-bot-v2 \
  --incoming-key-id governance-chair-dev-v2 \
  --effective-at 2026-04-24T00:00:00Z \
  --out ./build/rotation/governance-chair-receipt.json
./0aid signer-rotation-approve --root . \
  --receipt ./build/rotation/governance-chair-receipt.json \
  --role governance-ops \
  --signer-id governance-ops-bot \
  --approved-at 2026-04-23T00:00:00Z \
  --out ./build/rotation/governance-chair-governance-ops.json
./0aid signer-rotation-finalize --root . \
  --receipt ./build/rotation/governance-chair-receipt.json \
  --approvals ./build/rotation/governance-chair-governance-ops.json,./build/rotation/governance-chair-token-house.json,./build/rotation/governance-chair-telemetry.json \
  --out ./build/rotation/governance-chair-approved-bundle.json
./0aid signer-rotation-activate --root . \
  --bundle ./build/rotation/governance-chair-approved-bundle.json \
  --incoming-shared-secret dev-secret-governance-chair-v2 \
  --out ./build/rotation/governance-chair-activation-plan.json
./0aid signer-rotation-apply --root . \
  --plan ./build/rotation/governance-chair-activation-plan.json \
  --policy-out ./build/rotation/governance-chair-applied-policy.json \
  --out ./build/rotation/governance-chair-apply-result.json
./0aid signer-rotation-verify --root . \
  --plan ./build/rotation/governance-chair-activation-plan.json \
  --policy ./build/rotation/governance-chair-applied-policy.json \
  --verified-at 2026-04-24T00:15:00Z \
  --out ./build/rotation/governance-chair-verification.json
./0aid signer-rotation-ledger-append \
  --apply ./build/rotation/governance-chair-apply-result.json \
  --verification ./build/rotation/governance-chair-verification.json \
  --ledger-out ./build/rotation/activation-audit-ledger.json \
  --out ./build/rotation/governance-chair-ledger-append.json
./0aid signer-rotation-ledger-reconcile \
  --ledger ./build/rotation/activation-audit-ledger.json \
  --policy ./build/rotation/governance-chair-applied-policy.json \
  --out ./build/rotation/governance-chair-ledger-reconcile.json
./0aid init-node --root . --id val-3 --out ./build/nodes/validator-3
./0aid collect-validator --bundle ./build/nodes/validator-3 --out ./build/collection/validator-3.json
./0aid assemble-genesis --root . --collection ./build/collection --out ./build/assembled/genesis-plan.json
./0aid assemble-localnet --root . --collection ./build/collection --out ./build/assembled
PYTHONPATH=src python -m assurancectl.cli governance-sim \
  --proposal examples/proposals/emergency-pause.json
PYTHONPATH=src python -m assurancectl.cli governance-queue \
  --registry examples/proposals/registry.json
PYTHONPATH=src python -m assurancectl.cli governance-trends \
  --registry examples/proposals/registry.json \
  --history examples/proposals/history.json
PYTHONPATH=src python -m assurancectl.cli governance-remediation \
  --registry examples/proposals/registry.json \
  --history examples/proposals/history.json
PYTHONPATH=src python -m assurancectl.cli governance-remediation \
  --registry examples/proposals/registry.json \
  --history examples/proposals/history.json \
  --status examples/proposals/checkpoint-status.json
PYTHONPATH=src python -m assurancectl.cli governance-remediation \
  --registry examples/proposals/registry.json \
  --history examples/proposals/history.json \
  --artifact-out build/artifacts/governance-remediation.json
PYTHONPATH=src python -m assurancectl.cli governance-replay \
  --status examples/proposals/checkpoint-events.json
PYTHONPATH=src python -m assurancectl.cli governance-drift \
  --proposal examples/proposals/emergency-pause.json \
  --history examples/proposals/history.json
```

The identity and node-init paths are explicitly development-only. They emit
deterministic placeholder identity material so the local bootstrap flow can
advance without pretending to solve production validator custody.

The collection and assembly path is deterministic by design. Operators collect
each validator bundle into a normalized manifest, merge the full set into a
single genesis plan, and then render an assembled localnet bundle from that
same collection. Assembly fails closed on duplicate validator IDs, partial
collections, topology mismatches, and unexpected voting power.

The governance inference path is explicitly advisory. It helps classify and
score proposals, but it does not vote, sign, or bypass human review. The new
history-aware drift pass keeps that same posture while adding context from prior
governance behavior.

The first implementable chain milestone for registry and attestation scope is
captured in [config/modules/milestone-1.json](config/modules/milestone-1.json)
and documented in [docs/module-milestone-1.md](docs/module-milestone-1.md).
`0aid module-plan` renders that milestone with the current chain and validator
context so the next chain-code step has a bounded implementation sequence.

The permissioned actor and role bootstrap for that milestone is versioned in
[config/identity/bootstrap.json](config/identity/bootstrap.json). `0aid
identity-plan` renders the active actor set, required roles, role coverage, and
missing-role check so registry, attestation, governance, and validator rollout
hooks all depend on the same operator identity surface.

## Current Design

- `7` validators
- permissioned initial validator set
- dual-house governance metadata carried in config
- no public sale or transferability assumptions
- release guarded by legal, audit, and incident-response prerequisites
- launch-readiness report built into the operator CLI

## Next Engineering Step

Promote this skeleton into its own repository and add:

1. deterministic keygen and validator secrets handling
2. genesis init and validator config templates
3. module-specific state and CLI wiring
4. registry and attestation modules
5. external audit pipeline
