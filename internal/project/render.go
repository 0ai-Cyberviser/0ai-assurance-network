package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Plan struct {
	NetworkName      string          `json:"network_name"`
	ChainID          string          `json:"chain_id"`
	Mode             string          `json:"mode"`
	ValidatorCount   int             `json:"validator_count"`
	SeedCount        int             `json:"seed_count"`
	GovernanceHouses []string        `json:"governance_houses"`
	BaseDenom        string          `json:"base_denom"`
	Validators       []ValidatorNode `json:"validators"`
	ReleaseGuards    []string        `json:"release_guards"`
}

type ValidatorPlanOutput struct {
	ID                string   `json:"id"`
	Moniker           string   `json:"moniker"`
	Role              string   `json:"role"`
	Home              string   `json:"home"`
	RPCAddress        string   `json:"rpc_address"`
	P2PAddress        string   `json:"p2p_address"`
	AppAddress        string   `json:"app_address"`
	PrometheusAddress string   `json:"prometheus_address"`
	PersistentPeers   []string `json:"persistent_peers"`
	ContainerImage    string   `json:"container_image"`
	Binary            string   `json:"binary"`
}

func ModuleMap() map[string]any {
	return map[string]any{
		"execution_model": "appchain",
		"planned_modules": []map[string]string{
			{"name": "identity", "purpose": "validators, operators, auditors, organizations"},
			{"name": "registry", "purpose": "models, agents, services, and safety profiles"},
			{"name": "attest", "purpose": "attestations, reviewer stakes, appeals"},
			{"name": "incident", "purpose": "incident declarations and closure proofs"},
			{"name": "gov", "purpose": "bicameral governance flow"},
			{"name": "safety", "purpose": "policy baselines and emergency controls"},
			{"name": "treasury", "purpose": "fee routing, grants, reserve management"},
			{"name": "slash", "purpose": "slashing and accountability rules"},
			{"name": "reputation", "purpose": "non-transferable operator reputation"},
		},
	}
}

func NetworkPlan(bundle Bundle) Plan {
	return Plan{
		NetworkName:      bundle.Topology.NetworkName,
		ChainID:          bundle.Topology.ChainID,
		Mode:             bundle.Topology.Mode,
		ValidatorCount:   len(bundle.Topology.Validators),
		SeedCount:        len(bundle.Topology.SeedNodes),
		GovernanceHouses: bundle.Topology.Governance.Houses,
		BaseDenom:        bundle.Genesis.Denoms.Base,
		Validators:       bundle.Topology.Validators,
		ReleaseGuards:    bundle.Policy.RequiredBeforePublic,
	}
}

func RenderedGenesis(bundle Bundle) map[string]any {
	rendered := map[string]any{
		"chain_id":    bundle.Genesis.ChainID,
		"launch_mode": bundle.Genesis.LaunchMode,
		"denoms": map[string]any{
			"base":       bundle.Genesis.Denoms.Base,
			"display":    bundle.Genesis.Denoms.Display,
			"reputation": bundle.Genesis.Denoms.Reputation,
		},
		"staking": map[string]any{
			"minimum_self_bond":     bundle.Genesis.Staking.MinimumSelfBond,
			"unbonding_time_seconds": bundle.Genesis.Staking.UnbondingTimeSec,
			"max_validators":        bundle.Genesis.Staking.MaxValidators,
		},
		"governance": map[string]any{
			"proposal_bond":                  bundle.Genesis.Governance.ProposalBond,
			"standard_voting_period_seconds": bundle.Genesis.Governance.StandardVotingPeriodSec,
			"high_impact_timelock_seconds":   bundle.Genesis.Governance.HighImpactTimelockSec,
			"safety_critical_timelock_seconds": bundle.Genesis.Governance.SafetyCriticalSec,
			"dual_house_enabled":             bundle.Genesis.Governance.DualHouseEnabled,
		},
		"attestation": map[string]any{
			"minimum_auditor_bond": bundle.Genesis.Attest.MinimumAuditorBond,
			"appeal_window_seconds": bundle.Genesis.Attest.AppealWindowSec,
		},
		"incident": map[string]any{
			"postmortem_deadline_seconds": bundle.Genesis.Incident.PostmortemDeadlineSec,
			"public_reason_codes_required": bundle.Genesis.Incident.PublicReasonCodes,
		},
		"treasury": map[string]any{
			"fee_split_percent": bundle.Genesis.Treasury.FeeSplitPercent,
		},
		"localnet": map[string]any{
			"validators": validatorNames(bundle.Topology.Validators),
			"seed_nodes": seedNames(bundle.Topology.SeedNodes),
		},
	}
	return rendered
}

func ValidatorPlan(bundle Bundle, validatorID string) (ValidatorPlanOutput, error) {
	var validator *ValidatorNode
	for idx := range bundle.Topology.Validators {
		current := &bundle.Topology.Validators[idx]
		if current.ID == validatorID {
			validator = current
			break
		}
	}
	if validator == nil {
		return ValidatorPlanOutput{}, fmt.Errorf("unknown validator id: %s", validatorID)
	}

	peers := make([]string, 0, len(bundle.Topology.SeedNodes))
	for _, seed := range bundle.Topology.SeedNodes {
		peers = append(peers, fmt.Sprintf("%s@%s:%d", seed.ID, seed.P2PHost, seed.P2PPort))
	}

	return ValidatorPlanOutput{
		ID:                validator.ID,
		Moniker:           validator.Moniker,
		Role:              validator.Role,
		Home:              filepath.ToSlash(filepath.Join("/chain", validator.Moniker)),
		RPCAddress:        fmt.Sprintf("tcp://0.0.0.0:%d", validator.RPCPort),
		P2PAddress:        fmt.Sprintf("tcp://0.0.0.0:%d", validator.P2PPort),
		AppAddress:        fmt.Sprintf("tcp://0.0.0.0:%d", validator.AppPort),
		PrometheusAddress: fmt.Sprintf("tcp://0.0.0.0:%d", validator.PrometheusPort),
		PersistentPeers:   peers,
		ContainerImage:    bundle.Topology.ContainerImage,
		Binary:            bundle.Topology.Binary,
	}, nil
}

func DockerCompose(bundle Bundle) string {
	services := []string{
		"services:",
		"  seed-1:",
		fmt.Sprintf("    image: %s", bundle.Topology.ContainerImage),
		"    container_name: 0ai-seed-1",
		"    command:",
		fmt.Sprintf("      - %s", bundle.Topology.Binary),
		"      - start",
		"      - --home",
		"      - /chain/seed-1",
		"      - --p2p.laddr=tcp://0.0.0.0:26656",
		"    ports:",
		"      - \"26656:26656\"",
		"    volumes:",
		"      - seed-1-data:/chain/seed-1",
	}

	peerTargets := make([]string, 0, len(bundle.Topology.SeedNodes))
	for _, seed := range bundle.Topology.SeedNodes {
		peerTargets = append(peerTargets, fmt.Sprintf("%s@%s:%d", seed.ID, seed.P2PHost, seed.P2PPort))
	}
	peers := strings.Join(peerTargets, ",")

	for _, validator := range bundle.Topology.Validators {
		services = append(
			services,
			fmt.Sprintf("  %s:", validator.Moniker),
			fmt.Sprintf("    image: %s", bundle.Topology.ContainerImage),
			fmt.Sprintf("    container_name: 0ai-%s", validator.Moniker),
			"    command:",
			fmt.Sprintf("      - %s", bundle.Topology.Binary),
			"      - start",
			"      - --home",
			fmt.Sprintf("      - /chain/%s", validator.Moniker),
			"      - --rpc.laddr=tcp://0.0.0.0:26657",
			fmt.Sprintf("      - --p2p.laddr=tcp://0.0.0.0:%d", validator.P2PPort),
			fmt.Sprintf("      - --p2p.persistent_peers=%s", peers),
			"    ports:",
			fmt.Sprintf("      - \"%d:26657\"", validator.RPCPort),
			fmt.Sprintf("      - \"%d:%d\"", validator.P2PPort, validator.P2PPort),
			fmt.Sprintf("      - \"%d:1317\"", validator.AppPort),
			fmt.Sprintf("      - \"%d:26660\"", validator.PrometheusPort),
			"    volumes:",
			fmt.Sprintf("      - %s-data:/chain/%s", validator.Moniker, validator.Moniker),
			"    depends_on:",
			"      - seed-1",
		)
	}

	services = append(services, "volumes:", "  seed-1-data:")
	for _, validator := range bundle.Topology.Validators {
		services = append(services, fmt.Sprintf("  %s-data:", validator.Moniker))
	}

	return strings.Join(services, "\n") + "\n"
}

func WriteJSON(path string, value any) error {
	encoded, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(encoded, '\n'), 0o644)
}

func validatorNames(values []ValidatorNode) []string {
	names := make([]string, 0, len(values))
	for _, value := range values {
		names = append(names, value.Moniker)
	}
	return names
}

func seedNames(values []SeedNode) []string {
	names := make([]string, 0, len(values))
	for _, value := range values {
		names = append(names, value.Moniker)
	}
	return names
}
