# Threat Detection System

## Overview

The 0AI Assurance Network includes a zero-day threat detection system that analyzes governance proposals for security vulnerabilities, attack vectors, and exploit patterns. This system provides an additional security layer on top of the deterministic governance inference engine.

## Purpose

The threat detection system:

- Identifies potential security vulnerabilities in proposals
- Detects known attack patterns (governance attacks, economic exploits, smart contract vulnerabilities)
- Assigns threat levels and security remediation requirements
- Integrates seamlessly with existing governance workflows
- Maintains backward compatibility - it's opt-in and can be disabled

## Threat Detection Engine

The engine analyzes proposals against multiple threat categories:

### Vulnerability Categories

- **Smart Contract**: reentrancy, integer overflow, unchecked calls, delegatecall risks
- **Governance Attack**: flash loans, voting manipulation, quorum bypass, sybil attacks
- **Economic Exploit**: arbitrage manipulation, oracle attacks, front-running, MEV extraction
- **Infrastructure**: denial of service, network partition, eclipse attacks
- **Data Integrity**: replay attacks, signature forgery, state bloat

### Detection Rules

The system applies context-aware detection rules:

- **Validator Set Changes**: Escalates if emergency + validator changes (privilege escalation risk)
- **Treasury Operations**: Escalates for large amounts (economic exploit risk)
- **Emergency Actions**: Always requires security review
- **External Dependencies**: Checks for audit completion

## Threat Levels

| Level | Score Threshold | Security Review | Blocks Execution | Escalation |
|-------|----------------|-----------------|------------------|------------|
| **Critical** | ≥ 50 | Required | Yes | Required |
| **High** | ≥ 30 | Required | No | Required |
| **Elevated** | ≥ 15 | Required | No | No |
| **Low** | < 15 | No | No | No |

## Configuration

The threat detection system is configured via `config/governance/threat-detection-policy.json`:

```json
{
  "version": "threat-detection-2026-04-21",
  "engine": "0ai-threat-detection-v1",
  "enabled": true,
  "vulnerability_patterns": { ... },
  "threat_weights": { ... },
  "detection_rules": { ... }
}
```

### Enabling/Disabling

Set `"enabled": false` to disable threat detection entirely. When disabled, the system returns `None` and has no impact on governance workflows.

## Usage

### Command Line

Scan a proposal for threats:

```bash
PYTHONPATH=src python -m assurancectl.cli governance-threat-scan \
  --proposal examples/proposals/emergency-pause.json
```

With JSON output:

```bash
PYTHONPATH=src python -m assurancectl.cli governance-threat-scan \
  --proposal examples/proposals/emergency-pause.json \
  --json
```

Generate machine-readable artifact:

```bash
PYTHONPATH=src python -m assurancectl.cli governance-threat-scan \
  --proposal examples/proposals/emergency-pause.json \
  --artifact-out build/artifacts/threat-scan.json
```

### Make Targets

```bash
make governance-threat-scan PROPOSAL=examples/proposals/emergency-pause.json
```

## Report Structure

A threat detection report includes:

- `threat_level`: critical, high, elevated, or low
- `threat_score`: numerical score based on weighted detection
- `vulnerability_categories`: detected vulnerability types
- `attack_vectors`: potential exploit paths
- `triggered_patterns`: which detection rules fired
- `security_signals`: specific security concerns
- `rationale`: explanation of why threats were detected
- `security_remediation`: specific mitigation steps
- `requires_security_review`: boolean flag
- `blocks_execution`: whether execution should be blocked
- `requires_escalation`: whether to escalate to security council

## Integration with Governance Workflow

The threat detection system integrates with the existing governance inference:

1. **Independent Analysis**: Threat scanning runs independently of proposal classification
2. **Complementary Signals**: Adds security-specific signals to governance review
3. **Enhanced Remediation**: Security remediation items complement governance remediation
4. **Backward Compatible**: Existing workflows continue unchanged when threat detection is disabled

## Security Considerations

### False Positives

The pattern-matching approach may produce false positives. Security review requirements ensure human oversight for flagged proposals.

### Pattern Updates

Vulnerability patterns should be updated regularly as new attack vectors emerge. This is a configuration-level change that doesn't require code updates.

### Audit Trail

All threat detection results can be written to versioned artifacts for audit purposes using `--artifact-out`.

## Future Enhancements

Planned improvements:

1. Integration with external threat intelligence feeds
2. Machine learning-based anomaly detection
3. Historical attack pattern analysis
4. Integration with on-chain security monitors
5. Real-time vulnerability database lookups
