# Validator and Genesis Scaffolding

This document defines the first safe scaffolding layer for the 0AI Assurance Network appchain repository.

## Validator set

The initial skeleton uses a 7-validator permissioned set.

The example validator manifest contains public metadata only:

- validator ID
- operator label
- role
- public key digest placeholder
- voting power
- review status

It must never contain:

- production private keys
- signer mnemonics
- API tokens
- validator server credentials
- operator personal data

## Genesis template

The genesis template is deterministic and localnet-oriented. It carries:

- network identity
- permissioned validator count
- no public sale assumption
- no transferability assumption
- governance readiness gates
- module planning metadata
- provenance metadata

## Required gates before any release candidate

- legal review
- external security audit
- incident-response plan
- validator operations runbook
- key custody review
- governance configuration review
- registry and attestation review
- reproducible genesis digest review

## Safe next implementation

The next implementation PR should add validators for the YAML templates and should keep all execution local-only and deterministic.

Suggested commands:

```bash
make validate
make readiness
make render-localnet
```

No production secrets, live chain actions, public deployment, or token issuance are introduced by this scaffolding.
