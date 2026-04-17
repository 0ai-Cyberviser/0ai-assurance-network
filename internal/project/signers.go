package project

import (
	"fmt"
	"sort"
	"strconv"
	"time"
)

type SignerRoleCoverage struct {
	Role         string   `json:"role"`
	ReferencedBy []string `json:"referenced_by"`
}

type SignerManifestEntry struct {
	ActorID           string               `json:"actor_id"`
	ActorDisplayName  string               `json:"actor_display_name"`
	ActorStatus       string               `json:"actor_status"`
	OrganizationID    string               `json:"organization_id,omitempty"`
	OrganizationName  string               `json:"organization_name,omitempty"`
	SignerID          string               `json:"signer_id"`
	KeyID             string               `json:"key_id"`
	SignerStatus      string               `json:"signer_status"`
	Roles             []string             `json:"roles"`
	RoleCoverage      []SignerRoleCoverage `json:"role_coverage"`
	ProvisionedAt     string               `json:"provisioned_at"`
	RotateBy          string               `json:"rotate_by"`
	RotationStatus    string               `json:"rotation_status"`
	DaysUntilRotation int                  `json:"days_until_rotation"`
}

type SignerRotationPlanEntry struct {
	Order             int      `json:"order"`
	SignerID          string   `json:"signer_id"`
	ActorID           string   `json:"actor_id"`
	Roles             []string `json:"roles"`
	RotateBy          string   `json:"rotate_by"`
	RotationStatus    string   `json:"rotation_status"`
	DaysUntilRotation int      `json:"days_until_rotation"`
	RecommendedAction string   `json:"recommended_action"`
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

	roleUsage := governanceExecutionRoleUsage(bundle)
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

	rawSigners, ok := bundle.CheckpointSigners["signers"].([]any)
	if !ok || len(rawSigners) == 0 {
		return fmt.Errorf("at least one checkpoint signer must be configured")
	}

	signerIDs := make(map[string]struct{}, len(rawSigners))
	keyIDs := make(map[string]struct{}, len(rawSigners))
	activeActorIDs := make(map[string]struct{}, len(rawSigners))
	activeRoleCoverage := make(map[string]string)

	for _, rawSigner := range rawSigners {
		signer := stringMap(rawSigner)
		actorID := fmt.Sprint(signer["actor_id"])
		signerID := fmt.Sprint(signer["signer_id"])
		keyID := fmt.Sprint(signer["key_id"])
		sharedSecret := fmt.Sprint(signer["shared_secret"])
		status := fmt.Sprint(signer["status"])
		roles := stringSlice(signer["roles"])

		if actorID == "" {
			return fmt.Errorf("checkpoint signer %s must declare actor_id", signerID)
		}
		if signerID == "" {
			return fmt.Errorf("checkpoint signer signer_id must not be empty")
		}
		if _, exists := signerIDs[signerID]; exists {
			return fmt.Errorf("duplicate checkpoint signer_id: %s", signerID)
		}
		if _, exists := keyIDs[keyID]; exists {
			return fmt.Errorf("duplicate checkpoint key_id: %s", keyID)
		}
		if sharedSecret == "" {
			return fmt.Errorf("checkpoint signer %s must declare a shared_secret", signerID)
		}
		if status != "active" && status != "inactive" {
			return fmt.Errorf("checkpoint signer %s has invalid status", signerID)
		}
		if len(roles) == 0 {
			return fmt.Errorf("checkpoint signer %s must declare at least one role", signerID)
		}
		provisionedAt, err := parseRFC3339(signer["provisioned_at"], "checkpoint signer "+signerID+" provisioned_at")
		if err != nil {
			return err
		}
		rotateBy, err := parseRFC3339(signer["rotate_by"], "checkpoint signer "+signerID+" rotate_by")
		if err != nil {
			return err
		}
		if !rotateBy.After(provisionedAt) {
			return fmt.Errorf("checkpoint signer %s rotate_by must be after provisioned_at", signerID)
		}

		actor, exists := actors[actorID]
		if !exists {
			return fmt.Errorf("checkpoint signer %s references unknown actor %s", signerID, actorID)
		}
		if actor.Status != "active" {
			return fmt.Errorf("checkpoint signer %s references inactive actor %s", signerID, actorID)
		}

		for _, role := range roles {
			if _, exists := activeRoleBindings[actorID+"|"+role]; !exists {
				return fmt.Errorf(
					"checkpoint signer %s role %s is not backed by an active identity binding",
					signerID,
					role,
				)
			}
		}

		if status == "active" {
			if _, exists := activeActorIDs[actorID]; exists {
				return fmt.Errorf("duplicate checkpoint signer actor ownership: %s", actorID)
			}
			activeActorIDs[actorID] = struct{}{}
			if !rotateBy.After(referenceTime) {
				return fmt.Errorf("checkpoint signer %s has stale rotation metadata", signerID)
			}
			for _, role := range roles {
				if owner, exists := activeRoleCoverage[role]; exists {
					return fmt.Errorf("duplicate checkpoint signer role coverage: %s (%s, %s)", role, owner, signerID)
				}
				activeRoleCoverage[role] = signerID
			}
			if signerRotationStatus(referenceTime, warningWindowDays, status, rotateBy) == "stale" {
				return fmt.Errorf("checkpoint signer %s has stale rotation metadata", signerID)
			}
		}

		signerIDs[signerID] = struct{}{}
		keyIDs[keyID] = struct{}{}
	}

	missingRoles := make([]string, 0)
	for role := range roleUsage {
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

func SignerManifest(bundle Bundle) (SignerManifestOutput, error) {
	if err := ValidateSignerManifestInputs(bundle); err != nil {
		return SignerManifestOutput{}, err
	}

	rotationPolicy := stringMap(bundle.CheckpointSigners["rotation_policy"])
	referenceTime, _ := parseRFC3339(rotationPolicy["reference_time"], "checkpoint signer rotation_policy.reference_time")
	warningWindowDays, _ := intValue(rotationPolicy["warning_window_days"])
	roleUsage := governanceExecutionRoleUsage(bundle)

	actors := make(map[string]IdentityActor, len(bundle.Identity.Actors))
	for _, actor := range bundle.Identity.Actors {
		actors[actor.ActorID] = actor
	}

	entries := make([]SignerManifestEntry, 0)
	expiring := make([]string, 0)
	rotationPlan := make([]SignerRotationPlanEntry, 0)

	rawSigners := bundle.CheckpointSigners["signers"].([]any)
	parsedSigners := make([]parsedSignerEntry, 0, len(rawSigners))
	for _, rawSigner := range rawSigners {
		signer := stringMap(rawSigner)
		provisionedAt, _ := parseRFC3339(signer["provisioned_at"], "checkpoint signer provisioned_at")
		rotateBy, _ := parseRFC3339(signer["rotate_by"], "checkpoint signer rotate_by")
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

	sort.Slice(parsedSigners, func(i, j int) bool {
		if parsedSigners[i].RotateBy.Equal(parsedSigners[j].RotateBy) {
			return parsedSigners[i].SignerID < parsedSigners[j].SignerID
		}
		return parsedSigners[i].RotateBy.Before(parsedSigners[j].RotateBy)
	})

	activeCount := 0
	for _, signer := range parsedSigners {
		actor := actors[signer.ActorID]
		roleCoverage := make([]SignerRoleCoverage, 0, len(signer.Roles))
		for _, role := range signer.Roles {
			referencedBy := append([]string(nil), roleUsage[role]...)
			sort.Strings(referencedBy)
			roleCoverage = append(roleCoverage, SignerRoleCoverage{
				Role:         role,
				ReferencedBy: referencedBy,
			})
		}
		sort.Slice(roleCoverage, func(i, j int) bool {
			return roleCoverage[i].Role < roleCoverage[j].Role
		})

		organizationName := ""
		if actor.OrganizationID != "" {
			if organization, exists := actors[actor.OrganizationID]; exists {
				organizationName = organization.DisplayName
			}
		}

		entry := SignerManifestEntry{
			ActorID:           signer.ActorID,
			ActorDisplayName:  actor.DisplayName,
			ActorStatus:       actor.Status,
			OrganizationID:    actor.OrganizationID,
			OrganizationName:  organizationName,
			SignerID:          signer.SignerID,
			KeyID:             signer.KeyID,
			SignerStatus:      signer.Status,
			Roles:             append([]string(nil), signer.Roles...),
			RoleCoverage:      roleCoverage,
			ProvisionedAt:     signer.ProvisionedAt.Format(time.RFC3339),
			RotateBy:          signer.RotateBy.Format(time.RFC3339),
			RotationStatus:    signer.RotationStatus,
			DaysUntilRotation: signer.DaysUntilRotation,
		}
		entries = append(entries, entry)

		if signer.Status == "active" {
			activeCount++
			rotationPlan = append(rotationPlan, SignerRotationPlanEntry{
				SignerID:          signer.SignerID,
				ActorID:           signer.ActorID,
				Roles:             append([]string(nil), signer.Roles...),
				RotateBy:          signer.RotateBy.Format(time.RFC3339),
				RotationStatus:    signer.RotationStatus,
				DaysUntilRotation: signer.DaysUntilRotation,
				RecommendedAction: recommendedRotationAction(signer.RotationStatus),
			})
			if signer.RotationStatus == "expiring" {
				expiring = append(expiring, signer.SignerID)
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].SignerID < entries[j].SignerID
	})
	sort.Strings(expiring)
	for idx := range rotationPlan {
		rotationPlan[idx].Order = idx + 1
	}

	return SignerManifestOutput{
		Version:             "1.0.0",
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
