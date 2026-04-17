package project

import (
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
