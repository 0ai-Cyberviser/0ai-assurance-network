package project

import (
	"strings"
	"testing"
)

func TestNetworkPlan(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	plan := NetworkPlan(bundle)
	if plan.ValidatorCount != 7 {
		t.Fatalf("expected 7 validators, got %d", plan.ValidatorCount)
	}
	if plan.SeedCount != 1 {
		t.Fatalf("expected 1 seed, got %d", plan.SeedCount)
	}
}

func TestRenderedGenesisIncludesLocalnetMetadata(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	rendered := RenderedGenesis(bundle)
	localnet, ok := rendered["localnet"].(map[string]any)
	if !ok {
		t.Fatal("missing localnet metadata")
	}

	validators, ok := localnet["validators"].([]string)
	if !ok {
		t.Fatal("missing validator list")
	}
	if len(validators) != 7 {
		t.Fatalf("expected 7 validator names, got %d", len(validators))
	}
}

func TestValidatorPlan(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	plan, err := ValidatorPlan(bundle, "val-5")
	if err != nil {
		t.Fatalf("ValidatorPlan failed: %v", err)
	}
	if plan.Moniker != "validator-5" {
		t.Fatalf("unexpected moniker: %s", plan.Moniker)
	}
	if !strings.Contains(plan.Home, "/chain/validator-5") {
		t.Fatalf("unexpected home path: %s", plan.Home)
	}
}

func TestDockerComposeIncludesValidatorSeven(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	compose := DockerCompose(bundle)
	if !strings.Contains(compose, "validator-7") {
		t.Fatal("expected compose output to include validator-7")
	}
	if !strings.Contains(compose, "ghcr.io/0ai-cyberviser/0ai-assurance-network:dev") {
		t.Fatal("expected compose output to include container image")
	}
}
