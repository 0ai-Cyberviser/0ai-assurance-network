package project

import (
	"encoding/json"
	"strings"
	"testing"
)

func cloneSignerPolicy(t *testing.T, value map[string]any) map[string]any {
	t.Helper()

	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal signer policy: %v", err)
	}
	var cloned map[string]any
	if err := json.Unmarshal(encoded, &cloned); err != nil {
		t.Fatalf("unmarshal signer policy: %v", err)
	}
	return cloned
}

func TestSignerManifest(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	manifest, err := SignerManifest(bundle)
	if err != nil {
		t.Fatalf("SignerManifest failed: %v", err)
	}

	if manifest.ChainID != "0ai-assurance-1" {
		t.Fatalf("unexpected chain id: %s", manifest.ChainID)
	}
	if manifest.SignerCount != 14 {
		t.Fatalf("expected 14 signers, got %d", manifest.SignerCount)
	}
	if manifest.ActiveSignerCount != 14 {
		t.Fatalf("expected 14 active signers, got %d", manifest.ActiveSignerCount)
	}
	if len(manifest.ExpiringSignerIDs) != 1 || manifest.ExpiringSignerIDs[0] != "governance-chair-bot" {
		t.Fatalf("unexpected expiring signer set: %#v", manifest.ExpiringSignerIDs)
	}
	if len(manifest.RotationPlan) != 14 {
		t.Fatalf("expected 14 rotation plan entries, got %d", len(manifest.RotationPlan))
	}
	if manifest.RotationPlan[0].SignerID != "governance-chair-bot" {
		t.Fatalf("unexpected first rotation entry: %+v", manifest.RotationPlan[0])
	}

	foundGovernanceChair := false
	for _, signer := range manifest.Signers {
		if signer.SignerID != "governance-chair-bot" {
			continue
		}
		foundGovernanceChair = true
		if signer.ActorID != "op-governance-chair-1" {
			t.Fatalf("unexpected governance chair actor id: %s", signer.ActorID)
		}
		if signer.RotationStatus != "expiring" {
			t.Fatalf("expected governance chair signer to be expiring, got %s", signer.RotationStatus)
		}
		if len(signer.RoleCoverage) != 1 || len(signer.RoleCoverage[0].ReferencedBy) != 1 {
			t.Fatalf("unexpected governance chair role coverage: %+v", signer.RoleCoverage)
		}
		if signer.RoleCoverage[0].ReferencedBy[0] != "governance:phase_owner:release_blocker" {
			t.Fatalf("unexpected governance chair referenced_by: %+v", signer.RoleCoverage[0].ReferencedBy)
		}
	}
	if !foundGovernanceChair {
		t.Fatal("expected governance-chair-bot signer entry")
	}
}

func TestSignerManifestFailsOnDuplicateActorOwnership(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	policy := cloneSignerPolicy(t, bundle.CheckpointSigners)
	signers := policy["signers"].([]any)
	second := signers[1].(map[string]any)
	second["actor_id"] = "op-governance-chair-1"
	second["roles"] = []any{"governance-chair"}
	bundle.CheckpointSigners = policy

	_, err = SignerManifest(bundle)
	if err == nil || !strings.Contains(err.Error(), "duplicate checkpoint signer actor ownership") {
		t.Fatalf("expected duplicate actor ownership error, got %v", err)
	}
}

func TestSignerManifestFailsOnMissingRoleCoverage(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	policy := cloneSignerPolicy(t, bundle.CheckpointSigners)
	filtered := make([]any, 0)
	for _, rawSigner := range policy["signers"].([]any) {
		signer := rawSigner.(map[string]any)
		if signer["signer_id"] == "treasury-review-chair-bot" {
			continue
		}
		filtered = append(filtered, signer)
	}
	policy["signers"] = filtered
	bundle.CheckpointSigners = policy

	_, err = SignerManifest(bundle)
	if err == nil || !strings.Contains(err.Error(), "checkpoint signer coverage missing roles: treasury-review-chair") {
		t.Fatalf("expected missing signer coverage error, got %v", err)
	}
}

func TestSignerManifestFailsOnStaleRotationMetadata(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	policy := cloneSignerPolicy(t, bundle.CheckpointSigners)
	first := policy["signers"].([]any)[0].(map[string]any)
	first["rotate_by"] = "2026-04-01T00:00:00Z"
	bundle.CheckpointSigners = policy

	_, err = SignerManifest(bundle)
	if err == nil || !strings.Contains(err.Error(), "stale rotation metadata") {
		t.Fatalf("expected stale rotation error, got %v", err)
	}
}
