package project

import "testing"

func TestIdentityFoundationPlan(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	plan, err := IdentityFoundationPlan(bundle)
	if err != nil {
		t.Fatalf("IdentityFoundationPlan failed: %v", err)
	}

	if plan.Milestone != "permissioned-testnet-registry-attestation" {
		t.Fatalf("unexpected milestone: %s", plan.Milestone)
	}
	if plan.ChainID != "0ai-assurance-1" {
		t.Fatalf("unexpected chain id: %s", plan.ChainID)
	}
	if plan.ActorCount != 11 {
		t.Fatalf("expected 11 actors, got %d", plan.ActorCount)
	}
	if plan.ActiveActorCount != 11 {
		t.Fatalf("expected 11 active actors, got %d", plan.ActiveActorCount)
	}
	if len(plan.RequiredRoles) != 8 {
		t.Fatalf("expected 8 required roles, got %d", len(plan.RequiredRoles))
	}
	if len(plan.MissingRoles) != 0 {
		t.Fatalf("expected no missing roles, got %v", plan.MissingRoles)
	}
	foundNetworkAdmin := false
	for _, role := range plan.RolePlan {
		if role.Role == "network_admin" {
			foundNetworkAdmin = true
			if len(role.BoundActors) != 1 || role.BoundActors[0] != "op-network-admin-1" {
				t.Fatalf("unexpected network_admin binding: %+v", role)
			}
		}
	}
	if !foundNetworkAdmin {
		t.Fatal("expected network_admin role plan entry")
	}
}
