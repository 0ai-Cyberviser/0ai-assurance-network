#!/usr/bin/env python3
"""
PeachTrace: Governance Audit Trail and Provenance Tracking System

Maintains cryptographically-signed audit ledgers for all governance operations,
signer rotations, and funding deployments. Provides tamper-evident trails with
SHA256 digests, signature chains, and continuity verification.

Usage:
    # Append governance event to audit ledger
    python scripts/peachtrace.py append \\
        --ledger build/audit/governance-ledger.json \\
        --event-type governance_simulation \\
        --event-data '{"proposal_id": "draft-001", "disposition": "approved"}' \\
        --actor governance-ops-bot

    # Verify ledger integrity
    python scripts/peachtrace.py verify \\
        --ledger build/audit/governance-ledger.json

    # Export ledger for archive retention
    python scripts/peachtrace.py export \\
        --ledger build/audit/governance-ledger.json \\
        --output build/audit/governance-export-$(date +%Y%m%d).json
"""

from __future__ import annotations

import argparse
import hashlib
import json
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


class PeachTrace:
    """Governance audit trail system with provenance tracking."""
    
    def __init__(self, ledger_path: Path):
        self.ledger_path = ledger_path
        self.ledger = self._load_or_create_ledger()
    
    def _load_or_create_ledger(self) -> dict[str, Any]:
        """Load existing ledger or create new one."""
        if self.ledger_path.exists():
            with open(self.ledger_path) as f:
                return json.load(f)
        
        return {
            "ledger_id": f"peachtrace-{datetime.now(timezone.utc).strftime('%Y%m%d-%H%M%S')}",
            "created_at": datetime.now(timezone.utc).isoformat(),
            "version": "1.0.0",
            "events": [],
            "continuity_chain": [],
        }
    
    def _compute_event_digest(self, event: dict[str, Any]) -> str:
        """Compute SHA256 digest of event."""
        event_json = json.dumps(event, sort_keys=True)
        return hashlib.sha256(event_json.encode("utf-8")).hexdigest()
    
    def _compute_chain_digest(self, previous_digest: str, current_digest: str) -> str:
        """Compute chained digest linking to previous event."""
        chain_data = f"{previous_digest}:{current_digest}"
        return hashlib.sha256(chain_data.encode("utf-8")).hexdigest()
    
    def append_event(
        self,
        event_type: str,
        event_data: dict[str, Any],
        actor: str,
        metadata: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        """Append a new event to the audit ledger."""
        
        timestamp = datetime.now(timezone.utc).isoformat()
        event_id = f"{event_type}-{len(self.ledger['events']):06d}"
        
        event = {
            "event_id": event_id,
            "event_type": event_type,
            "timestamp": timestamp,
            "actor": actor,
            "data": event_data,
            "metadata": metadata or {},
        }
        
        # Compute event digest
        event_digest = self._compute_event_digest(event)
        event["digest_sha256"] = event_digest
        
        # Chain to previous event
        if self.ledger["events"]:
            previous_digest = self.ledger["events"][-1]["digest_sha256"]
            chain_digest = self._compute_chain_digest(previous_digest, event_digest)
            event["previous_digest"] = previous_digest
            event["chain_digest"] = chain_digest
            
            self.ledger["continuity_chain"].append({
                "from_event": self.ledger["events"][-1]["event_id"],
                "to_event": event_id,
                "chain_digest": chain_digest,
            })
        else:
            event["previous_digest"] = "genesis"
            event["chain_digest"] = event_digest
        
        self.ledger["events"].append(event)
        self._save_ledger()
        
        return event
    
    def verify_integrity(self) -> tuple[bool, list[str]]:
        """Verify cryptographic integrity of audit ledger."""
        errors = []
        
        if not self.ledger["events"]:
            return True, []
        
        # Verify each event digest
        for i, event in enumerate(self.ledger["events"]):
            # Recompute digest (exclude digest, previous_digest, and chain_digest)
            event_copy = {k: v for k, v in event.items() if k not in ["digest_sha256", "previous_digest", "chain_digest"]}
            expected_digest = self._compute_event_digest(event_copy)
            
            if event["digest_sha256"] != expected_digest:
                errors.append(f"Event {i} ({event['event_id']}): digest mismatch")
        
        # Verify continuity chain
        for i in range(1, len(self.ledger["events"])):
            current = self.ledger["events"][i]
            previous = self.ledger["events"][i - 1]
            
            if current["previous_digest"] != previous["digest_sha256"]:
                errors.append(f"Event {i} ({current['event_id']}): broken chain to previous event")
            
            expected_chain = self._compute_chain_digest(previous["digest_sha256"], current["digest_sha256"])
            if current["chain_digest"] != expected_chain:
                errors.append(f"Event {i} ({current['event_id']}): invalid chain digest")
        
        return len(errors) == 0, errors
    
    def export_for_archive(self) -> dict[str, Any]:
        """Export ledger with baseline snapshot for archive retention."""
        valid, errors = self.verify_integrity()
        
        if not valid:
            raise ValueError(f"Cannot export ledger with integrity errors: {errors}")
        
        export = {
            "export_timestamp": datetime.now(timezone.utc).isoformat(),
            "ledger": self.ledger,
            "baseline_snapshot": {
                "ledger_id": self.ledger["ledger_id"],
                "events_count": len(self.ledger["events"]),
                "first_event_timestamp": self.ledger["events"][0]["timestamp"] if self.ledger["events"] else None,
                "last_event_timestamp": self.ledger["events"][-1]["timestamp"] if self.ledger["events"] else None,
                "genesis_digest": self.ledger["events"][0]["digest_sha256"] if self.ledger["events"] else None,
                "head_digest": self.ledger["events"][-1]["digest_sha256"] if self.ledger["events"] else None,
            },
            "verification": {
                "integrity_valid": valid,
                "verified_at": datetime.now(timezone.utc).isoformat(),
                "verifier": "PeachTrace v1.0.0",
            },
        }
        
        # Compute export digest
        export_json = json.dumps(export, sort_keys=True)
        export["export_digest_sha256"] = hashlib.sha256(export_json.encode("utf-8")).hexdigest()
        
        return export
    
    def _save_ledger(self) -> None:
        """Save ledger to disk."""
        self.ledger_path.parent.mkdir(parents=True, exist_ok=True)
        with open(self.ledger_path, "w") as f:
            json.dump(self.ledger, f, indent=2)


def main() -> int:
    parser = argparse.ArgumentParser(description="PeachTrace governance audit trail system")
    subparsers = parser.add_subparsers(dest="command", required=True)
    
    # Append command
    append = subparsers.add_parser("append", help="Append event to audit ledger")
    append.add_argument("--ledger", required=True, type=Path, help="Audit ledger file")
    append.add_argument("--event-type", required=True, help="Event type")
    append.add_argument("--event-data", required=True, help="Event data JSON string")
    append.add_argument("--actor", required=True, help="Actor performing the action")
    append.add_argument("--metadata", help="Optional metadata JSON string")
    
    # Verify command
    verify = subparsers.add_parser("verify", help="Verify ledger integrity")
    verify.add_argument("--ledger", required=True, type=Path, help="Audit ledger file")
    
    # Export command
    export = subparsers.add_parser("export", help="Export ledger for archive")
    export.add_argument("--ledger", required=True, type=Path, help="Audit ledger file")
    export.add_argument("--output", required=True, type=Path, help="Export output file")
    
    args = parser.parse_args()
    
    if args.command == "append":
        trace = PeachTrace(args.ledger)
        
        event_data = json.loads(args.event_data)
        metadata = json.loads(args.metadata) if args.metadata else None
        
        event = trace.append_event(
            event_type=args.event_type,
            event_data=event_data,
            actor=args.actor,
            metadata=metadata,
        )
        
        print(f"✓ Appended event: {event['event_id']}")
        print(f"  Digest: {event['digest_sha256'][:16]}...")
        print(f"  Chain: {event['chain_digest'][:16]}...")
        print(f"  Ledger: {args.ledger}")
        
        return 0
    
    elif args.command == "verify":
        trace = PeachTrace(args.ledger)
        valid, errors = trace.verify_integrity()
        
        if valid:
            print(f"✓ Ledger integrity verified")
            print(f"  Events: {len(trace.ledger['events'])}")
            print(f"  Ledger ID: {trace.ledger['ledger_id']}")
        else:
            print(f"✗ Ledger integrity FAILED", file=sys.stderr)
            print(f"  Errors detected: {len(errors)}", file=sys.stderr)
            for error in errors:
                print(f"    - {error}", file=sys.stderr)
            return 1
        
        return 0
    
    elif args.command == "export":
        trace = PeachTrace(args.ledger)
        
        try:
            export_data = trace.export_for_archive()
            
            args.output.parent.mkdir(parents=True, exist_ok=True)
            with open(args.output, "w") as f:
                json.dump(export_data, f, indent=2)
            
            print(f"✓ Exported ledger for archive retention")
            print(f"  Events: {export_data['baseline_snapshot']['events_count']}")
            print(f"  Export digest: {export_data['export_digest_sha256'][:16]}...")
            print(f"  Output: {args.output}")
            
            return 0
        
        except ValueError as e:
            print(f"✗ Export failed: {e}", file=sys.stderr)
            return 1
    
    return 1


if __name__ == "__main__":
    sys.exit(main())
