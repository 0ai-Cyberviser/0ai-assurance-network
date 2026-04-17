.DEFAULT_GOAL := help
PYTHON ?= python3
GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/go-mod

.PHONY: help validate render-localnet readiness governance-sim governance-queue governance-trends governance-remediation governance-replay governance-drift go-build go-test init-node collect-validator assemble-genesis assemble-localnet clean

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
