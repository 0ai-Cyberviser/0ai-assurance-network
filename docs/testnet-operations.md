# Testnet Operations

## Scope

This runbook covers the permissioned testnet skeleton only. It does not assume
mainnet readiness.

## Goals

- render a reproducible local validator layout
- keep validator topology explicit
- validate policy and genesis assumptions before runtime exists

## Current Workflow

1. Edit `config/network-topology.json`
2. Edit `config/genesis/base-genesis.json`
3. Edit `config/policy/release-guards.json`
4. Run `make validate`
5. Run `make render-localnet`
6. Run `./0aid init-node --root . --id val-1 --out ./build/nodes/validator-1`
7. Review generated artifacts in `build/localnet/` and `build/nodes/`

## Generated Artifacts

- `docker-compose.yml`: validator and seed service definitions
- `network-summary.json`: simplified operator view of peers and ports
- `genesis.rendered.json`: carried-forward base genesis with localnet metadata
- `build/nodes/<validator>/manifest.json`: deterministic local node manifest
- `build/nodes/<validator>/config/identity.json`: development-only placeholder identity
- `build/nodes/<validator>/config/node.json`: node config bundle

## Operational Assumptions

- validators are permissioned
- seed nodes are fixed at launch
- governance defaults to dual-house mode
- public transferability remains disabled
- deterministic identities are for local bootstrap only

## Required Future Additions

- deterministic key management
- signing and release attestations
- validator allowlist tooling
- chain-binary startup hooks
- health checks and metrics wiring
