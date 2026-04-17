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
│   ├── governance/inference-policy.json
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
make -C projects/0ai-assurance-network governance-drift PROPOSAL=examples/proposals/emergency-pause.json HISTORY=examples/proposals/history.json
make -C projects/0ai-assurance-network go-build
make -C projects/0ai-assurance-network go-test
make -C projects/0ai-assurance-network init-node ID=val-3
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
owner roles, completion criteria, dependency ordering between phases, and
status-aware rollups for current execution readiness. That keeps the governance
path operationally useful once the engine detects an unstable cluster.

Status files are transition-aware. Each non-pending checkpoint update should
include `previous_status`, `updated_at`, and `recorded_by`, and the engine only
accepts lifecycle moves that stay inside the deterministic path
`pending -> in_progress -> completed`. It also validates dependency timestamp
ordering so downstream checkpoints cannot appear to complete before the latest
completed prerequisite.

`go-build` and `go-test` operate on the `0aid` binary skeleton and the internal
Go project package.

The generated compose file assumes a future `0aid` chain binary packaged in a
container image. It is intentionally parameterized so the image and binary path
can change without rewriting topology data.

The `0aid` binary now exists as a narrow operator-facing Go entrypoint. It
currently supports:

- `version`
- `module-map`
- `show-plan`
- `init-genesis`
- `render-validator`
- `render-identity`
- `init-node`

Bootstrap examples:

```bash
./0aid render-identity --root . --id val-3
./0aid init-node --root . --id val-3 --out ./build/nodes/validator-3
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
PYTHONPATH=src python -m assurancectl.cli governance-drift \
  --proposal examples/proposals/emergency-pause.json \
  --history examples/proposals/history.json
```

The identity and node-init paths are explicitly development-only. They emit
deterministic placeholder identity material so the local bootstrap flow can
advance without pretending to solve production validator custody.

The governance inference path is explicitly advisory. It helps classify and
score proposals, but it does not vote, sign, or bypass human review. The new
history-aware drift pass keeps that same posture while adding context from prior
governance behavior.

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
