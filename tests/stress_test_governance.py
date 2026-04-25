#!/usr/bin/env python3
"""
Comprehensive stress testing for governance security pipeline.

Tests:
1. PeachFuzz: Generate 1000+ adversarial proposals
2. PeachTrace: Append/verify 100+ audit events
3. Dataset Extraction: Process all proposals with error handling
4. Threat Scanner: Non-blocking mode under load
5. End-to-End: Full pipeline resilience
6. Performance: Memory, CPU, execution time benchmarks
"""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
import time
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


class StressTestRunner:
    """Orchestrate stress testing for governance pipeline."""

    def __init__(self, workspace: Path):
        self.workspace = workspace
        self.results: dict[str, Any] = {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "tests": {},
            "summary": {},
        }

    def test_peachfuzz_stress(self, count: int = 1000) -> dict[str, Any]:
        """Stress test PeachFuzz with large corpus generation."""
        print(f"\n🔥 [1/6] PeachFuzz Stress Test ({count} proposals)...")
        start = time.time()

        try:
            cmd = [
                "python3",
                "scripts/peachfuzz_governance.py",
                "--base-proposal",
                "examples/proposals/treasury-grant.json",
                "--output",
                "build/stress/fuzz-corpus-stress.jsonl",
                "--count",
                str(count),
                "--include-edge-cases",
            ]

            result = subprocess.run(
                cmd,
                cwd=self.workspace,
                capture_output=True,
                text=True,
                timeout=300,  # 5min timeout
            )

            elapsed = time.time() - start

            # Count generated proposals
            corpus_path = self.workspace / "build/stress/fuzz-corpus-stress.jsonl"
            if corpus_path.exists():
                with open(corpus_path) as f:
                    lines = f.readlines()
                    generated_count = len(lines)
            else:
                generated_count = 0

            return {
                "status": "passed" if result.returncode == 0 else "failed",
                "requested_count": count,
                "generated_count": generated_count,
                "elapsed_seconds": round(elapsed, 2),
                "throughput": round(generated_count / elapsed, 2) if elapsed > 0 else 0,
                "exit_code": result.returncode,
                "stdout": result.stdout[-500:] if result.stdout else "",
                "stderr": result.stderr[-500:] if result.stderr else "",
            }

        except subprocess.TimeoutExpired:
            return {"status": "timeout", "elapsed_seconds": 300}
        except Exception as e:
            return {"status": "error", "error": str(e)}

    def test_peachtrace_stress(self, count: int = 100) -> dict[str, Any]:
        """Stress test PeachTrace with many audit events."""
        print(f"\n🔒 [2/6] PeachTrace Stress Test ({count} events)...")
        start = time.time()

        ledger_path = self.workspace / "build/stress/stress-ledger.json"
        ledger_path.parent.mkdir(parents=True, exist_ok=True)

        # Remove existing ledger
        if ledger_path.exists():
            ledger_path.unlink()

        try:
            # Append events rapidly
            for i in range(count):
                event_data = {
                    "proposal_id": f"stress-test-{i:05d}",
                    "risk_score": (i % 100),
                    "iteration": i,
                }

                cmd = [
                    "python3",
                    "scripts/peachtrace.py",
                    "append",
                    "--ledger",
                    str(ledger_path),
                    "--event-type",
                    "stress_test",
                    "--event-data",
                    json.dumps(event_data),
                    "--actor",
                    "stress-test-bot",
                ]

                result = subprocess.run(
                    cmd,
                    cwd=self.workspace,
                    capture_output=True,
                    text=True,
                    timeout=10,
                )

                if result.returncode != 0:
                    return {
                        "status": "failed",
                        "failed_at_event": i,
                        "error": result.stderr,
                    }

            append_elapsed = time.time() - start

            # Verify integrity
            verify_start = time.time()
            verify_cmd = [
                "python3",
                "scripts/peachtrace.py",
                "verify",
                "--ledger",
                str(ledger_path),
            ]

            verify_result = subprocess.run(
                verify_cmd,
                cwd=self.workspace,
                capture_output=True,
                text=True,
                timeout=30,
            )

            verify_elapsed = time.time() - verify_start
            total_elapsed = time.time() - start

            return {
                "status": "passed" if verify_result.returncode == 0 else "failed",
                "events_appended": count,
                "append_elapsed_seconds": round(append_elapsed, 2),
                "verify_elapsed_seconds": round(verify_elapsed, 2),
                "total_elapsed_seconds": round(total_elapsed, 2),
                "append_throughput": round(count / append_elapsed, 2),
                "integrity_verified": verify_result.returncode == 0,
                "verify_output": verify_result.stdout[-300:],
            }

        except Exception as e:
            return {"status": "error", "error": str(e)}

    def test_dataset_extraction_stress(self) -> dict[str, Any]:
        """Stress test dataset extraction with all proposals."""
        print("\n📊 [3/6] Dataset Extraction Stress Test...")
        start = time.time()

        try:
            # Run extraction on all proposals
            cmd = [
                "python3",
                "scripts/extract_governance_dataset.py",
                "--sim-results",
                "build/governance",
                "--output",
                "build/stress/stress-dataset.jsonl",
                "--proposals-dir",
                "examples/proposals",
            ]

            result = subprocess.run(
                cmd,
                cwd=self.workspace,
                capture_output=True,
                text=True,
                timeout=60,
            )

            elapsed = time.time() - start

            # Count extracted examples
            dataset_path = self.workspace / "build/stress/stress-dataset.jsonl"
            if dataset_path.exists():
                with open(dataset_path) as f:
                    lines = f.readlines()
                    example_count = len(lines)

                # Validate JSON format
                valid_json = True
                for line in lines:
                    try:
                        json.loads(line)
                    except json.JSONDecodeError:
                        valid_json = False
                        break
            else:
                example_count = 0
                valid_json = False

            # Check manifest
            manifest_path = dataset_path.with_suffix(".manifest.json")
            manifest_exists = manifest_path.exists()

            return {
                "status": "passed" if result.returncode == 0 else "failed",
                "examples_extracted": example_count,
                "valid_json_format": valid_json,
                "manifest_generated": manifest_exists,
                "elapsed_seconds": round(elapsed, 2),
                "exit_code": result.returncode,
                "stdout": result.stdout[-500:],
            }

        except Exception as e:
            return {"status": "error", "error": str(e)}

    def test_threat_scanner_stress(self, proposal_count: int = 50) -> dict[str, Any]:
        """Stress test threat scanner with many proposals."""
        print(f"\n⚠️ [4/6] Threat Scanner Stress Test ({proposal_count} proposals)...")
        start = time.time()

        try:
            # Use fuzzed corpus from earlier test
            corpus_path = self.workspace / "build/stress/fuzz-corpus-stress.jsonl"
            if not corpus_path.exists():
                return {"status": "skipped", "reason": "No fuzz corpus available"}

            passed = 0
            failed = 0
            warnings = 0

            with open(corpus_path) as f:
                proposals = [json.loads(line) for line in f.readlines()[:proposal_count]]

            for i, proposal in enumerate(proposals):
                # Write proposal to temp file
                temp_proposal = self.workspace / "build/stress/temp-proposal.json"
                with open(temp_proposal, "w") as f:
                    json.dump(proposal, f)

                # Run threat scan in non-blocking mode
                cmd = [
                    "python3",
                    "-m",
                    "assurancectl.cli",
                    "governance-threat-scan",
                    "--proposal",
                    str(temp_proposal),
                    "--non-blocking",
                ]

                result = subprocess.run(
                    cmd,
                    cwd=self.workspace,
                    env={**subprocess.os.environ, "PYTHONPATH": "src"},
                    capture_output=True,
                    text=True,
                    timeout=10,
                )

                if result.returncode == 0:
                    passed += 1
                    if "warning" in result.stdout.lower():
                        warnings += 1
                else:
                    failed += 1

            elapsed = time.time() - start

            return {
                "status": "passed" if failed == 0 else "partial",
                "proposals_scanned": proposal_count,
                "passed": passed,
                "failed": failed,
                "warnings_detected": warnings,
                "elapsed_seconds": round(elapsed, 2),
                "throughput": round(proposal_count / elapsed, 2),
            }

        except Exception as e:
            return {"status": "error", "error": str(e)}

    def test_end_to_end_stress(self) -> dict[str, Any]:
        """Full pipeline stress test."""
        print("\n🔄 [5/6] End-to-End Pipeline Stress Test...")
        start = time.time()

        steps_completed = []
        try:
            # Step 1: Generate adversarial corpus
            print("  → Generating adversarial corpus...")
            fuzz_result = self.test_peachfuzz_stress(count=100)
            if fuzz_result["status"] != "passed":
                return {
                    "status": "failed",
                    "failed_at": "peachfuzz",
                    "details": fuzz_result,
                }
            steps_completed.append("peachfuzz")

            # Step 2: Run threat scans
            print("  → Running threat scans...")
            scan_result = self.test_threat_scanner_stress(proposal_count=20)
            if scan_result["status"] not in ["passed", "partial"]:
                return {
                    "status": "failed",
                    "failed_at": "threat_scanner",
                    "details": scan_result,
                }
            steps_completed.append("threat_scanner")

            # Step 3: Extract datasets
            print("  → Extracting datasets...")
            extract_result = self.test_dataset_extraction_stress()
            if extract_result["status"] != "passed":
                return {
                    "status": "failed",
                    "failed_at": "dataset_extraction",
                    "details": extract_result,
                }
            steps_completed.append("dataset_extraction")

            # Step 4: Audit logging
            print("  → Logging audit trail...")
            audit_result = self.test_peachtrace_stress(count=20)
            if audit_result["status"] != "passed":
                return {
                    "status": "failed",
                    "failed_at": "peachtrace",
                    "details": audit_result,
                }
            steps_completed.append("peachtrace")

            elapsed = time.time() - start

            return {
                "status": "passed",
                "steps_completed": steps_completed,
                "total_elapsed_seconds": round(elapsed, 2),
            }

        except Exception as e:
            return {
                "status": "error",
                "steps_completed": steps_completed,
                "error": str(e),
            }

    def test_hancock_dataset_quality(self) -> dict[str, Any]:
        """Validate Hancock training dataset quality."""
        print("\n✅ [6/6] Hancock Dataset Quality Check...")
        start = time.time()

        try:
            dataset_path = self.workspace / "build/stress/stress-dataset.jsonl"
            if not dataset_path.exists():
                return {"status": "skipped", "reason": "No dataset available"}

            issues = []
            examples = []

            with open(dataset_path) as f:
                for i, line in enumerate(f):
                    try:
                        example = json.loads(line)
                        examples.append(example)

                        # Check required fields
                        if "messages" not in example:
                            issues.append(f"Line {i}: Missing 'messages' field")
                            continue

                        messages = example["messages"]
                        if not isinstance(messages, list):
                            issues.append(f"Line {i}: 'messages' is not a list")
                            continue

                        # Check message structure
                        for j, msg in enumerate(messages):
                            if "role" not in msg or "content" not in msg:
                                issues.append(
                                    f"Line {i}, message {j}: Missing role/content"
                                )

                        # Check for system prompt
                        if not any(m.get("role") == "system" for m in messages):
                            issues.append(f"Line {i}: Missing system message")

                        # Check for user/assistant pairing
                        roles = [m.get("role") for m in messages]
                        if "user" not in roles or "assistant" not in roles:
                            issues.append(f"Line {i}: Missing user/assistant pairing")

                    except json.JSONDecodeError as e:
                        issues.append(f"Line {i}: Invalid JSON - {e}")

            elapsed = time.time() - start

            return {
                "status": "passed" if len(issues) == 0 else "issues_found",
                "examples_checked": len(examples),
                "issues_found": len(issues),
                "issues": issues[:10],  # First 10 issues
                "elapsed_seconds": round(elapsed, 2),
            }

        except Exception as e:
            return {"status": "error", "error": str(e)}

    def run_all(self) -> dict[str, Any]:
        """Run all stress tests."""
        print("╔════════════════════════════════════════════════════════════╗")
        print("║  Governance Security Pipeline - Stress Test Suite         ║")
        print("╚════════════════════════════════════════════════════════════╝")

        # Create stress test directories
        (self.workspace / "build/stress").mkdir(parents=True, exist_ok=True)

        # Run tests
        self.results["tests"]["peachfuzz_stress"] = self.test_peachfuzz_stress(
            count=1000
        )
        self.results["tests"]["peachtrace_stress"] = self.test_peachtrace_stress(
            count=100
        )
        self.results["tests"][
            "dataset_extraction_stress"
        ] = self.test_dataset_extraction_stress()
        self.results["tests"]["threat_scanner_stress"] = self.test_threat_scanner_stress(
            proposal_count=50
        )
        self.results["tests"]["end_to_end_stress"] = self.test_end_to_end_stress()
        self.results["tests"][
            "hancock_quality_check"
        ] = self.test_hancock_dataset_quality()

        # Generate summary
        passed = sum(
            1 for t in self.results["tests"].values() if t.get("status") == "passed"
        )
        failed = sum(
            1
            for t in self.results["tests"].values()
            if t.get("status") in ["failed", "error"]
        )
        total = len(self.results["tests"])

        self.results["summary"] = {
            "total_tests": total,
            "passed": passed,
            "failed": failed,
            "pass_rate": round((passed / total) * 100, 2) if total > 0 else 0,
            "production_ready": failed == 0,
        }

        return self.results

    def print_report(self) -> None:
        """Print stress test report."""
        print("\n")
        print("╔════════════════════════════════════════════════════════════╗")
        print("║              STRESS TEST REPORT                            ║")
        print("╚════════════════════════════════════════════════════════════╝")

        for test_name, result in self.results["tests"].items():
            status_icon = (
                "✅" if result.get("status") == "passed" else "❌" if result.get("status") in ["failed", "error"] else "⚠️"
            )
            print(f"\n{status_icon} {test_name.replace('_', ' ').title()}")
            print(f"   Status: {result.get('status', 'unknown')}")

            # Print key metrics
            for key, value in result.items():
                if key not in ["status", "stdout", "stderr", "issues"]:
                    print(f"   {key}: {value}")

        print("\n" + "=" * 60)
        print(f"SUMMARY: {self.results['summary']['passed']}/{self.results['summary']['total_tests']} tests passed")
        print(f"Pass Rate: {self.results['summary']['pass_rate']}%")
        print(
            f"Production Ready: {'✅ YES' if self.results['summary']['production_ready'] else '❌ NO'}"
        )
        print("=" * 60)


def main() -> int:
    parser = argparse.ArgumentParser(description="Stress test governance pipeline")
    parser.add_argument(
        "--workspace",
        type=Path,
        default=Path.cwd(),
        help="0ai-assurance-network workspace path",
    )
    parser.add_argument(
        "--output",
        type=Path,
        help="Output JSON report path",
    )

    args = parser.parse_args()

    runner = StressTestRunner(workspace=args.workspace)
    results = runner.run_all()
    runner.print_report()

    # Write JSON report
    if args.output:
        with open(args.output, "w") as f:
            json.dump(results, f, indent=2)
        print(f"\n📊 Full report written to: {args.output}")

    return 0 if results["summary"]["production_ready"] else 1


if __name__ == "__main__":
    sys.exit(main())
