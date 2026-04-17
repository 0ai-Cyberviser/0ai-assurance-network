# Identity Foundation

The first executable dependency surface for the permissioned testnet is the
identity bootstrap in [config/identity/bootstrap.json](../config/identity/bootstrap.json).

It exists to answer one narrow question for the initial network:

Who is allowed to perform registry, attestation, governance-hook, and validator
rollout actions before the full chain modules are implemented?

## Bootstrap Scope

The bootstrap currently models:

- organizations and councils
- operator identities
- bounded role bindings

It does not try to implement general-purpose on-chain identity yet. It is the
minimal operator surface required by the module milestone plan.

## Current Roles

The bootstrap covers the required roles derived from
[config/modules/milestone-1.json](../config/modules/milestone-1.json):

- `network_admin`
- `registry_operator`
- `registry_reviewer`
- `auditor_operator`
- `attestation_reviewer`
- `governance_admin`
- `validator_ops_lead`
- `safety_council_delegate`

Validation fails if any required role is missing an active binding or if a role
binding is duplicated.

## Operator Command

Render the current identity plan:

```bash
./0aid identity-plan --root .
```

Or via make:

```bash
make identity-plan
```

The plan reports:

- actor count
- active actor count
- required roles
- bound roles
- missing roles
- per-role actor bindings and module references

## Dependency Relationship

The identity bootstrap feeds the first chain milestone directly:

- registry uses it to authorize service and release operations
- attestation uses it to authorize auditors and reviewers
- governance hooks use it to authorize bounded admin and safety roles
- validator rollout uses it to verify that only approved operators can project
  release eligibility into validator activation gates

This keeps the initial permissioned testnet narrow and auditable while still
giving later chain code a concrete identity target to implement.
