# Governance Inference

## Purpose

The governance inference layer gives `0AI Assurance Network` a structured
decision-support engine for proposal review. It is intentionally explainable.

The current engine:

- reads a proposal JSON document
- classifies it as `standard`, `high_impact`, or `safety_critical`
- scores governance risk
- determines the required approval path
- produces a compact inference summary with remediation items
- can compare a proposal against recorded governance history and flag drift

This is not autonomous governance. It is inference support for human operators.

## Why Start with an Explainable Engine

Using a narrow, explainable engine first is the right engineering move:

- deterministic outputs are easier to test and audit
- governance reviewers can challenge the exact factors that drove a result
- later LLM or model-backed summarization can be added without giving up a
  stable control baseline

## Inputs

The engine uses:

- `config/governance/inference-policy.json`
- proposal JSON files such as:
  - `examples/proposals/treasury-grant.json`
  - `examples/proposals/emergency-pause.json`
- optional governance history in:
  - `examples/proposals/history.json`

## Command

```bash
PYTHONPATH=src python -m assurancectl.cli governance-sim \
  --proposal examples/proposals/emergency-pause.json
```

Or through make:

```bash
make governance-sim PROPOSAL=examples/proposals/emergency-pause.json
make governance-queue REGISTRY=examples/proposals/registry.json
make governance-trends REGISTRY=examples/proposals/registry.json HISTORY=examples/proposals/history.json
make governance-remediation REGISTRY=examples/proposals/registry.json HISTORY=examples/proposals/history.json
make governance-drift PROPOSAL=examples/proposals/emergency-pause.json HISTORY=examples/proposals/history.json
```

## Output

The report includes:

- proposal class
- confidence
- risk score
- required houses
- recommended disposition
- triggered signals
- rationale
- remediation items

For queue review, `governance-queue` reads `examples/proposals/registry.json`,
simulates every pending proposal, and sorts the queue by:

1. priority
2. proposal class
3. risk score

That gives operators a compact review list for governance meetings and incident
response windows.

With governance history enabled, the engine also evaluates:

- repeated emergency usage
- adverse outcomes for the same proposal kind
- treasury growth relative to historical precedent
- concentration of high-impact actions from one requester
- repeated validator-impacting governance churn
- stable-pattern suppression when a precedent cluster is routine and clean

That keeps the inference path explainable while adding historical context.

At the portfolio level, `governance-trends` groups the active queue into
cluster summaries keyed by proposal kind and requester. That lets operators see
whether the current governance load is dominated by one unstable cluster, even
before reading every proposal in detail.

The trend layer now uses time-windowed baselines:

- a recent window for current activity
- a broader baseline window for prior precedent
- an acceleration ratio threshold to decide whether governance churn is rising

That lets the queue distinguish ordinary quarterly maintenance from a cluster
that is suddenly compressing into a short time period.

On top of cluster velocity, the engine now reports proposal-kind seasonal
pressure:

- `quiet`
- `watch`
- `in_band`
- `above_norm`

That gives governance a second lens: whether a cluster is accelerating on its
own, and whether that proposal kind is above its normal seasonal cadence across
the portfolio.

`governance-remediation` sits on top of the queue and trend layers. It converts
cluster alerts into structured action bundles with:

- severity
- release readiness
- current release readiness
- release blockers
- immediate actions
- approval guardrails
- monitoring actions
- explicit checkpoints with owner roles, completion criteria, and dependency ordering

The remediation layer is still deterministic and policy-driven. It does not
invent actions from a model; it assembles them from the proposal class,
proposal kind, and detected systemic signals in `inference-policy.json`. The
checkpoint ownership and completion criteria are also policy-driven so operator
automation can consume them directly without hard-coding role mappings. Phase
ordering and dependencies are policy-driven too, so the engine can express
which blockers must clear before guardrails or monitoring work can begin.

When operators provide a checkpoint status file, the remediation layer also
computes a progress-aware rollup:

- checkpoint status counts
- current release readiness
- ready-to-start checkpoints once dependencies are satisfied

That makes the remediation output usable as a deterministic execution-progress
view instead of a static advisory bundle only.

`governance-replay` sits underneath that path. It reconstructs deterministic
checkpoint state from either:

- an append-only event log
- a derived snapshot of current checkpoint state

The event-log path is stricter and is the preferred long-term operator format.
Each event must include:

- `checkpoint_id`
- `previous_status`
- `new_status`
- `updated_at`
- `recorded_by`
- `rationale`

The status file is not treated as an arbitrary snapshot. For non-pending
updates, operators should provide `previous_status`, `updated_at`, and
`recorded_by`, and the engine validates the transition against the policy
lifecycle:

- `pending -> in_progress`
- `in_progress -> completed`
- idempotent same-state writes

Illegal moves are surfaced as transition alerts and force the current plan
readiness to `invalid`.

Audit gaps are treated the same way. The remediation engine rejects non-pending
updates without actor attribution or timestamps, and it rejects checkpoint
updates whose `updated_at` value predates the latest completed dependency.

Event logs add another deterministic guarantee: replay rejects duplicate,
contradictory, illegal, or out-of-order history. That means remediation can
consume an event log directly without trusting a mutable current-state file as
the source of truth.

## Future Direction

The next step is to add model-backed summarization on top of the deterministic
queue, trend, and remediation outputs. That adapter should remain advisory only
and should never replace the deterministic control path for classifying
approval requirements or mitigation bundles.
