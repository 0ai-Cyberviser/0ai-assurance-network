package project

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAssembleGenesisPlanDeterministic(t *testing.T) {
	bundle, err := LoadBundle(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("load bundle: %v", err)
	}

	tempDir := t.TempDir()
	collectionDir := filepath.Join(tempDir, "collection")

	for index := len(bundle.Topology.Validators) - 1; index >= 0; index-- {
		validator := bundle.Topology.Validators[index]
		nodeBundle, err := NodeInitPlan(bundle, validator.ID)
		if err != nil {
			t.Fatalf("node init plan for %s: %v", validator.ID, err)
		}
		nodeDir := filepath.Join(tempDir, "nodes", validator.ID)
		if err := WriteNodeInitBundle(nodeDir, nodeBundle); err != nil {
			t.Fatalf("write node bundle for %s: %v", validator.ID, err)
		}

		collected, err := CollectValidatorBundle(nodeDir)
		if err != nil {
			t.Fatalf("collect node bundle for %s: %v", validator.ID, err)
		}
		outPath := filepath.Join(collectionDir, "manifest-"+validator.Moniker+".json")
		if err := WriteCollectedValidator(outPath, collected); err != nil {
			t.Fatalf("write collected manifest for %s: %v", validator.ID, err)
		}
	}

	manifests1, err := LoadCollectedValidators(collectionDir)
	if err != nil {
		t.Fatalf("load collected validators (pass 1): %v", err)
	}
	plan1, err := AssembleGenesisPlan(bundle, manifests1)
	if err != nil {
		t.Fatalf("assemble genesis plan (pass 1): %v", err)
	}

	manifests2, err := LoadCollectedValidators(collectionDir)
	if err != nil {
		t.Fatalf("load collected validators (pass 2): %v", err)
	}
	plan2, err := AssembleGenesisPlan(bundle, manifests2)
	if err != nil {
		t.Fatalf("assemble genesis plan (pass 2): %v", err)
	}

	if !reflect.DeepEqual(plan1, plan2) {
		t.Fatalf("expected deterministic assembly plan output")
	}
	if len(plan1.CollectedValidators) != len(bundle.Topology.Validators) {
		t.Fatalf("expected %d validators, received %d", len(bundle.Topology.Validators), len(plan1.CollectedValidators))
	}
	if plan1.CommitQuorumPower != 67 {
		t.Fatalf("expected commit quorum power 67, received %d", plan1.CommitQuorumPower)
	}

	localnetBundle, err := RenderCollectedLocalnetBundle(bundle, manifests1)
	if err != nil {
		t.Fatalf("render collected localnet bundle: %v", err)
	}
	outDir := filepath.Join(tempDir, "assembled")
	if err := WriteCollectedLocalnetBundle(outDir, localnetBundle); err != nil {
		t.Fatalf("write collected localnet bundle: %v", err)
	}
}

func TestAssembleGenesisPlanRejectsDuplicateValidatorIDs(t *testing.T) {
	bundle, err := LoadBundle(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("load bundle: %v", err)
	}

	nodeBundle, err := NodeInitPlan(bundle, bundle.Topology.Validators[0].ID)
	if err != nil {
		t.Fatalf("node init plan: %v", err)
	}
	duplicate := CollectedValidatorManifest{
		Validator: nodeBundle.Validator,
		Identity:  nodeBundle.Identity,
		Config:    nodeBundle.Config,
	}

	_, err = AssembleGenesisPlan(bundle, []CollectedValidatorManifest{duplicate, duplicate})
	if err == nil {
		t.Fatal("expected duplicate validator id error")
	}
	if !strings.Contains(err.Error(), "duplicate collected validator id") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAssembleGenesisPlanRejectsTopologyMismatch(t *testing.T) {
	bundle, err := LoadBundle(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("load bundle: %v", err)
	}

	nodeBundle, err := NodeInitPlan(bundle, bundle.Topology.Validators[0].ID)
	if err != nil {
		t.Fatalf("node init plan: %v", err)
	}
	mismatch := CollectedValidatorManifest{
		Validator: nodeBundle.Validator,
		Identity:  nodeBundle.Identity,
		Config:    nodeBundle.Config,
	}
	mismatch.Validator.Moniker = "mismatched-moniker"

	_, err = AssembleGenesisPlan(bundle, []CollectedValidatorManifest{mismatch})
	if err == nil {
		t.Fatal("expected topology mismatch error")
	}
	if !strings.Contains(err.Error(), "collected validator plan mismatch") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAssembleGenesisPlanRejectsIncompleteValidatorSet(t *testing.T) {
	bundle, err := LoadBundle(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("load bundle: %v", err)
	}

	nodeBundle, err := NodeInitPlan(bundle, bundle.Topology.Validators[0].ID)
	if err != nil {
		t.Fatalf("node init plan: %v", err)
	}
	partial := CollectedValidatorManifest{
		Validator: nodeBundle.Validator,
		Identity:  nodeBundle.Identity,
		Config:    nodeBundle.Config,
	}

	_, err = AssembleGenesisPlan(bundle, []CollectedValidatorManifest{partial})
	if err == nil {
		t.Fatal("expected incomplete validator set error")
	}
	if !strings.Contains(err.Error(), "collected validator set is incomplete") {
		t.Fatalf("unexpected error: %v", err)
	}
}
