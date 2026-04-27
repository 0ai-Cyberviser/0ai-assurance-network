package project

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type SignerRoleCoverage struct {
	Role         string   `json:"role"`
	ReferencedBy []string `json:"referenced_by"`
}

type SignerManifestEntry struct {
	ActorID                string               `json:"actor_id"`
	ActorDisplayName       string               `json:"actor_display_name"`
	ActorStatus            string               `json:"actor_status"`
	OrganizationID         string               `json:"organization_id,omitempty"`
	OrganizationName       string               `json:"organization_name,omitempty"`
	SignerID               string               `json:"signer_id"`
	KeyID                  string               `json:"key_id"`
	SignerStatus           string               `json:"signer_status"`
	Roles                  []string             `json:"roles"`
	RoleCoverage           []SignerRoleCoverage `json:"role_coverage"`
	ProvisionedAt          string               `json:"provisioned_at"`
	RotateBy               string               `json:"rotate_by"`
	RotationStatus         string               `json:"rotation_status"`
	DaysUntilRotation      int                  `json:"days_until_rotation"`
	NextRotationReceiptID  string               `json:"next_rotation_receipt_id"`
	ReplacementManifestRef string               `json:"replacement_manifest_ref"`
}

type SignerRotationPlanEntry struct {
	Order                  int      `json:"order"`
	SignerID               string   `json:"signer_id"`
	ActorID                string   `json:"actor_id"`
	Roles                  []string `json:"roles"`
	RotateBy               string   `json:"rotate_by"`
	RotationStatus         string   `json:"rotation_status"`
	DaysUntilRotation      int      `json:"days_until_rotation"`
	RecommendedAction      string   `json:"recommended_action"`
	NextRotationReceiptID  string   `json:"next_rotation_receipt_id"`
	ReplacementManifestRef string   `json:"replacement_manifest_ref"`
}

type SignerManifestOutput struct {
	Version             string                    `json:"version"`
	ChainID             string                    `json:"chain_id"`
	PolicyVersion       string                    `json:"policy_version"`
	IdentityVersion     string                    `json:"identity_version"`
	ReferenceTime       string                    `json:"reference_time"`
	WarningWindowDays   int                       `json:"warning_window_days"`
	SignerCount         int                       `json:"signer_count"`
	ActiveSignerCount   int                       `json:"active_signer_count"`
	MissingRoleCoverage []string                  `json:"missing_role_coverage"`
	ExpiringSignerIDs   []string                  `json:"expiring_signer_ids"`
	RotationPlan        []SignerRotationPlanEntry `json:"rotation_plan"`
	Signers             []SignerManifestEntry     `json:"signers"`
}

type SignerRotationApprovalOption struct {
	ActorID          string `json:"actor_id"`
	ActorDisplayName string `json:"actor_display_name"`
	OrganizationID   string `json:"organization_id,omitempty"`
	OrganizationName string `json:"organization_name,omitempty"`
	SignerID         string `json:"signer_id"`
}

type SignerRotationApprovalRequirement struct {
	Role            string                         `json:"role"`
	EligibleSigners []SignerRotationApprovalOption `json:"eligible_signers"`
}

type SignerRotationReplacement struct {
	ActorID                string   `json:"actor_id"`
	ActorDisplayName       string   `json:"actor_display_name"`
	OrganizationID         string   `json:"organization_id,omitempty"`
	OrganizationName       string   `json:"organization_name,omitempty"`
	SignerID               string   `json:"signer_id"`
	KeyID                  string   `json:"key_id"`
	Roles                  []string `json:"roles"`
	ProvisionedAt          string   `json:"provisioned_at"`
	RotateBy               string   `json:"rotate_by"`
	EffectiveAt            string   `json:"effective_at"`
	RotationReceiptID      string   `json:"rotation_receipt_id"`
	ReplacementManifestRef string   `json:"replacement_manifest_ref"`
}

type SignerRotationReceiptRequest struct {
	OutgoingSignerID      string
	IncomingSignerID      string
	IncomingKeyID         string
	IncomingActorID       string
	IncomingRoles         []string
	IncomingProvisionedAt string
	IncomingRotateBy      string
	EffectiveAt           string
	ReceiptID             string
}

type SignerRotationReceiptOutput struct {
	Version                   string                              `json:"version"`
	ChainID                   string                              `json:"chain_id"`
	PolicyVersion             string                              `json:"policy_version"`
	IdentityVersion           string                              `json:"identity_version"`
	ReceiptID                 string                              `json:"receipt_id"`
	EffectiveAt               string                              `json:"effective_at"`
	OutgoingSigner            SignerManifestEntry                 `json:"outgoing_signer"`
	IncomingSigner            SignerRotationReplacement           `json:"incoming_signer"`
	ApprovalRequirements      []SignerRotationApprovalRequirement `json:"approval_requirements"`
	CoverageStatus            string                              `json:"coverage_status"`
	MissingRoleCoverage       []string                            `json:"missing_role_coverage"`
	ReplacementManifestRef    string                              `json:"replacement_manifest_ref"`
	ReplacementSignerManifest SignerManifestOutput                `json:"replacement_signer_manifest"`
}

type SignerRotationApprovalSignature struct {
	Format      string `json:"format"`
	SignatureID string `json:"signature_id"`
	SignedAt    string `json:"signed_at"`
	ExpiresAt   string `json:"expires_at"`
	Value       string `json:"value"`
}

type SignerRotationApproval struct {
	Version                string                          `json:"version"`
	ReceiptID              string                          `json:"receipt_id"`
	ReceiptDigest          string                          `json:"receipt_digest"`
	ApprovalRole           string                          `json:"approval_role"`
	ApprovedAt             string                          `json:"approved_at"`
	SignerID               string                          `json:"signer_id"`
	KeyID                  string                          `json:"key_id"`
	ActorID                string                          `json:"actor_id"`
	ActorDisplayName       string                          `json:"actor_display_name"`
	OrganizationID         string                          `json:"organization_id,omitempty"`
	OrganizationName       string                          `json:"organization_name,omitempty"`
	ReplacementManifestRef string                          `json:"replacement_manifest_ref"`
	Signature              SignerRotationApprovalSignature `json:"signature"`
}

type SignerRotationApprovalEnvelope struct {
	Approvals []SignerRotationApproval `json:"approvals"`
}

type SignerRotationApprovalRequest struct {
	Receipt      SignerRotationReceiptOutput
	ApprovalRole string
	SignerID     string
	ApprovedAt   string
	SignatureID  string
}

type SignerRotationFinalizeRequest struct {
	Receipt   SignerRotationReceiptOutput
	Approvals []SignerRotationApproval
}

type SignerRotationFinalizedBundle struct {
	Version                   string                              `json:"version"`
	Status                    string                              `json:"status"`
	ReceiptID                 string                              `json:"receipt_id"`
	ReceiptDigest             string                              `json:"receipt_digest"`
	ChainID                   string                              `json:"chain_id"`
	PolicyVersion             string                              `json:"policy_version"`
	IdentityVersion           string                              `json:"identity_version"`
	FinalizedAt               string                              `json:"finalized_at"`
	EffectiveAt               string                              `json:"effective_at"`
	OutgoingSigner            SignerManifestEntry                 `json:"outgoing_signer"`
	IncomingSigner            SignerRotationReplacement           `json:"incoming_signer"`
	ApprovalRequirements      []SignerRotationApprovalRequirement `json:"approval_requirements"`
	Approvals                 []SignerRotationApproval            `json:"approvals"`
	CoverageStatus            string                              `json:"coverage_status"`
	MissingRoleCoverage       []string                            `json:"missing_role_coverage"`
	ReplacementManifestRef    string                              `json:"replacement_manifest_ref"`
	ReplacementSignerManifest SignerManifestOutput                `json:"replacement_signer_manifest"`
}

type CheckpointSignerPolicySigner struct {
	ActorID       string   `json:"actor_id"`
	SignerID      string   `json:"signer_id"`
	KeyID         string   `json:"key_id"`
	Status        string   `json:"status"`
	ProvisionedAt string   `json:"provisioned_at"`
	RotateBy      string   `json:"rotate_by"`
	Roles         []string `json:"roles"`
	SharedSecret  string   `json:"shared_secret"`
}

type CheckpointSignerPolicyRotationPolicy struct {
	ReferenceTime     string   `json:"reference_time"`
	WarningWindowDays int      `json:"warning_window_days"`
	ApprovalRoles     []string `json:"approval_roles"`
}

type CheckpointSignerPolicyOutput struct {
	Version                       string                               `json:"version"`
	SignatureFormat               string                               `json:"signature_format"`
	RequireSignaturesForEventLogs bool                                 `json:"require_signatures_for_event_logs"`
	MaximumSignatureValidity      int                                  `json:"maximum_signature_validity_seconds"`
	RotationPolicy                CheckpointSignerPolicyRotationPolicy `json:"rotation_policy"`
	Signers                       []CheckpointSignerPolicySigner       `json:"signers"`
}

type SignerRotationPolicyPatch struct {
	PolicyPath           string                       `json:"policy_path"`
	PolicyVersionFrom    string                       `json:"policy_version_from"`
	PolicyVersionTo      string                       `json:"policy_version_to"`
	ReferenceTimeFrom    string                       `json:"reference_time_from"`
	ReferenceTimeTo      string                       `json:"reference_time_to"`
	RemoveSignerID       string                       `json:"remove_signer_id"`
	AddSigner            CheckpointSignerPolicySigner `json:"add_signer"`
	ResultingSignerCount int                          `json:"resulting_signer_count"`
}

type SignerRotationActivationRequest struct {
	FinalizedBundle      SignerRotationFinalizedBundle
	IncomingSharedSecret string
}

type SignerRotationActivationPlan struct {
	Version              string                       `json:"version"`
	Status               string                       `json:"status"`
	ReceiptID            string                       `json:"receipt_id"`
	ChainID              string                       `json:"chain_id"`
	PolicyPath           string                       `json:"policy_path"`
	CurrentPolicyVersion string                       `json:"current_policy_version"`
	TargetPolicyVersion  string                       `json:"target_policy_version"`
	EffectiveAt          string                       `json:"effective_at"`
	FinalizedAt          string                       `json:"finalized_at"`
	OutgoingSignerID     string                       `json:"outgoing_signer_id"`
	IncomingSignerID     string                       `json:"incoming_signer_id"`
	ActivationSteps      []string                     `json:"activation_steps"`
	PolicyPatch          SignerRotationPolicyPatch    `json:"policy_patch"`
	ResultingPolicy      CheckpointSignerPolicyOutput `json:"resulting_policy"`
}

type SignerRotationApplyRequest struct {
	ActivationPlan SignerRotationActivationPlan
}

type SignerRotationApplyResult struct {
	Version              string                       `json:"version"`
	Status               string                       `json:"status"`
	ReceiptID            string                       `json:"receipt_id"`
	ChainID              string                       `json:"chain_id"`
	PolicyPath           string                       `json:"policy_path"`
	TargetPolicyVersion  string                       `json:"target_policy_version"`
	ActivationPlanDigest string                       `json:"activation_plan_digest"`
	AppliedPolicyDigest  string                       `json:"applied_policy_digest"`
	AppliedPolicy        CheckpointSignerPolicyOutput `json:"applied_policy"`
	AppliedAtEffect      string                       `json:"applied_at_effective_time"`
}

type SignerRotationVerificationSignature struct {
	Format      string `json:"format"`
	SignatureID string `json:"signature_id"`
	SignedAt    string `json:"signed_at"`
	ExpiresAt   string `json:"expires_at"`
	Value       string `json:"value"`
}

type SignerRotationVerificationReceipt struct {
	Version              string                              `json:"version"`
	Status               string                              `json:"status"`
	ReceiptID            string                              `json:"receipt_id"`
	ChainID              string                              `json:"chain_id"`
	ActivationPlanDigest string                              `json:"activation_plan_digest"`
	PolicyDigest         string                              `json:"policy_digest"`
	VerifiedAt           string                              `json:"verified_at"`
	PolicyPath           string                              `json:"policy_path"`
	TargetPolicyVersion  string                              `json:"target_policy_version"`
	SignerID             string                              `json:"signer_id"`
	KeyID                string                              `json:"key_id"`
	ActorID              string                              `json:"actor_id"`
	ActorDisplayName     string                              `json:"actor_display_name"`
	OrganizationID       string                              `json:"organization_id,omitempty"`
	OrganizationName     string                              `json:"organization_name,omitempty"`
	Signature            SignerRotationVerificationSignature `json:"signature"`
}

type SignerRotationVerifyRequest struct {
	ActivationPlan SignerRotationActivationPlan
	Policy         CheckpointSignerPolicyOutput
	VerifiedAt     string
	SignatureID    string
}

type SignerRotationActivationAuditEntry struct {
	Version              string                              `json:"version"`
	Status               string                              `json:"status"`
	ReceiptID            string                              `json:"receipt_id"`
	ChainID              string                              `json:"chain_id"`
	PolicyPath           string                              `json:"policy_path"`
	TargetPolicyVersion  string                              `json:"target_policy_version"`
	EffectiveAt          string                              `json:"effective_at"`
	VerifiedAt           string                              `json:"verified_at"`
	ActivationPlanDigest string                              `json:"activation_plan_digest"`
	AppliedPolicyDigest  string                              `json:"applied_policy_digest"`
	SignerID             string                              `json:"signer_id"`
	KeyID                string                              `json:"key_id"`
	ActorID              string                              `json:"actor_id"`
	ActorDisplayName     string                              `json:"actor_display_name"`
	OrganizationID       string                              `json:"organization_id,omitempty"`
	OrganizationName     string                              `json:"organization_name,omitempty"`
	Signature            SignerRotationVerificationSignature `json:"signature"`
}

type SignerRotationActivationAuditLedger struct {
	Version    string                               `json:"version"`
	Status     string                               `json:"status"`
	ChainID    string                               `json:"chain_id"`
	PolicyPath string                               `json:"policy_path"`
	EntryCount int                                  `json:"entry_count"`
	Entries    []SignerRotationActivationAuditEntry `json:"entries"`
}

type SignerRotationActivationAuditAppendRequest struct {
	ApplyResult         SignerRotationApplyResult
	VerificationReceipt SignerRotationVerificationReceipt
	ExistingLedger      SignerRotationActivationAuditLedger
}

type SignerRotationActivationAuditAppendResult struct {
	Version       string                              `json:"version"`
	Status        string                              `json:"status"`
	AppendedIndex int                                 `json:"appended_index"`
	AppendedEntry SignerRotationActivationAuditEntry  `json:"appended_entry"`
	Ledger        SignerRotationActivationAuditLedger `json:"ledger"`
}

type SignerRotationActivationAuditReconcileRequest struct {
	Ledger     SignerRotationActivationAuditLedger
	Policy     CheckpointSignerPolicyOutput
	PolicyPath string
}

type SignerRotationActivationAuditReconcileReport struct {
	Version                string   `json:"version"`
	Status                 string   `json:"status"`
	ChainID                string   `json:"chain_id"`
	PolicyPath             string   `json:"policy_path"`
	CurrentPolicyVersion   string   `json:"current_policy_version"`
	CurrentPolicyDigest    string   `json:"current_policy_digest"`
	EntryCount             int      `json:"entry_count"`
	LatestReceiptID        string   `json:"latest_receipt_id,omitempty"`
	LatestTargetVersion    string   `json:"latest_target_policy_version,omitempty"`
	LatestPolicyDigest     string   `json:"latest_applied_policy_digest,omitempty"`
	CurrentPolicyExplained bool     `json:"current_policy_explained"`
	Issues                 []string `json:"issues"`
}

type SignerRotationActivationAuditBaselineSnapshot struct {
	Version                string                              `json:"version"`
	Status                 string                              `json:"status"`
	ChainID                string                              `json:"chain_id"`
	PolicyPath             string                              `json:"policy_path"`
	LedgerEntryCount       int                                 `json:"ledger_entry_count"`
	CurrentPolicyVersion   string                              `json:"current_policy_version"`
	CurrentPolicyDigest    string                              `json:"current_policy_digest"`
	CurrentPolicyExplained bool                                `json:"current_policy_explained"`
	LatestReceiptID        string                              `json:"latest_receipt_id,omitempty"`
	LatestTargetVersion    string                              `json:"latest_target_policy_version,omitempty"`
	LatestPolicyDigest     string                              `json:"latest_applied_policy_digest,omitempty"`
	LatestEntry            *SignerRotationActivationAuditEntry `json:"latest_entry,omitempty"`
	ReconciliationStatus   string                              `json:"reconciliation_status"`
	ContinuityIssues       []string                            `json:"continuity_issues"`
}

type SignerRotationActivationAuditExportRequest struct {
	Ledger         SignerRotationActivationAuditLedger
	Policy         CheckpointSignerPolicyOutput
	Reconciliation SignerRotationActivationAuditReconcileReport
	PolicyPath     string
}

type SignerRotationActivationAuditExportPackage struct {
	Version          string                                        `json:"version"`
	Status           string                                        `json:"status"`
	ChainID          string                                        `json:"chain_id"`
	PolicyPath       string                                        `json:"policy_path"`
	BaselineSnapshot SignerRotationActivationAuditBaselineSnapshot `json:"baseline_snapshot"`
	CurrentPolicy    CheckpointSignerPolicyOutput                  `json:"current_policy"`
	Ledger           SignerRotationActivationAuditLedger           `json:"ledger"`
	Reconciliation   SignerRotationActivationAuditReconcileReport  `json:"reconciliation"`
}

type SignerRotationActivationAuditExportVerificationRequest struct {
	ExportPackage SignerRotationActivationAuditExportPackage
}

type SignerRotationActivationAuditExportVerificationReport struct {
	Version              string   `json:"version"`
	Status               string   `json:"status"`
	ArchiveReady         bool     `json:"archive_ready"`
	ChainID              string   `json:"chain_id"`
	PolicyPath           string   `json:"policy_path"`
	CurrentPolicyVersion string   `json:"current_policy_version"`
	CurrentPolicyDigest  string   `json:"current_policy_digest"`
	EntryCount           int      `json:"entry_count"`
	LatestReceiptID      string   `json:"latest_receipt_id,omitempty"`
	LatestTargetVersion  string   `json:"latest_target_policy_version,omitempty"`
	VerificationIssues   []string `json:"verification_issues"`
}

type SignerRotationActivationAuditArchivePackage struct {
	PackagePath   string
	ExportPackage SignerRotationActivationAuditExportPackage
}

type SignerRotationActivationAuditArchiveIndexRequest struct {
	Packages []SignerRotationActivationAuditArchivePackage
}

type SignerRotationActivationAuditArchiveIndexEntry struct {
	PackagePath          string   `json:"package_path"`
	Status               string   `json:"status"`
	ArchiveReady         bool     `json:"archive_ready"`
	ChainID              string   `json:"chain_id"`
	PolicyPath           string   `json:"policy_path"`
	CurrentPolicyVersion string   `json:"current_policy_version"`
	CurrentPolicyDigest  string   `json:"current_policy_digest"`
	EntryCount           int      `json:"entry_count"`
	LatestReceiptID      string   `json:"latest_receipt_id,omitempty"`
	LatestTargetVersion  string   `json:"latest_target_policy_version,omitempty"`
	VerificationIssues   []string `json:"verification_issues,omitempty"`
}

type SignerRotationActivationAuditArchiveIndex struct {
	Version                    string                                           `json:"version"`
	Status                     string                                           `json:"status"`
	PackageCount               int                                              `json:"package_count"`
	ArchiveReadyCount          int                                              `json:"archive_ready_count"`
	ChainID                    string                                           `json:"chain_id"`
	PolicyPath                 string                                           `json:"policy_path"`
	LatestCurrentPolicyVersion string                                           `json:"latest_current_policy_version,omitempty"`
	LatestCurrentPolicyDigest  string                                           `json:"latest_current_policy_digest,omitempty"`
	Entries                    []SignerRotationActivationAuditArchiveIndexEntry `json:"entries"`
	Issues                     []string                                         `json:"issues"`
}

type SignerRotationActivationAuditArchivePromotionRequest struct {
	PackagePath        string
	ExportPackage      SignerRotationActivationAuditExportPackage
	VerificationReport SignerRotationActivationAuditExportVerificationReport
	ArchiveIndex       SignerRotationActivationAuditArchiveIndex
	PromotedAt         string
	PromotedBy         string
}

type SignerRotationActivationAuditArchivePromotionReceipt struct {
	Version              string `json:"version"`
	Status               string `json:"status"`
	ReceiptID            string `json:"receipt_id"`
	PackagePath          string `json:"package_path"`
	PromotedAt           string `json:"promoted_at"`
	PromotedBy           string `json:"promoted_by"`
	ChainID              string `json:"chain_id"`
	PolicyPath           string `json:"policy_path"`
	CurrentPolicyVersion string `json:"current_policy_version"`
	CurrentPolicyDigest  string `json:"current_policy_digest"`
	EntryCount           int    `json:"entry_count"`
	LatestReceiptID      string `json:"latest_receipt_id,omitempty"`
	LatestTargetVersion  string `json:"latest_target_policy_version,omitempty"`
	ExportPackageDigest  string `json:"export_package_digest"`
	VerificationDigest   string `json:"verification_digest"`
	ArchiveIndexDigest   string `json:"archive_index_digest"`
	ArchiveEntryDigest   string `json:"archive_entry_digest"`
}

type SignerRotationActivationAuditRetainedBaselineAttestation struct {
	Version              string   `json:"version"`
	Status               string   `json:"status"`
	AttestationID        string   `json:"attestation_id"`
	PromotionReceiptID   string   `json:"promotion_receipt_id"`
	PackagePath          string   `json:"package_path"`
	ArchiveEntryIndex    int      `json:"archive_entry_index"`
	PromotedAt           string   `json:"promoted_at"`
	PromotedBy           string   `json:"promoted_by"`
	ChainID              string   `json:"chain_id"`
	PolicyPath           string   `json:"policy_path"`
	CurrentPolicyVersion string   `json:"current_policy_version"`
	CurrentPolicyDigest  string   `json:"current_policy_digest"`
	LatestReceiptID      string   `json:"latest_receipt_id,omitempty"`
	LatestTargetVersion  string   `json:"latest_target_policy_version,omitempty"`
	ExportPackageDigest  string   `json:"export_package_digest"`
	VerificationDigest   string   `json:"verification_digest"`
	ArchiveIndexDigest   string   `json:"archive_index_digest"`
	ArchiveEntryDigest   string   `json:"archive_entry_digest"`
	Claims               []string `json:"claims"`
}

type SignerRotationActivationAuditArchivePromotionResult struct {
	Version                     string                                                   `json:"version"`
	Status                      string                                                   `json:"status"`
	PromotionReceipt            SignerRotationActivationAuditArchivePromotionReceipt     `json:"promotion_receipt"`
	RetainedBaselineAttestation SignerRotationActivationAuditRetainedBaselineAttestation `json:"retained_baseline_attestation"`
	Issues                      []string                                                 `json:"issues"`
}

type SignerRotationActivationAuditArchivePromotionVerificationRequest struct {
	PackagePath        string
	ExportPackage      SignerRotationActivationAuditExportPackage
	VerificationReport SignerRotationActivationAuditExportVerificationReport
	ArchiveIndex       SignerRotationActivationAuditArchiveIndex
	PromotionResult    SignerRotationActivationAuditArchivePromotionResult
	VerifiedAt         string
	VerifiedBy         string
}

type SignerRotationActivationAuditArchivePromotionVerificationReceipt struct {
	Version                string   `json:"version"`
	Status                 string   `json:"status"`
	VerificationReceiptID  string   `json:"verification_receipt_id"`
	PackagePath            string   `json:"package_path"`
	VerifiedAt             string   `json:"verified_at"`
	VerifiedBy             string   `json:"verified_by"`
	PromotionReceiptID     string   `json:"promotion_receipt_id"`
	AttestationID          string   `json:"attestation_id"`
	ChainID                string   `json:"chain_id"`
	PolicyPath             string   `json:"policy_path"`
	CurrentPolicyVersion   string   `json:"current_policy_version"`
	CurrentPolicyDigest    string   `json:"current_policy_digest"`
	LatestReceiptID        string   `json:"latest_receipt_id,omitempty"`
	LatestTargetVersion    string   `json:"latest_target_policy_version,omitempty"`
	PromotionResultDigest  string   `json:"promotion_result_digest"`
	PromotionReceiptDigest string   `json:"promotion_receipt_digest"`
	AttestationDigest      string   `json:"attestation_digest"`
	ArchiveIndexDigest     string   `json:"archive_index_digest"`
	ArchiveEntryDigest     string   `json:"archive_entry_digest"`
	VerificationIssues     []string `json:"verification_issues"`
}

type SignerRotationActivationAuditRetainedInventoryPackage struct {
	PromotionPath       string
	PromotionResult     SignerRotationActivationAuditArchivePromotionResult
	VerificationReceipt SignerRotationActivationAuditArchivePromotionVerificationReceipt
}

type SignerRotationActivationAuditRetainedInventorySnapshotRequest struct {
	Packages []SignerRotationActivationAuditRetainedInventoryPackage
}

type SignerRotationActivationAuditRetainedInventorySnapshotEntry struct {
	PromotionPath             string   `json:"promotion_path"`
	Status                    string   `json:"status"`
	Verified                  bool     `json:"verified"`
	VerificationReceiptID     string   `json:"verification_receipt_id"`
	PromotionReceiptID        string   `json:"promotion_receipt_id"`
	AttestationID             string   `json:"attestation_id"`
	ChainID                   string   `json:"chain_id"`
	PolicyPath                string   `json:"policy_path"`
	CurrentPolicyVersion      string   `json:"current_policy_version"`
	CurrentPolicyDigest       string   `json:"current_policy_digest"`
	LatestReceiptID           string   `json:"latest_receipt_id,omitempty"`
	LatestTargetVersion       string   `json:"latest_target_policy_version,omitempty"`
	PromotedAt                string   `json:"promoted_at"`
	PromotedBy                string   `json:"promoted_by"`
	VerifiedAt                string   `json:"verified_at"`
	VerifiedBy                string   `json:"verified_by"`
	PromotionResultDigest     string   `json:"promotion_result_digest"`
	VerificationReceiptDigest string   `json:"verification_receipt_digest"`
	VerificationIssues        []string `json:"verification_issues,omitempty"`
}

type SignerRotationActivationAuditRetainedInventorySnapshot struct {
	Version                    string                                                        `json:"version"`
	Status                     string                                                        `json:"status"`
	SnapshotReceiptID          string                                                        `json:"snapshot_receipt_id,omitempty"`
	PackageCount               int                                                           `json:"package_count"`
	VerifiedCount              int                                                           `json:"verified_count"`
	ChainID                    string                                                        `json:"chain_id"`
	PolicyPath                 string                                                        `json:"policy_path"`
	LatestCurrentPolicyVersion string                                                        `json:"latest_current_policy_version,omitempty"`
	LatestCurrentPolicyDigest  string                                                        `json:"latest_current_policy_digest,omitempty"`
	Entries                    []SignerRotationActivationAuditRetainedInventorySnapshotEntry `json:"entries"`
	Issues                     []string                                                      `json:"issues"`
}

type SignerRotationActivationAuditRetainedInventoryVerificationRequest struct {
	Snapshot   SignerRotationActivationAuditRetainedInventorySnapshot
	Packages   []SignerRotationActivationAuditRetainedInventoryPackage
	VerifiedAt string
	VerifiedBy string
}

type SignerRotationActivationAuditRetainedInventoryVerificationReceipt struct {
	Version                         string   `json:"version"`
	Status                          string   `json:"status"`
	VerificationReceiptID           string   `json:"verification_receipt_id"`
	VerifiedAt                      string   `json:"verified_at"`
	VerifiedBy                      string   `json:"verified_by"`
	ChainID                         string   `json:"chain_id"`
	PolicyPath                      string   `json:"policy_path"`
	PackageCount                    int      `json:"package_count"`
	VerifiedCount                   int      `json:"verified_count"`
	LatestCurrentPolicyVersion      string   `json:"latest_current_policy_version,omitempty"`
	LatestCurrentPolicyDigest       string   `json:"latest_current_policy_digest,omitempty"`
	InventorySnapshotDigest         string   `json:"inventory_snapshot_digest"`
	ExpectedInventorySnapshotDigest string   `json:"expected_inventory_snapshot_digest"`
	VerificationIssues              []string `json:"verification_issues"`
}

type SignerRotationActivationAuditRetainedInventoryContinuityPackage struct {
	SnapshotPath        string
	Snapshot            SignerRotationActivationAuditRetainedInventorySnapshot
	VerificationReceipt SignerRotationActivationAuditRetainedInventoryVerificationReceipt
}

type SignerRotationActivationAuditRetainedInventoryContinuityManifestRequest struct {
	Snapshots []SignerRotationActivationAuditRetainedInventoryContinuityPackage
}

type SignerRotationActivationAuditRetainedInventoryContinuityEntry struct {
	SnapshotPath               string `json:"snapshot_path"`
	Status                     string `json:"status"`
	Verified                   bool   `json:"verified"`
	VerificationReceiptID      string `json:"verification_receipt_id"`
	InventorySnapshotDigest    string `json:"inventory_snapshot_digest"`
	VerificationReceiptDigest  string `json:"verification_receipt_digest"`
	PackageCount               int    `json:"package_count"`
	VerifiedCount              int    `json:"verified_count"`
	ChainID                    string `json:"chain_id"`
	PolicyPath                 string `json:"policy_path"`
	LatestCurrentPolicyVersion string `json:"latest_current_policy_version,omitempty"`
	LatestCurrentPolicyDigest  string `json:"latest_current_policy_digest,omitempty"`
	VerifiedAt                 string `json:"verified_at"`
	VerifiedBy                 string `json:"verified_by"`
	SnapshotReceiptID          string `json:"snapshot_receipt_id,omitempty"`
}

type SignerRotationActivationAuditRetainedInventoryContinuityManifest struct {
	Version                    string                                                          `json:"version"`
	Status                     string                                                          `json:"status"`
	SnapshotCount              int                                                             `json:"snapshot_count"`
	VerifiedSnapshotCount      int                                                             `json:"verified_snapshot_count"`
	ChainID                    string                                                          `json:"chain_id"`
	PolicyPath                 string                                                          `json:"policy_path"`
	ManifestID                 string                                                          `json:"manifest_id,omitempty"`
	ChainID                    string                                                          `json:"chain_id"`
	PolicyPath                 string                                                          `json:"policy_path"`
	SnapshotCount              int                                                             `json:"snapshot_count"`
	LatestCurrentPolicyVersion string                                                          `json:"latest_current_policy_version,omitempty"`
	LatestCurrentPolicyDigest  string                                                          `json:"latest_current_policy_digest,omitempty"`
	Entries                    []SignerRotationActivationAuditRetainedInventoryContinuityEntry `json:"entries"`
	Issues                     []string                                                        `json:"issues"`
}

type SignerRotationActivationAuditRetainedInventoryContinuityRequest struct {
	Snapshots []SignerRotationActivationAuditRetainedInventorySnapshot
}

type parsedSignerEntry struct {
	ActorID           string
	SignerID          string
	KeyID             string
	Status            string
	Roles             []string
	ProvisionedAt     time.Time
	RotateBy          time.Time
	RotationStatus    string
	DaysUntilRotation int
}

func governanceExecutionRoleUsage(bundle Bundle) map[string][]string {
	roleUsage := make(map[string][]string)
	executionDefaults := stringMap(stringMap(stringMap(bundle.InferencePolicy["remediation"])["execution_defaults"]))
	for phase, role := range stringMap(executionDefaults["phase_owners"]) {
		appendIdentityRoleUsage(roleUsage, fmt.Sprint(role), "governance:phase_owner:"+phase)
	}
	for proposalKind, override := range stringMap(executionDefaults["owner_overrides"]) {
		for phase, role := range stringMap(override) {
			appendIdentityRoleUsage(roleUsage, fmt.Sprint(role), "governance:owner_override:"+proposalKind+":"+phase)
		}
	}
	return roleUsage
}

func parseRFC3339(value any, field string) (time.Time, error) {
	text := fmt.Sprint(value)
	if text == "" {
		return time.Time{}, fmt.Errorf("%s must be set", field)
	}
	parsed, err := time.Parse(time.RFC3339, text)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must be RFC3339 timestamp: %w", field, err)
	}
	return parsed.UTC(), nil
}

func intValue(value any) (int, error) {
	switch typed := value.(type) {
	case int:
		return typed, nil
	case int64:
		return int(typed), nil
	case float64:
		return int(typed), nil
	case string:
		return strconv.Atoi(typed)
	default:
		return 0, fmt.Errorf("unsupported integer type %T", value)
	}
}

func signerRotationStatus(referenceTime time.Time, warningWindowDays int, signerStatus string, rotateBy time.Time) string {
	if signerStatus != "active" {
		return "inactive"
	}
	if !rotateBy.After(referenceTime) {
		return "stale"
	}
	if rotateBy.Sub(referenceTime) <= (time.Duration(warningWindowDays) * 24 * time.Hour) {
		return "expiring"
	}
	return "current"
}

func signerReceiptArtifacts(signerID string, targetTime time.Time) (string, string) {
	stamp := strings.ToLower(targetTime.UTC().Format("20060102t150405z"))
	receiptID := fmt.Sprintf("rotation-%s-%s", signerID, stamp)
	return receiptID, fmt.Sprintf("build/rotation/%s/replacement-signer-manifest.json", receiptID)
}

func governanceApprovalRoles(bundle Bundle) ([]string, error) {
	rotationPolicy := stringMap(bundle.CheckpointSigners["rotation_policy"])
	rawRoles, ok := rotationPolicy["approval_roles"].([]any)
	if !ok || len(rawRoles) == 0 {
		return nil, fmt.Errorf("checkpoint signer rotation_policy.approval_roles must not be empty")
	}
	roles := make([]string, 0, len(rawRoles))
	for _, rawRole := range rawRoles {
		role := fmt.Sprint(rawRole)
		if role != "" {
			roles = append(roles, role)
		}
	}
	sort.Strings(roles)
	return roles, nil
}

func activeActorsAndBindings(bundle Bundle) (map[string]IdentityActor, map[string]struct{}) {
	actors := make(map[string]IdentityActor, len(bundle.Identity.Actors))
	activeRoleBindings := make(map[string]struct{}, len(bundle.Identity.RoleBindings))
	for _, actor := range bundle.Identity.Actors {
		actors[actor.ActorID] = actor
	}
	for _, binding := range bundle.Identity.RoleBindings {
		if binding.Status == "active" {
			activeRoleBindings[binding.ActorID+"|"+binding.Role] = struct{}{}
		}
	}
	return actors, activeRoleBindings
}

func governanceRequiredRoles(bundle Bundle) []string {
	required := make([]string, 0)
	roleUsage := governanceExecutionRoleUsage(bundle)
	for role := range roleUsage {
		required = append(required, role)
	}
	sort.Strings(required)
	return required
}

func parseCurrentSigners(bundle Bundle, referenceTime time.Time, warningWindowDays int) ([]parsedSignerEntry, error) {
	rawSigners, ok := bundle.CheckpointSigners["signers"].([]any)
	if !ok || len(rawSigners) == 0 {
		return nil, fmt.Errorf("at least one checkpoint signer must be configured")
	}

	parsedSigners := make([]parsedSignerEntry, 0, len(rawSigners))
	for _, rawSigner := range rawSigners {
		signer := stringMap(rawSigner)
		provisionedAt, err := parseRFC3339(signer["provisioned_at"], "checkpoint signer provisioned_at")
		if err != nil {
			return nil, err
		}
		rotateBy, err := parseRFC3339(signer["rotate_by"], "checkpoint signer rotate_by")
		if err != nil {
			return nil, err
		}
		status := fmt.Sprint(signer["status"])
		daysUntilRotation := int(rotateBy.Sub(referenceTime).Hours() / 24)
		parsedSigners = append(parsedSigners, parsedSignerEntry{
			ActorID:           fmt.Sprint(signer["actor_id"]),
			SignerID:          fmt.Sprint(signer["signer_id"]),
			KeyID:             fmt.Sprint(signer["key_id"]),
			Status:            status,
			Roles:             stringSlice(signer["roles"]),
			ProvisionedAt:     provisionedAt,
			RotateBy:          rotateBy,
			RotationStatus:    signerRotationStatus(referenceTime, warningWindowDays, status, rotateBy),
			DaysUntilRotation: daysUntilRotation,
		})
	}
	return parsedSigners, nil
}

func validateSignerSet(
	bundle Bundle,
	parsedSigners []parsedSignerEntry,
	referenceTime time.Time,
	warningWindowDays int,
	requiredRoles []string,
) error {
	actors, activeRoleBindings := activeActorsAndBindings(bundle)

	signerIDs := make(map[string]struct{}, len(parsedSigners))
	keyIDs := make(map[string]struct{}, len(parsedSigners))
	activeActorIDs := make(map[string]struct{}, len(parsedSigners))
	activeRoleCoverage := make(map[string]string)

	for _, signer := range parsedSigners {
		if signer.ActorID == "" {
			return fmt.Errorf("checkpoint signer %s must declare actor_id", signer.SignerID)
		}
		if signer.SignerID == "" {
			return fmt.Errorf("checkpoint signer signer_id must not be empty")
		}
		if _, exists := signerIDs[signer.SignerID]; exists {
			return fmt.Errorf("duplicate checkpoint signer_id: %s", signer.SignerID)
		}
		if _, exists := keyIDs[signer.KeyID]; exists {
			return fmt.Errorf("duplicate checkpoint key_id: %s", signer.KeyID)
		}
		if signer.Status != "active" && signer.Status != "inactive" {
			return fmt.Errorf("checkpoint signer %s has invalid status", signer.SignerID)
		}
		if len(signer.Roles) == 0 {
			return fmt.Errorf("checkpoint signer %s must declare at least one role", signer.SignerID)
		}
		if !signer.RotateBy.After(signer.ProvisionedAt) {
			return fmt.Errorf("checkpoint signer %s rotate_by must be after provisioned_at", signer.SignerID)
		}

		actor, exists := actors[signer.ActorID]
		if !exists {
			return fmt.Errorf("checkpoint signer %s references unknown actor %s", signer.SignerID, signer.ActorID)
		}
		if actor.Status != "active" {
			return fmt.Errorf("checkpoint signer %s references inactive actor %s", signer.SignerID, signer.ActorID)
		}

		for _, role := range signer.Roles {
			if _, exists := activeRoleBindings[signer.ActorID+"|"+role]; !exists {
				return fmt.Errorf(
					"checkpoint signer %s role %s is not backed by an active identity binding",
					signer.SignerID,
					role,
				)
			}
		}

		if signer.Status == "active" {
			if _, exists := activeActorIDs[signer.ActorID]; exists {
				return fmt.Errorf("duplicate checkpoint signer actor ownership: %s", signer.ActorID)
			}
			activeActorIDs[signer.ActorID] = struct{}{}
			if !signer.RotateBy.After(referenceTime) {
				return fmt.Errorf("checkpoint signer %s has stale rotation metadata", signer.SignerID)
			}
			for _, role := range signer.Roles {
				if owner, exists := activeRoleCoverage[role]; exists {
					return fmt.Errorf("duplicate checkpoint signer role coverage: %s (%s, %s)", role, owner, signer.SignerID)
				}
				activeRoleCoverage[role] = signer.SignerID
			}
			if signerRotationStatus(referenceTime, warningWindowDays, signer.Status, signer.RotateBy) == "stale" {
				return fmt.Errorf("checkpoint signer %s has stale rotation metadata", signer.SignerID)
			}
		}

		signerIDs[signer.SignerID] = struct{}{}
		keyIDs[signer.KeyID] = struct{}{}
	}

	missingRoles := make([]string, 0)
	for _, role := range requiredRoles {
		if _, exists := activeRoleCoverage[role]; !exists {
			missingRoles = append(missingRoles, role)
		}
	}
	sort.Strings(missingRoles)
	if len(missingRoles) > 0 {
		return fmt.Errorf("checkpoint signer coverage missing roles: %s", commaList(missingRoles))
	}

	return nil
}

func ValidateSignerManifestInputs(bundle Bundle) error {
	if err := ValidateIdentityBootstrap(bundle); err != nil {
		return err
	}

	if fmt.Sprint(bundle.CheckpointSigners["signature_format"]) != "0ai-hmac-sha256-v1" {
		return fmt.Errorf("checkpoint signer signature_format must be 0ai-hmac-sha256-v1")
	}
	if fmt.Sprint(bundle.CheckpointSigners["require_signatures_for_event_logs"]) != "true" {
		return fmt.Errorf("checkpoint signer policy must require signatures for event logs")
	}
	maxValidity, err := intValue(bundle.CheckpointSigners["maximum_signature_validity_seconds"])
	if err != nil || maxValidity <= 0 {
		return fmt.Errorf("checkpoint signer validity window must be positive")
	}
	rotationPolicy := stringMap(bundle.CheckpointSigners["rotation_policy"])
	referenceTime, err := parseRFC3339(rotationPolicy["reference_time"], "checkpoint signer rotation_policy.reference_time")
	if err != nil {
		return err
	}
	warningWindowDays, err := intValue(rotationPolicy["warning_window_days"])
	if err != nil || warningWindowDays <= 0 {
		return fmt.Errorf("checkpoint signer rotation_policy.warning_window_days must be positive")
	}
	approvalRoles, err := governanceApprovalRoles(bundle)
	if err != nil {
		return err
	}

	requiredRoles := governanceRequiredRoles(bundle)
	parsedSigners, err := parseCurrentSigners(bundle, referenceTime, warningWindowDays)
	if err != nil {
		return err
	}
	if err := validateSignerSet(bundle, parsedSigners, referenceTime, warningWindowDays, requiredRoles); err != nil {
		return err
	}

	activeRoleCoverage := make(map[string]struct{})
	for _, signer := range parsedSigners {
		if signer.Status != "active" {
			continue
		}
		for _, role := range signer.Roles {
			activeRoleCoverage[role] = struct{}{}
		}
	}
	missingApprovalRoles := make([]string, 0)
	for _, role := range approvalRoles {
		if _, exists := activeRoleCoverage[role]; !exists {
			missingApprovalRoles = append(missingApprovalRoles, role)
		}
	}
	if len(missingApprovalRoles) > 0 {
		return fmt.Errorf("checkpoint signer approval roles missing active coverage: %s", commaList(missingApprovalRoles))
	}
	return nil
}

func signerRoleCoverage(roleUsage map[string][]string, roles []string) []SignerRoleCoverage {
	coverage := make([]SignerRoleCoverage, 0, len(roles))
	for _, role := range roles {
		referencedBy := append([]string(nil), roleUsage[role]...)
		sort.Strings(referencedBy)
		coverage = append(coverage, SignerRoleCoverage{
			Role:         role,
			ReferencedBy: referencedBy,
		})
	}
	sort.Slice(coverage, func(i, j int) bool {
		return coverage[i].Role < coverage[j].Role
	})
	return coverage
}

func buildSignerManifestEntry(
	actors map[string]IdentityActor,
	roleUsage map[string][]string,
	signer parsedSignerEntry,
) SignerManifestEntry {
	actor := actors[signer.ActorID]
	organizationName := ""
	if actor.OrganizationID != "" {
		if organization, exists := actors[actor.OrganizationID]; exists {
			organizationName = organization.DisplayName
		}
	}
	receiptID, manifestRef := signerReceiptArtifacts(signer.SignerID, signer.RotateBy)
	return SignerManifestEntry{
		ActorID:                signer.ActorID,
		ActorDisplayName:       actor.DisplayName,
		ActorStatus:            actor.Status,
		OrganizationID:         actor.OrganizationID,
		OrganizationName:       organizationName,
		SignerID:               signer.SignerID,
		KeyID:                  signer.KeyID,
		SignerStatus:           signer.Status,
		Roles:                  append([]string(nil), signer.Roles...),
		RoleCoverage:           signerRoleCoverage(roleUsage, signer.Roles),
		ProvisionedAt:          signer.ProvisionedAt.Format(time.RFC3339),
		RotateBy:               signer.RotateBy.Format(time.RFC3339),
		RotationStatus:         signer.RotationStatus,
		DaysUntilRotation:      signer.DaysUntilRotation,
		NextRotationReceiptID:  receiptID,
		ReplacementManifestRef: manifestRef,
	}
}

func buildSignerManifestOutput(
	bundle Bundle,
	parsedSigners []parsedSignerEntry,
	referenceTime time.Time,
	warningWindowDays int,
) SignerManifestOutput {
	roleUsage := governanceExecutionRoleUsage(bundle)
	actors, _ := activeActorsAndBindings(bundle)

	sortedForPlan := append([]parsedSignerEntry(nil), parsedSigners...)
	sort.Slice(sortedForPlan, func(i, j int) bool {
		if sortedForPlan[i].RotateBy.Equal(sortedForPlan[j].RotateBy) {
			return sortedForPlan[i].SignerID < sortedForPlan[j].SignerID
		}
		return sortedForPlan[i].RotateBy.Before(sortedForPlan[j].RotateBy)
	})

	entries := make([]SignerManifestEntry, 0, len(sortedForPlan))
	expiring := make([]string, 0)
	rotationPlan := make([]SignerRotationPlanEntry, 0)
	activeCount := 0

	for _, signer := range sortedForPlan {
		entry := buildSignerManifestEntry(actors, roleUsage, signer)
		entries = append(entries, entry)
		if signer.Status != "active" {
			continue
		}
		activeCount++
		if signer.RotationStatus == "expiring" {
			expiring = append(expiring, signer.SignerID)
		}
		rotationPlan = append(rotationPlan, SignerRotationPlanEntry{
			SignerID:               signer.SignerID,
			ActorID:                signer.ActorID,
			Roles:                  append([]string(nil), signer.Roles...),
			RotateBy:               signer.RotateBy.Format(time.RFC3339),
			RotationStatus:         signer.RotationStatus,
			DaysUntilRotation:      signer.DaysUntilRotation,
			RecommendedAction:      recommendedRotationAction(signer.RotationStatus),
			NextRotationReceiptID:  entry.NextRotationReceiptID,
			ReplacementManifestRef: entry.ReplacementManifestRef,
		})
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].SignerID < entries[j].SignerID
	})
	sort.Strings(expiring)
	for idx := range rotationPlan {
		rotationPlan[idx].Order = idx + 1
	}

	return SignerManifestOutput{
		Version:             "1.1.0",
		ChainID:             bundle.Topology.ChainID,
		PolicyVersion:       fmt.Sprint(bundle.CheckpointSigners["version"]),
		IdentityVersion:     bundle.Identity.Version,
		ReferenceTime:       referenceTime.Format(time.RFC3339),
		WarningWindowDays:   warningWindowDays,
		SignerCount:         len(entries),
		ActiveSignerCount:   activeCount,
		MissingRoleCoverage: []string{},
		ExpiringSignerIDs:   expiring,
		RotationPlan:        rotationPlan,
		Signers:             entries,
	}
}

func SignerManifest(bundle Bundle) (SignerManifestOutput, error) {
	if err := ValidateSignerManifestInputs(bundle); err != nil {
		return SignerManifestOutput{}, err
	}

	rotationPolicy := stringMap(bundle.CheckpointSigners["rotation_policy"])
	referenceTime, _ := parseRFC3339(rotationPolicy["reference_time"], "checkpoint signer rotation_policy.reference_time")
	warningWindowDays, _ := intValue(rotationPolicy["warning_window_days"])
	parsedSigners, err := parseCurrentSigners(bundle, referenceTime, warningWindowDays)
	if err != nil {
		return SignerManifestOutput{}, err
	}
	return buildSignerManifestOutput(bundle, parsedSigners, referenceTime, warningWindowDays), nil
}

func findParsedSignerByID(signers []parsedSignerEntry, signerID string) (parsedSignerEntry, bool) {
	for _, signer := range signers {
		if signer.SignerID == signerID {
			return signer, true
		}
	}
	return parsedSignerEntry{}, false
}

func replacementSigners(signers []parsedSignerEntry, outgoingSignerID string, incoming parsedSignerEntry) []parsedSignerEntry {
	replaced := make([]parsedSignerEntry, 0, len(signers))
	for _, signer := range signers {
		if signer.SignerID == outgoingSignerID {
			continue
		}
		replaced = append(replaced, signer)
	}
	replaced = append(replaced, incoming)
	return replaced
}

func approvalRequirements(bundle Bundle, parsedSigners []parsedSignerEntry) ([]SignerRotationApprovalRequirement, error) {
	approvalRoles, err := governanceApprovalRoles(bundle)
	if err != nil {
		return nil, err
	}
	actors, _ := activeActorsAndBindings(bundle)
	optionsByRole := make(map[string][]SignerRotationApprovalOption)
	for _, signer := range parsedSigners {
		if signer.Status != "active" {
			continue
		}
		actor := actors[signer.ActorID]
		organizationName := ""
		if actor.OrganizationID != "" {
			if organization, exists := actors[actor.OrganizationID]; exists {
				organizationName = organization.DisplayName
			}
		}
		for _, role := range signer.Roles {
			optionsByRole[role] = append(optionsByRole[role], SignerRotationApprovalOption{
				ActorID:          actor.ActorID,
				ActorDisplayName: actor.DisplayName,
				OrganizationID:   actor.OrganizationID,
				OrganizationName: organizationName,
				SignerID:         signer.SignerID,
			})
		}
	}

	requirements := make([]SignerRotationApprovalRequirement, 0, len(approvalRoles))
	for _, role := range approvalRoles {
		options := append([]SignerRotationApprovalOption(nil), optionsByRole[role]...)
		sort.Slice(options, func(i, j int) bool {
			if options[i].ActorID == options[j].ActorID {
				return options[i].SignerID < options[j].SignerID
			}
			return options[i].ActorID < options[j].ActorID
		})
		requirements = append(requirements, SignerRotationApprovalRequirement{
			Role:            role,
			EligibleSigners: options,
		})
	}
	return requirements, nil
}

func signerRotationReceiptDigest(receipt SignerRotationReceiptOutput) (string, error) {
	encoded, err := json.Marshal(receipt)
	if err != nil {
		return "", fmt.Errorf("marshal signer rotation receipt: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func signerRotationApprovalMessage(
	receipt SignerRotationReceiptOutput,
	receiptDigest string,
	approvalRole string,
	signerID string,
	keyID string,
	approvedAt time.Time,
) string {
	parts := []string{
		"0ai-assurance-network/signer-rotation-approval/v1",
		receipt.Version,
		receipt.ChainID,
		receipt.PolicyVersion,
		receipt.IdentityVersion,
		receipt.ReceiptID,
		receiptDigest,
		receipt.OutgoingSigner.SignerID,
		receipt.IncomingSigner.SignerID,
		receipt.IncomingSigner.KeyID,
		receipt.EffectiveAt,
		approvalRole,
		signerID,
		keyID,
		approvedAt.UTC().Format(time.RFC3339),
		receipt.ReplacementManifestRef,
	}
	return joinStrings(parts, "\n")
}

func buildReceiptRequestFromOutput(receipt SignerRotationReceiptOutput) SignerRotationReceiptRequest {
	return SignerRotationReceiptRequest{
		OutgoingSignerID:      receipt.OutgoingSigner.SignerID,
		IncomingSignerID:      receipt.IncomingSigner.SignerID,
		IncomingKeyID:         receipt.IncomingSigner.KeyID,
		IncomingActorID:       receipt.IncomingSigner.ActorID,
		IncomingRoles:         append([]string(nil), receipt.IncomingSigner.Roles...),
		IncomingProvisionedAt: receipt.IncomingSigner.ProvisionedAt,
		IncomingRotateBy:      receipt.IncomingSigner.RotateBy,
		EffectiveAt:           receipt.EffectiveAt,
		ReceiptID:             receipt.ReceiptID,
	}
}

func ensureReceiptMatchesBundle(bundle Bundle, receipt SignerRotationReceiptOutput) (SignerRotationReceiptOutput, string, error) {
	expected, err := SignerRotationReceipt(bundle, buildReceiptRequestFromOutput(receipt))
	if err != nil {
		return SignerRotationReceiptOutput{}, "", err
	}
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		return SignerRotationReceiptOutput{}, "", fmt.Errorf("marshal expected signer rotation receipt: %w", err)
	}
	actualJSON, err := json.Marshal(receipt)
	if err != nil {
		return SignerRotationReceiptOutput{}, "", fmt.Errorf("marshal actual signer rotation receipt: %w", err)
	}
	if !bytesEqual(expectedJSON, actualJSON) {
		return SignerRotationReceiptOutput{}, "", fmt.Errorf("signer rotation receipt drift detected")
	}
	receiptDigest, err := signerRotationReceiptDigest(expected)
	if err != nil {
		return SignerRotationReceiptOutput{}, "", err
	}
	return expected, receiptDigest, nil
}

func bytesEqual(a []byte, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for idx := range a {
		if a[idx] != b[idx] {
			return false
		}
	}
	return true
}

func signerPolicyEntry(bundle Bundle, signerID string, keyID string) (map[string]any, error) {
	rawSigners, ok := bundle.CheckpointSigners["signers"].([]any)
	if !ok {
		return nil, fmt.Errorf("checkpoint signer list must be configured")
	}
	for _, rawSigner := range rawSigners {
		signer := stringMap(rawSigner)
		if fmt.Sprint(signer["signer_id"]) == signerID && fmt.Sprint(signer["key_id"]) == keyID {
			return signer, nil
		}
	}
	return nil, fmt.Errorf("unknown signer_id/key_id pair: %s/%s", signerID, keyID)
}

func findKeyIDForSigner(bundle Bundle, signerID string) string {
	rawSigners, ok := bundle.CheckpointSigners["signers"].([]any)
	if !ok {
		return ""
	}
	for _, rawSigner := range rawSigners {
		signer := stringMap(rawSigner)
		if fmt.Sprint(signer["signer_id"]) == signerID {
			return fmt.Sprint(signer["key_id"])
		}
	}
	return ""
}

func eligibleApprovalOption(
	requirements []SignerRotationApprovalRequirement,
	role string,
	signerID string,
) (SignerRotationApprovalOption, error) {
	for _, requirement := range requirements {
		if requirement.Role != role {
			continue
		}
		for _, option := range requirement.EligibleSigners {
			if option.SignerID == signerID {
				return option, nil
			}
		}
		return SignerRotationApprovalOption{}, fmt.Errorf("signer %s is not eligible to approve role %s", signerID, role)
	}
	return SignerRotationApprovalOption{}, fmt.Errorf("unknown approval role: %s", role)
}

func signerRotationApprovalExpiry(
	signerRotateBy time.Time,
	effectiveAt time.Time,
	approvedAt time.Time,
	maxValiditySeconds int,
) time.Time {
	expiresAt := approvedAt.Add(time.Duration(maxValiditySeconds) * time.Second)
	if expiresAt.After(signerRotateBy) {
		expiresAt = signerRotateBy
	}
	if expiresAt.After(effectiveAt) {
		expiresAt = effectiveAt
	}
	return expiresAt.UTC()
}

func verifySignerRotationApproval(
	bundle Bundle,
	receipt SignerRotationReceiptOutput,
	receiptDigest string,
	approval SignerRotationApproval,
	seenSignatureIDs map[string]struct{},
) (SignerRotationApproval, error) {
	if approval.ReceiptID != receipt.ReceiptID {
		return SignerRotationApproval{}, fmt.Errorf("approval receipt_id mismatch for role %s", approval.ApprovalRole)
	}
	if approval.ReceiptDigest != receiptDigest {
		return SignerRotationApproval{}, fmt.Errorf("approval receipt_digest mismatch for role %s", approval.ApprovalRole)
	}
	if approval.ReplacementManifestRef != receipt.ReplacementManifestRef {
		return SignerRotationApproval{}, fmt.Errorf("approval replacement_manifest_ref mismatch for role %s", approval.ApprovalRole)
	}
	option, err := eligibleApprovalOption(receipt.ApprovalRequirements, approval.ApprovalRole, approval.SignerID)
	if err != nil {
		return SignerRotationApproval{}, err
	}
	signerEntry, err := signerPolicyEntry(bundle, approval.SignerID, approval.KeyID)
	if err != nil {
		return SignerRotationApproval{}, err
	}
	if fmt.Sprint(signerEntry["status"]) != "active" {
		return SignerRotationApproval{}, fmt.Errorf("approval signer %s must be active", approval.SignerID)
	}
	approvedAt, err := parseRFC3339(approval.ApprovedAt, "approval approved_at")
	if err != nil {
		return SignerRotationApproval{}, err
	}
	signedAt, err := parseRFC3339(approval.Signature.SignedAt, "approval signature signed_at")
	if err != nil {
		return SignerRotationApproval{}, err
	}
	expiresAt, err := parseRFC3339(approval.Signature.ExpiresAt, "approval signature expires_at")
	if err != nil {
		return SignerRotationApproval{}, err
	}
	if !approvedAt.Equal(signedAt) {
		return SignerRotationApproval{}, fmt.Errorf("approval signed_at must equal approved_at for role %s", approval.ApprovalRole)
	}
	effectiveAt, err := parseRFC3339(receipt.EffectiveAt, "receipt effective_at")
	if err != nil {
		return SignerRotationApproval{}, err
	}
	provisionedAt, err := parseRFC3339(signerEntry["provisioned_at"], "checkpoint signer provisioned_at")
	if err != nil {
		return SignerRotationApproval{}, err
	}
	rotateBy, err := parseRFC3339(signerEntry["rotate_by"], "checkpoint signer rotate_by")
	if err != nil {
		return SignerRotationApproval{}, err
	}
	if approvedAt.Before(provisionedAt) {
		return SignerRotationApproval{}, fmt.Errorf("approval approved_at must be on or after signer provisioned_at for role %s", approval.ApprovalRole)
	}
	if approvedAt.After(rotateBy) {
		return SignerRotationApproval{}, fmt.Errorf("approval approved_at must be on or before signer rotate_by for role %s", approval.ApprovalRole)
	}
	if approvedAt.After(effectiveAt) {
		return SignerRotationApproval{}, fmt.Errorf("approval approved_at must be on or before receipt effective_at for role %s", approval.ApprovalRole)
	}
	maxValiditySeconds, err := intValue(bundle.CheckpointSigners["maximum_signature_validity_seconds"])
	if err != nil || maxValiditySeconds <= 0 {
		return SignerRotationApproval{}, fmt.Errorf("checkpoint signer validity window must be positive")
	}
	expectedExpiry := signerRotationApprovalExpiry(rotateBy, effectiveAt, approvedAt, maxValiditySeconds)
	if !expiresAt.Equal(expectedExpiry) {
		return SignerRotationApproval{}, fmt.Errorf("approval signature expires_at mismatch for role %s", approval.ApprovalRole)
	}
	if approval.Signature.Format != fmt.Sprint(bundle.CheckpointSigners["signature_format"]) {
		return SignerRotationApproval{}, fmt.Errorf("approval signature format mismatch for role %s", approval.ApprovalRole)
	}
	if approval.Signature.SignatureID == "" {
		return SignerRotationApproval{}, fmt.Errorf("approval signature_id must be set for role %s", approval.ApprovalRole)
	}
	if _, exists := seenSignatureIDs[approval.Signature.SignatureID]; exists {
		return SignerRotationApproval{}, fmt.Errorf("duplicate approval signature_id: %s", approval.Signature.SignatureID)
	}
	if approval.ActorID != option.ActorID {
		return SignerRotationApproval{}, fmt.Errorf("approval actor mismatch for role %s", approval.ApprovalRole)
	}
	if approval.ActorDisplayName != option.ActorDisplayName {
		return SignerRotationApproval{}, fmt.Errorf("approval actor display name mismatch for role %s", approval.ApprovalRole)
	}
	if approval.OrganizationID != option.OrganizationID || approval.OrganizationName != option.OrganizationName {
		return SignerRotationApproval{}, fmt.Errorf("approval organization mismatch for role %s", approval.ApprovalRole)
	}

	signingMessage := signerRotationApprovalMessage(
		receipt,
		receiptDigest,
		approval.ApprovalRole,
		approval.SignerID,
		approval.KeyID,
		approvedAt,
	)
	expectedSignature := hmac.New(sha256.New, []byte(fmt.Sprint(signerEntry["shared_secret"])))
	expectedSignature.Write([]byte(signingMessage))
	expectedValue := hex.EncodeToString(expectedSignature.Sum(nil))
	if !hmac.Equal([]byte(expectedValue), []byte(approval.Signature.Value)) {
		return SignerRotationApproval{}, fmt.Errorf("approval signature verification failed for role %s", approval.ApprovalRole)
	}
	seenSignatureIDs[approval.Signature.SignatureID] = struct{}{}
	return approval, nil
}

func finalizedAt(approvals []SignerRotationApproval) (string, error) {
	latest := time.Time{}
	for _, approval := range approvals {
		approvedAt, err := parseRFC3339(approval.ApprovedAt, "approval approved_at")
		if err != nil {
			return "", err
		}
		if approvedAt.After(latest) {
			latest = approvedAt
		}
	}
	if latest.IsZero() {
		return "", fmt.Errorf("at least one approval is required")
	}
	return latest.Format(time.RFC3339), nil
}

func buildFinalizeRequestFromBundle(bundle SignerRotationFinalizedBundle) SignerRotationFinalizeRequest {
	return SignerRotationFinalizeRequest{
		Receipt: SignerRotationReceiptOutput{
			Version:                   bundle.Version,
			ChainID:                   bundle.ChainID,
			PolicyVersion:             bundle.PolicyVersion,
			IdentityVersion:           bundle.IdentityVersion,
			ReceiptID:                 bundle.ReceiptID,
			EffectiveAt:               bundle.EffectiveAt,
			OutgoingSigner:            bundle.OutgoingSigner,
			IncomingSigner:            bundle.IncomingSigner,
			ApprovalRequirements:      append([]SignerRotationApprovalRequirement(nil), bundle.ApprovalRequirements...),
			CoverageStatus:            "ready",
			MissingRoleCoverage:       append([]string(nil), bundle.MissingRoleCoverage...),
			ReplacementManifestRef:    bundle.ReplacementManifestRef,
			ReplacementSignerManifest: bundle.ReplacementSignerManifest,
		},
		Approvals: append([]SignerRotationApproval(nil), bundle.Approvals...),
	}
}

func ensureFinalizedBundleMatchesBundle(bundle Bundle, finalized SignerRotationFinalizedBundle) (SignerRotationFinalizedBundle, error) {
	request := buildFinalizeRequestFromBundle(finalized)
	expectedReceipt, err := SignerRotationReceipt(bundle, buildReceiptRequestFromOutput(request.Receipt))
	if err != nil {
		return SignerRotationFinalizedBundle{}, err
	}
	expected, err := SignerRotationFinalize(bundle, SignerRotationFinalizeRequest{
		Receipt:   expectedReceipt,
		Approvals: request.Approvals,
	})
	if err != nil {
		return SignerRotationFinalizedBundle{}, fmt.Errorf("signer rotation finalized bundle drift detected: %w", err)
	}
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		return SignerRotationFinalizedBundle{}, fmt.Errorf("marshal expected signer rotation finalized bundle: %w", err)
	}
	actualJSON, err := json.Marshal(finalized)
	if err != nil {
		return SignerRotationFinalizedBundle{}, fmt.Errorf("marshal actual signer rotation finalized bundle: %w", err)
	}
	if !bytesEqual(expectedJSON, actualJSON) {
		return SignerRotationFinalizedBundle{}, fmt.Errorf("signer rotation finalized bundle drift detected")
	}
	return expected, nil
}

func currentSignerPolicy(bundle Bundle) (CheckpointSignerPolicyOutput, error) {
	rotationPolicy := stringMap(bundle.CheckpointSigners["rotation_policy"])
	warningWindowDays, err := intValue(rotationPolicy["warning_window_days"])
	if err != nil {
		return CheckpointSignerPolicyOutput{}, fmt.Errorf("checkpoint signer rotation_policy.warning_window_days must be positive")
	}
	approvalRoles, err := governanceApprovalRoles(bundle)
	if err != nil {
		return CheckpointSignerPolicyOutput{}, err
	}
	rawSigners, ok := bundle.CheckpointSigners["signers"].([]any)
	if !ok {
		return CheckpointSignerPolicyOutput{}, fmt.Errorf("checkpoint signer list must be configured")
	}
	signers := make([]CheckpointSignerPolicySigner, 0, len(rawSigners))
	for _, rawSigner := range rawSigners {
		signer := stringMap(rawSigner)
		signers = append(signers, CheckpointSignerPolicySigner{
			ActorID:       fmt.Sprint(signer["actor_id"]),
			SignerID:      fmt.Sprint(signer["signer_id"]),
			KeyID:         fmt.Sprint(signer["key_id"]),
			Status:        fmt.Sprint(signer["status"]),
			ProvisionedAt: fmt.Sprint(signer["provisioned_at"]),
			RotateBy:      fmt.Sprint(signer["rotate_by"]),
			Roles:         append([]string(nil), stringSlice(signer["roles"])...),
			SharedSecret:  fmt.Sprint(signer["shared_secret"]),
		})
	}
	sort.Slice(signers, func(i, j int) bool {
		return signers[i].SignerID < signers[j].SignerID
	})
	return CheckpointSignerPolicyOutput{
		Version:                       fmt.Sprint(bundle.CheckpointSigners["version"]),
		SignatureFormat:               fmt.Sprint(bundle.CheckpointSigners["signature_format"]),
		RequireSignaturesForEventLogs: fmt.Sprint(bundle.CheckpointSigners["require_signatures_for_event_logs"]) == "true",
		MaximumSignatureValidity:      maxIntValue(bundle.CheckpointSigners["maximum_signature_validity_seconds"]),
		RotationPolicy: CheckpointSignerPolicyRotationPolicy{
			ReferenceTime:     fmt.Sprint(rotationPolicy["reference_time"]),
			WarningWindowDays: warningWindowDays,
			ApprovalRoles:     append([]string(nil), approvalRoles...),
		},
		Signers: signers,
	}, nil
}

func maxIntValue(value any) int {
	parsed, _ := intValue(value)
	return parsed
}

func activationPolicyVersion(currentVersion string, receiptID string) string {
	return fmt.Sprintf("%s+%s", currentVersion, receiptID)
}

func policySignerFromReplacement(replacement SignerRotationReplacement, sharedSecret string) CheckpointSignerPolicySigner {
	return CheckpointSignerPolicySigner{
		ActorID:       replacement.ActorID,
		SignerID:      replacement.SignerID,
		KeyID:         replacement.KeyID,
		Status:        "active",
		ProvisionedAt: replacement.ProvisionedAt,
		RotateBy:      replacement.RotateBy,
		Roles:         append([]string(nil), replacement.Roles...),
		SharedSecret:  sharedSecret,
	}
}

func policyOutputToCheckpointSignerMap(policy CheckpointSignerPolicyOutput) (map[string]any, error) {
	encoded, err := json.Marshal(policy)
	if err != nil {
		return nil, fmt.Errorf("marshal checkpoint signer policy output: %w", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		return nil, fmt.Errorf("decode checkpoint signer policy output: %w", err)
	}
	return decoded, nil
}

func activationPlanDigest(plan SignerRotationActivationPlan) (string, error) {
	encoded, err := json.Marshal(plan)
	if err != nil {
		return "", fmt.Errorf("marshal signer rotation activation plan: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func checkpointSignerPolicyDigest(policy CheckpointSignerPolicyOutput) (string, error) {
	encoded, err := json.Marshal(policy)
	if err != nil {
		return "", fmt.Errorf("marshal checkpoint signer policy: %w", err)
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func digestJSON(value any, description string) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal %s: %w", description, err)
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func deterministicArtifactID(prefix string, values ...string) string {
	joined := strings.Join(values, "|")
	sum := sha256.Sum256([]byte(joined))
	digest := hex.EncodeToString(sum[:])
	if len(digest) > 16 {
		digest = digest[:16]
	}
	return prefix + "-" + digest
}

func signerRotationVerificationMessage(
	plan SignerRotationActivationPlan,
	planDigest string,
	policyDigest string,
	policyPath string,
	signerID string,
	keyID string,
	verifiedAt time.Time,
) string {
	parts := []string{
		"0ai-assurance-network/signer-rotation-verification/v1",
		plan.Version,
		plan.ChainID,
		plan.ReceiptID,
		planDigest,
		policyDigest,
		policyPath,
		plan.TargetPolicyVersion,
		signerID,
		keyID,
		verifiedAt.UTC().Format(time.RFC3339),
	}
	return joinStrings(parts, "\n")
}

func signerRotationVerificationExpiry(
	signerRotateBy time.Time,
	verifiedAt time.Time,
	maxValiditySeconds int,
) time.Time {
	expiresAt := verifiedAt.Add(time.Duration(maxValiditySeconds) * time.Second)
	if expiresAt.After(signerRotateBy) {
		expiresAt = signerRotateBy
	}
	return expiresAt.UTC()
}

func policySignerByID(policy CheckpointSignerPolicyOutput, signerID string) (CheckpointSignerPolicySigner, bool) {
	for _, signer := range policy.Signers {
		if signer.SignerID == signerID {
			return signer, true
		}
	}
	return CheckpointSignerPolicySigner{}, false
}

func actorContext(bundle Bundle, actorID string) (IdentityActor, string) {
	actors, _ := activeActorsAndBindings(bundle)
	actor := actors[actorID]
	organizationName := ""
	if actor.OrganizationID != "" {
		if organization, exists := actors[actor.OrganizationID]; exists {
			organizationName = organization.DisplayName
		}
	}
	return actor, organizationName
}

func ensureActivationPlanMatchesBundle(bundle Bundle, plan SignerRotationActivationPlan) (SignerRotationActivationPlan, error) {
	currentPolicy, err := currentSignerPolicy(bundle)
	if err != nil {
		return SignerRotationActivationPlan{}, err
	}
	if plan.PolicyPath != "config/governance/checkpoint-signers.json" {
		return SignerRotationActivationPlan{}, fmt.Errorf("unexpected activation policy path: %s", plan.PolicyPath)
	}
	if plan.CurrentPolicyVersion != currentPolicy.Version || plan.PolicyPatch.PolicyVersionFrom != currentPolicy.Version {
		return SignerRotationActivationPlan{}, fmt.Errorf("signer rotation activation plan drift detected")
	}
	if plan.PolicyPatch.ReferenceTimeFrom != currentPolicy.RotationPolicy.ReferenceTime {
		return SignerRotationActivationPlan{}, fmt.Errorf("signer rotation activation plan drift detected")
	}
	if strings.TrimSpace(plan.PolicyPatch.AddSigner.SharedSecret) == "" {
		return SignerRotationActivationPlan{}, fmt.Errorf("activation plan add_signer.shared_secret must be set")
	}
	signers := make([]CheckpointSignerPolicySigner, 0, len(currentPolicy.Signers))
	foundOutgoing := false
	foundIncoming := false
	for _, signer := range currentPolicy.Signers {
		if signer.SignerID == plan.PolicyPatch.RemoveSignerID {
			foundOutgoing = true
			continue
		}
		if signer.SignerID == plan.PolicyPatch.AddSigner.SignerID {
			foundIncoming = true
		}
		signers = append(signers, signer)
	}
	if !foundOutgoing || foundIncoming {
		return SignerRotationActivationPlan{}, fmt.Errorf("signer rotation activation plan drift detected")
	}
	signers = append(signers, plan.PolicyPatch.AddSigner)
	sort.Slice(signers, func(i, j int) bool {
		return signers[i].SignerID < signers[j].SignerID
	})
	expectedPolicy := currentPolicy
	expectedPolicy.Version = plan.PolicyPatch.PolicyVersionTo
	expectedPolicy.RotationPolicy.ReferenceTime = plan.PolicyPatch.ReferenceTimeTo
	expectedPolicy.Signers = signers
	expectedJSON, err := json.Marshal(expectedPolicy)
	if err != nil {
		return SignerRotationActivationPlan{}, fmt.Errorf("marshal expected signer rotation activation policy: %w", err)
	}
	actualJSON, err := json.Marshal(plan.ResultingPolicy)
	if err != nil {
		return SignerRotationActivationPlan{}, fmt.Errorf("marshal actual signer rotation activation policy: %w", err)
	}
	if !bytesEqual(expectedJSON, actualJSON) {
		return SignerRotationActivationPlan{}, fmt.Errorf("signer rotation activation plan drift detected")
	}
	validationMap, err := policyOutputToCheckpointSignerMap(expectedPolicy)
	if err != nil {
		return SignerRotationActivationPlan{}, err
	}
	validationBundle := bundle
	validationBundle.CheckpointSigners = validationMap
	if err := ValidateSignerManifestInputs(validationBundle); err != nil {
		return SignerRotationActivationPlan{}, err
	}
	return plan, nil
}

func SignerRotationReceipt(bundle Bundle, request SignerRotationReceiptRequest) (SignerRotationReceiptOutput, error) {
	if err := ValidateSignerManifestInputs(bundle); err != nil {
		return SignerRotationReceiptOutput{}, err
	}

	rotationPolicy := stringMap(bundle.CheckpointSigners["rotation_policy"])
	referenceTime, _ := parseRFC3339(rotationPolicy["reference_time"], "checkpoint signer rotation_policy.reference_time")
	warningWindowDays, _ := intValue(rotationPolicy["warning_window_days"])
	parsedSigners, err := parseCurrentSigners(bundle, referenceTime, warningWindowDays)
	if err != nil {
		return SignerRotationReceiptOutput{}, err
	}
	requiredRoles := governanceRequiredRoles(bundle)

	outgoing, exists := findParsedSignerByID(parsedSigners, request.OutgoingSignerID)
	if !exists {
		return SignerRotationReceiptOutput{}, fmt.Errorf("unknown outgoing signer id: %s", request.OutgoingSignerID)
	}
	if outgoing.Status != "active" {
		return SignerRotationReceiptOutput{}, fmt.Errorf("outgoing signer %s must be active", outgoing.SignerID)
	}
	if request.IncomingSignerID == "" {
		return SignerRotationReceiptOutput{}, fmt.Errorf("incoming signer id must be set")
	}
	if request.IncomingKeyID == "" {
		return SignerRotationReceiptOutput{}, fmt.Errorf("incoming key id must be set")
	}
	if request.IncomingSignerID == outgoing.SignerID {
		return SignerRotationReceiptOutput{}, fmt.Errorf("incoming signer id must differ from outgoing signer id")
	}

	incomingActorID := request.IncomingActorID
	if incomingActorID == "" {
		incomingActorID = outgoing.ActorID
	}
	incomingRoles := append([]string(nil), request.IncomingRoles...)
	if len(incomingRoles) == 0 {
		incomingRoles = append([]string(nil), outgoing.Roles...)
	}
	sort.Strings(incomingRoles)

	actors, activeRoleBindings := activeActorsAndBindings(bundle)
	incomingActor, exists := actors[incomingActorID]
	if !exists {
		return SignerRotationReceiptOutput{}, fmt.Errorf("incoming signer references unknown actor %s", incomingActorID)
	}
	if incomingActor.Status != "active" {
		return SignerRotationReceiptOutput{}, fmt.Errorf("incoming signer references inactive actor %s", incomingActorID)
	}
	for _, role := range incomingRoles {
		if _, exists := activeRoleBindings[incomingActorID+"|"+role]; !exists {
			return SignerRotationReceiptOutput{}, fmt.Errorf(
				"incoming signer role %s is not backed by an active identity binding",
				role,
			)
		}
	}

	effectiveAt, err := parseRFC3339(request.EffectiveAt, "signer rotation effective_at")
	if err != nil {
		return SignerRotationReceiptOutput{}, err
	}
	if !effectiveAt.After(referenceTime) {
		return SignerRotationReceiptOutput{}, fmt.Errorf("signer rotation effective_at must be after the rotation reference time")
	}
	if effectiveAt.After(outgoing.RotateBy) {
		return SignerRotationReceiptOutput{}, fmt.Errorf("signer rotation effective_at must be on or before the outgoing signer rotate_by time")
	}

	incomingProvisionedAt := referenceTime
	if request.IncomingProvisionedAt != "" {
		incomingProvisionedAt, err = parseRFC3339(request.IncomingProvisionedAt, "incoming signer provisioned_at")
		if err != nil {
			return SignerRotationReceiptOutput{}, err
		}
	}
	if incomingProvisionedAt.After(effectiveAt) {
		return SignerRotationReceiptOutput{}, fmt.Errorf("incoming signer provisioned_at must be on or before effective_at")
	}

	incomingRotateBy := effectiveAt.Add(90 * 24 * time.Hour)
	if request.IncomingRotateBy != "" {
		incomingRotateBy, err = parseRFC3339(request.IncomingRotateBy, "incoming signer rotate_by")
		if err != nil {
			return SignerRotationReceiptOutput{}, err
		}
	}
	if !incomingRotateBy.After(effectiveAt) {
		return SignerRotationReceiptOutput{}, fmt.Errorf("incoming signer rotate_by must be after effective_at")
	}

	receiptID := request.ReceiptID
	replacementManifestRef := ""
	if receiptID == "" {
		receiptID, replacementManifestRef = signerReceiptArtifacts(outgoing.SignerID, effectiveAt)
	} else {
		replacementManifestRef = fmt.Sprintf("build/rotation/%s/replacement-signer-manifest.json", receiptID)
	}

	incoming := parsedSignerEntry{
		ActorID:           incomingActorID,
		SignerID:          request.IncomingSignerID,
		KeyID:             request.IncomingKeyID,
		Status:            "active",
		Roles:             append([]string(nil), incomingRoles...),
		ProvisionedAt:     incomingProvisionedAt,
		RotateBy:          incomingRotateBy,
		RotationStatus:    signerRotationStatus(referenceTime, warningWindowDays, "active", incomingRotateBy),
		DaysUntilRotation: int(incomingRotateBy.Sub(referenceTime).Hours() / 24),
	}
	replacement := replacementSigners(parsedSigners, outgoing.SignerID, incoming)
	if err := validateSignerSet(bundle, replacement, referenceTime, warningWindowDays, requiredRoles); err != nil {
		return SignerRotationReceiptOutput{}, err
	}

	previewManifest := buildSignerManifestOutput(bundle, replacement, referenceTime, warningWindowDays)
	approvalReqs, err := approvalRequirements(bundle, parsedSigners)
	if err != nil {
		return SignerRotationReceiptOutput{}, err
	}

	roleUsage := governanceExecutionRoleUsage(bundle)
	outgoingEntry := buildSignerManifestEntry(actors, roleUsage, outgoing)

	incomingOrganizationName := ""
	if incomingActor.OrganizationID != "" {
		if organization, exists := actors[incomingActor.OrganizationID]; exists {
			incomingOrganizationName = organization.DisplayName
		}
	}

	return SignerRotationReceiptOutput{
		Version:         "1.0.0",
		ChainID:         bundle.Topology.ChainID,
		PolicyVersion:   fmt.Sprint(bundle.CheckpointSigners["version"]),
		IdentityVersion: bundle.Identity.Version,
		ReceiptID:       receiptID,
		EffectiveAt:     effectiveAt.Format(time.RFC3339),
		OutgoingSigner:  outgoingEntry,
		IncomingSigner: SignerRotationReplacement{
			ActorID:                incomingActorID,
			ActorDisplayName:       incomingActor.DisplayName,
			OrganizationID:         incomingActor.OrganizationID,
			OrganizationName:       incomingOrganizationName,
			SignerID:               incoming.SignerID,
			KeyID:                  incoming.KeyID,
			Roles:                  append([]string(nil), incoming.Roles...),
			ProvisionedAt:          incoming.ProvisionedAt.Format(time.RFC3339),
			RotateBy:               incoming.RotateBy.Format(time.RFC3339),
			EffectiveAt:            effectiveAt.Format(time.RFC3339),
			RotationReceiptID:      receiptID,
			ReplacementManifestRef: replacementManifestRef,
		},
		ApprovalRequirements:      approvalReqs,
		CoverageStatus:            "ready",
		MissingRoleCoverage:       []string{},
		ReplacementManifestRef:    replacementManifestRef,
		ReplacementSignerManifest: previewManifest,
	}, nil
}

func GenerateSignerRotationApproval(bundle Bundle, request SignerRotationApprovalRequest) (SignerRotationApproval, error) {
	if err := ValidateSignerManifestInputs(bundle); err != nil {
		return SignerRotationApproval{}, err
	}
	receipt, receiptDigest, err := ensureReceiptMatchesBundle(bundle, request.Receipt)
	if err != nil {
		return SignerRotationApproval{}, err
	}
	if strings.TrimSpace(request.ApprovalRole) == "" {
		return SignerRotationApproval{}, fmt.Errorf("approval role must be set")
	}
	if strings.TrimSpace(request.SignerID) == "" {
		return SignerRotationApproval{}, fmt.Errorf("approval signer id must be set")
	}
	option, err := eligibleApprovalOption(receipt.ApprovalRequirements, request.ApprovalRole, request.SignerID)
	if err != nil {
		return SignerRotationApproval{}, err
	}
	signerEntry, err := signerPolicyEntry(bundle, request.SignerID, findKeyIDForSigner(bundle, request.SignerID))
	if err != nil {
		return SignerRotationApproval{}, err
	}
	approvedAt, err := parseRFC3339(request.ApprovedAt, "approval approved_at")
	if err != nil {
		return SignerRotationApproval{}, err
	}
	effectiveAt, err := parseRFC3339(receipt.EffectiveAt, "receipt effective_at")
	if err != nil {
		return SignerRotationApproval{}, err
	}
	provisionedAt, err := parseRFC3339(signerEntry["provisioned_at"], "checkpoint signer provisioned_at")
	if err != nil {
		return SignerRotationApproval{}, err
	}
	rotateBy, err := parseRFC3339(signerEntry["rotate_by"], "checkpoint signer rotate_by")
	if err != nil {
		return SignerRotationApproval{}, err
	}
	if approvedAt.Before(provisionedAt) {
		return SignerRotationApproval{}, fmt.Errorf("approval approved_at must be on or after signer provisioned_at for role %s", request.ApprovalRole)
	}
	if approvedAt.After(rotateBy) {
		return SignerRotationApproval{}, fmt.Errorf("approval approved_at must be on or before signer rotate_by for role %s", request.ApprovalRole)
	}
	if approvedAt.After(effectiveAt) {
		return SignerRotationApproval{}, fmt.Errorf("approval approved_at must be on or before receipt effective_at for role %s", request.ApprovalRole)
	}
	maxValiditySeconds, err := intValue(bundle.CheckpointSigners["maximum_signature_validity_seconds"])
	if err != nil || maxValiditySeconds <= 0 {
		return SignerRotationApproval{}, fmt.Errorf("checkpoint signer validity window must be positive")
	}
	signatureID := request.SignatureID
	if signatureID == "" {
		signatureID = fmt.Sprintf(
			"approve-%s-%s-%s",
			receipt.ReceiptID,
			request.ApprovalRole,
			strings.ToLower(approvedAt.UTC().Format("20060102t150405z")),
		)
	}
	expiresAt := signerRotationApprovalExpiry(rotateBy, effectiveAt, approvedAt, maxValiditySeconds)
	signingMessage := signerRotationApprovalMessage(
		receipt,
		receiptDigest,
		request.ApprovalRole,
		fmt.Sprint(signerEntry["signer_id"]),
		fmt.Sprint(signerEntry["key_id"]),
		approvedAt,
	)
	mac := hmac.New(sha256.New, []byte(fmt.Sprint(signerEntry["shared_secret"])))
	mac.Write([]byte(signingMessage))
	organizationID := option.OrganizationID
	organizationName := option.OrganizationName
	return SignerRotationApproval{
		Version:                "1.0.0",
		ReceiptID:              receipt.ReceiptID,
		ReceiptDigest:          receiptDigest,
		ApprovalRole:           request.ApprovalRole,
		ApprovedAt:             approvedAt.Format(time.RFC3339),
		SignerID:               fmt.Sprint(signerEntry["signer_id"]),
		KeyID:                  fmt.Sprint(signerEntry["key_id"]),
		ActorID:                option.ActorID,
		ActorDisplayName:       option.ActorDisplayName,
		OrganizationID:         organizationID,
		OrganizationName:       organizationName,
		ReplacementManifestRef: receipt.ReplacementManifestRef,
		Signature: SignerRotationApprovalSignature{
			Format:      fmt.Sprint(bundle.CheckpointSigners["signature_format"]),
			SignatureID: signatureID,
			SignedAt:    approvedAt.Format(time.RFC3339),
			ExpiresAt:   expiresAt.Format(time.RFC3339),
			Value:       hex.EncodeToString(mac.Sum(nil)),
		},
	}, nil
}

func SignerRotationFinalize(bundle Bundle, request SignerRotationFinalizeRequest) (SignerRotationFinalizedBundle, error) {
	if err := ValidateSignerManifestInputs(bundle); err != nil {
		return SignerRotationFinalizedBundle{}, err
	}
	if len(request.Approvals) == 0 {
		return SignerRotationFinalizedBundle{}, fmt.Errorf("at least one approval artifact is required")
	}
	receipt, receiptDigest, err := ensureReceiptMatchesBundle(bundle, request.Receipt)
	if err != nil {
		return SignerRotationFinalizedBundle{}, err
	}
	requiredRoles := make(map[string]struct{}, len(receipt.ApprovalRequirements))
	for _, requirement := range receipt.ApprovalRequirements {
		requiredRoles[requirement.Role] = struct{}{}
	}
	seenRoles := make(map[string]struct{}, len(requiredRoles))
	seenSigners := make(map[string]struct{}, len(request.Approvals))
	seenActors := make(map[string]struct{}, len(request.Approvals))
	seenSignatureIDs := make(map[string]struct{}, len(request.Approvals))
	verifiedApprovals := make([]SignerRotationApproval, 0, len(request.Approvals))
	for _, approval := range request.Approvals {
		if _, exists := requiredRoles[approval.ApprovalRole]; !exists {
			return SignerRotationFinalizedBundle{}, fmt.Errorf("unexpected approval role: %s", approval.ApprovalRole)
		}
		if _, exists := seenRoles[approval.ApprovalRole]; exists {
			return SignerRotationFinalizedBundle{}, fmt.Errorf("duplicate approval role: %s", approval.ApprovalRole)
		}
		if _, exists := seenSigners[approval.SignerID]; exists {
			return SignerRotationFinalizedBundle{}, fmt.Errorf("duplicate approval signer: %s", approval.SignerID)
		}
		if _, exists := seenActors[approval.ActorID]; exists {
			return SignerRotationFinalizedBundle{}, fmt.Errorf("duplicate approval actor ownership: %s", approval.ActorID)
		}
		verified, err := verifySignerRotationApproval(bundle, receipt, receiptDigest, approval, seenSignatureIDs)
		if err != nil {
			return SignerRotationFinalizedBundle{}, err
		}
		seenRoles[verified.ApprovalRole] = struct{}{}
		seenSigners[verified.SignerID] = struct{}{}
		seenActors[verified.ActorID] = struct{}{}
		verifiedApprovals = append(verifiedApprovals, verified)
	}
	missingRoles := make([]string, 0)
	for role := range requiredRoles {
		if _, exists := seenRoles[role]; !exists {
			missingRoles = append(missingRoles, role)
		}
	}
	sort.Strings(missingRoles)
	if len(missingRoles) > 0 {
		return SignerRotationFinalizedBundle{}, fmt.Errorf("missing approval coverage for roles: %s", commaList(missingRoles))
	}
	sort.Slice(verifiedApprovals, func(i, j int) bool {
		if verifiedApprovals[i].ApprovalRole == verifiedApprovals[j].ApprovalRole {
			return verifiedApprovals[i].SignerID < verifiedApprovals[j].SignerID
		}
		return verifiedApprovals[i].ApprovalRole < verifiedApprovals[j].ApprovalRole
	})
	finalizedAtValue, err := finalizedAt(verifiedApprovals)
	if err != nil {
		return SignerRotationFinalizedBundle{}, err
	}
	return SignerRotationFinalizedBundle{
		Version:                   "1.0.0",
		Status:                    "approved",
		ReceiptID:                 receipt.ReceiptID,
		ReceiptDigest:             receiptDigest,
		ChainID:                   receipt.ChainID,
		PolicyVersion:             receipt.PolicyVersion,
		IdentityVersion:           receipt.IdentityVersion,
		FinalizedAt:               finalizedAtValue,
		EffectiveAt:               receipt.EffectiveAt,
		OutgoingSigner:            receipt.OutgoingSigner,
		IncomingSigner:            receipt.IncomingSigner,
		ApprovalRequirements:      receipt.ApprovalRequirements,
		Approvals:                 verifiedApprovals,
		CoverageStatus:            "approved",
		MissingRoleCoverage:       []string{},
		ReplacementManifestRef:    receipt.ReplacementManifestRef,
		ReplacementSignerManifest: receipt.ReplacementSignerManifest,
	}, nil
}

func SignerRotationActivation(bundle Bundle, request SignerRotationActivationRequest) (SignerRotationActivationPlan, error) {
	if err := ValidateSignerManifestInputs(bundle); err != nil {
		return SignerRotationActivationPlan{}, err
	}
	if strings.TrimSpace(request.IncomingSharedSecret) == "" {
		return SignerRotationActivationPlan{}, fmt.Errorf("incoming shared secret must be set")
	}
	finalized, err := ensureFinalizedBundleMatchesBundle(bundle, request.FinalizedBundle)
	if err != nil {
		return SignerRotationActivationPlan{}, err
	}
	currentPolicy, err := currentSignerPolicy(bundle)
	if err != nil {
		return SignerRotationActivationPlan{}, err
	}
	signers := make([]CheckpointSignerPolicySigner, 0, len(currentPolicy.Signers))
	foundOutgoing := false
	foundIncoming := false
	for _, signer := range currentPolicy.Signers {
		if signer.SignerID == finalized.OutgoingSigner.SignerID {
			foundOutgoing = true
			continue
		}
		if signer.SignerID == finalized.IncomingSigner.SignerID {
			foundIncoming = true
		}
		signers = append(signers, signer)
	}
	if !foundOutgoing {
		return SignerRotationActivationPlan{}, fmt.Errorf("outgoing signer %s is not present in current checkpoint signer policy", finalized.OutgoingSigner.SignerID)
	}
	if foundIncoming {
		return SignerRotationActivationPlan{}, fmt.Errorf("incoming signer %s already exists in current checkpoint signer policy", finalized.IncomingSigner.SignerID)
	}

	incomingSigner := policySignerFromReplacement(finalized.IncomingSigner, request.IncomingSharedSecret)
	signers = append(signers, incomingSigner)
	sort.Slice(signers, func(i, j int) bool {
		return signers[i].SignerID < signers[j].SignerID
	})

	resultingPolicy := currentPolicy
	resultingPolicy.Version = activationPolicyVersion(currentPolicy.Version, finalized.ReceiptID)
	resultingPolicy.RotationPolicy.ReferenceTime = finalized.EffectiveAt
	resultingPolicy.Signers = signers

	updatedSignerMap, err := policyOutputToCheckpointSignerMap(resultingPolicy)
	if err != nil {
		return SignerRotationActivationPlan{}, err
	}
	validationBundle := bundle
	validationBundle.CheckpointSigners = updatedSignerMap
	if err := ValidateSignerManifestInputs(validationBundle); err != nil {
		return SignerRotationActivationPlan{}, err
	}

	steps := []string{
		"Verify the finalized signer rotation bundle remains bound to the current checkpoint signer policy lineage.",
		fmt.Sprintf("Replace signer %s with %s in config/governance/checkpoint-signers.json.", finalized.OutgoingSigner.SignerID, finalized.IncomingSigner.SignerID),
		fmt.Sprintf("Set checkpoint signer policy version to %s and advance rotation reference_time to %s.", resultingPolicy.Version, finalized.EffectiveAt),
		fmt.Sprintf("Publish the approved bundle and replacement manifest at %s before activation.", finalized.ReplacementManifestRef),
		"Reload signer-manifest validation against the patched policy before any governance replay or remediation signing resumes.",
	}

	return SignerRotationActivationPlan{
		Version:              "1.0.0",
		Status:               "ready",
		ReceiptID:            finalized.ReceiptID,
		ChainID:              finalized.ChainID,
		PolicyPath:           "config/governance/checkpoint-signers.json",
		CurrentPolicyVersion: currentPolicy.Version,
		TargetPolicyVersion:  resultingPolicy.Version,
		EffectiveAt:          finalized.EffectiveAt,
		FinalizedAt:          finalized.FinalizedAt,
		OutgoingSignerID:     finalized.OutgoingSigner.SignerID,
		IncomingSignerID:     finalized.IncomingSigner.SignerID,
		ActivationSteps:      steps,
		PolicyPatch: SignerRotationPolicyPatch{
			PolicyPath:           "config/governance/checkpoint-signers.json",
			PolicyVersionFrom:    currentPolicy.Version,
			PolicyVersionTo:      resultingPolicy.Version,
			ReferenceTimeFrom:    currentPolicy.RotationPolicy.ReferenceTime,
			ReferenceTimeTo:      resultingPolicy.RotationPolicy.ReferenceTime,
			RemoveSignerID:       finalized.OutgoingSigner.SignerID,
			AddSigner:            incomingSigner,
			ResultingSignerCount: len(resultingPolicy.Signers),
		},
		ResultingPolicy: resultingPolicy,
	}, nil
}

func SignerRotationApply(bundle Bundle, request SignerRotationApplyRequest) (SignerRotationApplyResult, error) {
	if err := ValidateSignerManifestInputs(bundle); err != nil {
		return SignerRotationApplyResult{}, err
	}
	plan, err := ensureActivationPlanMatchesBundle(bundle, request.ActivationPlan)
	if err != nil {
		return SignerRotationApplyResult{}, err
	}
	planDigest, err := activationPlanDigest(plan)
	if err != nil {
		return SignerRotationApplyResult{}, err
	}
	policyDigest, err := checkpointSignerPolicyDigest(plan.ResultingPolicy)
	if err != nil {
		return SignerRotationApplyResult{}, err
	}
	return SignerRotationApplyResult{
		Version:              "1.0.0",
		Status:               "applied",
		ReceiptID:            plan.ReceiptID,
		ChainID:              plan.ChainID,
		PolicyPath:           plan.PolicyPath,
		TargetPolicyVersion:  plan.TargetPolicyVersion,
		ActivationPlanDigest: planDigest,
		AppliedPolicyDigest:  policyDigest,
		AppliedPolicy:        plan.ResultingPolicy,
		AppliedAtEffect:      plan.EffectiveAt,
	}, nil
}

func SignerRotationVerify(bundle Bundle, request SignerRotationVerifyRequest) (SignerRotationVerificationReceipt, error) {
	if err := ValidateSignerManifestInputs(bundle); err != nil {
		return SignerRotationVerificationReceipt{}, err
	}
	plan, err := ensureActivationPlanMatchesBundle(bundle, request.ActivationPlan)
	if err != nil {
		return SignerRotationVerificationReceipt{}, err
	}
	planDigest, err := activationPlanDigest(plan)
	if err != nil {
		return SignerRotationVerificationReceipt{}, err
	}
	expectedPolicyJSON, err := json.Marshal(plan.ResultingPolicy)
	if err != nil {
		return SignerRotationVerificationReceipt{}, fmt.Errorf("marshal expected activation policy: %w", err)
	}
	actualPolicyJSON, err := json.Marshal(request.Policy)
	if err != nil {
		return SignerRotationVerificationReceipt{}, fmt.Errorf("marshal verification policy: %w", err)
	}
	if !bytesEqual(expectedPolicyJSON, actualPolicyJSON) {
		return SignerRotationVerificationReceipt{}, fmt.Errorf("signer rotation applied policy drift detected")
	}
	validationMap, err := policyOutputToCheckpointSignerMap(request.Policy)
	if err != nil {
		return SignerRotationVerificationReceipt{}, err
	}
	validationBundle := bundle
	validationBundle.CheckpointSigners = validationMap
	if err := ValidateSignerManifestInputs(validationBundle); err != nil {
		return SignerRotationVerificationReceipt{}, err
	}
	if request.Policy.Version != plan.TargetPolicyVersion {
		return SignerRotationVerificationReceipt{}, fmt.Errorf("verification policy version mismatch")
	}
	if request.Policy.RotationPolicy.ReferenceTime != plan.EffectiveAt {
		return SignerRotationVerificationReceipt{}, fmt.Errorf("verification policy reference_time mismatch")
	}
	if _, exists := policySignerByID(request.Policy, plan.OutgoingSignerID); exists {
		return SignerRotationVerificationReceipt{}, fmt.Errorf("verification policy still contains outgoing signer %s", plan.OutgoingSignerID)
	}
	incomingSigner, exists := policySignerByID(request.Policy, plan.IncomingSignerID)
	if !exists {
		return SignerRotationVerificationReceipt{}, fmt.Errorf("verification policy missing incoming signer %s", plan.IncomingSignerID)
	}
	if strings.TrimSpace(incomingSigner.SharedSecret) == "" {
		return SignerRotationVerificationReceipt{}, fmt.Errorf("verification signer shared_secret must be set")
	}
	verifiedAt, err := parseRFC3339(request.VerifiedAt, "verification verified_at")
	if err != nil {
		return SignerRotationVerificationReceipt{}, err
	}
	effectiveAt, err := parseRFC3339(plan.EffectiveAt, "activation effective_at")
	if err != nil {
		return SignerRotationVerificationReceipt{}, err
	}
	provisionedAt, err := parseRFC3339(incomingSigner.ProvisionedAt, "verification signer provisioned_at")
	if err != nil {
		return SignerRotationVerificationReceipt{}, err
	}
	rotateBy, err := parseRFC3339(incomingSigner.RotateBy, "verification signer rotate_by")
	if err != nil {
		return SignerRotationVerificationReceipt{}, err
	}
	if verifiedAt.Before(effectiveAt) {
		return SignerRotationVerificationReceipt{}, fmt.Errorf("verification verified_at must be on or after activation effective_at")
	}
	if verifiedAt.Before(provisionedAt) {
		return SignerRotationVerificationReceipt{}, fmt.Errorf("verification verified_at must be on or after signer provisioned_at")
	}
	if verifiedAt.After(rotateBy) {
		return SignerRotationVerificationReceipt{}, fmt.Errorf("verification verified_at must be on or before signer rotate_by")
	}
	maxValiditySeconds := request.Policy.MaximumSignatureValidity
	if maxValiditySeconds <= 0 {
		return SignerRotationVerificationReceipt{}, fmt.Errorf("checkpoint signer validity window must be positive")
	}
	policyDigest, err := checkpointSignerPolicyDigest(request.Policy)
	if err != nil {
		return SignerRotationVerificationReceipt{}, err
	}
	signatureID := strings.TrimSpace(request.SignatureID)
	if signatureID == "" {
		signatureID = fmt.Sprintf(
			"verify-%s-%s",
			plan.ReceiptID,
			strings.ToLower(verifiedAt.UTC().Format("20060102t150405z")),
		)
	}
	signatureMessage := signerRotationVerificationMessage(
		plan,
		planDigest,
		policyDigest,
		plan.PolicyPath,
		incomingSigner.SignerID,
		incomingSigner.KeyID,
		verifiedAt,
	)
	mac := hmac.New(sha256.New, []byte(incomingSigner.SharedSecret))
	mac.Write([]byte(signatureMessage))
	expiresAt := signerRotationVerificationExpiry(rotateBy, verifiedAt, maxValiditySeconds)
	actor, organizationName := actorContext(bundle, incomingSigner.ActorID)
	return SignerRotationVerificationReceipt{
		Version:              "1.0.0",
		Status:               "verified",
		ReceiptID:            plan.ReceiptID,
		ChainID:              plan.ChainID,
		ActivationPlanDigest: planDigest,
		PolicyDigest:         policyDigest,
		VerifiedAt:           verifiedAt.Format(time.RFC3339),
		PolicyPath:           plan.PolicyPath,
		TargetPolicyVersion:  plan.TargetPolicyVersion,
		SignerID:             incomingSigner.SignerID,
		KeyID:                incomingSigner.KeyID,
		ActorID:              actor.ActorID,
		ActorDisplayName:     actor.DisplayName,
		OrganizationID:       actor.OrganizationID,
		OrganizationName:     organizationName,
		Signature: SignerRotationVerificationSignature{
			Format:      request.Policy.SignatureFormat,
			SignatureID: signatureID,
			SignedAt:    verifiedAt.Format(time.RFC3339),
			ExpiresAt:   expiresAt.Format(time.RFC3339),
			Value:       hex.EncodeToString(mac.Sum(nil)),
		},
	}, nil
}

func activationAuditEntry(
	applyResult SignerRotationApplyResult,
	verification SignerRotationVerificationReceipt,
) SignerRotationActivationAuditEntry {
	return SignerRotationActivationAuditEntry{
		Version:              "1.0.0",
		Status:               "verified",
		ReceiptID:            verification.ReceiptID,
		ChainID:              verification.ChainID,
		PolicyPath:           verification.PolicyPath,
		TargetPolicyVersion:  verification.TargetPolicyVersion,
		EffectiveAt:          applyResult.AppliedAtEffect,
		VerifiedAt:           verification.VerifiedAt,
		ActivationPlanDigest: verification.ActivationPlanDigest,
		AppliedPolicyDigest:  verification.PolicyDigest,
		SignerID:             verification.SignerID,
		KeyID:                verification.KeyID,
		ActorID:              verification.ActorID,
		ActorDisplayName:     verification.ActorDisplayName,
		OrganizationID:       verification.OrganizationID,
		OrganizationName:     verification.OrganizationName,
		Signature:            verification.Signature,
	}
}

func SignerRotationActivationAuditAppend(
	request SignerRotationActivationAuditAppendRequest,
) (SignerRotationActivationAuditAppendResult, error) {
	if request.ApplyResult.Status != "applied" {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("signer rotation apply result must have status applied")
	}
	if request.VerificationReceipt.Status != "verified" {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("signer rotation verification receipt must have status verified")
	}
	if request.ApplyResult.ReceiptID != request.VerificationReceipt.ReceiptID {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit receipt_id mismatch")
	}
	if request.ApplyResult.ChainID != request.VerificationReceipt.ChainID {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit chain_id mismatch")
	}
	if request.ApplyResult.PolicyPath != request.VerificationReceipt.PolicyPath {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit policy_path mismatch")
	}
	if request.ApplyResult.TargetPolicyVersion != request.VerificationReceipt.TargetPolicyVersion {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit target_policy_version mismatch")
	}
	if request.ApplyResult.ActivationPlanDigest != request.VerificationReceipt.ActivationPlanDigest {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit activation_plan_digest mismatch")
	}
	if request.ApplyResult.AppliedPolicyDigest != request.VerificationReceipt.PolicyDigest {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit policy_digest mismatch")
	}
	if strings.TrimSpace(request.VerificationReceipt.Signature.SignatureID) == "" {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit verification signature_id must be set")
	}
	if strings.TrimSpace(request.VerificationReceipt.Signature.Value) == "" {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit verification signature must be set")
	}
	effectiveAt, err := parseRFC3339(request.ApplyResult.AppliedAtEffect, "activation audit effective_at")
	if err != nil {
		return SignerRotationActivationAuditAppendResult{}, err
	}
	verifiedAt, err := parseRFC3339(request.VerificationReceipt.VerifiedAt, "activation audit verified_at")
	if err != nil {
		return SignerRotationActivationAuditAppendResult{}, err
	}
	if verifiedAt.Before(effectiveAt) {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit verified_at must be on or after effective_at")
	}
	ledger := request.ExistingLedger
	if ledger.Version == "" {
		ledger = SignerRotationActivationAuditLedger{
			Version:    "1.0.0",
			Status:     "active",
			ChainID:    request.ApplyResult.ChainID,
			PolicyPath: request.ApplyResult.PolicyPath,
			Entries:    []SignerRotationActivationAuditEntry{},
		}
	}
	if ledger.Version != "1.0.0" {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("unexpected activation audit ledger version: %s", ledger.Version)
	}
	if ledger.Status == "" {
		ledger.Status = "active"
	}
	if ledger.ChainID != request.ApplyResult.ChainID {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit ledger chain_id mismatch")
	}
	if ledger.PolicyPath != request.ApplyResult.PolicyPath {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit ledger policy_path mismatch")
	}

	var lastEffectiveAt time.Time
	var lastVerifiedAt time.Time
	seenReceiptIDs := make(map[string]struct{}, len(ledger.Entries))
	seenTargetVersions := make(map[string]struct{}, len(ledger.Entries))
	seenSignatureIDs := make(map[string]struct{}, len(ledger.Entries))
	for _, entry := range ledger.Entries {
		if _, exists := seenReceiptIDs[entry.ReceiptID]; exists {
			return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("duplicate activation audit receipt_id in ledger: %s", entry.ReceiptID)
		}
		seenReceiptIDs[entry.ReceiptID] = struct{}{}
		if _, exists := seenTargetVersions[entry.TargetPolicyVersion]; exists {
			return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("duplicate activation audit target_policy_version in ledger: %s", entry.TargetPolicyVersion)
		}
		seenTargetVersions[entry.TargetPolicyVersion] = struct{}{}
		if _, exists := seenSignatureIDs[entry.Signature.SignatureID]; exists {
			return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("duplicate activation audit signature_id in ledger: %s", entry.Signature.SignatureID)
		}
		seenSignatureIDs[entry.Signature.SignatureID] = struct{}{}

		entryEffectiveAt, err := parseRFC3339(entry.EffectiveAt, "activation audit ledger effective_at")
		if err != nil {
			return SignerRotationActivationAuditAppendResult{}, err
		}
		entryVerifiedAt, err := parseRFC3339(entry.VerifiedAt, "activation audit ledger verified_at")
		if err != nil {
			return SignerRotationActivationAuditAppendResult{}, err
		}
		if entryVerifiedAt.Before(entryEffectiveAt) {
			return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit ledger entry verified_at must be on or after effective_at")
		}
		if entryEffectiveAt.After(lastEffectiveAt) {
			lastEffectiveAt = entryEffectiveAt
		}
		if entryVerifiedAt.After(lastVerifiedAt) {
			lastVerifiedAt = entryVerifiedAt
		}
	}

	if _, exists := seenReceiptIDs[request.ApplyResult.ReceiptID]; exists {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit receipt_id already recorded: %s", request.ApplyResult.ReceiptID)
	}
	if _, exists := seenTargetVersions[request.ApplyResult.TargetPolicyVersion]; exists {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit target_policy_version already recorded: %s", request.ApplyResult.TargetPolicyVersion)
	}
	if _, exists := seenSignatureIDs[request.VerificationReceipt.Signature.SignatureID]; exists {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit signature_id already recorded: %s", request.VerificationReceipt.Signature.SignatureID)
	}
	if !lastEffectiveAt.IsZero() && !effectiveAt.After(lastEffectiveAt) {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit effective_at must be strictly increasing")
	}
	if !lastVerifiedAt.IsZero() && !verifiedAt.After(lastVerifiedAt) {
		return SignerRotationActivationAuditAppendResult{}, fmt.Errorf("activation audit verified_at must be strictly increasing")
	}

	entry := activationAuditEntry(request.ApplyResult, request.VerificationReceipt)
	ledger.Entries = append(ledger.Entries, entry)
	ledger.EntryCount = len(ledger.Entries)
	return SignerRotationActivationAuditAppendResult{
		Version:       "1.0.0",
		Status:        "appended",
		AppendedIndex: len(ledger.Entries) - 1,
		AppendedEntry: entry,
		Ledger:        ledger,
	}, nil
}

func SignerRotationActivationAuditReconcile(
	request SignerRotationActivationAuditReconcileRequest,
) (SignerRotationActivationAuditReconcileReport, error) {
	policyDigest, err := checkpointSignerPolicyDigest(request.Policy)
	if err != nil {
		return SignerRotationActivationAuditReconcileReport{}, err
	}
	policyPath := strings.TrimSpace(request.PolicyPath)
	if policyPath == "" {
		policyPath = request.Ledger.PolicyPath
	}
	report := SignerRotationActivationAuditReconcileReport{
		Version:                "1.0.0",
		Status:                 "consistent",
		ChainID:                "",
		PolicyPath:             policyPath,
		CurrentPolicyVersion:   request.Policy.Version,
		CurrentPolicyDigest:    policyDigest,
		EntryCount:             len(request.Ledger.Entries),
		CurrentPolicyExplained: false,
		Issues:                 []string{},
	}
	if request.Policy.Version == "" {
		return SignerRotationActivationAuditReconcileReport{}, fmt.Errorf("reconciliation policy version must be set")
	}
	report.ChainID = ""
	if len(request.Ledger.Entries) > 0 {
		report.ChainID = request.Ledger.Entries[0].ChainID
	}
	if report.ChainID == "" {
		report.ChainID = request.Ledger.ChainID
	}
	if report.ChainID == "" {
		report.ChainID = "unknown"
	}
	if request.Ledger.Version != "" && request.Ledger.Version != "1.0.0" {
		report.Issues = append(report.Issues, fmt.Sprintf("unexpected activation audit ledger version: %s", request.Ledger.Version))
	}
	if request.Ledger.ChainID != "" && request.Ledger.ChainID != report.ChainID {
		report.Issues = append(report.Issues, "activation audit ledger chain_id does not match current policy lineage")
	}
	if request.Ledger.PolicyPath != "" && policyPath != "" && request.Ledger.PolicyPath != policyPath {
		report.Issues = append(report.Issues, "activation audit ledger policy_path does not match reconciliation policy path")
	}
	if len(request.Ledger.Entries) == 0 {
		if strings.Contains(request.Policy.Version, "+rotation-") {
			report.Issues = append(report.Issues, "current checkpoint signer policy appears rotated but activation audit ledger is empty")
			report.Status = "gap"
		} else {
			report.CurrentPolicyExplained = true
		}
		return report, nil
	}

	seenReceiptIDs := make(map[string]struct{}, len(request.Ledger.Entries))
	seenTargetVersions := make(map[string]struct{}, len(request.Ledger.Entries))
	seenSignatureIDs := make(map[string]struct{}, len(request.Ledger.Entries))
	lastEffectiveAt := time.Time{}
	lastVerifiedAt := time.Time{}
	latest := request.Ledger.Entries[len(request.Ledger.Entries)-1]
	report.LatestReceiptID = latest.ReceiptID
	report.LatestTargetVersion = latest.TargetPolicyVersion
	report.LatestPolicyDigest = latest.AppliedPolicyDigest

	for _, entry := range request.Ledger.Entries {
		if _, exists := seenReceiptIDs[entry.ReceiptID]; exists {
			report.Issues = append(report.Issues, fmt.Sprintf("duplicate activation audit receipt_id in ledger: %s", entry.ReceiptID))
		}
		seenReceiptIDs[entry.ReceiptID] = struct{}{}
		if _, exists := seenTargetVersions[entry.TargetPolicyVersion]; exists {
			report.Issues = append(report.Issues, fmt.Sprintf("duplicate activation audit target_policy_version in ledger: %s", entry.TargetPolicyVersion))
		}
		seenTargetVersions[entry.TargetPolicyVersion] = struct{}{}
		if _, exists := seenSignatureIDs[entry.Signature.SignatureID]; exists {
			report.Issues = append(report.Issues, fmt.Sprintf("duplicate activation audit signature_id in ledger: %s", entry.Signature.SignatureID))
		}
		seenSignatureIDs[entry.Signature.SignatureID] = struct{}{}
		if entry.ChainID != "" && entry.ChainID != report.ChainID {
			report.Issues = append(report.Issues, fmt.Sprintf("activation audit entry %s chain_id mismatch", entry.ReceiptID))
		}
		if policyPath != "" && entry.PolicyPath != "" && entry.PolicyPath != policyPath {
			report.Issues = append(report.Issues, fmt.Sprintf("activation audit entry %s policy_path mismatch", entry.ReceiptID))
		}
		effectiveAt, err := parseRFC3339(entry.EffectiveAt, "activation audit entry effective_at")
		if err != nil {
			return SignerRotationActivationAuditReconcileReport{}, err
		}
		verifiedAt, err := parseRFC3339(entry.VerifiedAt, "activation audit entry verified_at")
		if err != nil {
			return SignerRotationActivationAuditReconcileReport{}, err
		}
		if verifiedAt.Before(effectiveAt) {
			report.Issues = append(report.Issues, fmt.Sprintf("activation audit entry %s verified_at is before effective_at", entry.ReceiptID))
		}
		if !lastEffectiveAt.IsZero() && !effectiveAt.After(lastEffectiveAt) {
			report.Issues = append(report.Issues, "activation audit ledger effective_at is not strictly increasing")
		}
		if !lastVerifiedAt.IsZero() && !verifiedAt.After(lastVerifiedAt) {
			report.Issues = append(report.Issues, "activation audit ledger verified_at is not strictly increasing")
		}
		lastEffectiveAt = effectiveAt
		lastVerifiedAt = verifiedAt
	}

	if latest.TargetPolicyVersion != request.Policy.Version {
		report.Issues = append(report.Issues, "current checkpoint signer policy version is not explained by the latest activation audit entry")
	}
	if latest.AppliedPolicyDigest != policyDigest {
		report.Issues = append(report.Issues, "current checkpoint signer policy digest is not explained by the latest activation audit entry")
	}
	report.CurrentPolicyExplained = latest.TargetPolicyVersion == request.Policy.Version && latest.AppliedPolicyDigest == policyDigest
	if !report.CurrentPolicyExplained && strings.Contains(request.Policy.Version, "+rotation-") {
		report.Issues = append(report.Issues, "current checkpoint signer policy cannot be explained by the activation audit ledger lineage")
	}

	if len(report.Issues) == 0 {
		report.Status = "consistent"
		return report, nil
	}
	report.Status = "gap"
	for _, issue := range report.Issues {
		if strings.Contains(issue, "duplicate activation audit") || strings.Contains(issue, "not strictly increasing") || strings.Contains(issue, "mismatch") {
			report.Status = "invalid"
			break
		}
	}
	return report, nil
}

func SignerRotationActivationAuditExport(
	request SignerRotationActivationAuditExportRequest,
) (SignerRotationActivationAuditExportPackage, error) {
	if request.Policy.Version == "" {
		return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("activation audit export policy version must be set")
	}
	report := request.Reconciliation
	if report.Version != "1.0.0" {
		return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("unexpected activation audit reconciliation report version: %s", report.Version)
	}
	if request.Ledger.Version != "" && request.Ledger.Version != "1.0.0" {
		return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("unexpected activation audit ledger version: %s", request.Ledger.Version)
	}
	policyDigest, err := checkpointSignerPolicyDigest(request.Policy)
	if err != nil {
		return SignerRotationActivationAuditExportPackage{}, err
	}
	policyPath := strings.TrimSpace(request.PolicyPath)
	if policyPath == "" {
		policyPath = strings.TrimSpace(report.PolicyPath)
	}
	if policyPath == "" {
		policyPath = strings.TrimSpace(request.Ledger.PolicyPath)
	}
	if policyPath == "" {
		return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("activation audit export policy path must be set")
	}
	if report.PolicyPath != "" && report.PolicyPath != policyPath {
		return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("activation audit reconciliation policy_path mismatch")
	}
	if request.Ledger.PolicyPath != "" && request.Ledger.PolicyPath != policyPath {
		return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("activation audit ledger policy_path mismatch")
	}
	if report.CurrentPolicyVersion != request.Policy.Version {
		return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("activation audit reconciliation current_policy_version mismatch")
	}
	if report.CurrentPolicyDigest != policyDigest {
		return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("activation audit reconciliation current_policy_digest mismatch")
	}
	if report.EntryCount != len(request.Ledger.Entries) {
		return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("activation audit reconciliation entry_count mismatch")
	}

	chainID := strings.TrimSpace(report.ChainID)
	if chainID == "" {
		if len(request.Ledger.Entries) > 0 {
			chainID = request.Ledger.Entries[0].ChainID
		}
		if chainID == "" {
			chainID = request.Ledger.ChainID
		}
	}
	if chainID == "" {
		chainID = "unknown"
	}
	if request.Ledger.ChainID != "" && request.Ledger.ChainID != chainID {
		return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("activation audit ledger chain_id mismatch")
	}
	if report.ChainID != "" && report.ChainID != chainID {
		return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("activation audit reconciliation chain_id mismatch")
	}

	var latestEntry *SignerRotationActivationAuditEntry
	if len(request.Ledger.Entries) > 0 {
		entry := request.Ledger.Entries[len(request.Ledger.Entries)-1]
		if report.LatestReceiptID != entry.ReceiptID {
			return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("activation audit reconciliation latest_receipt_id mismatch")
		}
		if report.LatestTargetVersion != entry.TargetPolicyVersion {
			return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("activation audit reconciliation latest_target_policy_version mismatch")
		}
		if report.LatestPolicyDigest != entry.AppliedPolicyDigest {
			return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("activation audit reconciliation latest_applied_policy_digest mismatch")
		}
		entryCopy := entry
		latestEntry = &entryCopy
	} else if report.LatestReceiptID != "" || report.LatestTargetVersion != "" || report.LatestPolicyDigest != "" {
		return SignerRotationActivationAuditExportPackage{}, fmt.Errorf("activation audit reconciliation latest entry metadata mismatch")
	}

	exportStatus := report.Status
	if exportStatus == "" {
		exportStatus = "unknown"
	}
	return SignerRotationActivationAuditExportPackage{
		Version:    "1.0.0",
		Status:     exportStatus,
		ChainID:    chainID,
		PolicyPath: policyPath,
		BaselineSnapshot: SignerRotationActivationAuditBaselineSnapshot{
			Version:                "1.0.0",
			Status:                 exportStatus,
			ChainID:                chainID,
			PolicyPath:             policyPath,
			LedgerEntryCount:       len(request.Ledger.Entries),
			CurrentPolicyVersion:   request.Policy.Version,
			CurrentPolicyDigest:    policyDigest,
			CurrentPolicyExplained: report.CurrentPolicyExplained,
			LatestReceiptID:        report.LatestReceiptID,
			LatestTargetVersion:    report.LatestTargetVersion,
			LatestPolicyDigest:     report.LatestPolicyDigest,
			LatestEntry:            latestEntry,
			ReconciliationStatus:   report.Status,
			ContinuityIssues:       append([]string(nil), report.Issues...),
		},
		CurrentPolicy:  request.Policy,
		Ledger:         request.Ledger,
		Reconciliation: report,
	}, nil
}

func SignerRotationActivationAuditVerifyExport(
	request SignerRotationActivationAuditExportVerificationRequest,
) (SignerRotationActivationAuditExportVerificationReport, error) {
	pkg := request.ExportPackage
	report := SignerRotationActivationAuditExportVerificationReport{
		Version:              "1.0.0",
		Status:               "consistent",
		ArchiveReady:         false,
		ChainID:              pkg.ChainID,
		PolicyPath:           pkg.PolicyPath,
		CurrentPolicyVersion: pkg.CurrentPolicy.Version,
		EntryCount:           len(pkg.Ledger.Entries),
		LatestReceiptID:      pkg.BaselineSnapshot.LatestReceiptID,
		LatestTargetVersion:  pkg.BaselineSnapshot.LatestTargetVersion,
		VerificationIssues:   []string{},
	}
	policyDigest, err := checkpointSignerPolicyDigest(pkg.CurrentPolicy)
	if err != nil {
		return SignerRotationActivationAuditExportVerificationReport{}, err
	}
	report.CurrentPolicyDigest = policyDigest
	if pkg.Version != "1.0.0" {
		report.VerificationIssues = append(report.VerificationIssues, fmt.Sprintf("unexpected activation audit export package version: %s", pkg.Version))
	}
	expectedReconciliation, err := SignerRotationActivationAuditReconcile(SignerRotationActivationAuditReconcileRequest{
		Ledger:     pkg.Ledger,
		Policy:     pkg.CurrentPolicy,
		PolicyPath: pkg.PolicyPath,
	})
	if err != nil {
		report.VerificationIssues = append(report.VerificationIssues, err.Error())
	} else {
		expectedExport, err := SignerRotationActivationAuditExport(SignerRotationActivationAuditExportRequest{
			Ledger:         pkg.Ledger,
			Policy:         pkg.CurrentPolicy,
			Reconciliation: expectedReconciliation,
			PolicyPath:     pkg.PolicyPath,
		})
		if err != nil {
			report.VerificationIssues = append(report.VerificationIssues, err.Error())
		} else {
			expectedJSON, err := json.Marshal(expectedExport)
			if err != nil {
				return SignerRotationActivationAuditExportVerificationReport{}, fmt.Errorf("marshal expected activation audit export package: %w", err)
			}
			actualJSON, err := json.Marshal(pkg)
			if err != nil {
				return SignerRotationActivationAuditExportVerificationReport{}, fmt.Errorf("marshal actual activation audit export package: %w", err)
			}
			if !bytesEqual(expectedJSON, actualJSON) {
				report.VerificationIssues = append(report.VerificationIssues, "activation audit export package drift detected")
			}
		}
	}
	if len(report.VerificationIssues) > 0 {
		report.Status = "invalid"
		return report, nil
	}
	if pkg.Reconciliation.Status == "consistent" {
		report.ArchiveReady = true
		report.Status = "consistent"
		return report, nil
	}
	report.Status = "review"
	return report, nil
}

func BuildSignerRotationActivationAuditArchiveIndex(
	request SignerRotationActivationAuditArchiveIndexRequest,
) (SignerRotationActivationAuditArchiveIndex, error) {
	if len(request.Packages) == 0 {
		return SignerRotationActivationAuditArchiveIndex{}, fmt.Errorf("activation audit archive index requires at least one export package")
	}
	index := SignerRotationActivationAuditArchiveIndex{
		Version:    "1.0.0",
		Status:     "consistent",
		ChainID:    "",
		PolicyPath: "",
		Entries:    []SignerRotationActivationAuditArchiveIndexEntry{},
		Issues:     []string{},
	}
	type indexedEntry struct {
		entry       SignerRotationActivationAuditArchiveIndexEntry
		latestOrder string
	}
	indexedEntries := make([]indexedEntry, 0, len(request.Packages))
	seenPackagePaths := make(map[string]struct{}, len(request.Packages))
	seenVersions := make(map[string]string, len(request.Packages))
	seenReceipts := make(map[string]string, len(request.Packages))
	for _, item := range request.Packages {
		if _, exists := seenPackagePaths[item.PackagePath]; exists {
			index.Issues = append(index.Issues, fmt.Sprintf("duplicate archive package_path %s", item.PackagePath))
		} else {
			seenPackagePaths[item.PackagePath] = struct{}{}
		}
		verification, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
			ExportPackage: item.ExportPackage,
		})
		if err != nil {
			return SignerRotationActivationAuditArchiveIndex{}, err
		}
		if index.ChainID == "" {
			index.ChainID = verification.ChainID
		} else if verification.ChainID != "" && verification.ChainID != index.ChainID {
			index.Issues = append(index.Issues, fmt.Sprintf("archive package %s chain_id mismatch", item.PackagePath))
		}
		if index.PolicyPath == "" {
			index.PolicyPath = verification.PolicyPath
		} else if verification.PolicyPath != "" && verification.PolicyPath != index.PolicyPath {
			index.Issues = append(index.Issues, fmt.Sprintf("archive package %s policy_path mismatch", item.PackagePath))
		}
		if existing, exists := seenVersions[verification.CurrentPolicyVersion]; exists && verification.CurrentPolicyVersion != "" {
			index.Issues = append(index.Issues, fmt.Sprintf("duplicate archive current_policy_version %s in %s and %s", verification.CurrentPolicyVersion, existing, item.PackagePath))
		} else if verification.CurrentPolicyVersion != "" {
			seenVersions[verification.CurrentPolicyVersion] = item.PackagePath
		}
		if existing, exists := seenReceipts[verification.LatestReceiptID]; exists && verification.LatestReceiptID != "" {
			index.Issues = append(index.Issues, fmt.Sprintf("duplicate archive latest_receipt_id %s in %s and %s", verification.LatestReceiptID, existing, item.PackagePath))
		} else if verification.LatestReceiptID != "" {
			seenReceipts[verification.LatestReceiptID] = item.PackagePath
		}
		entry := SignerRotationActivationAuditArchiveIndexEntry{
			PackagePath:          item.PackagePath,
			Status:               verification.Status,
			ArchiveReady:         verification.ArchiveReady,
			ChainID:              verification.ChainID,
			PolicyPath:           verification.PolicyPath,
			CurrentPolicyVersion: verification.CurrentPolicyVersion,
			CurrentPolicyDigest:  verification.CurrentPolicyDigest,
			EntryCount:           verification.EntryCount,
			LatestReceiptID:      verification.LatestReceiptID,
			LatestTargetVersion:  verification.LatestTargetVersion,
			VerificationIssues:   append([]string(nil), verification.VerificationIssues...),
		}
		if verification.ArchiveReady {
			index.ArchiveReadyCount++
		}
		indexedEntries = append(indexedEntries, indexedEntry{
			entry:       entry,
			latestOrder: archiveEntrySortKey(item.ExportPackage),
		})
	}
	sort.Slice(indexedEntries, func(i, j int) bool {
		if indexedEntries[i].latestOrder == indexedEntries[j].latestOrder {
			return indexedEntries[i].entry.PackagePath < indexedEntries[j].entry.PackagePath
		}
		return indexedEntries[i].latestOrder < indexedEntries[j].latestOrder
	})
	for _, item := range indexedEntries {
		index.Entries = append(index.Entries, item.entry)
	}
	index.PackageCount = len(index.Entries)
	if len(index.Entries) > 0 {
		latest := index.Entries[len(index.Entries)-1]
		index.LatestCurrentPolicyVersion = latest.CurrentPolicyVersion
		index.LatestCurrentPolicyDigest = latest.CurrentPolicyDigest
	}
	if len(index.Issues) > 0 {
		index.Status = "invalid"
		return index, nil
	}
	for _, entry := range index.Entries {
		if !entry.ArchiveReady {
			index.Status = "review"
			return index, nil
		}
	}
	index.Status = "consistent"
	return index, nil
}

func activationAuditArchiveIndexConsistencyIssues(index SignerRotationActivationAuditArchiveIndex) []string {
	issues := []string{}
	if index.Version != "1.0.0" {
		issues = append(issues, fmt.Sprintf("unexpected activation audit archive index version: %s", index.Version))
	}
	if index.Status != "consistent" {
		issues = append(issues, fmt.Sprintf("activation audit archive index status must be consistent, got %s", index.Status))
	}
	if len(index.Issues) > 0 {
		issues = append(issues, "activation audit archive index contains issues")
	}
	if len(index.Entries) != index.PackageCount {
		issues = append(issues, "activation audit archive index package_count does not match entries")
	}
	if len(index.Entries) == 0 {
		issues = append(issues, "activation audit archive index must include at least one entry")
		return issues
	}

	seenPackagePaths := make(map[string]struct{}, len(index.Entries))
	archiveReadyCount := 0
	for _, entry := range index.Entries {
		if _, exists := seenPackagePaths[entry.PackagePath]; exists {
			issues = append(issues, fmt.Sprintf("activation audit archive index contains duplicate package_path: %s", entry.PackagePath))
		} else {
			seenPackagePaths[entry.PackagePath] = struct{}{}
		}
		if !entry.ArchiveReady {
			issues = append(issues, fmt.Sprintf("activation audit archive index contains non-archive-ready entry for package_path: %s", entry.PackagePath))
		}
		if entry.Status != "consistent" {
			issues = append(issues, fmt.Sprintf("activation audit archive index contains non-consistent entry for package_path: %s", entry.PackagePath))
		}
		if entry.ArchiveReady {
			archiveReadyCount++
		}
	}
	if archiveReadyCount != index.ArchiveReadyCount || archiveReadyCount != len(index.Entries) {
		issues = append(issues, "activation audit archive index archive_ready_count mismatch")
	}
	latestEntry := index.Entries[len(index.Entries)-1]
	if index.LatestCurrentPolicyVersion != latestEntry.CurrentPolicyVersion {
		issues = append(issues, "activation audit archive index latest_current_policy_version mismatch")
	}
	if index.LatestCurrentPolicyDigest != latestEntry.CurrentPolicyDigest {
		issues = append(issues, "activation audit archive index latest_current_policy_digest mismatch")
	}
	return issues
}

func archiveEntrySortKey(pkg SignerRotationActivationAuditExportPackage) string {
	if pkg.BaselineSnapshot.LatestEntry != nil {
		return pkg.BaselineSnapshot.LatestEntry.EffectiveAt + "|" + pkg.CurrentPolicy.Version
	}
	return "0000-00-00T00:00:00Z|" + pkg.CurrentPolicy.Version
}

func BuildSignerRotationActivationAuditArchivePromotion(
	request SignerRotationActivationAuditArchivePromotionRequest,
) (SignerRotationActivationAuditArchivePromotionResult, error) {
	packagePath := strings.TrimSpace(request.PackagePath)
	if packagePath == "" {
		return SignerRotationActivationAuditArchivePromotionResult{}, fmt.Errorf("activation audit archive promotion package_path must be set")
	}
	promotedAt, err := parseRFC3339(request.PromotedAt, "activation audit archive promotion promoted_at")
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionResult{}, err
	}
	promotedBy := strings.TrimSpace(request.PromotedBy)
	if promotedBy == "" {
		return SignerRotationActivationAuditArchivePromotionResult{}, fmt.Errorf("activation audit archive promotion promoted_by must be set")
	}

	result := SignerRotationActivationAuditArchivePromotionResult{
		Version: "1.0.0",
		Status:  "invalid",
		Issues:  []string{},
	}

	expectedVerification, err := SignerRotationActivationAuditVerifyExport(SignerRotationActivationAuditExportVerificationRequest{
		ExportPackage: request.ExportPackage,
	})
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionResult{}, err
	}
	expectedVerificationJSON, err := json.Marshal(expectedVerification)
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionResult{}, fmt.Errorf("marshal expected activation audit export verification report: %w", err)
	}
	actualVerificationJSON, err := json.Marshal(request.VerificationReport)
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionResult{}, fmt.Errorf("marshal provided activation audit export verification report: %w", err)
	}
	if !bytesEqual(expectedVerificationJSON, actualVerificationJSON) {
		result.Issues = append(result.Issues, "activation audit export verification report drift detected")
	}
	if expectedVerification.Status != "consistent" || !expectedVerification.ArchiveReady {
		result.Issues = append(result.Issues, "activation audit export package is not archive_ready")
	}
	result.Issues = append(result.Issues, activationAuditArchiveIndexConsistencyIssues(request.ArchiveIndex)...)
	if request.ArchiveIndex.ChainID != "" && request.ArchiveIndex.ChainID != expectedVerification.ChainID {
		result.Issues = append(result.Issues, "activation audit archive index chain_id mismatch")
	}
	if request.ArchiveIndex.PolicyPath != "" && request.ArchiveIndex.PolicyPath != expectedVerification.PolicyPath {
		result.Issues = append(result.Issues, "activation audit archive index policy_path mismatch")
	}

	matchedIndex := -1
	for idx, entry := range request.ArchiveIndex.Entries {
		if entry.PackagePath == packagePath {
			matchedIndex = idx
			break
		}
	}
	if matchedIndex == -1 {
		result.Issues = append(result.Issues, fmt.Sprintf("archive package path %s not found in archive index", packagePath))
	}

	var matchedEntry SignerRotationActivationAuditArchiveIndexEntry
	if matchedIndex >= 0 {
		matchedEntry = request.ArchiveIndex.Entries[matchedIndex]
		if matchedEntry.Status != expectedVerification.Status {
			result.Issues = append(result.Issues, "archive index entry status mismatch")
		}
		if matchedEntry.ArchiveReady != expectedVerification.ArchiveReady {
			result.Issues = append(result.Issues, "archive index entry archive_ready mismatch")
		}
		if matchedEntry.ChainID != expectedVerification.ChainID {
			result.Issues = append(result.Issues, "archive index entry chain_id mismatch")
		}
		if matchedEntry.PolicyPath != expectedVerification.PolicyPath {
			result.Issues = append(result.Issues, "archive index entry policy_path mismatch")
		}
		if matchedEntry.CurrentPolicyVersion != expectedVerification.CurrentPolicyVersion {
			result.Issues = append(result.Issues, "archive index entry current_policy_version mismatch")
		}
		if matchedEntry.CurrentPolicyDigest != expectedVerification.CurrentPolicyDigest {
			result.Issues = append(result.Issues, "archive index entry current_policy_digest mismatch")
		}
		if matchedEntry.EntryCount != expectedVerification.EntryCount {
			result.Issues = append(result.Issues, "archive index entry entry_count mismatch")
		}
		if matchedEntry.LatestReceiptID != expectedVerification.LatestReceiptID {
			result.Issues = append(result.Issues, "archive index entry latest_receipt_id mismatch")
		}
		if matchedEntry.LatestTargetVersion != expectedVerification.LatestTargetVersion {
			result.Issues = append(result.Issues, "archive index entry latest_target_policy_version mismatch")
		}
	}

	if len(result.Issues) > 0 {
		return result, nil
	}

	exportPackageDigest, err := digestJSON(request.ExportPackage, "activation audit export package")
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionResult{}, err
	}
	verificationDigest, err := digestJSON(expectedVerification, "activation audit export verification report")
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionResult{}, err
	}
	archiveIndexDigest, err := digestJSON(request.ArchiveIndex, "activation audit archive index")
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionResult{}, err
	}
	archiveEntryDigest, err := digestJSON(matchedEntry, "activation audit archive index entry")
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionResult{}, err
	}

	receiptID := deterministicArtifactID(
		"archive-promotion",
		packagePath,
		expectedVerification.CurrentPolicyVersion,
		expectedVerification.CurrentPolicyDigest,
		promotedAt.Format(time.RFC3339),
		promotedBy,
		archiveIndexDigest,
	)
	attestationID := deterministicArtifactID(
		"retained-baseline",
		receiptID,
		archiveEntryDigest,
		verificationDigest,
	)

	result.PromotionReceipt = SignerRotationActivationAuditArchivePromotionReceipt{
		Version:              "1.0.0",
		Status:               "promoted",
		ReceiptID:            receiptID,
		PackagePath:          packagePath,
		PromotedAt:           promotedAt.Format(time.RFC3339),
		PromotedBy:           promotedBy,
		ChainID:              expectedVerification.ChainID,
		PolicyPath:           expectedVerification.PolicyPath,
		CurrentPolicyVersion: expectedVerification.CurrentPolicyVersion,
		CurrentPolicyDigest:  expectedVerification.CurrentPolicyDigest,
		EntryCount:           expectedVerification.EntryCount,
		LatestReceiptID:      expectedVerification.LatestReceiptID,
		LatestTargetVersion:  expectedVerification.LatestTargetVersion,
		ExportPackageDigest:  exportPackageDigest,
		VerificationDigest:   verificationDigest,
		ArchiveIndexDigest:   archiveIndexDigest,
		ArchiveEntryDigest:   archiveEntryDigest,
	}
	result.RetainedBaselineAttestation = SignerRotationActivationAuditRetainedBaselineAttestation{
		Version:              "1.0.0",
		Status:               "retained",
		AttestationID:        attestationID,
		PromotionReceiptID:   receiptID,
		PackagePath:          packagePath,
		ArchiveEntryIndex:    matchedIndex,
		PromotedAt:           promotedAt.Format(time.RFC3339),
		PromotedBy:           promotedBy,
		ChainID:              expectedVerification.ChainID,
		PolicyPath:           expectedVerification.PolicyPath,
		CurrentPolicyVersion: expectedVerification.CurrentPolicyVersion,
		CurrentPolicyDigest:  expectedVerification.CurrentPolicyDigest,
		LatestReceiptID:      expectedVerification.LatestReceiptID,
		LatestTargetVersion:  expectedVerification.LatestTargetVersion,
		ExportPackageDigest:  exportPackageDigest,
		VerificationDigest:   verificationDigest,
		ArchiveIndexDigest:   archiveIndexDigest,
		ArchiveEntryDigest:   archiveEntryDigest,
		Claims: []string{
			"verified export package matches the retained archive entry",
			"archive index lineage matches the promoted export verification report",
			"retained baseline promotion is bound to the current policy digest and receipt lineage",
		},
	}
	result.Status = "promoted"
	return result, nil
}

func VerifySignerRotationActivationAuditArchivePromotion(
	request SignerRotationActivationAuditArchivePromotionVerificationRequest,
) (SignerRotationActivationAuditArchivePromotionVerificationReceipt, error) {
	packagePath := strings.TrimSpace(request.PackagePath)
	if packagePath == "" {
		return SignerRotationActivationAuditArchivePromotionVerificationReceipt{}, fmt.Errorf("activation audit archive promotion verification package_path must be set")
	}
	verifiedAt, err := parseRFC3339(request.VerifiedAt, "activation audit archive promotion verification verified_at")
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionVerificationReceipt{}, err
	}
	verifiedBy := strings.TrimSpace(request.VerifiedBy)
	if verifiedBy == "" {
		return SignerRotationActivationAuditArchivePromotionVerificationReceipt{}, fmt.Errorf("activation audit archive promotion verification verified_by must be set")
	}

	receipt := SignerRotationActivationAuditArchivePromotionVerificationReceipt{
		Version:            "1.0.0",
		Status:             "invalid",
		PackagePath:        packagePath,
		VerifiedAt:         verifiedAt.Format(time.RFC3339),
		VerifiedBy:         verifiedBy,
		VerificationIssues: []string{},
	}

	expectedPromotion, err := BuildSignerRotationActivationAuditArchivePromotion(SignerRotationActivationAuditArchivePromotionRequest{
		PackagePath:        packagePath,
		ExportPackage:      request.ExportPackage,
		VerificationReport: request.VerificationReport,
		ArchiveIndex:       request.ArchiveIndex,
		PromotedAt:         request.PromotionResult.PromotionReceipt.PromotedAt,
		PromotedBy:         request.PromotionResult.PromotionReceipt.PromotedBy,
	})
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionVerificationReceipt{}, err
	}
	expectedJSON, err := json.Marshal(expectedPromotion)
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionVerificationReceipt{}, fmt.Errorf("marshal expected activation audit archive promotion: %w", err)
	}
	actualJSON, err := json.Marshal(request.PromotionResult)
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionVerificationReceipt{}, fmt.Errorf("marshal actual activation audit archive promotion: %w", err)
	}
	if !bytesEqual(expectedJSON, actualJSON) {
		receipt.VerificationIssues = append(receipt.VerificationIssues, "activation audit archive promotion drift detected")
	}
	if expectedPromotion.Status != "promoted" {
		receipt.VerificationIssues = append(receipt.VerificationIssues, "activation audit archive promotion is not promoted")
	}

	promotionResultDigest, err := digestJSON(request.PromotionResult, "activation audit archive promotion result")
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionVerificationReceipt{}, err
	}
	promotionReceiptDigest, err := digestJSON(request.PromotionResult.PromotionReceipt, "activation audit archive promotion receipt")
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionVerificationReceipt{}, err
	}
	attestationDigest, err := digestJSON(request.PromotionResult.RetainedBaselineAttestation, "activation audit retained baseline attestation")
	if err != nil {
		return SignerRotationActivationAuditArchivePromotionVerificationReceipt{}, err
	}

	receipt.PromotionReceiptID = request.PromotionResult.PromotionReceipt.ReceiptID
	receipt.AttestationID = request.PromotionResult.RetainedBaselineAttestation.AttestationID
	receipt.ChainID = request.PromotionResult.PromotionReceipt.ChainID
	receipt.PolicyPath = request.PromotionResult.PromotionReceipt.PolicyPath
	receipt.CurrentPolicyVersion = request.PromotionResult.PromotionReceipt.CurrentPolicyVersion
	receipt.CurrentPolicyDigest = request.PromotionResult.PromotionReceipt.CurrentPolicyDigest
	receipt.LatestReceiptID = request.PromotionResult.PromotionReceipt.LatestReceiptID
	receipt.LatestTargetVersion = request.PromotionResult.PromotionReceipt.LatestTargetVersion
	receipt.PromotionResultDigest = promotionResultDigest
	receipt.PromotionReceiptDigest = promotionReceiptDigest
	receipt.AttestationDigest = attestationDigest
	receipt.ArchiveIndexDigest = request.PromotionResult.PromotionReceipt.ArchiveIndexDigest
	receipt.ArchiveEntryDigest = request.PromotionResult.PromotionReceipt.ArchiveEntryDigest

	if request.PromotionResult.PromotionReceipt.Status != "promoted" {
		receipt.VerificationIssues = append(receipt.VerificationIssues, "promotion receipt status must be promoted")
	}
	if request.PromotionResult.RetainedBaselineAttestation.Status != "retained" {
		receipt.VerificationIssues = append(receipt.VerificationIssues, "retained baseline attestation status must be retained")
	}
	if request.PromotionResult.RetainedBaselineAttestation.PromotionReceiptID != request.PromotionResult.PromotionReceipt.ReceiptID {
		receipt.VerificationIssues = append(receipt.VerificationIssues, "retained baseline attestation promotion_receipt_id mismatch")
	}
	if request.PromotionResult.RetainedBaselineAttestation.ArchiveIndexDigest != request.PromotionResult.PromotionReceipt.ArchiveIndexDigest {
		receipt.VerificationIssues = append(receipt.VerificationIssues, "retained baseline attestation archive_index_digest mismatch")
	}
	if request.PromotionResult.RetainedBaselineAttestation.ArchiveEntryDigest != request.PromotionResult.PromotionReceipt.ArchiveEntryDigest {
		receipt.VerificationIssues = append(receipt.VerificationIssues, "retained baseline attestation archive_entry_digest mismatch")
	}
	if request.PromotionResult.RetainedBaselineAttestation.ExportPackageDigest != request.PromotionResult.PromotionReceipt.ExportPackageDigest {
		receipt.VerificationIssues = append(receipt.VerificationIssues, "retained baseline attestation export_package_digest mismatch")
	}
	if request.PromotionResult.RetainedBaselineAttestation.VerificationDigest != request.PromotionResult.PromotionReceipt.VerificationDigest {
		receipt.VerificationIssues = append(receipt.VerificationIssues, "retained baseline attestation verification_digest mismatch")
	}

	if len(receipt.VerificationIssues) > 0 {
		return receipt, nil
	}
	receipt.VerificationReceiptID = deterministicArtifactID(
		"archive-verification",
		receipt.PromotionReceiptID,
		receipt.AttestationID,
		packagePath,
		receipt.CurrentPolicyVersion,
		receipt.VerifiedAt,
		receipt.VerifiedBy,
	)
	receipt.Status = "verified"
	return receipt, nil
}

func BuildSignerRotationActivationAuditRetainedInventorySnapshot(
	request SignerRotationActivationAuditRetainedInventorySnapshotRequest,
) (SignerRotationActivationAuditRetainedInventorySnapshot, error) {
	if len(request.Packages) == 0 {
		return SignerRotationActivationAuditRetainedInventorySnapshot{}, fmt.Errorf("retained inventory snapshot requires at least one promoted package")
	}
	snapshot := SignerRotationActivationAuditRetainedInventorySnapshot{
		Version: "1.0.0",
		Status:  "consistent",
		Entries: []SignerRotationActivationAuditRetainedInventorySnapshotEntry{},
		Issues:  []string{},
	}
	type orderedEntry struct {
		entry   SignerRotationActivationAuditRetainedInventorySnapshotEntry
		sortKey string
	}
	ordered := make([]orderedEntry, 0, len(request.Packages))
	seenPolicyVersions := make(map[string]string, len(request.Packages))
	seenPromotionReceipts := make(map[string]string, len(request.Packages))
	seenAttestations := make(map[string]string, len(request.Packages))
	seenVerificationReceipts := make(map[string]string, len(request.Packages))
	for _, item := range request.Packages {
		verificationReceipt := item.VerificationReceipt
		promotion := item.PromotionResult
		path := strings.TrimSpace(item.PromotionPath)
		if path == "" {
			snapshot.Issues = append(snapshot.Issues, "retained inventory promotion_path must be set")
			continue
		}
		promotionResultDigest, err := digestJSON(promotion, "activation audit archive promotion result")
		if err != nil {
			return SignerRotationActivationAuditRetainedInventorySnapshot{}, err
		}
		promotionReceiptDigest, err := digestJSON(promotion.PromotionReceipt, "activation audit archive promotion receipt")
		if err != nil {
			return SignerRotationActivationAuditRetainedInventorySnapshot{}, err
		}
		attestationDigest, err := digestJSON(promotion.RetainedBaselineAttestation, "activation audit retained baseline attestation")
		if err != nil {
			return SignerRotationActivationAuditRetainedInventorySnapshot{}, err
		}
		verificationReceiptDigest, err := digestJSON(verificationReceipt, "activation audit archive promotion verification receipt")
		if err != nil {
			return SignerRotationActivationAuditRetainedInventorySnapshot{}, err
		}
		entry := SignerRotationActivationAuditRetainedInventorySnapshotEntry{
			PromotionPath:             path,
			Status:                    verificationReceipt.Status,
			Verified:                  verificationReceipt.Status == "verified",
			VerificationReceiptID:     verificationReceipt.VerificationReceiptID,
			PromotionReceiptID:        promotion.PromotionReceipt.ReceiptID,
			AttestationID:             promotion.RetainedBaselineAttestation.AttestationID,
			ChainID:                   promotion.PromotionReceipt.ChainID,
			PolicyPath:                promotion.PromotionReceipt.PolicyPath,
			CurrentPolicyVersion:      promotion.PromotionReceipt.CurrentPolicyVersion,
			CurrentPolicyDigest:       promotion.PromotionReceipt.CurrentPolicyDigest,
			LatestReceiptID:           promotion.PromotionReceipt.LatestReceiptID,
			LatestTargetVersion:       promotion.PromotionReceipt.LatestTargetVersion,
			PromotedAt:                promotion.PromotionReceipt.PromotedAt,
			PromotedBy:                promotion.PromotionReceipt.PromotedBy,
			VerifiedAt:                verificationReceipt.VerifiedAt,
			VerifiedBy:                verificationReceipt.VerifiedBy,
			PromotionResultDigest:     promotionResultDigest,
			VerificationReceiptDigest: verificationReceiptDigest,
			VerificationIssues:        append([]string(nil), verificationReceipt.VerificationIssues...),
		}
		if verificationReceipt.Status != "verified" {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s is not verified", path))
		}
		if verificationReceipt.PromotionResultDigest != promotionResultDigest {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s promotion_result_digest mismatch", path))
		}
		if verificationReceipt.PromotionReceiptDigest != promotionReceiptDigest {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s promotion_receipt_digest mismatch", path))
		}
		if verificationReceipt.AttestationDigest != attestationDigest {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s attestation_digest mismatch", path))
		}
		if verificationReceipt.PromotionReceiptID != promotion.PromotionReceipt.ReceiptID {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s promotion_receipt_id mismatch", path))
		}
		if verificationReceipt.AttestationID != promotion.RetainedBaselineAttestation.AttestationID {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s attestation_id mismatch", path))
		}
		if verificationReceipt.ChainID != promotion.PromotionReceipt.ChainID {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s chain_id mismatch", path))
		}
		if verificationReceipt.PolicyPath != promotion.PromotionReceipt.PolicyPath {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s policy_path mismatch", path))
		}
		if verificationReceipt.CurrentPolicyVersion != promotion.PromotionReceipt.CurrentPolicyVersion {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s current_policy_version mismatch", path))
		}
		if verificationReceipt.CurrentPolicyDigest != promotion.PromotionReceipt.CurrentPolicyDigest {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s current_policy_digest mismatch", path))
		}
		if verificationReceipt.ArchiveIndexDigest != promotion.PromotionReceipt.ArchiveIndexDigest {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s archive_index_digest mismatch", path))
		}
		if verificationReceipt.ArchiveEntryDigest != promotion.PromotionReceipt.ArchiveEntryDigest {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s archive_entry_digest mismatch", path))
		}
		if existing, exists := seenPolicyVersions[entry.CurrentPolicyVersion]; exists && entry.CurrentPolicyVersion != "" {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("duplicate retained inventory current_policy_version %s in %s and %s", entry.CurrentPolicyVersion, existing, path))
		} else if entry.CurrentPolicyVersion != "" {
			seenPolicyVersions[entry.CurrentPolicyVersion] = path
		}
		if existing, exists := seenPromotionReceipts[entry.PromotionReceiptID]; exists && entry.PromotionReceiptID != "" {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("duplicate retained inventory promotion_receipt_id %s in %s and %s", entry.PromotionReceiptID, existing, path))
		} else if entry.PromotionReceiptID != "" {
			seenPromotionReceipts[entry.PromotionReceiptID] = path
		}
		if existing, exists := seenAttestations[entry.AttestationID]; exists && entry.AttestationID != "" {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("duplicate retained inventory attestation_id %s in %s and %s", entry.AttestationID, existing, path))
		} else if entry.AttestationID != "" {
			seenAttestations[entry.AttestationID] = path
		}
		if existing, exists := seenVerificationReceipts[entry.VerificationReceiptID]; exists && entry.VerificationReceiptID != "" {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("duplicate retained inventory verification_receipt_id %s in %s and %s", entry.VerificationReceiptID, existing, path))
		} else if entry.VerificationReceiptID != "" {
			seenVerificationReceipts[entry.VerificationReceiptID] = path
		}
		if snapshot.ChainID == "" {
			snapshot.ChainID = entry.ChainID
		} else if entry.ChainID != "" && snapshot.ChainID != entry.ChainID {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s chain_id mismatch", path))
		}
		if snapshot.PolicyPath == "" {
			snapshot.PolicyPath = entry.PolicyPath
		} else if entry.PolicyPath != "" && snapshot.PolicyPath != entry.PolicyPath {
			snapshot.Issues = append(snapshot.Issues, fmt.Sprintf("retained inventory package %s policy_path mismatch", path))
		}
		if entry.Verified {
			snapshot.VerifiedCount++
		}
		ordered = append(ordered, orderedEntry{
			entry:   entry,
			sortKey: entry.PromotedAt + "|" + entry.CurrentPolicyVersion,
		})
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].sortKey == ordered[j].sortKey {
			return ordered[i].entry.PromotionPath < ordered[j].entry.PromotionPath
		}
		return ordered[i].sortKey < ordered[j].sortKey
	})
	for _, item := range ordered {
		snapshot.Entries = append(snapshot.Entries, item.entry)
	}
	snapshot.PackageCount = len(snapshot.Entries)
	if len(snapshot.Entries) > 0 {
		latest := snapshot.Entries[len(snapshot.Entries)-1]
		snapshot.LatestCurrentPolicyVersion = latest.CurrentPolicyVersion
		snapshot.LatestCurrentPolicyDigest = latest.CurrentPolicyDigest
	}
	if len(snapshot.Issues) > 0 {
		snapshot.Status = "invalid"
		return snapshot, nil
	}
	receiptValues := make([]string, 0, 4+len(snapshot.Entries))
	receiptValues = append(receiptValues,
		snapshot.ChainID,
		snapshot.PolicyPath,
		snapshot.LatestCurrentPolicyVersion,
		snapshot.LatestCurrentPolicyDigest,
		fmt.Sprintf("%d", snapshot.PackageCount),
		fmt.Sprintf("%d", snapshot.VerifiedCount),
	)
	for _, entry := range snapshot.Entries {
		receiptValues = append(receiptValues, entry.VerificationReceiptID)
	}
	snapshot.SnapshotReceiptID = deterministicArtifactID("retained-inventory-snapshot", receiptValues...)
	snapshot.Status = "consistent"
	return snapshot, nil
}

func VerifySignerRotationActivationAuditRetainedInventorySnapshot(
	request SignerRotationActivationAuditRetainedInventoryVerificationRequest,
) (SignerRotationActivationAuditRetainedInventoryVerificationReceipt, error) {
	verifiedAt, err := parseRFC3339(request.VerifiedAt, "retained inventory verification verified_at")
	if err != nil {
		return SignerRotationActivationAuditRetainedInventoryVerificationReceipt{}, err
	}
	verifiedBy := strings.TrimSpace(request.VerifiedBy)
	if verifiedBy == "" {
		return SignerRotationActivationAuditRetainedInventoryVerificationReceipt{}, fmt.Errorf("retained inventory verification verified_by must be set")
	}

	receipt := SignerRotationActivationAuditRetainedInventoryVerificationReceipt{
		Version:            "1.0.0",
		Status:             "invalid",
		VerifiedAt:         verifiedAt.Format(time.RFC3339),
		VerifiedBy:         verifiedBy,
		VerificationIssues: []string{},
	}

	expectedSnapshot, err := BuildSignerRotationActivationAuditRetainedInventorySnapshot(SignerRotationActivationAuditRetainedInventorySnapshotRequest{
		Packages: request.Packages,
	})
	if err != nil {
		return SignerRotationActivationAuditRetainedInventoryVerificationReceipt{}, err
	}
	snapshotDigest, err := digestJSON(request.Snapshot, "retained inventory snapshot")
	if err != nil {
		return SignerRotationActivationAuditRetainedInventoryVerificationReceipt{}, err
	}
	expectedSnapshotDigest, err := digestJSON(expectedSnapshot, "expected retained inventory snapshot")
	if err != nil {
		return SignerRotationActivationAuditRetainedInventoryVerificationReceipt{}, err
	}

	receipt.ChainID = request.Snapshot.ChainID
	receipt.PolicyPath = request.Snapshot.PolicyPath
	receipt.PackageCount = request.Snapshot.PackageCount
	receipt.VerifiedCount = request.Snapshot.VerifiedCount
	receipt.LatestCurrentPolicyVersion = request.Snapshot.LatestCurrentPolicyVersion
	receipt.LatestCurrentPolicyDigest = request.Snapshot.LatestCurrentPolicyDigest
	receipt.InventorySnapshotDigest = snapshotDigest
	receipt.ExpectedInventorySnapshotDigest = expectedSnapshotDigest

	if snapshotDigest != expectedSnapshotDigest {
		receipt.VerificationIssues = append(receipt.VerificationIssues, "retained inventory snapshot drift detected")
	}
	if expectedSnapshot.Status != "consistent" {
		receipt.VerificationIssues = append(receipt.VerificationIssues, "expected retained inventory snapshot is not consistent")
	}
	if request.Snapshot.Status != "consistent" {
		receipt.VerificationIssues = append(receipt.VerificationIssues, "retained inventory snapshot status must be consistent")
	}
	if len(request.Snapshot.Issues) > 0 {
		receipt.VerificationIssues = append(receipt.VerificationIssues, "retained inventory snapshot carries unresolved issues")
	}

	if len(receipt.VerificationIssues) > 0 {
		return receipt, nil
	}
	receipt.VerificationReceiptID = deterministicArtifactID(
		"retained-inventory-verification",
		snapshotDigest,
		receipt.LatestCurrentPolicyVersion,
		receipt.VerifiedAt,
		receipt.VerifiedBy,
	)
	receipt.Status = "verified"
	return receipt, nil
}

func BuildSignerRotationActivationAuditRetainedInventoryContinuityManifest(
	request SignerRotationActivationAuditRetainedInventoryContinuityManifestRequest,
) (SignerRotationActivationAuditRetainedInventoryContinuityManifest, error) {
	if len(request.Snapshots) == 0 {
		return SignerRotationActivationAuditRetainedInventoryContinuityManifest{}, fmt.Errorf("retained inventory continuity manifest requires at least one inventory snapshot")
func BuildSignerRotationActivationAuditRetainedInventoryContinuityManifest(
	request SignerRotationActivationAuditRetainedInventoryContinuityRequest,
) (SignerRotationActivationAuditRetainedInventoryContinuityManifest, error) {
	if len(request.Snapshots) == 0 {
		return SignerRotationActivationAuditRetainedInventoryContinuityManifest{}, fmt.Errorf("continuity manifest requires at least one retained inventory snapshot")
	}
	manifest := SignerRotationActivationAuditRetainedInventoryContinuityManifest{
		Version: "1.0.0",
		Status:  "continuous",
		Entries: []SignerRotationActivationAuditRetainedInventoryContinuityEntry{},
		Issues:  []string{},
	}
	type orderedContinuityEntry struct {
		entry    SignerRotationActivationAuditRetainedInventoryContinuityEntry
		snapshot SignerRotationActivationAuditRetainedInventorySnapshot
		sortKey  string
	}
	ordered := make([]orderedContinuityEntry, 0, len(request.Snapshots))
	seenSnapshotPaths := make(map[string]struct{}, len(request.Snapshots))
	seenReceiptIDs := make(map[string]string, len(request.Snapshots))
	seenSnapshotDigests := make(map[string]string, len(request.Snapshots))
	seenLatestVersions := make(map[string]string, len(request.Snapshots))

	for _, item := range request.Snapshots {
		path := strings.TrimSpace(item.SnapshotPath)
		if path == "" {
			manifest.Issues = append(manifest.Issues, "retained inventory continuity snapshot_path must be set")
			continue
		}
		if _, exists := seenSnapshotPaths[path]; exists {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("duplicate retained inventory snapshot_path %s", path))
		}
		seenSnapshotPaths[path] = struct{}{}

		snapshotDigest, err := digestJSON(item.Snapshot, "retained inventory snapshot")
		if err != nil {
			return SignerRotationActivationAuditRetainedInventoryContinuityManifest{}, err
		}
		receiptDigest, err := digestJSON(item.VerificationReceipt, "retained inventory verification receipt")
		if err != nil {
			return SignerRotationActivationAuditRetainedInventoryContinuityManifest{}, err
		}
		entry := SignerRotationActivationAuditRetainedInventoryContinuityEntry{
			SnapshotPath:               path,
			Status:                     item.VerificationReceipt.Status,
			Verified:                   item.VerificationReceipt.Status == "verified",
			VerificationReceiptID:      item.VerificationReceipt.VerificationReceiptID,
			InventorySnapshotDigest:    snapshotDigest,
			VerificationReceiptDigest:  receiptDigest,
			PackageCount:               item.Snapshot.PackageCount,
			VerifiedCount:              item.Snapshot.VerifiedCount,
			ChainID:                    item.Snapshot.ChainID,
			PolicyPath:                 item.Snapshot.PolicyPath,
			LatestCurrentPolicyVersion: item.Snapshot.LatestCurrentPolicyVersion,
			LatestCurrentPolicyDigest:  item.Snapshot.LatestCurrentPolicyDigest,
			VerifiedAt:                 item.VerificationReceipt.VerifiedAt,
			VerifiedBy:                 item.VerificationReceipt.VerifiedBy,
		}

		expectedReceiptID := deterministicArtifactID(
			"retained-inventory-verification",
			snapshotDigest,
			item.Snapshot.LatestCurrentPolicyVersion,
			item.VerificationReceipt.VerifiedAt,
			item.VerificationReceipt.VerifiedBy,
		)
		if item.Snapshot.Version != "1.0.0" {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s has unexpected version %s", path, item.Snapshot.Version))
		}
		if item.Snapshot.Status != "consistent" {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s status must be consistent", path))
		}
		if len(item.Snapshot.Issues) > 0 {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s carries unresolved issues", path))
		}
		if len(item.Snapshot.Entries) != item.Snapshot.PackageCount {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s package_count does not match entries", path))
		}
		actualVerifiedCount := 0
		for _, retainedEntry := range item.Snapshot.Entries {
			if retainedEntry.Verified {
				actualVerifiedCount++
			}
		}
		if actualVerifiedCount != item.Snapshot.VerifiedCount {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s verified_count does not match entries", path))
		}
		if item.VerificationReceipt.Version != "1.0.0" {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s verification receipt has unexpected version %s", path, item.VerificationReceipt.Version))
		}
		if item.VerificationReceipt.Status != "verified" {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s is not verified", path))
		}
		if _, err := parseRFC3339(item.VerificationReceipt.VerifiedAt, "retained inventory verification verified_at"); err != nil {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s verification receipt verified_at is invalid", path))
		}
		if strings.TrimSpace(item.VerificationReceipt.VerifiedBy) == "" {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s verification receipt verified_by must be set", path))
		}
		if item.VerificationReceipt.VerificationReceiptID != expectedReceiptID {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s verification_receipt_id mismatch", path))
		}
		if item.VerificationReceipt.InventorySnapshotDigest != snapshotDigest {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s digest mismatch", path))
		}
		if item.VerificationReceipt.ExpectedInventorySnapshotDigest != snapshotDigest {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s expected digest mismatch", path))
		}
		if item.VerificationReceipt.PackageCount != item.Snapshot.PackageCount {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s package_count mismatch", path))
		}
		if item.VerificationReceipt.VerifiedCount != item.Snapshot.VerifiedCount {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s verified_count mismatch", path))
		}
		if item.VerificationReceipt.ChainID != item.Snapshot.ChainID {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s chain_id mismatch", path))
		}
		if item.VerificationReceipt.PolicyPath != item.Snapshot.PolicyPath {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s policy_path mismatch", path))
		}
		if item.VerificationReceipt.LatestCurrentPolicyVersion != item.Snapshot.LatestCurrentPolicyVersion {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s latest_current_policy_version mismatch", path))
		}
		if item.VerificationReceipt.LatestCurrentPolicyDigest != item.Snapshot.LatestCurrentPolicyDigest {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s latest_current_policy_digest mismatch", path))
		}
		if len(item.VerificationReceipt.VerificationIssues) > 0 {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s verification receipt carries issues", path))
		}
		if existing, exists := seenReceiptIDs[entry.VerificationReceiptID]; exists && entry.VerificationReceiptID != "" {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("duplicate retained inventory verification_receipt_id %s in %s and %s", entry.VerificationReceiptID, existing, path))
		} else if entry.VerificationReceiptID != "" {
			seenReceiptIDs[entry.VerificationReceiptID] = path
		}
		if existing, exists := seenSnapshotDigests[snapshotDigest]; exists {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("duplicate retained inventory snapshot digest in %s and %s", existing, path))
		} else {
			seenSnapshotDigests[snapshotDigest] = path
		}
		if existing, exists := seenLatestVersions[entry.LatestCurrentPolicyVersion]; exists && entry.LatestCurrentPolicyVersion != "" {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("duplicate retained inventory latest_current_policy_version %s in %s and %s", entry.LatestCurrentPolicyVersion, existing, path))
		} else if entry.LatestCurrentPolicyVersion != "" {
			seenLatestVersions[entry.LatestCurrentPolicyVersion] = path
		}
		if manifest.ChainID == "" {
			manifest.ChainID = entry.ChainID
		} else if entry.ChainID != "" && manifest.ChainID != entry.ChainID {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s chain_id mismatch", path))
		}
		if manifest.PolicyPath == "" {
			manifest.PolicyPath = entry.PolicyPath
		} else if entry.PolicyPath != "" && manifest.PolicyPath != entry.PolicyPath {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s policy_path mismatch", path))
		}
		if entry.Verified {
			manifest.VerifiedSnapshotCount++
		}
		ordered = append(ordered, orderedContinuityEntry{
			entry:    entry,
			snapshot: item.Snapshot,
			sortKey:  entry.VerifiedAt + "|" + path,
		})
	}

	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].sortKey == ordered[j].sortKey {
			return ordered[i].entry.SnapshotPath < ordered[j].entry.SnapshotPath
		}
		return ordered[i].sortKey < ordered[j].sortKey
	})
	var previous *orderedContinuityEntry
	for idx := range ordered {
		current := ordered[idx]
		if previous != nil {
			if current.entry.PackageCount < previous.entry.PackageCount {
				manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s package_count regressed from %s", current.entry.SnapshotPath, previous.entry.SnapshotPath))
			}
			currentEntries := retainedInventoryEntriesByReceipt(current.snapshot)
			for _, previousEntry := range previous.snapshot.Entries {
				currentEntry, exists := currentEntries[previousEntry.PromotionReceiptID]
				if !exists {
					manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s dropped promotion_receipt_id %s from %s", current.entry.SnapshotPath, previousEntry.PromotionReceiptID, previous.entry.SnapshotPath))
					continue
				}
				previousDigest, err := digestJSON(previousEntry, "previous retained inventory entry")
				if err != nil {
					return SignerRotationActivationAuditRetainedInventoryContinuityManifest{}, err
				}
				currentDigest, err := digestJSON(currentEntry, "current retained inventory entry")
				if err != nil {
					return SignerRotationActivationAuditRetainedInventoryContinuityManifest{}, err
				}
				if previousDigest != currentDigest {
					manifest.Issues = append(manifest.Issues, fmt.Sprintf("retained inventory snapshot %s changed promotion_receipt_id %s from %s", current.entry.SnapshotPath, previousEntry.PromotionReceiptID, previous.entry.SnapshotPath))
				}
			}
		}
		manifest.Entries = append(manifest.Entries, current.entry)
		previous = &ordered[idx]
	}
	manifest.SnapshotCount = len(manifest.Entries)
	if len(manifest.Entries) > 0 {
		latest := manifest.Entries[len(manifest.Entries)-1]
	seenReceiptIDs := make(map[string]int, len(request.Snapshots))
	manifestReceiptValues := make([]string, 0, len(request.Snapshots))
	for idx, snapshot := range request.Snapshots {
		entry := SignerRotationActivationAuditRetainedInventoryContinuityEntry{
			SnapshotReceiptID:          snapshot.SnapshotReceiptID,
			PackageCount:               snapshot.PackageCount,
			VerifiedCount:              snapshot.VerifiedCount,
			LatestCurrentPolicyVersion: snapshot.LatestCurrentPolicyVersion,
			LatestCurrentPolicyDigest:  snapshot.LatestCurrentPolicyDigest,
			Status:                     snapshot.Status,
		}
		if snapshot.Status != "consistent" {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("continuity snapshot %d status must be consistent, got %s", idx, snapshot.Status))
		}
		if snapshot.SnapshotReceiptID == "" {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("continuity snapshot %d is missing snapshot_receipt_id", idx))
		} else if prior, exists := seenReceiptIDs[snapshot.SnapshotReceiptID]; exists {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("duplicate snapshot_receipt_id %s in snapshot %d and snapshot %d", snapshot.SnapshotReceiptID, prior, idx))
		} else {
			seenReceiptIDs[snapshot.SnapshotReceiptID] = idx
		}
		if snapshot.ChainID == "" {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("continuity snapshot %d is missing chain_id", idx))
		}
		if manifest.ChainID == "" {
			manifest.ChainID = snapshot.ChainID
		} else if manifest.ChainID != snapshot.ChainID {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("continuity snapshot %d chain_id mismatch", idx))
		}
		if snapshot.PolicyPath == "" {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("continuity snapshot %d is missing policy_path", idx))
		}
		if manifest.PolicyPath == "" {
			manifest.PolicyPath = snapshot.PolicyPath
		} else if manifest.PolicyPath != snapshot.PolicyPath {
			manifest.Issues = append(manifest.Issues, fmt.Sprintf("continuity snapshot %d policy_path mismatch", idx))
		}
		if idx > 0 {
			prev := request.Snapshots[idx-1]
			if snapshot.PackageCount < prev.PackageCount {
				manifest.Issues = append(manifest.Issues, fmt.Sprintf("continuity snapshot %d package_count regressed from %d to %d", idx, prev.PackageCount, snapshot.PackageCount))
			}
			if snapshot.VerifiedCount < prev.VerifiedCount {
				manifest.Issues = append(manifest.Issues, fmt.Sprintf("continuity snapshot %d verified_count regressed from %d to %d", idx, prev.VerifiedCount, snapshot.VerifiedCount))
			}
			if snapshot.LatestCurrentPolicyVersion != "" && prev.LatestCurrentPolicyVersion != "" &&
				snapshot.LatestCurrentPolicyVersion < prev.LatestCurrentPolicyVersion {
				manifest.Issues = append(manifest.Issues, fmt.Sprintf("continuity snapshot %d latest_current_policy_version regressed from %s to %s", idx, prev.LatestCurrentPolicyVersion, snapshot.LatestCurrentPolicyVersion))
			}
		}
		manifestReceiptValues = append(manifestReceiptValues, snapshot.SnapshotReceiptID)
		manifest.Entries = append(manifest.Entries, entry)
	}
	manifest.SnapshotCount = len(manifest.Entries)
	if manifest.SnapshotCount > 0 {
		latest := manifest.Entries[manifest.SnapshotCount-1]
		manifest.LatestCurrentPolicyVersion = latest.LatestCurrentPolicyVersion
		manifest.LatestCurrentPolicyDigest = latest.LatestCurrentPolicyDigest
	}
	if len(manifest.Issues) > 0 {
		manifest.Status = "invalid"
		return manifest, nil
	}
	manifest.ManifestID = deterministicArtifactID("retained-inventory-continuity", manifestReceiptValues...)
	manifest.Status = "continuous"
	return manifest, nil
}

func retainedInventoryEntriesByReceipt(snapshot SignerRotationActivationAuditRetainedInventorySnapshot) map[string]SignerRotationActivationAuditRetainedInventorySnapshotEntry {
	entries := make(map[string]SignerRotationActivationAuditRetainedInventorySnapshotEntry, len(snapshot.Entries))
	for _, entry := range snapshot.Entries {
		key := entry.PromotionReceiptID
		if key == "" {
			key = entry.PromotionPath
		}
		entries[key] = entry
	}
	return entries
}

func recommendedRotationAction(rotationStatus string) string {
	switch rotationStatus {
	case "expiring":
		return "rotate signer key before the current rotate_by deadline and publish the replacement manifest"
	case "inactive":
		return "retain retired signer metadata for audit history only"
	default:
		return "maintain signer custody and validate the next scheduled rotation checkpoint"
	}
}

func commaList(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return fmt.Sprintf("%s", sortAndJoin(values))
}

func sortAndJoin(values []string) string {
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	return joinStrings(sorted, ", ")
}

func joinStrings(values []string, separator string) string {
	switch len(values) {
	case 0:
		return ""
	case 1:
		return values[0]
	}
	output := values[0]
	for _, value := range values[1:] {
		output += separator + value
	}
	return output
}
