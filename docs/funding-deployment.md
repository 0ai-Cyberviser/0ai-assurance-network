# Blockchain Funding Deployment

This document describes the blockchain funding deployment process for the 0AI Assurance Network.

## Overview

The funding deployment system manages the allocation and distribution of network resources across multiple funding pools, including:

- **Validator Rewards**: Pool for validator staking rewards
- **Treasury Grants**: Pool for community treasury grants
- **Development Fund**: Pool for protocol development funding
- **Security Reserve**: Emergency security and incident response reserve

## Configuration

Funding configuration is defined in `config/governance/funding-config.json`:

```json
{
  "funding_pools": {
    "pool_name": {
      "allocation_percent": 40.0,
      "min_balance": 1000000,
      "description": "Pool description"
    }
  },
  "allocation_strategy": "proportional",
  "treasury_address": "0ai-treasury-main",
  "initial_validator_stake": 100000,
  "grant_approval_threshold": 0.67,
  "funding_cycle_days": 30,
  "max_grant_per_cycle": 100000
}
```

### Configuration Fields

- **funding_pools**: Dictionary of funding pools with allocation percentages and minimum balances
  - `allocation_percent`: Percentage of total funding allocated to this pool (must sum to 100%)
  - `min_balance`: Minimum balance required in the pool
  - `description`: Human-readable pool description

- **allocation_strategy**: Strategy for fund allocation (`proportional`, `fixed`, or `dynamic`)

- **treasury_address**: Main treasury address identifier

- **initial_validator_stake**: Initial stake amount for each validator

- **grant_approval_threshold**: Minimum approval ratio for treasury grants (0-1)

- **funding_cycle_days**: Number of days in each funding cycle

- **max_grant_per_cycle**: Maximum grant amount per funding cycle

## Validation Rules

The funding configuration must pass the following validation checks:

1. **Required Fields**: All required fields must be present
   - `funding_pools`
   - `allocation_strategy`
   - `treasury_address`

2. **Allocation Percentages**: Pool allocation percentages must sum to exactly 100%

3. **Allocation Strategy**: Must be one of: `proportional`, `fixed`, `dynamic`

4. **Pool Configuration**: Each pool must have:
   - `allocation_percent` (numeric)
   - `min_balance` (numeric)

## Deployment Process

### 1. Validate Configuration

Before deployment, validate the funding configuration:

```bash
make funding-validate
```

This performs a dry-run deployment that validates the configuration without making changes.

### 2. Generate Deployment Configuration

Generate a deployment configuration file:

```bash
make funding-deploy FUNDING_CONFIG=config/governance/funding-config.json OUT=build/funding-deployment.json DRY_RUN=true
```

This creates a deployment configuration that includes:
- Network metadata derived from the repository configuration
- Validator funding allocations
- Pool configurations
- Chain/network identifiers such as `chain_id` where available

### 3. Deploy to Blockchain

To perform actual deployment (when blockchain runtime is available):

```bash
make funding-deploy FUNDING_CONFIG=config/governance/funding-config.json OUT=build/funding-deployment.json
```

**Note**: Actual blockchain deployment requires the chain runtime to be operational. The current implementation validates configuration and generates deployment artifacts.

## Testing

### Run Funding Tests

Execute the funding deployment test suite:

```bash
make funding-test
```

This runs tests that validate:
- Configuration validation logic
- Invalid allocation rejection
- Missing field detection
- Deployment output generation
- Validator funding allocation
- Invalid strategy rejection

### Manual Testing

You can manually test the deployment script:

```bash
python scripts/deploy_funding.py \
  --root . \
  --funding-config config/governance/funding-config.json \
  --output build/test-deployment.json \
  --dry-run
```

## Continuous Integration

The CI/CD pipeline automatically:
1. Validates funding configuration on every push
2. Runs funding deployment tests
3. Ensures configuration changes don't break deployment

See `.github/workflows/ci.yml` for the complete CI configuration.

## Deployment Artifacts

Successful deployment generates the following artifacts:

### Deployment Configuration (`build/funding-deployment.json`)

```json
{
  "deployment_version": "1.0.0",
  "network_id": "0ai-testnet",
  "genesis_time": "...",
  "funding_config": { ... },
  "validators": [
    {
      "validator_id": "val-1",
      "address": "...",
      "initial_stake": 100000,
      "funding_pool": "validator_rewards"
    }
  ]
}
```

## Integration with Governance

The funding deployment integrates with the governance system:

- **Treasury Grants**: Managed through governance proposals (see `examples/proposals/treasury-grant.json`)
- **Approval Workflow**: Grants require approval based on `grant_approval_threshold`
- **Audit Trail**: All funding decisions are recorded in governance checkpoints

## Security Considerations

1. **Configuration Validation**: Always validate configuration before deployment
2. **Allocation Limits**: Enforce maximum grant amounts per cycle
3. **Approval Thresholds**: Require multi-party approval for large grants
4. **Minimum Balances**: Maintain minimum balances in each pool
5. **Audit Logging**: All funding operations are logged for audit

## Troubleshooting

### Configuration Validation Fails

If validation fails with "Total allocation must equal 100%":
- Verify all pool `allocation_percent` values sum to exactly 100
- Check for rounding errors in percentage values

### Missing Required Fields

If validation fails with "Missing required field":
- Ensure all required fields are present in the configuration
- Verify JSON syntax is valid

### Invalid Allocation Strategy

If validation fails with "Invalid allocation_strategy":
- Use one of: `proportional`, `fixed`, `dynamic`
- Check for typos in the strategy name

## Future Enhancements

Planned enhancements to the funding deployment system:

1. **Multi-chain Support**: Deploy funding configuration across multiple chains
2. **Dynamic Rebalancing**: Automatic pool rebalancing based on usage
3. **Historical Analytics**: Track funding usage and efficiency over time
4. **Smart Contract Integration**: Direct integration with on-chain treasury contracts
5. **Automated Disbursement**: Scheduled automatic grant disbursements

## References

- [Governance Inference](governance-inference.md)
- [Network Topology](../config/network-topology.json)
- [Funding Configuration](../config/governance/funding-config.json)
- [Treasury Grant Example](../examples/proposals/treasury-grant.json)
