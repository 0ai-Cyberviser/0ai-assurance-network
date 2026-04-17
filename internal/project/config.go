package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Bundle struct {
	Root     string
	Topology TopologyConfig
	Genesis  GenesisConfig
	Policy   PolicyConfig
}

type TopologyConfig struct {
	NetworkName     string             `json:"network_name"`
	ChainID         string             `json:"chain_id"`
	Mode            string             `json:"mode"`
	Binary          string             `json:"binary"`
	ContainerImage  string             `json:"container_image"`
	SeedNodes       []SeedNode         `json:"seed_nodes"`
	Validators      []ValidatorNode    `json:"validators"`
	Governance      GovernanceTopology `json:"governance"`
	PersistentPeers bool               `json:"persistent_peers"`
}

type SeedNode struct {
	ID      string `json:"id"`
	Moniker string `json:"moniker"`
	P2PHost string `json:"p2p_host"`
	P2PPort int    `json:"p2p_port"`
}

type ValidatorNode struct {
	ID             string `json:"id"`
	Moniker        string `json:"moniker"`
	Role           string `json:"role"`
	IP             string `json:"ip"`
	RPCPort        int    `json:"rpc_port"`
	P2PPort        int    `json:"p2p_port"`
	AppPort        int    `json:"app_port"`
	PrometheusPort int    `json:"prometheus_port"`
	VotingPower    int    `json:"voting_power"`
}

type GovernanceTopology struct {
	Houses                      []string `json:"houses"`
	CriticalActionsRequireDual  bool     `json:"critical_actions_require_dual_approval"`
}

type GenesisConfig struct {
	ChainID    string          `json:"chain_id"`
	LaunchMode string          `json:"launch_mode"`
	Denoms     GenesisDenoms   `json:"denoms"`
	Staking    GenesisStaking  `json:"staking"`
	Governance GenesisGov      `json:"governance"`
	Attest     GenesisAttest   `json:"attestation"`
	Incident   GenesisIncident `json:"incident"`
	Treasury   GenesisTreasury `json:"treasury"`
}

type GenesisDenoms struct {
	Base       string `json:"base"`
	Display    string `json:"display"`
	Reputation string `json:"reputation"`
}

type GenesisStaking struct {
	MinimumSelfBond    string `json:"minimum_self_bond"`
	UnbondingTimeSec   int    `json:"unbonding_time_seconds"`
	MaxValidators      int    `json:"max_validators"`
}

type GenesisGov struct {
	ProposalBond            string `json:"proposal_bond"`
	StandardVotingPeriodSec int    `json:"standard_voting_period_seconds"`
	HighImpactTimelockSec   int    `json:"high_impact_timelock_seconds"`
	SafetyCriticalSec       int    `json:"safety_critical_timelock_seconds"`
	DualHouseEnabled        bool   `json:"dual_house_enabled"`
}

type GenesisAttest struct {
	MinimumAuditorBond string `json:"minimum_auditor_bond"`
	AppealWindowSec    int    `json:"appeal_window_seconds"`
}

type GenesisIncident struct {
	PostmortemDeadlineSec int  `json:"postmortem_deadline_seconds"`
	PublicReasonCodes     bool `json:"public_reason_codes_required"`
}

type GenesisTreasury struct {
	FeeSplitPercent map[string]int `json:"fee_split_percent"`
}

type PolicyConfig struct {
	Version                  string            `json:"version"`
	RequiredBeforePublic     []string          `json:"required_before_public_launch"`
	ProhibitedShortcuts      []string          `json:"prohibited_shortcuts"`
	SafeModeDefaults         SafeModeDefaults  `json:"safe_mode_defaults"`
}

type SafeModeDefaults struct {
	PermissionedTestnet             bool `json:"permissioned_testnet"`
	PublicTransferability           bool `json:"public_transferability"`
	EmergencyPauseRequiresPostmortem bool `json:"emergency_pause_requires_postmortem"`
}

func LoadBundle(root string) (Bundle, error) {
	resolvedRoot, err := filepath.Abs(root)
	if err != nil {
		return Bundle{}, err
	}

	var topology TopologyConfig
	if err := loadJSON(filepath.Join(resolvedRoot, "config", "network-topology.json"), &topology); err != nil {
		return Bundle{}, err
	}

	var genesis GenesisConfig
	if err := loadJSON(filepath.Join(resolvedRoot, "config", "genesis", "base-genesis.json"), &genesis); err != nil {
		return Bundle{}, err
	}

	var policy PolicyConfig
	if err := loadJSON(filepath.Join(resolvedRoot, "config", "policy", "release-guards.json"), &policy); err != nil {
		return Bundle{}, err
	}

	return Bundle{
		Root:     resolvedRoot,
		Topology: topology,
		Genesis:  genesis,
		Policy:   policy,
	}, nil
}

func loadJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}
