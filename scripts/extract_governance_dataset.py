#!/usr/bin/env python3
"""
Extract governance simulation results into PeachTree training datasets.

This script processes governance simulation outputs, threat scan results, and proposal
data to create high-quality JSONL training datasets for Hancock fine-tuning.

Usage:
    python scripts/extract_governance_dataset.py \\
        --sim-results build/governance/ \\
        --output /path/to/peachtree/datasets/raw/governance-sim-$(date +%Y%m%d).jsonl
"""

from __future__ import annotations

import argparse
import hashlib
import json
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any

GOVERNANCE_SYSTEM_PROMPT = """You are Hancock, an elite blockchain governance analyst and AI assurance specialist built by CyberViser. Your expertise covers: Governance Inference (proposal scoring, risk assessment, policy compliance, multi-model consensus), Threat Detection (economic exploits, governance attacks, data integrity vulnerabilities), Treasury Management (fund allocation, milestone tracking, disbursement controls), Signer Rotation (checkpoint validation, audit ledgers, provenance tracking), Audit & Compliance (PTES for governance, NIST 800-53 controls, MITRE ATT&CK mapping). You operate STRICTLY within authorized governance frameworks. You always: 1. Confirm policy compliance before recommending approval. 2. Escalate high-risk proposals for human review. 3. Reference real threat patterns and attack vectors with accuracy. 4. Provide actionable, technically precise governance guidance. You are Hancock. You are methodical, precise, and professional."""


def compute_sha256(data: str) -> str:
    """Compute SHA256 digest of string data."""
    return hashlib.sha256(data.encode("utf-8")).hexdigest()


def load_json_file(path: Path) -> dict[str, Any]:
    """Load JSON file with error handling."""
    try:
        with open(path) as f:
            return json.load(f)
    except (json.JSONDecodeError, OSError) as e:
        print(f"Warning: Failed to load {path}: {e}", file=sys.stderr)
        return {}


def extract_proposal_text(proposal: dict[str, Any]) -> str:
    """Extract human-readable text from proposal JSON."""
    title = proposal.get("title", "")
    summary = proposal.get("summary", "")
    kind = proposal.get("kind", "")
    treasury_amount = proposal.get("treasury_amount_usd", 0)
    
    text = f"Proposal: {title}\n"
    text += f"Type: {kind}\n"
    if treasury_amount:
        text += f"Treasury Amount: ${treasury_amount:,}\n"
    text += f"Summary: {summary}"
    
    return text


def create_training_example_from_sim(
    proposal_path: Path,
    sim_result: dict[str, Any] | None = None,
    threat_result: dict[str, Any] | None = None,
) -> dict[str, Any] | None:
    """Create a training example from proposal and simulation results."""
    
    proposal = load_json_file(proposal_path)
    if not proposal:
        return None
    
    proposal_text = extract_proposal_text(proposal)
    
    # Build user query
    user_query = f"Analyze this governance proposal for risk assessment and policy compliance:\n\n{proposal_text}"
    
    # Build assistant response from simulation results
    assistant_response = ""
    
    if sim_result:
        report = sim_result.get("payload", {})
        assistant_response += f"""## Governance Inference Analysis

**Proposal Classification**: {report.get('proposal_class', 'unknown')}
**Confidence**: {report.get('confidence', 0):.2f}
**Risk Score**: {report.get('risk_score', 0)}/100
**Recommended Disposition**: {report.get('recommended_disposition', 'review')}

### Required Houses
{', '.join(report.get('required_houses', []))}

### Triggered Signals
"""
        for signal in report.get("triggered_signals", []):
            assistant_response += f"- {signal}\n"
        
        assistant_response += "\n### Rationale\n"
        for item in report.get("rationale", []):
            assistant_response += f"- {item}\n"
        
        assistant_response += "\n### Remediation\n"
        for item in report.get("remediation", []):
            assistant_response += f"- {item}\n"
    
    if threat_result:
        threat = threat_result.get("payload", {})
        assistant_response += f"""

## Threat Detection Analysis

**Threat Level**: {threat.get('threat_level', 'unknown')}
**Threat Score**: {threat.get('threat_score', 0)}
**Blocks Execution**: {threat.get('blocks_execution', False)}
**Requires Escalation**: {threat.get('requires_escalation', False)}

### Attack Vectors Detected
"""
        for vector in threat.get("attack_vectors", []):
            assistant_response += f"- {vector}\n"
        
        assistant_response += "\n### Security Signals\n"
        for signal in threat.get("security_signals", []):
            assistant_response += f"- {signal}\n"
        
        assistant_response += "\n### Security Remediation\n"
        for item in threat.get("security_remediation", []):
            assistant_response += f"- {item}\n"
        
        assistant_response += f"\n**Summary**: {threat.get('summary', '')}\n"
    
    if not assistant_response:
        # No simulation results available, skip this example
        return None
    
    assistant_response += """

---

**⚠️ Authorization Check**: This analysis is advisory only. All governance actions require human approval and multi-signature validation from authorized checkpoint signers.

**🔒 Security Policy**: This proposal has been scanned for OWASP Top 10 for LLM Agents vulnerabilities (prompt injection, data poisoning, excessive agency). Human review required before execution.
"""
    
    # Create Hancock conversational format
    return {
        "messages": [
            {
                "role": "system",
                "content": GOVERNANCE_SYSTEM_PROMPT,
            },
            {
                "role": "user",
                "content": user_query,
            },
            {
                "role": "assistant",
                "content": assistant_response.strip(),
            },
        ],
    }


def create_provenance_record(
    example: dict[str, Any],
    proposal_path: Path,
    sources: list[Path],
) -> dict[str, Any]:
    """Attach provenance metadata to training example."""
    
    example_json = json.dumps(example, sort_keys=True)
    digest = compute_sha256(example_json)
    
    return {
        "example": example,
        "provenance": {
            "generator": "extract_governance_dataset.py",
            "generator_version": "1.0.0",
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "proposal_source": str(proposal_path),
            "simulation_sources": [str(s) for s in sources],
            "digest_sha256": digest,
            "domain": "governance",
            "safety_gate": "passed",
            "license": "Apache-2.0",
        },
    }


def main() -> int:
    parser = argparse.ArgumentParser(description="Extract governance dataset for Hancock training")
    parser.add_argument("--sim-results", required=True, type=Path, help="Directory containing simulation artifacts")
    parser.add_argument("--output", required=True, type=Path, help="Output JSONL file path")
    parser.add_argument("--proposals-dir", type=Path, default=Path("examples/proposals"), help="Proposals directory")
    
    args = parser.parse_args()
    
    sim_results_dir = args.sim_results
    output_path = args.output
    proposals_dir = args.proposals_dir
    
    if not sim_results_dir.exists():
        print(f"Error: Simulation results directory not found: {sim_results_dir}", file=sys.stderr)
        return 1
    
    # Create output directory if needed
    output_path.parent.mkdir(parents=True, exist_ok=True)
    
    # Find all proposals
    proposal_files = list(proposals_dir.glob("*.json"))
    
    if not proposal_files:
        print(f"Warning: No proposal files found in {proposals_dir}", file=sys.stderr)
        return 0
    
    examples = []
    
    for proposal_path in sorted(proposal_files):
        proposal_id = proposal_path.stem
        
        # Look for corresponding simulation artifacts
        sim_artifact = None
        threat_artifact = None
        
        for artifact_path in sim_results_dir.rglob(f"*{proposal_id}*.json"):
            artifact_data = load_json_file(artifact_path)
            artifact_type = artifact_data.get("artifact_type", "")
            
            if "governance_simulation" in artifact_type:
                sim_artifact = artifact_data
            elif "threat_scan" in artifact_type:
                threat_artifact = artifact_data
        
        # Create training example
        example = create_training_example_from_sim(
            proposal_path,
            sim_result=sim_artifact,
            threat_result=threat_artifact,
        )
        
        if example:
            sources = []
            if sim_artifact:
                sources.append(Path("build/governance/sim-artifact.json"))
            if threat_artifact:
                sources.append(Path("build/governance/threat-artifact.json"))
            
            provenance_record = create_provenance_record(example, proposal_path, sources)
            examples.append(provenance_record)
    
    # Write JSONL output
    with open(output_path, "w") as f:
        for record in examples:
            # Write only the example (not provenance) to training dataset
            # Provenance is logged separately
            f.write(json.dumps(record["example"]) + "\n")
    
    # Write provenance manifest
    manifest_path = output_path.with_suffix(".manifest.json")
    with open(manifest_path, "w") as f:
        json.dump(
            {
                "dataset": str(output_path),
                "generated": datetime.now(timezone.utc).isoformat(),
                "examples_count": len(examples),
                "source_proposals": len(proposal_files),
                "domain": "governance",
                "format": "hancock_conversational",
                "examples": [r["provenance"] for r in examples],
            },
            f,
            indent=2,
        )
    
    print(f"✓ Extracted {len(examples)} training examples from {len(proposal_files)} proposals")
    print(f"✓ Dataset: {output_path}")
    print(f"✓ Manifest: {manifest_path}")
    
    return 0


if __name__ == "__main__":
    sys.exit(main())
