# 0AI Assurance Network Appchain Hardening Roadmap

This roadmap promotes the current 0AI Assurance Network skeleton into a deterministic, audit-ready, permissioned appchain repository.

## Current design baseline

- 7 validators
- Permissioned initial validator set
- Dual-house governance metadata carried in config
- No public sale assumptions
- No transferability assumptions
- Release guarded by legal, audit, and incident-response prerequisites
- Launch-readiness report built into the operator CLI

## Safety boundary

This repository remains pre-launch and validation-focused.

The engineering program must not introduce:

- public token issuance
- public sale flows
- autonomous live-chain deployment
- production signer operation from CI
- production private-key generation or storage
- transferability assumptions
- unsandboxed fuzzing or mutation execution

Every release-bound artifact must include provenance, digest validation, and human approval gates.

## Phase 1: repository hardening

Create a clean operator-facing structure:

```text
cmd/0aid/
config/
  validators/
  governance/
  genesis/
  funding/
internal/
  keygen/
  genesis/
  registry/
  attestation/
  audit/
scripts/
tests/
docs/
.github/workflows/
```

Acceptance criteria:

- `make validate` remains the primary static gate.
- `make readiness` reports launch blockers.
- `make render-localnet` remains deterministic.
- CI validates fixtures without requiring secrets or network access.

## Phase 2: deterministic local/testnet keygen

Add deterministic key generation only for localnet and testnet use.

Rules:

- Deterministic keys are local/dev/testnet only.
- Production validator keys are never committed.
- Generated secret material is written only to ignored local paths.
- Public manifests include validator IDs, public keys, digests, and provenance.
- Production signing remains manual, external, and human-approved.

Candidate files:

```text
internal/keygen/deterministic.py
internal/keygen/manifest.py
config/validators/validators.example.yaml
docs/validator-secrets.md
```

## Phase 3: genesis and validator templates

Add deterministic genesis rendering and validator config templates.

Candidate files:

```text
config/genesis/genesis.template.yaml
config/validators/validator-01.example.yaml
scripts/render_genesis.py
tests/test_genesis_render.py
```

Candidate commands:

```bash
make init-genesis
make render-localnet
make validate-genesis
make validator-configs
```

## Phase 4: module-specific state and CLI wiring

Define module plans before implementation:

- registry module
- attestation module
- governance metadata module
- audit-ledger module
- funding config module
- incident-response gate module

Candidate commands:

```bash
0aid module-plan registry
0aid module-plan attestation
0aid readiness
0aid render-localnet
0aid audit-ledger verify
```

## Phase 5: registry and attestation modules

Minimum registry record:

```json
{
  "schema_version": "0ai.registry.v1",
  "subject_id": "validator-001",
  "subject_type": "validator",
  "public_key_digest": "sha256:example",
  "attestation_digest": "sha256:example",
  "status": "pending_review",
  "provenance": {
    "source": "localnet",
    "generated_by": "0aid",
    "created_at": "1970-01-01T00:00:00Z"
  }
}
```

## Phase 6: external audit pipeline

Release readiness requires:

- static validation
- unit tests
- genesis reproducibility check
- validator manifest digest check
- governance config lint
- SBOM generation
- secret scan
- audit-readiness report
- release blocker report

Candidate workflows:

```text
.github/workflows/ci.yml
.github/workflows/audit-readiness.yml
.github/workflows/release-gate.yml
```

## Launch-readiness blocker policy

A release candidate is blocked if any of the following are incomplete:

- legal review
- external security audit
- incident-response plan
- validator operations runbook
- key custody review
- genesis reproducibility verification
- governance configuration review
- registry and attestation digest review

## Next PRs

1. Add deterministic local/testnet keygen skeleton.
2. Add validator manifest schema and validator templates.
3. Add deterministic genesis renderer.
4. Add registry and attestation schema validators.
5. Add audit-readiness and release-gate workflows.
