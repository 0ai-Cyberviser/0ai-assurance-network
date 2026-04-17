package project

import (
	"fmt"
	"sort"
)

type IdentityRolePlan struct {
	Role         string   `json:"role"`
	BoundActors  []string `json:"bound_actors"`
	ReferencedBy []string `json:"referenced_by"`
}

type IdentityFoundationPlanOutput struct {
	Version          string                `json:"version"`
	ChainID          string                `json:"chain_id"`
	Milestone        string                `json:"milestone"`
	ActorCount       int                   `json:"actor_count"`
	ActiveActorCount int                   `json:"active_actor_count"`
	RequiredRoles    []string              `json:"required_roles"`
	BoundRoles       []string              `json:"bound_roles"`
	MissingRoles     []string              `json:"missing_roles"`
	Actors           []IdentityActor       `json:"actors"`
	RoleBindings     []IdentityRoleBinding `json:"role_bindings"`
	RolePlan         []IdentityRolePlan    `json:"role_plan"`
}

func stringMap(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return map[string]any{}
}

func stringSlice(value any) []string {
	raw, ok := value.([]any)
	if !ok {
		return nil
	}
	items := make([]string, 0, len(raw))
	for _, entry := range raw {
		text := fmt.Sprint(entry)
		if text != "" {
			items = append(items, text)
		}
	}
	return items
}

func appendIdentityRoleUsage(roleUsage map[string][]string, role string, source string) {
	if role == "" {
		return
	}
	for _, existing := range roleUsage[role] {
		if existing == source {
			return
		}
	}
	roleUsage[role] = append(roleUsage[role], source)
	sort.Strings(roleUsage[role])
}

func requiredIdentityRoles(bundle Bundle) (map[string][]string, []string) {
	roleUsage := make(map[string][]string)

	recordBoundary := func(prefix string, boundary ModuleBoundary) {
		source := prefix + ":" + boundary.Name
		for _, tx := range boundary.Transactions {
			for _, role := range tx.ActorRoles {
				appendIdentityRoleUsage(roleUsage, role, source)
			}
		}
		for _, permission := range boundary.OperatorPermissions {
			appendIdentityRoleUsage(roleUsage, permission.Role, source)
		}
	}

	for _, boundary := range bundle.Modules.MVPModules {
		recordBoundary("mvp", boundary)
	}
	for _, boundary := range bundle.Modules.DependencySurfaces {
		recordBoundary("dependency", boundary)
	}
	executionDefaults := stringMap(stringMap(stringMap(bundle.InferencePolicy["remediation"])["execution_defaults"]))
	for phase, role := range stringMap(executionDefaults["phase_owners"]) {
		appendIdentityRoleUsage(roleUsage, fmt.Sprint(role), "governance:phase_owner:"+phase)
	}
	for proposalKind, override := range stringMap(executionDefaults["owner_overrides"]) {
		for phase, role := range stringMap(override) {
			appendIdentityRoleUsage(roleUsage, fmt.Sprint(role), "governance:owner_override:"+proposalKind+":"+phase)
		}
	}
	if rawSigners, ok := bundle.CheckpointSigners["signers"].([]any); ok {
		for _, rawSigner := range rawSigners {
			signer := stringMap(rawSigner)
			signerID := fmt.Sprint(signer["signer_id"])
			for _, role := range stringSlice(signer["roles"]) {
				appendIdentityRoleUsage(roleUsage, role, "governance:signer:"+signerID)
			}
		}
	}

	roles := make([]string, 0, len(roleUsage))
	for role := range roleUsage {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	return roleUsage, roles
}

func ValidateIdentityBootstrap(bundle Bundle) error {
	if bundle.Identity.Version == "" {
		return fmt.Errorf("identity bootstrap version must be set")
	}
	if bundle.Identity.ChainID != bundle.Topology.ChainID {
		return fmt.Errorf(
			"identity bootstrap chain id %s does not match topology chain id %s",
			bundle.Identity.ChainID,
			bundle.Topology.ChainID,
		)
	}
	if len(bundle.Identity.Actors) == 0 {
		return fmt.Errorf("identity bootstrap must declare actors")
	}
	if len(bundle.Identity.RoleBindings) == 0 {
		return fmt.Errorf("identity bootstrap must declare role bindings")
	}

	allowedActorTypes := map[string]struct{}{
		"organization":    {},
		"operator":        {},
		"council":         {},
		"service_account": {},
	}
	allowedStatus := map[string]struct{}{
		"active":   {},
		"inactive": {},
	}

	actors := make(map[string]IdentityActor, len(bundle.Identity.Actors))
	for _, actor := range bundle.Identity.Actors {
		if actor.ActorID == "" {
			return fmt.Errorf("identity actor id must not be empty")
		}
		if _, exists := actors[actor.ActorID]; exists {
			return fmt.Errorf("duplicate identity actor id: %s", actor.ActorID)
		}
		if _, ok := allowedActorTypes[actor.ActorType]; !ok {
			return fmt.Errorf("identity actor %s has invalid actor type %s", actor.ActorID, actor.ActorType)
		}
		if _, ok := allowedStatus[actor.Status]; !ok {
			return fmt.Errorf("identity actor %s has invalid status %s", actor.ActorID, actor.Status)
		}
		if actor.DisplayName == "" {
			return fmt.Errorf("identity actor %s must declare a display name", actor.ActorID)
		}
		actors[actor.ActorID] = actor
	}
	for _, actor := range bundle.Identity.Actors {
		if actor.OrganizationID == "" {
			continue
		}
		organization, exists := actors[actor.OrganizationID]
		if !exists {
			return fmt.Errorf(
				"identity actor %s references unknown organization %s",
				actor.ActorID,
				actor.OrganizationID,
			)
		}
		if organization.ActorType != "organization" && organization.ActorType != "council" {
			return fmt.Errorf(
				"identity actor %s organization %s must be organization or council",
				actor.ActorID,
				actor.OrganizationID,
			)
		}
	}

	roleUsage, requiredRoles := requiredIdentityRoles(bundle)
	boundRoles := make(map[string]struct{})
	activeActorRoleBindings := make(map[string]struct{})
	seenBindings := make(map[string]struct{}, len(bundle.Identity.RoleBindings))
	for _, binding := range bundle.Identity.RoleBindings {
		if binding.ActorID == "" || binding.Role == "" || binding.Scope == "" || binding.GrantedBy == "" {
			return fmt.Errorf("identity role binding must declare actor_id, role, scope, and granted_by")
		}
		if _, exists := actors[binding.ActorID]; !exists {
			return fmt.Errorf("identity role binding references unknown actor %s", binding.ActorID)
		}
		if _, ok := allowedStatus[binding.Status]; !ok {
			return fmt.Errorf(
				"identity role binding %s/%s has invalid status %s",
				binding.ActorID,
				binding.Role,
				binding.Status,
			)
		}
		if _, ok := roleUsage[binding.Role]; !ok {
			return fmt.Errorf("identity role binding uses undeclared role %s", binding.Role)
		}
		bindingKey := binding.ActorID + "|" + binding.Role + "|" + binding.Scope
		if _, exists := seenBindings[bindingKey]; exists {
			return fmt.Errorf("duplicate identity role binding: %s", bindingKey)
		}
		seenBindings[bindingKey] = struct{}{}
		if binding.Status == "active" {
			boundRoles[binding.Role] = struct{}{}
			activeActorRoleBindings[binding.ActorID+"|"+binding.Role] = struct{}{}
		}
	}
	for _, role := range requiredRoles {
		if _, exists := boundRoles[role]; !exists {
			return fmt.Errorf("identity bootstrap missing active binding for required role %s", role)
		}
	}
	if rawSigners, ok := bundle.CheckpointSigners["signers"].([]any); ok {
		for _, rawSigner := range rawSigners {
			signer := stringMap(rawSigner)
			signerID := fmt.Sprint(signer["signer_id"])
			actorID := fmt.Sprint(signer["actor_id"])
			if actorID == "" {
				return fmt.Errorf("checkpoint signer %s must declare actor_id", signerID)
			}
			actor, exists := actors[actorID]
			if !exists {
				return fmt.Errorf("checkpoint signer %s references unknown actor %s", signerID, actorID)
			}
			if actor.Status != "active" {
				return fmt.Errorf("checkpoint signer %s references inactive actor %s", signerID, actorID)
			}
			for _, role := range stringSlice(signer["roles"]) {
				if _, exists := activeActorRoleBindings[actorID+"|"+role]; !exists {
					return fmt.Errorf(
						"checkpoint signer %s role %s is not backed by an active identity binding",
						signerID,
						role,
					)
				}
			}
		}
	}
	return nil
}

func IdentityFoundationPlan(bundle Bundle) (IdentityFoundationPlanOutput, error) {
	if err := ValidateIdentityBootstrap(bundle); err != nil {
		return IdentityFoundationPlanOutput{}, err
	}

	roleUsage, requiredRoles := requiredIdentityRoles(bundle)
	activeActors := 0
	for _, actor := range bundle.Identity.Actors {
		if actor.Status == "active" {
			activeActors++
		}
	}

	roleBindings := append([]IdentityRoleBinding(nil), bundle.Identity.RoleBindings...)
	sort.Slice(roleBindings, func(i, j int) bool {
		if roleBindings[i].Role == roleBindings[j].Role {
			if roleBindings[i].ActorID == roleBindings[j].ActorID {
				return roleBindings[i].Scope < roleBindings[j].Scope
			}
			return roleBindings[i].ActorID < roleBindings[j].ActorID
		}
		return roleBindings[i].Role < roleBindings[j].Role
	})

	actors := append([]IdentityActor(nil), bundle.Identity.Actors...)
	sort.Slice(actors, func(i, j int) bool {
		return actors[i].ActorID < actors[j].ActorID
	})

	boundRolesSet := make(map[string][]string)
	for _, binding := range roleBindings {
		if binding.Status != "active" {
			continue
		}
		boundRolesSet[binding.Role] = append(boundRolesSet[binding.Role], binding.ActorID)
	}

	boundRoles := make([]string, 0, len(boundRolesSet))
	rolePlan := make([]IdentityRolePlan, 0, len(requiredRoles))
	missingRoles := make([]string, 0)
	for _, role := range requiredRoles {
		actorsForRole := boundRolesSet[role]
		sort.Strings(actorsForRole)
		if len(actorsForRole) == 0 {
			missingRoles = append(missingRoles, role)
		} else {
			boundRoles = append(boundRoles, role)
		}
		rolePlan = append(rolePlan, IdentityRolePlan{
			Role:         role,
			BoundActors:  actorsForRole,
			ReferencedBy: roleUsage[role],
		})
	}
	sort.Strings(boundRoles)

	return IdentityFoundationPlanOutput{
		Version:          bundle.Identity.Version,
		ChainID:          bundle.Identity.ChainID,
		Milestone:        bundle.Modules.Milestone,
		ActorCount:       len(actors),
		ActiveActorCount: activeActors,
		RequiredRoles:    requiredRoles,
		BoundRoles:       boundRoles,
		MissingRoles:     missingRoles,
		Actors:           actors,
		RoleBindings:     roleBindings,
		RolePlan:         rolePlan,
	}, nil
}
