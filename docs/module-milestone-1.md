# Module Milestone 1

The first implementable chain milestone is intentionally narrow:

- permissioned operator identity
- governed service registry
- release safety attestations
- governance and validator rollout hooks only where required to make those flows enforceable

This milestone is not a full public protocol surface. It is the smallest chain
scope that lets the permissioned testnet prove:

- governed services can be registered
- release candidates can be attested
- validator rollout can be blocked on missing or disputed attestations
- governance can freeze or escalate release actions when safety requires it

## Source Of Truth

The machine-readable source of truth is
[config/modules/milestone-1.json](../config/modules/milestone-1.json).

`0aid module-plan --root .` renders that config together with the current chain
context:

- chain id
- network mode
- governance houses
- validator count
- implementation sequence

## MVP Modules

### `registry`

- state:
  - service records
  - release channels
  - registry action log
- key transactions:
  - `register_service`
  - `propose_release`
  - `retire_service`
- operator roles:
  - `registry_operator`
  - `registry_reviewer`

### `attestation`

- state:
  - attestation records
  - reviewer bonds
  - appeal windows
- key transactions:
  - `submit_attestation`
  - `finalize_attestation`
  - `appeal_attestation`
- operator roles:
  - `auditor_operator`
  - `attestation_reviewer`

## Dependency Surfaces

These are included only to support the MVP modules:

- `identity_core`
- `governance_hooks`
- `validator_ops_hooks`

They should stay narrow and serve the registry/attestation flow rather than
expanding into general-purpose modules during the first milestone.

## Rollout Order

1. `identity-foundation`
2. `registry-mvp`
3. `attestation-mvp`
4. `governance-and-validator-hooks`

That order is deliberate:

- registry and attestation cannot exist without actor and role bindings
- attestation depends on registry release proposals
- governance and validator gating should attach after the core records and
  attestation lifecycle exist

## Validator Interaction Rule

Validators must not activate a governed release unless:

- the service is visible in the registry
- the release proposal exists
- the attestation is finalized and not under appeal
- governance has not frozen the rollout path

That is the operational bridge between safety review and validator behavior for
the first permissioned testnet.
