.DEFAULT_GOAL := help
PYTHON ?= python3
GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/go-mod

.PHONY: help validate render-localnet readiness governance-sim governance-queue governance-trends governance-remediation governance-replay governance-drift go-build go-test module-plan identity-plan signer-manifest signer-rotation-receipt signer-rotation-approve signer-rotation-finalize signer-rotation-activate signer-rotation-apply signer-rotation-verify signer-rotation-ledger-append signer-rotation-ledger-reconcile signer-rotation-ledger-export signer-rotation-ledger-verify-export signer-rotation-ledger-archive-index init-node collect-validator assemble-genesis assemble-localnet clean

help:
	@echo ""
	@echo "0AI Assurance Network repo skeleton"
	@echo ""
	@echo "Targets:"
	@echo "  validate         Validate topology, genesis, and policy config"
	@echo "  render-localnet  Generate build/localnet artifacts from config"
	@echo "  readiness        Generate a launch-readiness report"
	@echo "  governance-sim   Simulate governance inference: make governance-sim PROPOSAL=examples/proposals/treasury-grant.json"
	@echo "  governance-queue Score a registry of proposals: make governance-queue REGISTRY=examples/proposals/registry.json"
	@echo "  governance-trends Cluster portfolio-level governance trends"
	@echo "  governance-remediation Emit structured mitigation bundles for active trends"
	@echo "  governance-replay Replay checkpoint event logs into deterministic current state"
	@echo "  governance-drift Compare a proposal against governance history"
	@echo "  go-build         Build the 0aid binary"
	@echo "  go-test          Run Go unit tests"
	@echo "  module-plan      Render the first registry/attestation milestone plan"
	@echo "  identity-plan    Render the permissioned actor and role bootstrap plan"
	@echo "  signer-manifest  Render checkpoint signer ownership and rotation plan"
	@echo "  signer-rotation-receipt Render a replacement-ready signer manifest and rotation receipt stub"
	@echo "  signer-rotation-approve Generate a signed approval artifact for a rotation receipt"
	@echo "  signer-rotation-finalize Validate approvals and render a finalized rotation bundle"
	@echo "  signer-rotation-activate Render an activation plan and checkpoint-signer policy patch"
	@echo "  signer-rotation-apply Validate an activation plan and emit the applied checkpoint signer policy"
	@echo "  signer-rotation-verify Sign a post-activation verification receipt against the applied policy"
	@echo "  signer-rotation-ledger-append Append a verified activation record into the audit ledger"
	@echo "  signer-rotation-ledger-reconcile Reconcile the activation audit ledger against the current signer policy"
	@echo "  signer-rotation-ledger-export Export the activation audit ledger with baseline snapshot and continuity report"
	@echo "  signer-rotation-ledger-verify-export Verify an exported activation audit package for archive retention"
	@echo "  signer-rotation-ledger-archive-index Build an archive index manifest over exported audit packages"
	@echo "  init-node        Generate a development node bundle: make init-node ID=val-1"
	@echo "  collect-validator Normalize a node bundle into a collected manifest"
	@echo "  assemble-genesis Merge collected manifests into a deterministic genesis plan"
	@echo "  assemble-localnet Render a reproducible localnet bundle from collected manifests"
	@echo "  clean            Remove generated localnet artifacts"
	@echo ""

validate:
	PYTHONPATH=src $(PYTHON) -m assurancectl.cli validate

render-localnet:
	PYTHONPATH=src $(PYTHON) -m assurancectl.cli render-localnet

readiness:
	PYTHONPATH=src $(PYTHON) -m assurancectl.cli readiness-report

governance-sim:
	@test -n "$(PROPOSAL)" || (echo "Usage: make governance-sim PROPOSAL=examples/proposals/treasury-grant.json" && exit 1)
	PYTHONPATH=src $(PYTHON) -m assurancectl.cli governance-sim --proposal $(PROPOSAL)

governance-queue:
	@test -n "$(REGISTRY)" || (echo "Usage: make governance-queue REGISTRY=examples/proposals/registry.json" && exit 1)
	PYTHONPATH=src $(PYTHON) -m assurancectl.cli governance-queue --registry $(REGISTRY)

governance-trends:
	@test -n "$(REGISTRY)" || (echo "Usage: make governance-trends REGISTRY=examples/proposals/registry.json HISTORY=examples/proposals/history.json" && exit 1)
	@test -n "$(HISTORY)" || (echo "Usage: make governance-trends REGISTRY=examples/proposals/registry.json HISTORY=examples/proposals/history.json" && exit 1)
	PYTHONPATH=src $(PYTHON) -m assurancectl.cli governance-trends --registry $(REGISTRY) --history $(HISTORY)

governance-remediation:
	@test -n "$(REGISTRY)" || (echo "Usage: make governance-remediation REGISTRY=examples/proposals/registry.json HISTORY=examples/proposals/history.json" && exit 1)
	@test -n "$(HISTORY)" || (echo "Usage: make governance-remediation REGISTRY=examples/proposals/registry.json HISTORY=examples/proposals/history.json" && exit 1)
	PYTHONPATH=src $(PYTHON) -m assurancectl.cli governance-remediation --registry $(REGISTRY) --history $(HISTORY) $(if $(STATUS),--status $(STATUS),)

governance-replay:
	@test -n "$(STATUS)" || (echo "Usage: make governance-replay STATUS=examples/proposals/checkpoint-events.json" && exit 1)
	PYTHONPATH=src $(PYTHON) -m assurancectl.cli governance-replay --status $(STATUS)

governance-drift:
	@test -n "$(PROPOSAL)" || (echo "Usage: make governance-drift PROPOSAL=examples/proposals/emergency-pause.json HISTORY=examples/proposals/history.json" && exit 1)
	@test -n "$(HISTORY)" || (echo "Usage: make governance-drift PROPOSAL=examples/proposals/emergency-pause.json HISTORY=examples/proposals/history.json" && exit 1)
	PYTHONPATH=src $(PYTHON) -m assurancectl.cli governance-drift --proposal $(PROPOSAL) --history $(HISTORY)

go-build:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go build ./cmd/0aid

go-test:
	GOCACHE=$(GOCACHE) GOMODCACHE=$(GOMODCACHE) go test ./...

module-plan:
	./0aid module-plan --root . $(if $(OUT),--out $(OUT),)

identity-plan:
	./0aid identity-plan --root . $(if $(OUT),--out $(OUT),)

signer-manifest:
	./0aid signer-manifest --root . $(if $(OUT),--out $(OUT),)

signer-rotation-receipt:
	@test -n "$(OUTGOING_SIGNER_ID)" || (echo "Usage: make signer-rotation-receipt OUTGOING_SIGNER_ID=governance-chair-bot INCOMING_SIGNER_ID=governance-chair-bot-v2 INCOMING_KEY_ID=governance-chair-dev-v2 EFFECTIVE_AT=2026-04-24T00:00:00Z" && exit 1)
	@test -n "$(INCOMING_SIGNER_ID)" || (echo "Usage: make signer-rotation-receipt OUTGOING_SIGNER_ID=governance-chair-bot INCOMING_SIGNER_ID=governance-chair-bot-v2 INCOMING_KEY_ID=governance-chair-dev-v2 EFFECTIVE_AT=2026-04-24T00:00:00Z" && exit 1)
	@test -n "$(INCOMING_KEY_ID)" || (echo "Usage: make signer-rotation-receipt OUTGOING_SIGNER_ID=governance-chair-bot INCOMING_SIGNER_ID=governance-chair-bot-v2 INCOMING_KEY_ID=governance-chair-dev-v2 EFFECTIVE_AT=2026-04-24T00:00:00Z" && exit 1)
	@test -n "$(EFFECTIVE_AT)" || (echo "Usage: make signer-rotation-receipt OUTGOING_SIGNER_ID=governance-chair-bot INCOMING_SIGNER_ID=governance-chair-bot-v2 INCOMING_KEY_ID=governance-chair-dev-v2 EFFECTIVE_AT=2026-04-24T00:00:00Z" && exit 1)
	./0aid signer-rotation-receipt --root . --outgoing-signer-id $(OUTGOING_SIGNER_ID) --incoming-signer-id $(INCOMING_SIGNER_ID) --incoming-key-id $(INCOMING_KEY_ID) --effective-at $(EFFECTIVE_AT) $(if $(INCOMING_ACTOR_ID),--incoming-actor-id $(INCOMING_ACTOR_ID),) $(if $(INCOMING_ROLES),--incoming-roles $(INCOMING_ROLES),) $(if $(INCOMING_PROVISIONED_AT),--incoming-provisioned-at $(INCOMING_PROVISIONED_AT),) $(if $(INCOMING_ROTATE_BY),--incoming-rotate-by $(INCOMING_ROTATE_BY),) $(if $(RECEIPT_ID),--receipt-id $(RECEIPT_ID),) $(if $(OUT),--out $(OUT),)

signer-rotation-approve:
	@test -n "$(RECEIPT)" || (echo "Usage: make signer-rotation-approve RECEIPT=build/rotation/governance-chair-receipt.json ROLE=governance-ops SIGNER_ID=governance-ops-bot APPROVED_AT=2026-04-23T00:00:00Z" && exit 1)
	@test -n "$(ROLE)" || (echo "Usage: make signer-rotation-approve RECEIPT=build/rotation/governance-chair-receipt.json ROLE=governance-ops SIGNER_ID=governance-ops-bot APPROVED_AT=2026-04-23T00:00:00Z" && exit 1)
	@test -n "$(SIGNER_ID)" || (echo "Usage: make signer-rotation-approve RECEIPT=build/rotation/governance-chair-receipt.json ROLE=governance-ops SIGNER_ID=governance-ops-bot APPROVED_AT=2026-04-23T00:00:00Z" && exit 1)
	@test -n "$(APPROVED_AT)" || (echo "Usage: make signer-rotation-approve RECEIPT=build/rotation/governance-chair-receipt.json ROLE=governance-ops SIGNER_ID=governance-ops-bot APPROVED_AT=2026-04-23T00:00:00Z" && exit 1)
	./0aid signer-rotation-approve --root . --receipt $(RECEIPT) --role $(ROLE) --signer-id $(SIGNER_ID) --approved-at $(APPROVED_AT) $(if $(SIGNATURE_ID),--signature-id $(SIGNATURE_ID),) $(if $(OUT),--out $(OUT),)

signer-rotation-finalize:
	@test -n "$(RECEIPT)" || (echo "Usage: make signer-rotation-finalize RECEIPT=build/rotation/governance-chair-receipt.json APPROVALS=build/rotation/governance-chair-approvals.json" && exit 1)
	@test -n "$(APPROVALS)" || (echo "Usage: make signer-rotation-finalize RECEIPT=build/rotation/governance-chair-receipt.json APPROVALS=build/rotation/governance-chair-approvals.json" && exit 1)
	./0aid signer-rotation-finalize --root . --receipt $(RECEIPT) --approvals $(APPROVALS) $(if $(OUT),--out $(OUT),)

signer-rotation-activate:
	@test -n "$(BUNDLE)" || (echo "Usage: make signer-rotation-activate BUNDLE=build/rotation/governance-chair-approved-bundle.json INCOMING_SHARED_SECRET=dev-secret-governance-chair-v2" && exit 1)
	@test -n "$(INCOMING_SHARED_SECRET)" || (echo "Usage: make signer-rotation-activate BUNDLE=build/rotation/governance-chair-approved-bundle.json INCOMING_SHARED_SECRET=dev-secret-governance-chair-v2" && exit 1)
	./0aid signer-rotation-activate --root . --bundle $(BUNDLE) --incoming-shared-secret $(INCOMING_SHARED_SECRET) $(if $(OUT),--out $(OUT),)

signer-rotation-apply:
	@test -n "$(PLAN)" || (echo "Usage: make signer-rotation-apply PLAN=build/rotation/governance-chair-activation-plan.json" && exit 1)
	./0aid signer-rotation-apply --root . --plan $(PLAN) $(if $(POLICY_OUT),--policy-out $(POLICY_OUT),) $(if $(OUT),--out $(OUT),)

signer-rotation-verify:
	@test -n "$(PLAN)" || (echo "Usage: make signer-rotation-verify PLAN=build/rotation/governance-chair-activation-plan.json VERIFIED_AT=2026-04-24T00:15:00Z" && exit 1)
	@test -n "$(VERIFIED_AT)" || (echo "Usage: make signer-rotation-verify PLAN=build/rotation/governance-chair-activation-plan.json VERIFIED_AT=2026-04-24T00:15:00Z" && exit 1)
	./0aid signer-rotation-verify --root . --plan $(PLAN) --verified-at $(VERIFIED_AT) $(if $(POLICY),--policy $(POLICY),) $(if $(SIGNATURE_ID),--signature-id $(SIGNATURE_ID),) $(if $(OUT),--out $(OUT),)

signer-rotation-ledger-append:
	@test -n "$(APPLY)" || (echo "Usage: make signer-rotation-ledger-append APPLY=build/rotation/governance-chair-apply-result.json VERIFICATION=build/rotation/governance-chair-verification.json" && exit 1)
	@test -n "$(VERIFICATION)" || (echo "Usage: make signer-rotation-ledger-append APPLY=build/rotation/governance-chair-apply-result.json VERIFICATION=build/rotation/governance-chair-verification.json" && exit 1)
	./0aid signer-rotation-ledger-append --apply $(APPLY) --verification $(VERIFICATION) $(if $(LEDGER),--ledger $(LEDGER),) $(if $(LEDGER_OUT),--ledger-out $(LEDGER_OUT),) $(if $(OUT),--out $(OUT),)

signer-rotation-ledger-reconcile:
	./0aid signer-rotation-ledger-reconcile --root . $(if $(LEDGER),--ledger $(LEDGER),) $(if $(POLICY),--policy $(POLICY),) $(if $(OUT),--out $(OUT),)

signer-rotation-ledger-export:
	./0aid signer-rotation-ledger-export --root . $(if $(LEDGER),--ledger $(LEDGER),) $(if $(POLICY),--policy $(POLICY),) $(if $(RECONCILE),--reconcile $(RECONCILE),) $(if $(OUT),--out $(OUT),)

signer-rotation-ledger-verify-export:
	@test -n "$(EXPORT)" || (echo "Usage: make signer-rotation-ledger-verify-export EXPORT=build/rotation/governance-chair-audit-export.json" && exit 1)
	./0aid signer-rotation-ledger-verify-export --export $(EXPORT) $(if $(OUT),--out $(OUT),)

signer-rotation-ledger-archive-index:
	@test -n "$(EXPORTS)" || (echo "Usage: make signer-rotation-ledger-archive-index EXPORTS=build/rotation/current-audit-export.json,build/rotation/governance-chair-audit-export.json" && exit 1)
	./0aid signer-rotation-ledger-archive-index --exports $(EXPORTS) $(if $(OUT),--out $(OUT),)

init-node:
	@test -n "$(ID)" || (echo "Usage: make init-node ID=val-1" && exit 1)
	./0aid init-node --root . --id $(ID) --out ./build/nodes/$(ID)

collect-validator:
	@test -n "$(BUNDLE)" || (echo "Usage: make collect-validator BUNDLE=build/nodes/val-1 OUT=build/collection/val-1.json" && exit 1)
	@test -n "$(OUT)" || (echo "Usage: make collect-validator BUNDLE=build/nodes/val-1 OUT=build/collection/val-1.json" && exit 1)
	./0aid collect-validator --bundle $(BUNDLE) --out $(OUT)

assemble-genesis:
	@test -n "$(COLLECTION)" || (echo "Usage: make assemble-genesis COLLECTION=build/collection OUT=build/assembled/genesis-plan.json" && exit 1)
	./0aid assemble-genesis --root . --collection $(COLLECTION) $(if $(OUT),--out $(OUT),)

assemble-localnet:
	@test -n "$(COLLECTION)" || (echo "Usage: make assemble-localnet COLLECTION=build/collection OUT=build/assembled" && exit 1)
	@test -n "$(OUT)" || (echo "Usage: make assemble-localnet COLLECTION=build/collection OUT=build/assembled" && exit 1)
	./0aid assemble-localnet --root . --collection $(COLLECTION) --out $(OUT)

clean:
	rm -rf build/localnet .cache 0aid
