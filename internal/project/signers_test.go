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
	if manifest.Version != "1.1.0" {
		t.Fatalf("unexpected manifest version: %s", manifest.Version)
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
		if signer.NextRotationReceiptID == "" || signer.ReplacementManifestRef == "" {
			t.Fatalf("expected governance chair signer to include receipt metadata: %+v", signer)
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

func TestSignerRotationReceipt(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	receipt, err := SignerRotationReceipt(bundle, SignerRotationReceiptRequest{
		OutgoingSignerID: "governance-chair-bot",
		IncomingSignerID: "governance-chair-bot-v2",
		IncomingKeyID:    "governance-chair-dev-v2",
		EffectiveAt:      "2026-04-24T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("SignerRotationReceipt failed: %v", err)
	}

	if receipt.CoverageStatus != "ready" {
		t.Fatalf("unexpected coverage status: %s", receipt.CoverageStatus)
	}
	if receipt.ReceiptID == "" || receipt.ReplacementManifestRef == "" {
		t.Fatalf("expected receipt metadata, got %+v", receipt)
	}
	if receipt.OutgoingSigner.SignerID != "governance-chair-bot" {
		t.Fatalf("unexpected outgoing signer: %+v", receipt.OutgoingSigner)
	}
	if receipt.IncomingSigner.SignerID != "governance-chair-bot-v2" {
		t.Fatalf("unexpected incoming signer: %+v", receipt.IncomingSigner)
	}
	if len(receipt.ApprovalRequirements) != 3 {
		t.Fatalf("expected 3 approval requirements, got %d", len(receipt.ApprovalRequirements))
	}
	if receipt.ReplacementSignerManifest.SignerCount != 14 {
		t.Fatalf("expected replacement manifest with 14 signers, got %d", receipt.ReplacementSignerManifest.SignerCount)
	}
}

func TestSignerRotationReceiptFailsOnReplacementCoverageGap(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	bundle.Identity.RoleBindings = append(bundle.Identity.RoleBindings, IdentityRoleBinding{
		ActorID:   "op-governance-chair-1",
		Role:      "network_admin",
		Scope:     "network",
		GrantedBy: "genesis",
		Status:    "active",
	})

	_, err = SignerRotationReceipt(bundle, SignerRotationReceiptRequest{
		OutgoingSignerID: "governance-chair-bot",
		IncomingSignerID: "governance-chair-bot-v2",
		IncomingKeyID:    "governance-chair-dev-v2",
		IncomingRoles:    []string{"network_admin"},
		EffectiveAt:      "2026-04-24T00:00:00Z",
	})
	if err == nil || !strings.Contains(err.Error(), "checkpoint signer coverage missing roles: governance-chair") {
		t.Fatalf("expected coverage gap error, got %v", err)
	}
}

func TestSignerRotationReceiptFailsOnDuplicateReplacementOwnership(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	bundle.Identity.RoleBindings = append(bundle.Identity.RoleBindings, IdentityRoleBinding{
		ActorID:   "op-governance-ops-1",
		Role:      "governance-chair",
		Scope:     "governance",
		GrantedBy: "genesis",
		Status:    "active",
	})

	_, err = SignerRotationReceipt(bundle, SignerRotationReceiptRequest{
		OutgoingSignerID: "governance-chair-bot",
		IncomingSignerID: "governance-chair-bot-v2",
		IncomingKeyID:    "governance-chair-dev-v2",
		IncomingActorID:  "op-governance-ops-1",
		IncomingRoles:    []string{"governance-chair"},
		EffectiveAt:      "2026-04-24T00:00:00Z",
	})
	if err == nil || !strings.Contains(err.Error(), "duplicate checkpoint signer actor ownership") {
		t.Fatalf("expected duplicate replacement ownership error, got %v", err)
	}
}

func TestSignerRotationReceiptFailsOnInvalidEffectiveAtOrdering(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	_, err = SignerRotationReceipt(bundle, SignerRotationReceiptRequest{
		OutgoingSignerID: "governance-chair-bot",
		IncomingSignerID: "governance-chair-bot-v2",
		IncomingKeyID:    "governance-chair-dev-v2",
		EffectiveAt:      "2026-05-10T00:00:00Z",
	})
	if err == nil || !strings.Contains(err.Error(), "effective_at must be on or before the outgoing signer rotate_by time") {
		t.Fatalf("expected effective_at ordering error, got %v", err)
	}
}

func mustSignerRotationReceipt(t *testing.T, bundle Bundle) SignerRotationReceiptOutput {
	t.Helper()

	receipt, err := SignerRotationReceipt(bundle, SignerRotationReceiptRequest{
		OutgoingSignerID: "governance-chair-bot",
		IncomingSignerID: "governance-chair-bot-v2",
		IncomingKeyID:    "governance-chair-dev-v2",
		EffectiveAt:      "2026-04-24T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("SignerRotationReceipt failed: %v", err)
	}
	return receipt
}

func TestSignerRotationApproval(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	receipt := mustSignerRotationReceipt(t, bundle)
	approval, err := GenerateSignerRotationApproval(bundle, SignerRotationApprovalRequest{
		Receipt:      receipt,
		ApprovalRole: "governance-ops",
		SignerID:     "governance-ops-bot",
		ApprovedAt:   "2026-04-23T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("GenerateSignerRotationApproval failed: %v", err)
	}
	if approval.ReceiptID != receipt.ReceiptID {
		t.Fatalf("unexpected receipt id: %s", approval.ReceiptID)
	}
	if approval.ApprovalRole != "governance-ops" {
		t.Fatalf("unexpected approval role: %s", approval.ApprovalRole)
	}
	if approval.Signature.Value == "" || approval.Signature.SignatureID == "" {
		t.Fatalf("expected signature metadata, got %+v", approval.Signature)
	}
}

func TestSignerRotationFinalize(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	receipt := mustSignerRotationReceipt(t, bundle)
	approvals := []SignerRotationApproval{}
	for _, item := range []struct {
		role     string
		signerID string
	}{
		{role: "governance-ops", signerID: "governance-ops-bot"},
		{role: "token-house-secretariat", signerID: "token-house-secretariat-bot"},
		{role: "telemetry-ops", signerID: "telemetry-ops-bot"},
	} {
		approval, err := GenerateSignerRotationApproval(bundle, SignerRotationApprovalRequest{
			Receipt:      receipt,
			ApprovalRole: item.role,
			SignerID:     item.signerID,
			ApprovedAt:   "2026-04-23T00:00:00Z",
		})
		if err != nil {
			t.Fatalf("GenerateSignerRotationApproval(%s) failed: %v", item.role, err)
		}
		approvals = append(approvals, approval)
	}

	finalized, err := SignerRotationFinalize(bundle, SignerRotationFinalizeRequest{
		Receipt:   receipt,
		Approvals: approvals,
	})
	if err != nil {
		t.Fatalf("SignerRotationFinalize failed: %v", err)
	}
	if finalized.Status != "approved" {
		t.Fatalf("unexpected finalized status: %s", finalized.Status)
	}
	if len(finalized.Approvals) != 3 {
		t.Fatalf("expected 3 approvals, got %d", len(finalized.Approvals))
	}
	if finalized.ReplacementManifestRef != receipt.ReplacementManifestRef {
		t.Fatalf("unexpected replacement manifest ref: %s", finalized.ReplacementManifestRef)
	}
}

func TestSignerRotationFinalizeFailsOnMissingApprovalCoverage(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	receipt := mustSignerRotationReceipt(t, bundle)
	approval, err := GenerateSignerRotationApproval(bundle, SignerRotationApprovalRequest{
		Receipt:      receipt,
		ApprovalRole: "governance-ops",
		SignerID:     "governance-ops-bot",
		ApprovedAt:   "2026-04-23T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("GenerateSignerRotationApproval failed: %v", err)
	}

	_, err = SignerRotationFinalize(bundle, SignerRotationFinalizeRequest{
		Receipt:   receipt,
		Approvals: []SignerRotationApproval{approval},
	})
	if err == nil || !strings.Contains(err.Error(), "missing approval coverage for roles") {
		t.Fatalf("expected missing approval coverage error, got %v", err)
	}
}

func TestSignerRotationFinalizeFailsOnApprovalReceiptDrift(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	receipt := mustSignerRotationReceipt(t, bundle)
	approval, err := GenerateSignerRotationApproval(bundle, SignerRotationApprovalRequest{
		Receipt:      receipt,
		ApprovalRole: "governance-ops",
		SignerID:     "governance-ops-bot",
		ApprovedAt:   "2026-04-23T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("GenerateSignerRotationApproval failed: %v", err)
	}
	approval.ReceiptDigest = "deadbeef"

	_, err = SignerRotationFinalize(bundle, SignerRotationFinalizeRequest{
		Receipt:   receipt,
		Approvals: []SignerRotationApproval{approval},
	})
	if err == nil || !strings.Contains(err.Error(), "approval receipt_digest mismatch") {
		t.Fatalf("expected approval receipt drift error, got %v", err)
	}
}

func mustSignerRotationFinalizedBundle(t *testing.T, bundle Bundle) SignerRotationFinalizedBundle {
	t.Helper()

	receipt := mustSignerRotationReceipt(t, bundle)
	approvals := []SignerRotationApproval{}
	for _, item := range []struct {
		role     string
		signerID string
	}{
		{role: "governance-ops", signerID: "governance-ops-bot"},
		{role: "token-house-secretariat", signerID: "token-house-secretariat-bot"},
		{role: "telemetry-ops", signerID: "telemetry-ops-bot"},
	} {
		approval, err := GenerateSignerRotationApproval(bundle, SignerRotationApprovalRequest{
			Receipt:      receipt,
			ApprovalRole: item.role,
			SignerID:     item.signerID,
			ApprovedAt:   "2026-04-23T00:00:00Z",
		})
		if err != nil {
			t.Fatalf("GenerateSignerRotationApproval(%s) failed: %v", item.role, err)
		}
		approvals = append(approvals, approval)
	}

	finalized, err := SignerRotationFinalize(bundle, SignerRotationFinalizeRequest{
		Receipt:   receipt,
		Approvals: approvals,
	})
	if err != nil {
		t.Fatalf("SignerRotationFinalize failed: %v", err)
	}
	return finalized
}

func TestSignerRotationActivation(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	finalized := mustSignerRotationFinalizedBundle(t, bundle)
	activation, err := SignerRotationActivation(bundle, SignerRotationActivationRequest{
		FinalizedBundle:      finalized,
		IncomingSharedSecret: "dev-secret-governance-chair-v2",
	})
	if err != nil {
		t.Fatalf("SignerRotationActivation failed: %v", err)
	}
	if activation.Status != "ready" {
		t.Fatalf("unexpected activation status: %s", activation.Status)
	}
	if activation.PolicyPatch.RemoveSignerID != "governance-chair-bot" {
		t.Fatalf("unexpected removed signer: %s", activation.PolicyPatch.RemoveSignerID)
	}
	if activation.PolicyPatch.AddSigner.SignerID != "governance-chair-bot-v2" {
		t.Fatalf("unexpected replacement signer: %s", activation.PolicyPatch.AddSigner.SignerID)
	}
	if activation.ResultingPolicy.RotationPolicy.ReferenceTime != finalized.EffectiveAt {
		t.Fatalf("unexpected resulting reference time: %s", activation.ResultingPolicy.RotationPolicy.ReferenceTime)
	}
	if activation.TargetPolicyVersion == activation.CurrentPolicyVersion {
		t.Fatalf("expected target policy version to differ from current version")
	}
}

func TestSignerRotationActivationFailsWithoutIncomingSharedSecret(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	finalized := mustSignerRotationFinalizedBundle(t, bundle)
	_, err = SignerRotationActivation(bundle, SignerRotationActivationRequest{
		FinalizedBundle: finalized,
	})
	if err == nil || !strings.Contains(err.Error(), "incoming shared secret must be set") {
		t.Fatalf("expected missing shared secret error, got %v", err)
	}
}

func TestSignerRotationActivationFailsOnBundleDrift(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	finalized := mustSignerRotationFinalizedBundle(t, bundle)
	bundle.CheckpointSigners["version"] = "checkpoint-signers-2026-04-18"
	_, err = SignerRotationActivation(bundle, SignerRotationActivationRequest{
		FinalizedBundle:      finalized,
		IncomingSharedSecret: "dev-secret-governance-chair-v2",
	})
	if err == nil || !strings.Contains(err.Error(), "signer rotation finalized bundle drift detected") {
		t.Fatalf("expected finalized bundle drift error, got %v", err)
	}
}
