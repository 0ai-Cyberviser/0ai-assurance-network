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

func mustSignerRotationActivationPlan(t *testing.T, bundle Bundle) SignerRotationActivationPlan {
	t.Helper()

	finalized := mustSignerRotationFinalizedBundle(t, bundle)
	activation, err := SignerRotationActivation(bundle, SignerRotationActivationRequest{
		FinalizedBundle:      finalized,
		IncomingSharedSecret: "dev-secret-governance-chair-v2",
	})
	if err != nil {
		t.Fatalf("SignerRotationActivation failed: %v", err)
	}
	return activation
}

func TestSignerRotationApply(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	plan := mustSignerRotationActivationPlan(t, bundle)
	result, err := SignerRotationApply(bundle, SignerRotationApplyRequest{
		ActivationPlan: plan,
	})
	if err != nil {
		t.Fatalf("SignerRotationApply failed: %v", err)
	}
	if result.Status != "applied" {
		t.Fatalf("unexpected apply status: %s", result.Status)
	}
	if result.TargetPolicyVersion != plan.TargetPolicyVersion {
		t.Fatalf("unexpected target policy version: %s", result.TargetPolicyVersion)
	}
	if result.ActivationPlanDigest == "" || result.AppliedPolicyDigest == "" {
		t.Fatalf("expected apply digests, got %+v", result)
	}
	encodedExpected, err := json.Marshal(plan.ResultingPolicy)
	if err != nil {
		t.Fatalf("marshal expected applied policy: %v", err)
	}
	encodedActual, err := json.Marshal(result.AppliedPolicy)
	if err != nil {
		t.Fatalf("marshal actual applied policy: %v", err)
	}
	if !bytesEqual(encodedExpected, encodedActual) {
		t.Fatalf("applied policy does not match activation result")
	}
}

func TestSignerRotationApplyFailsOnPlanDrift(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	plan := mustSignerRotationActivationPlan(t, bundle)
	plan.PolicyPatch.RemoveSignerID = "nonexistent-signer"
	_, err = SignerRotationApply(bundle, SignerRotationApplyRequest{
		ActivationPlan: plan,
	})
	if err == nil || !strings.Contains(err.Error(), "signer rotation activation plan drift detected") {
		t.Fatalf("expected activation plan drift error, got %v", err)
	}
}

func TestSignerRotationVerify(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	plan := mustSignerRotationActivationPlan(t, bundle)
	applied, err := SignerRotationApply(bundle, SignerRotationApplyRequest{
		ActivationPlan: plan,
	})
	if err != nil {
		t.Fatalf("SignerRotationApply failed: %v", err)
	}
	receipt, err := SignerRotationVerify(bundle, SignerRotationVerifyRequest{
		ActivationPlan: plan,
		Policy:         applied.AppliedPolicy,
		VerifiedAt:     "2026-04-24T00:15:00Z",
	})
	if err != nil {
		t.Fatalf("SignerRotationVerify failed: %v", err)
	}
	if receipt.Status != "verified" {
		t.Fatalf("unexpected verification status: %s", receipt.Status)
	}
	if receipt.SignerID != "governance-chair-bot-v2" {
		t.Fatalf("unexpected verification signer: %s", receipt.SignerID)
	}
	if receipt.Signature.Value == "" || receipt.Signature.SignatureID == "" {
		t.Fatalf("expected verification signature metadata, got %+v", receipt.Signature)
	}
}

func TestSignerRotationVerifyFailsOnPolicyDrift(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	plan := mustSignerRotationActivationPlan(t, bundle)
	policy, err := currentSignerPolicy(bundle)
	if err != nil {
		t.Fatalf("currentSignerPolicy failed: %v", err)
	}
	_, err = SignerRotationVerify(bundle, SignerRotationVerifyRequest{
		ActivationPlan: plan,
		Policy:         policy,
		VerifiedAt:     "2026-04-24T00:15:00Z",
	})
	if err == nil || !strings.Contains(err.Error(), "signer rotation applied policy drift detected") {
		t.Fatalf("expected applied policy drift error, got %v", err)
	}
}

func mustSignerRotationApplyResult(t *testing.T, bundle Bundle) SignerRotationApplyResult {
	t.Helper()

	plan := mustSignerRotationActivationPlan(t, bundle)
	result, err := SignerRotationApply(bundle, SignerRotationApplyRequest{
		ActivationPlan: plan,
	})
	if err != nil {
		t.Fatalf("SignerRotationApply failed: %v", err)
	}
	return result
}

func mustSignerRotationVerificationReceipt(t *testing.T, bundle Bundle) SignerRotationVerificationReceipt {
	t.Helper()

	plan := mustSignerRotationActivationPlan(t, bundle)
	applied := mustSignerRotationApplyResult(t, bundle)
	receipt, err := SignerRotationVerify(bundle, SignerRotationVerifyRequest{
		ActivationPlan: plan,
		Policy:         applied.AppliedPolicy,
		VerifiedAt:     "2026-04-24T00:15:00Z",
	})
	if err != nil {
		t.Fatalf("SignerRotationVerify failed: %v", err)
	}
	return receipt
}

func TestSignerRotationActivationAuditAppend(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	appendResult, err := SignerRotationActivationAuditAppend(SignerRotationActivationAuditAppendRequest{
		ApplyResult:         mustSignerRotationApplyResult(t, bundle),
		VerificationReceipt: mustSignerRotationVerificationReceipt(t, bundle),
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditAppend failed: %v", err)
	}
	if appendResult.Status != "appended" {
		t.Fatalf("unexpected append status: %s", appendResult.Status)
	}
	if appendResult.AppendedIndex != 0 {
		t.Fatalf("unexpected appended index: %d", appendResult.AppendedIndex)
	}
	if appendResult.Ledger.EntryCount != 1 || len(appendResult.Ledger.Entries) != 1 {
		t.Fatalf("expected one ledger entry, got %+v", appendResult.Ledger)
	}
	if appendResult.AppendedEntry.ReceiptID != "rotation-governance-chair-bot-20260424t000000z" {
		t.Fatalf("unexpected appended receipt id: %s", appendResult.AppendedEntry.ReceiptID)
	}
}

func TestSignerRotationActivationAuditAppendFailsOnDigestMismatch(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	applyResult := mustSignerRotationApplyResult(t, bundle)
	verification := mustSignerRotationVerificationReceipt(t, bundle)
	verification.PolicyDigest = "deadbeef"
	_, err = SignerRotationActivationAuditAppend(SignerRotationActivationAuditAppendRequest{
		ApplyResult:         applyResult,
		VerificationReceipt: verification,
	})
	if err == nil || !strings.Contains(err.Error(), "activation audit policy_digest mismatch") {
		t.Fatalf("expected activation audit digest mismatch error, got %v", err)
	}
}

func TestSignerRotationActivationAuditAppendFailsOnDuplicateReceipt(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	applyResult := mustSignerRotationApplyResult(t, bundle)
	verification := mustSignerRotationVerificationReceipt(t, bundle)
	existing := SignerRotationActivationAuditLedger{
		Version:    "1.0.0",
		Status:     "active",
		ChainID:    applyResult.ChainID,
		PolicyPath: applyResult.PolicyPath,
		Entries: []SignerRotationActivationAuditEntry{
			activationAuditEntry(applyResult, verification),
		},
		EntryCount: 1,
	}
	_, err = SignerRotationActivationAuditAppend(SignerRotationActivationAuditAppendRequest{
		ApplyResult:         applyResult,
		VerificationReceipt: verification,
		ExistingLedger:      existing,
	})
	if err == nil || !strings.Contains(err.Error(), "activation audit receipt_id already recorded") {
		t.Fatalf("expected duplicate receipt error, got %v", err)
	}
}

func TestSignerRotationActivationAuditAppendFailsOnOutOfOrderVerification(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	applyResult := mustSignerRotationApplyResult(t, bundle)
	verification := mustSignerRotationVerificationReceipt(t, bundle)
	existing := SignerRotationActivationAuditLedger{
		Version:    "1.0.0",
		Status:     "active",
		ChainID:    applyResult.ChainID,
		PolicyPath: applyResult.PolicyPath,
		Entries: []SignerRotationActivationAuditEntry{
			{
				Version:              "1.0.0",
				Status:               "verified",
				ReceiptID:            "rotation-older-20260423t000000z",
				ChainID:              applyResult.ChainID,
				PolicyPath:           applyResult.PolicyPath,
				TargetPolicyVersion:  "checkpoint-signers-older",
				EffectiveAt:          "2026-04-23T23:59:00Z",
				VerifiedAt:           "2026-04-24T00:16:00Z",
				ActivationPlanDigest: "older-plan",
				AppliedPolicyDigest:  "older-policy",
				SignerID:             "governance-chair-bot-v2",
				KeyID:                "governance-chair-dev-v2",
				ActorID:              "op-governance-chair-1",
				ActorDisplayName:     "Governance Chair 1",
				OrganizationID:       "org-0ai-core",
				OrganizationName:     "0AI Core",
				Signature: SignerRotationVerificationSignature{
					Format:      "0ai-hmac-sha256-v1",
					SignatureID: "older-signature",
					SignedAt:    "2026-04-24T00:16:00Z",
					ExpiresAt:   "2026-04-25T00:16:00Z",
					Value:       "older-value",
				},
			},
		},
		EntryCount: 1,
	}
	_, err = SignerRotationActivationAuditAppend(SignerRotationActivationAuditAppendRequest{
		ApplyResult:         applyResult,
		VerificationReceipt: verification,
		ExistingLedger:      existing,
	})
	if err == nil || !strings.Contains(err.Error(), "activation audit verified_at must be strictly increasing") {
		t.Fatalf("expected out-of-order verification error, got %v", err)
	}
}

func mustSignerRotationActivationAuditLedger(t *testing.T, bundle Bundle) SignerRotationActivationAuditLedger {
	t.Helper()

	appendResult, err := SignerRotationActivationAuditAppend(SignerRotationActivationAuditAppendRequest{
		ApplyResult:         mustSignerRotationApplyResult(t, bundle),
		VerificationReceipt: mustSignerRotationVerificationReceipt(t, bundle),
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditAppend failed: %v", err)
	}
	return appendResult.Ledger
}

func TestSignerRotationActivationAuditReconcile(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	applied := mustSignerRotationApplyResult(t, bundle)
	ledger := mustSignerRotationActivationAuditLedger(t, bundle)
	report, err := SignerRotationActivationAuditReconcile(SignerRotationActivationAuditReconcileRequest{
		Ledger:     ledger,
		Policy:     applied.AppliedPolicy,
		PolicyPath: "config/governance/checkpoint-signers.json",
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditReconcile failed: %v", err)
	}
	if report.Status != "consistent" {
		t.Fatalf("unexpected reconcile status: %s", report.Status)
	}
	if !report.CurrentPolicyExplained {
		t.Fatalf("expected current policy to be explained")
	}
	if len(report.Issues) != 0 {
		t.Fatalf("expected no reconciliation issues, got %+v", report.Issues)
	}
}

func TestSignerRotationActivationAuditReconcileFlagsEmptyLedgerGap(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	applied := mustSignerRotationApplyResult(t, bundle)
	report, err := SignerRotationActivationAuditReconcile(SignerRotationActivationAuditReconcileRequest{
		Ledger:     SignerRotationActivationAuditLedger{},
		Policy:     applied.AppliedPolicy,
		PolicyPath: "config/governance/checkpoint-signers.json",
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditReconcile failed: %v", err)
	}
	if report.Status != "gap" {
		t.Fatalf("unexpected reconcile status: %s", report.Status)
	}
	if report.CurrentPolicyExplained {
		t.Fatalf("expected unexplained current policy for empty ledger")
	}
	if len(report.Issues) == 0 || !strings.Contains(report.Issues[0], "activation audit ledger is empty") {
		t.Fatalf("expected empty-ledger issue, got %+v", report.Issues)
	}
}

func TestSignerRotationActivationAuditReconcileFlagsCurrentPolicyMismatch(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	ledger := mustSignerRotationActivationAuditLedger(t, bundle)
	currentPolicy, err := currentSignerPolicy(bundle)
	if err != nil {
		t.Fatalf("currentSignerPolicy failed: %v", err)
	}
	report, err := SignerRotationActivationAuditReconcile(SignerRotationActivationAuditReconcileRequest{
		Ledger:     ledger,
		Policy:     currentPolicy,
		PolicyPath: "config/governance/checkpoint-signers.json",
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditReconcile failed: %v", err)
	}
	if report.Status != "gap" {
		t.Fatalf("unexpected reconcile status: %s", report.Status)
	}
	if report.CurrentPolicyExplained {
		t.Fatalf("expected current policy mismatch to remain unexplained")
	}
	if len(report.Issues) == 0 {
		t.Fatalf("expected reconciliation issues for current policy mismatch")
	}
}

func mustSignerRotationActivationAuditReconcileReport(
	t *testing.T,
	ledger SignerRotationActivationAuditLedger,
	policy CheckpointSignerPolicyOutput,
) SignerRotationActivationAuditReconcileReport {
	t.Helper()

	report, err := SignerRotationActivationAuditReconcile(SignerRotationActivationAuditReconcileRequest{
		Ledger:     ledger,
		Policy:     policy,
		PolicyPath: "config/governance/checkpoint-signers.json",
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditReconcile failed: %v", err)
	}
	return report
}

func TestSignerRotationActivationAuditExport(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	applied := mustSignerRotationApplyResult(t, bundle)
	ledger := mustSignerRotationActivationAuditLedger(t, bundle)
	report := mustSignerRotationActivationAuditReconcileReport(t, ledger, applied.AppliedPolicy)
	exportPackage, err := SignerRotationActivationAuditExport(SignerRotationActivationAuditExportRequest{
		Ledger:         ledger,
		Policy:         applied.AppliedPolicy,
		Reconciliation: report,
		PolicyPath:     "config/governance/checkpoint-signers.json",
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditExport failed: %v", err)
	}
	if exportPackage.Status != "consistent" {
		t.Fatalf("unexpected export status: %s", exportPackage.Status)
	}
	if exportPackage.BaselineSnapshot.CurrentPolicyVersion != applied.AppliedPolicy.Version {
		t.Fatalf("unexpected baseline current policy version: %s", exportPackage.BaselineSnapshot.CurrentPolicyVersion)
	}
	if exportPackage.BaselineSnapshot.LedgerEntryCount != 1 {
		t.Fatalf("unexpected baseline ledger entry count: %d", exportPackage.BaselineSnapshot.LedgerEntryCount)
	}
	if exportPackage.BaselineSnapshot.LatestEntry == nil {
		t.Fatal("expected latest baseline entry")
	}
	if exportPackage.BaselineSnapshot.LatestEntry.ReceiptID != report.LatestReceiptID {
		t.Fatalf("unexpected latest baseline receipt id: %s", exportPackage.BaselineSnapshot.LatestEntry.ReceiptID)
	}
	if len(exportPackage.BaselineSnapshot.ContinuityIssues) != 0 {
		t.Fatalf("expected no continuity issues, got %+v", exportPackage.BaselineSnapshot.ContinuityIssues)
	}
}

func TestSignerRotationActivationAuditExportFailsOnReconciliationDigestMismatch(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	applied := mustSignerRotationApplyResult(t, bundle)
	ledger := mustSignerRotationActivationAuditLedger(t, bundle)
	report := mustSignerRotationActivationAuditReconcileReport(t, ledger, applied.AppliedPolicy)
	report.CurrentPolicyDigest = "deadbeef"
	_, err = SignerRotationActivationAuditExport(SignerRotationActivationAuditExportRequest{
		Ledger:         ledger,
		Policy:         applied.AppliedPolicy,
		Reconciliation: report,
		PolicyPath:     "config/governance/checkpoint-signers.json",
	})
	if err == nil || !strings.Contains(err.Error(), "activation audit reconciliation current_policy_digest mismatch") {
		t.Fatalf("expected reconciliation digest mismatch error, got %v", err)
	}
}

func TestSignerRotationActivationAuditExportFailsOnLatestReceiptMismatch(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	applied := mustSignerRotationApplyResult(t, bundle)
	ledger := mustSignerRotationActivationAuditLedger(t, bundle)
	report := mustSignerRotationActivationAuditReconcileReport(t, ledger, applied.AppliedPolicy)
	report.LatestReceiptID = "rotation-other-20260424t000000z"
	_, err = SignerRotationActivationAuditExport(SignerRotationActivationAuditExportRequest{
		Ledger:         ledger,
		Policy:         applied.AppliedPolicy,
		Reconciliation: report,
		PolicyPath:     "config/governance/checkpoint-signers.json",
	})
	if err == nil || !strings.Contains(err.Error(), "activation audit reconciliation latest_receipt_id mismatch") {
		t.Fatalf("expected latest receipt mismatch error, got %v", err)
	}
}

func mustCurrentSignerRotationActivationAuditExportPackage(
	t *testing.T,
	bundle Bundle,
) SignerRotationActivationAuditExportPackage {
	t.Helper()

	currentPolicy, err := currentSignerPolicy(bundle)
	if err != nil {
		t.Fatalf("currentSignerPolicy failed: %v", err)
	}
	ledger := SignerRotationActivationAuditLedger{
		Version:    "1.0.0",
		Status:     "active",
		ChainID:    "0ai-assurance-1",
		PolicyPath: "config/governance/checkpoint-signers.json",
		Entries:    []SignerRotationActivationAuditEntry{},
		EntryCount: 0,
	}
	report := mustSignerRotationActivationAuditReconcileReport(t, ledger, currentPolicy)
	exportPackage, err := SignerRotationActivationAuditExport(SignerRotationActivationAuditExportRequest{
		Ledger:         ledger,
		Policy:         currentPolicy,
		Reconciliation: report,
		PolicyPath:     "config/governance/checkpoint-signers.json",
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditExport failed: %v", err)
	}
	return exportPackage
}

func mustRotatedSignerRotationActivationAuditExportPackage(
	t *testing.T,
	bundle Bundle,
) SignerRotationActivationAuditExportPackage {
	t.Helper()

	applied := mustSignerRotationApplyResult(t, bundle)
	ledger := mustSignerRotationActivationAuditLedger(t, bundle)
	report := mustSignerRotationActivationAuditReconcileReport(t, ledger, applied.AppliedPolicy)
	exportPackage, err := SignerRotationActivationAuditExport(SignerRotationActivationAuditExportRequest{
		Ledger:         ledger,
		Policy:         applied.AppliedPolicy,
		Reconciliation: report,
		PolicyPath:     "config/governance/checkpoint-signers.json",
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditExport failed: %v", err)
	}
	return exportPackage
}

func mustSignerRotationActivationAuditRetainedInventoryPackage(
	t *testing.T,
	packagePath string,
	promotionPath string,
	exportPackage SignerRotationActivationAuditExportPackage,
	archiveIndex SignerRotationActivationAuditArchiveIndex,
	promotedAt string,
	verifiedAt string,
) SignerRotationActivationAuditRetainedInventoryPackage {
	t.Helper()

	verification, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
		ExportPackage: exportPackage,
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditVerifyExport failed: %v", err)
	}
	promotion, err := BuildSignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionRequest{
		PackagePath:        packagePath,
		ExportPackage:      exportPackage,
		VerificationReport: verification,
		ArchiveIndex:       archiveIndex,
		PromotedAt:         promotedAt,
		PromotedBy:         "governance-archive-bot",
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchivePromotion failed: %v", err)
	}
	receipt, err := VerifySignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionVerificationRequest{
		PackagePath:        packagePath,
		ExportPackage:      exportPackage,
		VerificationReport: verification,
		ArchiveIndex:       archiveIndex,
		PromotionResult:    promotion,
		VerifiedAt:         verifiedAt,
		VerifiedBy:         "governance-audit-bot",
	})
	if err != nil {
		t.Fatalf("VerifySignerRotationActivationAuditArchivePromotion failed: %v", err)
	}
	return SignerRotationActivationAuditRetainedInventoryPackage{
		PromotionPath:       promotionPath,
		PromotionResult:     promotion,
		VerificationReceipt: receipt,
	}
}

func mustSignerRotationActivationAuditRetainedInventorySnapshot(
	t *testing.T,
	packages ...SignerRotationActivationAuditRetainedInventoryPackage,
) SignerRotationActivationAuditRetainedInventorySnapshot {
	t.Helper()

	snapshot, err := BuildSignerRotationActivationAuditRetainedInventorySnapshot(SignerRotationActivationAuditRetainedInventorySnapshotRequest{
		Packages: packages,
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditRetainedInventorySnapshot failed: %v", err)
	}
	return snapshot
}

func TestSignerRotationActivationAuditVerifyExport(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	exportPackage := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	report, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
		ExportPackage: exportPackage,
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditVerifyExport failed: %v", err)
	}
	if report.Status != "consistent" {
		t.Fatalf("unexpected verification status: %s", report.Status)
	}
	if !report.ArchiveReady {
		t.Fatal("expected export package to be archive ready")
	}
	if len(report.VerificationIssues) != 0 {
		t.Fatalf("expected no verification issues, got %+v", report.VerificationIssues)
	}
}

func TestSignerRotationActivationAuditVerifyExportFlagsTamper(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	exportPackage := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	exportPackage.BaselineSnapshot.CurrentPolicyDigest = "deadbeef"
	report, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
		ExportPackage: exportPackage,
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditVerifyExport failed: %v", err)
	}
	if report.Status != "invalid" {
		t.Fatalf("expected invalid verification status, got %s", report.Status)
	}
	if report.ArchiveReady {
		t.Fatal("expected tampered export package to be non-archive-ready")
	}
	if len(report.VerificationIssues) == 0 || !strings.Contains(report.VerificationIssues[0], "activation audit export package drift detected") {
		t.Fatalf("expected export drift issue, got %+v", report.VerificationIssues)
	}
}

func TestSignerRotationActivationAuditArchiveIndex(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	if index.Status != "consistent" {
		t.Fatalf("unexpected archive index status: %s", index.Status)
	}
	if index.PackageCount != 2 || index.ArchiveReadyCount != 2 {
		t.Fatalf("unexpected archive counts: %+v", index)
	}
	if index.LatestCurrentPolicyVersion != rotatedExport.CurrentPolicy.Version {
		t.Fatalf("unexpected latest archive policy version: %s", index.LatestCurrentPolicyVersion)
	}
}

func TestSignerRotationActivationAuditArchiveIndexFlagsDuplicateBaseline(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/export-a.json", ExportPackage: rotatedExport},
			{PackagePath: "build/rotation/export-b.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	if index.Status != "invalid" {
		t.Fatalf("expected invalid archive index status, got %s", index.Status)
	}
	if len(index.Issues) == 0 || !strings.Contains(index.Issues[0], "duplicate archive current_policy_version") {
		t.Fatalf("expected duplicate baseline issue, got %+v", index.Issues)
	}
}

func TestSignerRotationActivationAuditArchiveIndexFlagsDuplicatePackagePath(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/shared-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/shared-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	if index.Status != "invalid" {
		t.Fatalf("expected invalid archive index status, got %s", index.Status)
	}
	if len(index.Issues) == 0 || !strings.Contains(strings.Join(index.Issues, " | "), "duplicate archive package_path") {
		t.Fatalf("expected duplicate package path issue, got %+v", index.Issues)
	}
}

func TestSignerRotationActivationAuditArchivePromotion(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	verification, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
		ExportPackage: rotatedExport,
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditVerifyExport failed: %v", err)
	}
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}

	promotion, err := BuildSignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionRequest{
		PackagePath:        "build/rotation/governance-chair-audit-export.json",
		ExportPackage:      rotatedExport,
		VerificationReport: verification,
		ArchiveIndex:       index,
		PromotedAt:         "2026-04-24T00:20:00Z",
		PromotedBy:         "governance-archive-bot",
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchivePromotion failed: %v", err)
	}
	if promotion.Status != "promoted" {
		t.Fatalf("unexpected promotion status: %s", promotion.Status)
	}
	if promotion.PromotionReceipt.Status != "promoted" {
		t.Fatalf("unexpected promotion receipt status: %s", promotion.PromotionReceipt.Status)
	}
	if promotion.RetainedBaselineAttestation.Status != "retained" {
		t.Fatalf("unexpected retained baseline attestation status: %s", promotion.RetainedBaselineAttestation.Status)
	}
	if promotion.PromotionReceipt.PackagePath != "build/rotation/governance-chair-audit-export.json" {
		t.Fatalf("unexpected promotion receipt package path: %s", promotion.PromotionReceipt.PackagePath)
	}
	if promotion.RetainedBaselineAttestation.ArchiveEntryIndex != 1 {
		t.Fatalf("unexpected archive entry index: %d", promotion.RetainedBaselineAttestation.ArchiveEntryIndex)
	}
	if promotion.PromotionReceipt.ExportPackageDigest == "" || promotion.PromotionReceipt.ArchiveIndexDigest == "" {
		t.Fatalf("expected populated promotion digests, got %+v", promotion.PromotionReceipt)
	}
	if promotion.RetainedBaselineAttestation.PromotionReceiptID != promotion.PromotionReceipt.ReceiptID {
		t.Fatalf("attestation receipt binding mismatch: %+v", promotion)
	}
}

func TestSignerRotationActivationAuditArchivePromotionFailsOnLineageMismatch(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	verification, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
		ExportPackage: currentExport,
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditVerifyExport failed: %v", err)
	}
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}

	promotion, err := BuildSignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionRequest{
		PackagePath:        "build/rotation/governance-chair-audit-export.json",
		ExportPackage:      rotatedExport,
		VerificationReport: verification,
		ArchiveIndex:       index,
		PromotedAt:         "2026-04-24T00:20:00Z",
		PromotedBy:         "governance-archive-bot",
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchivePromotion failed: %v", err)
	}
	if promotion.Status != "invalid" {
		t.Fatalf("expected invalid promotion status, got %s", promotion.Status)
	}
	if len(promotion.Issues) == 0 || !strings.Contains(strings.Join(promotion.Issues, " | "), "verification report drift detected") {
		t.Fatalf("expected verification drift issue, got %+v", promotion.Issues)
	}
}

func TestSignerRotationActivationAuditArchivePromotionFlagsNonReadyArchiveIndexEntry(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	verification, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
		ExportPackage: rotatedExport,
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditVerifyExport failed: %v", err)
	}
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	index.Status = "consistent"
	index.Entries[0].ArchiveReady = false
	index.ArchiveReadyCount = 1

	promotion, err := BuildSignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionRequest{
		PackagePath:        "build/rotation/governance-chair-audit-export.json",
		ExportPackage:      rotatedExport,
		VerificationReport: verification,
		ArchiveIndex:       index,
		PromotedAt:         "2026-04-24T00:20:00Z",
		PromotedBy:         "governance-archive-bot",
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchivePromotion failed: %v", err)
	}
	if promotion.Status != "invalid" {
		t.Fatalf("expected invalid promotion status, got %s", promotion.Status)
	}
	if len(promotion.Issues) == 0 || !strings.Contains(strings.Join(promotion.Issues, " | "), "non-archive-ready entry") {
		t.Fatalf("expected non-archive-ready entry issue, got %+v", promotion.Issues)
	}
}

func TestSignerRotationActivationAuditArchivePromotionFlagsNonConsistentArchiveIndexEntry(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	verification, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
		ExportPackage: rotatedExport,
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditVerifyExport failed: %v", err)
	}
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	index.Status = "consistent"
	index.Entries[0].Status = "review"

	promotion, err := BuildSignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionRequest{
		PackagePath:        "build/rotation/governance-chair-audit-export.json",
		ExportPackage:      rotatedExport,
		VerificationReport: verification,
		ArchiveIndex:       index,
		PromotedAt:         "2026-04-24T00:20:00Z",
		PromotedBy:         "governance-archive-bot",
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchivePromotion failed: %v", err)
	}
	if promotion.Status != "invalid" {
		t.Fatalf("expected invalid promotion status, got %s", promotion.Status)
	}
	if len(promotion.Issues) == 0 || !strings.Contains(strings.Join(promotion.Issues, " | "), "non-consistent entry") {
		t.Fatalf("expected non-consistent entry issue, got %+v", promotion.Issues)
	}
}

func TestSignerRotationActivationAuditArchivePromotionFlagsDuplicateArchivePackagePath(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	verification, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
		ExportPackage: rotatedExport,
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditVerifyExport failed: %v", err)
	}
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	index.Status = "consistent"
	index.Entries = append(index.Entries, index.Entries[len(index.Entries)-1])
	index.PackageCount = len(index.Entries)
	index.ArchiveReadyCount = len(index.Entries)

	promotion, err := BuildSignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionRequest{
		PackagePath:        "build/rotation/governance-chair-audit-export.json",
		ExportPackage:      rotatedExport,
		VerificationReport: verification,
		ArchiveIndex:       index,
		PromotedAt:         "2026-04-24T00:20:00Z",
		PromotedBy:         "governance-archive-bot",
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchivePromotion failed: %v", err)
	}
	if promotion.Status != "invalid" {
		t.Fatalf("expected invalid promotion status, got %s", promotion.Status)
	}
	if len(promotion.Issues) == 0 || !strings.Contains(strings.Join(promotion.Issues, " | "), "duplicate package_path") {
		t.Fatalf("expected duplicate package_path issue, got %+v", promotion.Issues)
	}
}

func TestVerifySignerRotationActivationAuditArchivePromotion(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	verification, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
		ExportPackage: rotatedExport,
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditVerifyExport failed: %v", err)
	}
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	promotion, err := BuildSignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionRequest{
		PackagePath:        "build/rotation/governance-chair-audit-export.json",
		ExportPackage:      rotatedExport,
		VerificationReport: verification,
		ArchiveIndex:       index,
		PromotedAt:         "2026-04-24T00:20:00Z",
		PromotedBy:         "governance-archive-bot",
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchivePromotion failed: %v", err)
	}

	receipt, err := VerifySignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionVerificationRequest{
		PackagePath:        "build/rotation/governance-chair-audit-export.json",
		ExportPackage:      rotatedExport,
		VerificationReport: verification,
		ArchiveIndex:       index,
		PromotionResult:    promotion,
		VerifiedAt:         "2026-04-24T00:25:00Z",
		VerifiedBy:         "governance-audit-bot",
	})
	if err != nil {
		t.Fatalf("VerifySignerRotationActivationAuditArchivePromotion failed: %v", err)
	}
	if receipt.Status != "verified" {
		t.Fatalf("unexpected promotion verification status: %s", receipt.Status)
	}
	if receipt.VerificationReceiptID == "" || receipt.PromotionResultDigest == "" {
		t.Fatalf("expected populated verification receipt digests, got %+v", receipt)
	}
	if receipt.PromotionReceiptID != promotion.PromotionReceipt.ReceiptID {
		t.Fatalf("unexpected promotion receipt binding: %+v", receipt)
	}
}

func TestVerifySignerRotationActivationAuditArchivePromotionFlagsDrift(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	verification, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
		ExportPackage: rotatedExport,
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditVerifyExport failed: %v", err)
	}
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	promotion, err := BuildSignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionRequest{
		PackagePath:        "build/rotation/governance-chair-audit-export.json",
		ExportPackage:      rotatedExport,
		VerificationReport: verification,
		ArchiveIndex:       index,
		PromotedAt:         "2026-04-24T00:20:00Z",
		PromotedBy:         "governance-archive-bot",
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchivePromotion failed: %v", err)
	}
	promotion.PromotionReceipt.CurrentPolicyDigest = "deadbeef"

	receipt, err := VerifySignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionVerificationRequest{
		PackagePath:        "build/rotation/governance-chair-audit-export.json",
		ExportPackage:      rotatedExport,
		VerificationReport: verification,
		ArchiveIndex:       index,
		PromotionResult:    promotion,
		VerifiedAt:         "2026-04-24T00:25:00Z",
		VerifiedBy:         "governance-audit-bot",
	})
	if err != nil {
		t.Fatalf("VerifySignerRotationActivationAuditArchivePromotion failed: %v", err)
	}
	if receipt.Status != "invalid" {
		t.Fatalf("expected invalid promotion verification status, got %s", receipt.Status)
	}
	if len(receipt.VerificationIssues) == 0 || !strings.Contains(strings.Join(receipt.VerificationIssues, " | "), "promotion drift detected") {
		t.Fatalf("expected promotion drift issue, got %+v", receipt.VerificationIssues)
	}
}

func TestSignerRotationActivationAuditRetainedInventorySnapshot(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	verification, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
		ExportPackage: rotatedExport,
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditVerifyExport failed: %v", err)
	}
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	promotion, err := BuildSignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionRequest{
		PackagePath:        "build/rotation/governance-chair-audit-export.json",
		ExportPackage:      rotatedExport,
		VerificationReport: verification,
		ArchiveIndex:       index,
		PromotedAt:         "2026-04-24T00:20:00Z",
		PromotedBy:         "governance-archive-bot",
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchivePromotion failed: %v", err)
	}
	receipt, err := VerifySignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionVerificationRequest{
		PackagePath:        "build/rotation/governance-chair-audit-export.json",
		ExportPackage:      rotatedExport,
		VerificationReport: verification,
		ArchiveIndex:       index,
		PromotionResult:    promotion,
		VerifiedAt:         "2026-04-24T00:25:00Z",
		VerifiedBy:         "governance-audit-bot",
	})
	if err != nil {
		t.Fatalf("VerifySignerRotationActivationAuditArchivePromotion failed: %v", err)
	}

	snapshot, err := BuildSignerRotationActivationAuditRetainedInventorySnapshot(SignerRotationActivationAuditRetainedInventorySnapshotRequest{
		Packages: []SignerRotationActivationAuditRetainedInventoryPackage{
			{
				PromotionPath:       "build/rotation/governance-chair-archive-promotion.json",
				PromotionResult:     promotion,
				VerificationReceipt: receipt,
			},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditRetainedInventorySnapshot failed: %v", err)
	}
	if snapshot.Status != "consistent" {
		t.Fatalf("unexpected retained inventory status: %s", snapshot.Status)
	}
	if snapshot.PackageCount != 1 || snapshot.VerifiedCount != 1 {
		t.Fatalf("unexpected retained inventory counts: %+v", snapshot)
	}
	if snapshot.LatestCurrentPolicyVersion != promotion.PromotionReceipt.CurrentPolicyVersion {
		t.Fatalf("unexpected retained inventory latest policy version: %s", snapshot.LatestCurrentPolicyVersion)
	}
}

func TestSignerRotationActivationAuditRetainedInventorySnapshotFlagsReceiptMismatch(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	verification, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
		ExportPackage: rotatedExport,
	})
	if err != nil {
		t.Fatalf("SignerRotationActivationAuditVerifyExport failed: %v", err)
	}
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	promotion, err := BuildSignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionRequest{
		PackagePath:        "build/rotation/governance-chair-audit-export.json",
		ExportPackage:      rotatedExport,
		VerificationReport: verification,
		ArchiveIndex:       index,
		PromotedAt:         "2026-04-24T00:20:00Z",
		PromotedBy:         "governance-archive-bot",
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchivePromotion failed: %v", err)
	}
	receipt, err := VerifySignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionVerificationRequest{
		PackagePath:        "build/rotation/governance-chair-audit-export.json",
		ExportPackage:      rotatedExport,
		VerificationReport: verification,
		ArchiveIndex:       index,
		PromotionResult:    promotion,
		VerifiedAt:         "2026-04-24T00:25:00Z",
		VerifiedBy:         "governance-audit-bot",
	})
	if err != nil {
		t.Fatalf("VerifySignerRotationActivationAuditArchivePromotion failed: %v", err)
	}
	receipt.CurrentPolicyDigest = "deadbeef"

	snapshot, err := BuildSignerRotationActivationAuditRetainedInventorySnapshot(SignerRotationActivationAuditRetainedInventorySnapshotRequest{
		Packages: []SignerRotationActivationAuditRetainedInventoryPackage{
			{
				PromotionPath:       "build/rotation/governance-chair-archive-promotion.json",
				PromotionResult:     promotion,
				VerificationReceipt: receipt,
			},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditRetainedInventorySnapshot failed: %v", err)
	}
	if snapshot.Status != "invalid" {
		t.Fatalf("expected invalid retained inventory status, got %s", snapshot.Status)
	}
	if len(snapshot.Issues) == 0 || !strings.Contains(strings.Join(snapshot.Issues, " | "), "current_policy_digest mismatch") {
		t.Fatalf("expected retained inventory digest mismatch issue, got %+v", snapshot.Issues)
	}
}

func TestVerifySignerRotationActivationAuditRetainedInventorySnapshot(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	rotatedPackage := mustSignerRotationActivationAuditRetainedInventoryPackage(
		t,
		"build/rotation/governance-chair-audit-export.json",
		"build/rotation/governance-chair-archive-promotion.json",
		rotatedExport,
		index,
		"2026-04-24T00:20:00Z",
		"2026-04-24T00:25:00Z",
	)
	snapshot := mustSignerRotationActivationAuditRetainedInventorySnapshot(t, rotatedPackage)

	receipt, err := VerifySignerRotationActivationAuditRetainedInventorySnapshot(SignerRotationActivationAuditRetainedInventoryVerificationRequest{
		Snapshot:   snapshot,
		Packages:   []SignerRotationActivationAuditRetainedInventoryPackage{rotatedPackage},
		VerifiedAt: "2026-04-24T00:30:00Z",
		VerifiedBy: "governance-audit-bot",
	})
	if err != nil {
		t.Fatalf("VerifySignerRotationActivationAuditRetainedInventorySnapshot failed: %v", err)
	}
	if receipt.Status != "verified" {
		t.Fatalf("unexpected retained inventory verification status: %s", receipt.Status)
	}
	if receipt.VerificationReceiptID == "" || receipt.InventorySnapshotDigest == "" {
		t.Fatalf("expected populated retained inventory verification receipt, got %+v", receipt)
	}
	if receipt.PackageCount != snapshot.PackageCount || receipt.LatestCurrentPolicyVersion != snapshot.LatestCurrentPolicyVersion {
		t.Fatalf("retained inventory receipt did not mirror snapshot summary: %+v", receipt)
	}
}

func TestVerifySignerRotationActivationAuditRetainedInventorySnapshotFlagsDrift(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	rotatedPackage := mustSignerRotationActivationAuditRetainedInventoryPackage(
		t,
		"build/rotation/governance-chair-audit-export.json",
		"build/rotation/governance-chair-archive-promotion.json",
		rotatedExport,
		index,
		"2026-04-24T00:20:00Z",
		"2026-04-24T00:25:00Z",
	)
	snapshot := mustSignerRotationActivationAuditRetainedInventorySnapshot(t, rotatedPackage)
	snapshot.LatestCurrentPolicyDigest = "deadbeef"

	receipt, err := VerifySignerRotationActivationAuditRetainedInventorySnapshot(SignerRotationActivationAuditRetainedInventoryVerificationRequest{
		Snapshot:   snapshot,
		Packages:   []SignerRotationActivationAuditRetainedInventoryPackage{rotatedPackage},
		VerifiedAt: "2026-04-24T00:30:00Z",
		VerifiedBy: "governance-audit-bot",
	})
	if err != nil {
		t.Fatalf("VerifySignerRotationActivationAuditRetainedInventorySnapshot failed: %v", err)
	}
	if receipt.Status != "invalid" {
		t.Fatalf("expected invalid retained inventory verification status, got %s", receipt.Status)
	}
	if len(receipt.VerificationIssues) == 0 || !strings.Contains(strings.Join(receipt.VerificationIssues, " | "), "snapshot drift detected") {
		t.Fatalf("expected retained inventory drift issue, got %+v", receipt.VerificationIssues)
	}
}

func TestSignerRotationActivationAuditRetainedInventoryContinuityManifest(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	currentPackage := mustSignerRotationActivationAuditRetainedInventoryPackage(
		t,
		"build/rotation/current-audit-export.json",
		"build/rotation/current-archive-promotion.json",
		currentExport,
		index,
		"2026-04-17T00:20:00Z",
		"2026-04-17T00:25:00Z",
	)
	rotatedPackage := mustSignerRotationActivationAuditRetainedInventoryPackage(
		t,
		"build/rotation/governance-chair-audit-export.json",
		"build/rotation/governance-chair-archive-promotion.json",
		rotatedExport,
		index,
		"2026-04-24T00:20:00Z",
		"2026-04-24T00:25:00Z",
	)
	currentSnapshot := mustSignerRotationActivationAuditRetainedInventorySnapshot(t, currentPackage)
	rotatedSnapshot := mustSignerRotationActivationAuditRetainedInventorySnapshot(t, currentPackage, rotatedPackage)
	currentReceipt, err := VerifySignerRotationActivationAuditRetainedInventorySnapshot(SignerRotationActivationAuditRetainedInventoryVerificationRequest{
		Snapshot:   currentSnapshot,
		Packages:   []SignerRotationActivationAuditRetainedInventoryPackage{currentPackage},
		VerifiedAt: "2026-04-17T00:30:00Z",
		VerifiedBy: "governance-audit-bot",
	})
	if err != nil {
		t.Fatalf("VerifySignerRotationActivationAuditRetainedInventorySnapshot current failed: %v", err)
	}
	rotatedReceipt, err := VerifySignerRotationActivationAuditRetainedInventorySnapshot(SignerRotationActivationAuditRetainedInventoryVerificationRequest{
		Snapshot:   rotatedSnapshot,
		Packages:   []SignerRotationActivationAuditRetainedInventoryPackage{currentPackage, rotatedPackage},
		VerifiedAt: "2026-04-24T00:30:00Z",
		VerifiedBy: "governance-audit-bot",
	})
	if err != nil {
		t.Fatalf("VerifySignerRotationActivationAuditRetainedInventorySnapshot rotated failed: %v", err)
	}

	manifest, err := BuildSignerRotationActivationAuditRetainedInventoryContinuityManifest(SignerRotationActivationAuditRetainedInventoryContinuityManifestRequest{
		Snapshots: []SignerRotationActivationAuditRetainedInventoryContinuityPackage{
			{SnapshotPath: "build/rotation/current-retained-inventory.json", Snapshot: currentSnapshot, VerificationReceipt: currentReceipt},
			{SnapshotPath: "build/rotation/retained-archive-inventory.json", Snapshot: rotatedSnapshot, VerificationReceipt: rotatedReceipt},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditRetainedInventoryContinuityManifest failed: %v", err)
	}
	if manifest.Status != "continuous" {
		t.Fatalf("unexpected retained inventory continuity status: %s (%+v)", manifest.Status, manifest.Issues)
	}
	if manifest.SnapshotCount != 2 || manifest.VerifiedSnapshotCount != 2 {
		t.Fatalf("unexpected retained inventory continuity counts: %+v", manifest)
	}
	if manifest.LatestCurrentPolicyVersion != rotatedSnapshot.LatestCurrentPolicyVersion {
		t.Fatalf("unexpected retained inventory continuity latest policy version: %s", manifest.LatestCurrentPolicyVersion)
	}
}

func TestSignerRotationActivationAuditRetainedInventoryContinuityManifestFlagsVerificationReceiptDrift(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	rotatedPackage := mustSignerRotationActivationAuditRetainedInventoryPackage(
		t,
		"build/rotation/governance-chair-audit-export.json",
		"build/rotation/governance-chair-archive-promotion.json",
		rotatedExport,
		index,
		"2026-04-24T00:20:00Z",
		"2026-04-24T00:25:00Z",
	)
	snapshot := mustSignerRotationActivationAuditRetainedInventorySnapshot(t, rotatedPackage)
	receipt, err := VerifySignerRotationActivationAuditRetainedInventorySnapshot(SignerRotationActivationAuditRetainedInventoryVerificationRequest{
		Snapshot:   snapshot,
		Packages:   []SignerRotationActivationAuditRetainedInventoryPackage{rotatedPackage},
		VerifiedAt: "2026-04-24T00:30:00Z",
		VerifiedBy: "governance-audit-bot",
	})
	if err != nil {
		t.Fatalf("VerifySignerRotationActivationAuditRetainedInventorySnapshot failed: %v", err)
	}
	receipt.VerificationReceiptID = "retained-inventory-verification-deadbeef"

	manifest, err := BuildSignerRotationActivationAuditRetainedInventoryContinuityManifest(SignerRotationActivationAuditRetainedInventoryContinuityManifestRequest{
		Snapshots: []SignerRotationActivationAuditRetainedInventoryContinuityPackage{
			{SnapshotPath: "build/rotation/retained-archive-inventory.json", Snapshot: snapshot, VerificationReceipt: receipt},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditRetainedInventoryContinuityManifest failed: %v", err)
	}
	if manifest.Status != "invalid" {
		t.Fatalf("expected invalid retained inventory continuity status, got %s", manifest.Status)
	}
	if len(manifest.Issues) == 0 || !strings.Contains(strings.Join(manifest.Issues, " | "), "verification_receipt_id mismatch") {
		t.Fatalf("expected verification receipt drift issue, got %+v", manifest.Issues)
	}
}

func TestSignerRotationActivationAuditRetainedInventoryContinuityManifestFlagsDroppedEntry(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	currentExport := mustCurrentSignerRotationActivationAuditExportPackage(t, bundle)
	rotatedExport := mustRotatedSignerRotationActivationAuditExportPackage(t, bundle)
	index, err := BuildSignerRotationActivationAuditArchiveIndex(SignerRotationActivationAuditArchiveIndexRequest{
		Packages: []SignerRotationActivationAuditArchivePackage{
			{PackagePath: "build/rotation/current-audit-export.json", ExportPackage: currentExport},
			{PackagePath: "build/rotation/governance-chair-audit-export.json", ExportPackage: rotatedExport},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditArchiveIndex failed: %v", err)
	}
	currentPackage := mustSignerRotationActivationAuditRetainedInventoryPackage(
		t,
		"build/rotation/current-audit-export.json",
		"build/rotation/current-archive-promotion.json",
		currentExport,
		index,
		"2026-04-17T00:20:00Z",
		"2026-04-17T00:25:00Z",
	)
	rotatedPackage := mustSignerRotationActivationAuditRetainedInventoryPackage(
		t,
		"build/rotation/governance-chair-audit-export.json",
		"build/rotation/governance-chair-archive-promotion.json",
		rotatedExport,
		index,
		"2026-04-24T00:20:00Z",
		"2026-04-24T00:25:00Z",
	)
	currentSnapshot := mustSignerRotationActivationAuditRetainedInventorySnapshot(t, currentPackage)
	rotatedSnapshot := mustSignerRotationActivationAuditRetainedInventorySnapshot(t, rotatedPackage)
	currentReceipt, err := VerifySignerRotationActivationAuditRetainedInventorySnapshot(SignerRotationActivationAuditRetainedInventoryVerificationRequest{
		Snapshot:   currentSnapshot,
		Packages:   []SignerRotationActivationAuditRetainedInventoryPackage{currentPackage},
		VerifiedAt: "2026-04-17T00:30:00Z",
		VerifiedBy: "governance-audit-bot",
	})
	if err != nil {
		t.Fatalf("VerifySignerRotationActivationAuditRetainedInventorySnapshot current failed: %v", err)
	}
	rotatedReceipt, err := VerifySignerRotationActivationAuditRetainedInventorySnapshot(SignerRotationActivationAuditRetainedInventoryVerificationRequest{
		Snapshot:   rotatedSnapshot,
		Packages:   []SignerRotationActivationAuditRetainedInventoryPackage{rotatedPackage},
		VerifiedAt: "2026-04-24T00:30:00Z",
		VerifiedBy: "governance-audit-bot",
	})
	if err != nil {
		t.Fatalf("VerifySignerRotationActivationAuditRetainedInventorySnapshot rotated failed: %v", err)
	}

	manifest, err := BuildSignerRotationActivationAuditRetainedInventoryContinuityManifest(SignerRotationActivationAuditRetainedInventoryContinuityManifestRequest{
		Snapshots: []SignerRotationActivationAuditRetainedInventoryContinuityPackage{
			{SnapshotPath: "build/rotation/current-retained-inventory.json", Snapshot: currentSnapshot, VerificationReceipt: currentReceipt},
			{SnapshotPath: "build/rotation/retained-archive-inventory.json", Snapshot: rotatedSnapshot, VerificationReceipt: rotatedReceipt},
		},
	})
	if err != nil {
		t.Fatalf("BuildSignerRotationActivationAuditRetainedInventoryContinuityManifest failed: %v", err)
	}
	if manifest.Status != "invalid" {
		t.Fatalf("expected invalid retained inventory continuity status, got %s", manifest.Status)
	}
	if len(manifest.Issues) == 0 || !strings.Contains(strings.Join(manifest.Issues, " | "), "dropped promotion_receipt_id") {
		t.Fatalf("expected dropped entry continuity issue, got %+v", manifest.Issues)
	}
}
