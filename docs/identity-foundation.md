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
- governance execution actors used by remediation checkpoints and signer policy

It does not try to implement general-purpose on-chain identity yet. It is the
minimal operator surface required by the module milestone plan.

## Current Roles

The bootstrap covers the required roles derived from
[config/modules/milestone-1.json](../config/modules/milestone-1.json), plus the
governance execution roles referenced by
[config/governance/inference-policy.json](../config/governance/inference-policy.json)
and [config/governance/checkpoint-signers.json](../config/governance/checkpoint-signers.json):

- `network_admin`
- `registry_operator`
- `registry_reviewer`
- `auditor_operator`
- `attestation_reviewer`
- `governance_admin`
- `validator_ops_lead`
- `safety_council_delegate`
- `governance-chair`
- `governance-ops`
- `token-house-secretariat`
- `telemetry-ops`
- `treasury-program-manager`
- `treasury-review-chair`
- `finance-telemetry-lead`
- `validator-ops-lead`
- `staking-governance-reviewer`
- `network-reliability-lead`
- `safety-council-chair`
- `incident-commander`
- `safety-council-secretariat`
- `security-telemetry-lead`

Validation fails if any required role is missing an active binding or if a role
binding is duplicated. It also fails if a configured checkpoint signer is not
bound to an active bootstrap actor for the role it claims to sign.

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
- governance replay uses it to bind signed checkpoint events to active actors
- governance remediation uses it to resolve eligible owner actors per checkpoint
- validator rollout uses it to verify that only approved operators can project
  release eligibility into validator activation gates

This keeps the initial permissioned testnet narrow and auditable while still
giving later chain code a concrete identity target to implement.
