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
