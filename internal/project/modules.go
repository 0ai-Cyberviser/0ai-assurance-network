package project

type ModulePlanOutput struct {
	Version                string           `json:"version"`
	Milestone              string           `json:"milestone"`
	Scope                  string           `json:"scope"`
	ChainID                string           `json:"chain_id"`
	NetworkMode            string           `json:"network_mode"`
	GovernanceHouses       []string         `json:"governance_houses"`
	ValidatorCount         int              `json:"validator_count"`
	MVPModules             []ModuleBoundary `json:"mvp_modules"`
	DependencySurfaces     []ModuleBoundary `json:"dependency_surfaces"`
	Rollout                []ModuleRollout  `json:"rollout"`
	ImplementationSequence []string         `json:"implementation_sequence"`
}

func MilestoneModulePlan(bundle Bundle) ModulePlanOutput {
	sequence := make([]string, 0, len(bundle.Modules.Rollout))
	for _, phase := range bundle.Modules.Rollout {
		sequence = append(sequence, phase.Name)
	}

	return ModulePlanOutput{
		Version:                bundle.Modules.Version,
		Milestone:              bundle.Modules.Milestone,
		Scope:                  bundle.Modules.Scope,
		ChainID:                bundle.Topology.ChainID,
		NetworkMode:            bundle.Topology.Mode,
		GovernanceHouses:       bundle.Topology.Governance.Houses,
		ValidatorCount:         len(bundle.Topology.Validators),
		MVPModules:             bundle.Modules.MVPModules,
		DependencySurfaces:     bundle.Modules.DependencySurfaces,
		Rollout:                bundle.Modules.Rollout,
		ImplementationSequence: sequence,
	}
}
