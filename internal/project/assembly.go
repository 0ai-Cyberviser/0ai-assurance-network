package project

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
)

type CollectedValidatorManifest struct {
	Validator ValidatorPlanOutput `json:"validator"`
	Identity  DevelopmentIdentity `json:"identity"`
	Config    NodeConfigBundle    `json:"config"`
}

type GenesisAssemblyValidator struct {
	ID              string `json:"id"`
	Moniker         string `json:"moniker"`
	Role            string `json:"role"`
	VotingPower     int    `json:"voting_power"`
	PeerID          string `json:"peer_id"`
	ConsensusPubKey string `json:"consensus_public_key"`
	OperatorAddress string `json:"operator_address"`
	RewardAddress   string `json:"reward_address"`
}

type GenesisAssemblyPlan struct {
	NetworkName           string                     `json:"network_name"`
	ChainID               string                     `json:"chain_id"`
	CollectedValidatorIDs []string                   `json:"collected_validator_ids"`
	CollectedValidators   []GenesisAssemblyValidator `json:"validators"`
	CollectedVotingPower  int                        `json:"collected_voting_power"`
	CommitQuorumPower     int                        `json:"commit_quorum_power"`
	Genesis               map[string]any             `json:"genesis"`
}

type CollectedLocalnetBundle struct {
	Plan           GenesisAssemblyPlan `json:"plan"`
	NetworkSummary map[string]any      `json:"network_summary"`
	DockerCompose  string              `json:"docker_compose"`
}

func CollectValidatorBundle(bundlePath string) (CollectedValidatorManifest, error) {
	var manifest CollectedValidatorManifest
	if err := loadJSON(filepath.Join(bundlePath, "manifest.json"), &manifest); err != nil {
		return CollectedValidatorManifest{}, err
	}
	return manifest, nil
}

func WriteCollectedValidator(path string, manifest CollectedValidatorManifest) error {
	return WriteJSON(path, manifest)
}

func LoadCollectedValidators(collectionDir string) ([]CollectedValidatorManifest, error) {
	entries, err := os.ReadDir(collectionDir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		files = append(files, filepath.Join(collectionDir, entry.Name()))
	}
	sort.Strings(files)

	manifests := make([]CollectedValidatorManifest, 0, len(files))
	for _, file := range files {
		var manifest CollectedValidatorManifest
		if err := loadJSON(file, &manifest); err != nil {
			return nil, err
		}
		manifests = append(manifests, manifest)
	}
	return manifests, nil
}

func AssembleGenesisPlan(bundle Bundle, manifests []CollectedValidatorManifest) (GenesisAssemblyPlan, error) {
	if len(manifests) == 0 {
		return GenesisAssemblyPlan{}, fmt.Errorf("no collected validator manifests were supplied")
	}

	topologyByID := make(map[string]ValidatorNode, len(bundle.Topology.Validators))
	topologyIDs := make([]string, 0, len(bundle.Topology.Validators))
	totalVotingPower := 0
	for _, validator := range bundle.Topology.Validators {
		topologyByID[validator.ID] = validator
		topologyIDs = append(topologyIDs, validator.ID)
		totalVotingPower += validator.VotingPower
	}

	seenIDs := make(map[string]struct{}, len(manifests))
	collected := make([]GenesisAssemblyValidator, 0, len(manifests))
	collectedIDs := make([]string, 0, len(manifests))
	for _, manifest := range manifests {
		validatorID := manifest.Identity.ValidatorID
		if validatorID == "" {
			validatorID = manifest.Validator.ID
		}
		if validatorID == "" {
			return GenesisAssemblyPlan{}, fmt.Errorf("collected validator manifest is missing validator id")
		}
		if _, exists := seenIDs[validatorID]; exists {
			return GenesisAssemblyPlan{}, fmt.Errorf("duplicate collected validator id: %s", validatorID)
		}
		seenIDs[validatorID] = struct{}{}

		topologyValidator, exists := topologyByID[validatorID]
		if !exists {
			return GenesisAssemblyPlan{}, fmt.Errorf("collected validator %s does not exist in topology", validatorID)
		}

		expectedPlan, err := ValidatorPlan(bundle, validatorID)
		if err != nil {
			return GenesisAssemblyPlan{}, err
		}
		if !reflect.DeepEqual(expectedPlan, manifest.Validator) {
			return GenesisAssemblyPlan{}, fmt.Errorf("collected validator plan mismatch for %s", validatorID)
		}
		if manifest.Identity.ChainID != bundle.Topology.ChainID {
			return GenesisAssemblyPlan{}, fmt.Errorf(
				"collected validator %s has chain id %s but topology expects %s",
				validatorID,
				manifest.Identity.ChainID,
				bundle.Topology.ChainID,
			)
		}
		if manifest.Identity.Moniker != topologyValidator.Moniker {
			return GenesisAssemblyPlan{}, fmt.Errorf(
				"collected validator %s has moniker %s but topology expects %s",
				validatorID,
				manifest.Identity.Moniker,
				topologyValidator.Moniker,
			)
		}
		if manifest.Identity.Mode != "development_only" {
			return GenesisAssemblyPlan{}, fmt.Errorf(
				"collected validator %s identity mode must be development_only",
				validatorID,
			)
		}

		collectedIDs = append(collectedIDs, validatorID)
		collected = append(collected, GenesisAssemblyValidator{
			ID:              validatorID,
			Moniker:         topologyValidator.Moniker,
			Role:            topologyValidator.Role,
			VotingPower:     topologyValidator.VotingPower,
			PeerID:          manifest.Identity.PeerID,
			ConsensusPubKey: manifest.Identity.ConsensusPublicKey,
			OperatorAddress: manifest.Identity.OperatorAddress,
			RewardAddress:   manifest.Identity.RewardAddress,
		})
	}

	sort.Strings(topologyIDs)
	sort.Strings(collectedIDs)
	if len(collectedIDs) != len(topologyIDs) {
		return GenesisAssemblyPlan{}, fmt.Errorf(
			"collected validator set is incomplete: expected %d validators, received %d",
			len(topologyIDs),
			len(collectedIDs),
		)
	}
	for index := range topologyIDs {
		if topologyIDs[index] != collectedIDs[index] {
			return GenesisAssemblyPlan{}, fmt.Errorf(
				"collected validator set does not match topology: expected %v, received %v",
				topologyIDs,
				collectedIDs,
			)
		}
	}

	sort.Slice(collected, func(i, j int) bool {
		return collected[i].ID < collected[j].ID
	})

	collectedVotingPower := 0
	for _, validator := range collected {
		collectedVotingPower += validator.VotingPower
	}
	if collectedVotingPower != totalVotingPower {
		return GenesisAssemblyPlan{}, fmt.Errorf(
			"collected voting power %d does not match topology voting power %d",
			collectedVotingPower,
			totalVotingPower,
		)
	}

	genesis := RenderedGenesis(bundle)
	genesis["validators"] = collected

	return GenesisAssemblyPlan{
		NetworkName:           bundle.Topology.NetworkName,
		ChainID:               bundle.Topology.ChainID,
		CollectedValidatorIDs: collectedIDs,
		CollectedValidators:   collected,
		CollectedVotingPower:  collectedVotingPower,
		CommitQuorumPower:     ((collectedVotingPower * 2) / 3) + 1,
		Genesis:               genesis,
	}, nil
}

func WriteCollectedLocalnetBundle(path string, bundle CollectedLocalnetBundle) error {
	root := filepath.Clean(path)
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(root, "genesis.assembled.json"), bundle.Plan.Genesis); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(root, "network-summary.assembled.json"), bundle.NetworkSummary); err != nil {
		return err
	}
	if err := WriteJSON(filepath.Join(root, "validator-collection.json"), bundle.Plan.CollectedValidators); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(root, "docker-compose.assembled.yml"), []byte(bundle.DockerCompose), 0o644)
}

func RenderCollectedLocalnetBundle(bundle Bundle, manifests []CollectedValidatorManifest) (CollectedLocalnetBundle, error) {
	plan, err := AssembleGenesisPlan(bundle, manifests)
	if err != nil {
		return CollectedLocalnetBundle{}, err
	}
	return CollectedLocalnetBundle{
		Plan:           plan,
		NetworkSummary: NetworkSummary(bundle),
		DockerCompose:  DockerCompose(bundle),
	}, nil
}
