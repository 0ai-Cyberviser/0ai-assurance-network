from __future__ import annotations

import json
import sys
from pathlib import Path


SUPPORTED_SCHEMA = "0ai.assurance.governance.artifact"
SUPPORTED_MAJOR = "1"


def main(path: str) -> int:
    artifact = json.loads(Path(path).read_text(encoding="utf-8"))
    if artifact.get("schema") != SUPPORTED_SCHEMA:
        print("unsupported schema", file=sys.stderr)
        return 1

    version = str(artifact.get("schema_version", "0.0.0"))
    if version.split(".", 1)[0] != SUPPORTED_MAJOR:
        print(f"unsupported major version: {version}", file=sys.stderr)
        return 1

    if artifact.get("artifact_type") != "governance_remediation":
        print("artifact is not a governance remediation payload", file=sys.stderr)
        return 1

    blocking_clusters = [
        plan["trend_cluster"]
        for plan in artifact.get("payload", [])
        if plan.get("current_release_readiness") not in {"monitoring", "complete"}
    ]
    print(json.dumps({"blocking_clusters": blocking_clusters}, indent=2))
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1]))
