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
