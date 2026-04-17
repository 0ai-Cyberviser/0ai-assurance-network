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
}
