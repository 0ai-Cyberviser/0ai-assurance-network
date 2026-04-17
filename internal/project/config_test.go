package project

import "testing"

func TestLoadBundle(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	if bundle.Topology.ChainID != "0ai-assurance-1" {
		t.Fatalf("unexpected chain id: %s", bundle.Topology.ChainID)
	}
	if len(bundle.Topology.Validators) != 7 {
		t.Fatalf("expected 7 validators, got %d", len(bundle.Topology.Validators))
	}
	if !bundle.Topology.Governance.CriticalActionsRequireDual {
		t.Fatal("expected dual approval to be enabled")
	}
	if !bundle.Policy.SafeModeDefaults.PermissionedTestnet {
		t.Fatal("expected permissioned_testnet safe mode")
	}
	if bundle.Modules.Milestone != "permissioned-testnet-registry-attestation" {
		t.Fatalf("unexpected module milestone: %s", bundle.Modules.Milestone)
	}
	if len(bundle.Modules.MVPModules) != 2 {
		t.Fatalf("expected 2 mvp modules, got %d", len(bundle.Modules.MVPModules))
	}
	if bundle.Identity.Version != "1.0.0" {
		t.Fatalf("unexpected identity bootstrap version: %s", bundle.Identity.Version)
	}
	if len(bundle.Identity.RoleBindings) != 22 {
		t.Fatalf("expected 22 identity role bindings, got %d", len(bundle.Identity.RoleBindings))
	}
	if len(bundle.CheckpointSigners) == 0 {
		t.Fatal("expected checkpoint signer policy to be loaded")
	}
	if len(bundle.CheckpointSigners["signers"].([]any)) != 14 {
		t.Fatalf("expected 14 checkpoint signers, got %d", len(bundle.CheckpointSigners["signers"].([]any)))
	}
	rotationPolicy := bundle.CheckpointSigners["rotation_policy"].(map[string]any)
	if rotationPolicy["reference_time"] != "2026-04-17T00:00:00Z" {
		t.Fatalf("unexpected signer rotation reference time: %v", rotationPolicy["reference_time"])
	}
	if len(rotationPolicy["approval_roles"].([]any)) != 3 {
		t.Fatalf("expected 3 signer rotation approval roles, got %d", len(rotationPolicy["approval_roles"].([]any)))
	}
	if len(bundle.InferencePolicy) == 0 {
		t.Fatal("expected inference policy to be loaded")
	}
}
