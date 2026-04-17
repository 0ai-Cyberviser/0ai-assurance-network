package project

import "testing"

func TestMilestoneModulePlan(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	plan := MilestoneModulePlan(bundle)
	if plan.Milestone != "permissioned-testnet-registry-attestation" {
		t.Fatalf("unexpected milestone: %s", plan.Milestone)
	}
	if plan.ChainID != "0ai-assurance-1" {
		t.Fatalf("unexpected chain id: %s", plan.ChainID)
	}
	if plan.NetworkMode != "permissioned_testnet" {
		t.Fatalf("unexpected network mode: %s", plan.NetworkMode)
	}
	if len(plan.MVPModules) != 2 {
		t.Fatalf("expected 2 mvp modules, got %d", len(plan.MVPModules))
	}
	if len(plan.DependencySurfaces) != 3 {
		t.Fatalf("expected 3 dependency surfaces, got %d", len(plan.DependencySurfaces))
	}
	if len(plan.Rollout) != 4 {
		t.Fatalf("expected 4 rollout phases, got %d", len(plan.Rollout))
	}
	if plan.ImplementationSequence[0] != "identity-foundation" {
		t.Fatalf("unexpected first rollout phase: %s", plan.ImplementationSequence[0])
	}
	if plan.ImplementationSequence[len(plan.ImplementationSequence)-1] != "governance-and-validator-hooks" {
		t.Fatalf("unexpected final rollout phase: %s", plan.ImplementationSequence[len(plan.ImplementationSequence)-1])
	}
}
