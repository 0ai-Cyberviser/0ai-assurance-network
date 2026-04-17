package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDevelopmentIdentityForValidatorIsDeterministic(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	first, err := DevelopmentIdentityForValidator(bundle, "val-2")
	if err != nil {
		t.Fatalf("DevelopmentIdentityForValidator failed: %v", err)
	}
	second, err := DevelopmentIdentityForValidator(bundle, "val-2")
	if err != nil {
		t.Fatalf("DevelopmentIdentityForValidator failed: %v", err)
	}

	if first != second {
		t.Fatal("expected deterministic identity output")
	}
	if first.Mode != "development_only" {
		t.Fatalf("unexpected mode: %s", first.Mode)
	}
	if !strings.Contains(first.Warning, "Do not use") && !strings.Contains(strings.ToLower(first.Warning), "do not use") {
		t.Fatalf("unexpected warning: %s", first.Warning)
	}
}

func TestNodeInitPlan(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	plan, err := NodeInitPlan(bundle, "val-6")
	if err != nil {
		t.Fatalf("NodeInitPlan failed: %v", err)
	}

	if plan.Validator.Moniker != "validator-6" {
		t.Fatalf("unexpected moniker: %s", plan.Validator.Moniker)
	}
	if len(plan.Config.PersistentPeers) != 1 {
		t.Fatalf("expected one persistent peer, got %d", len(plan.Config.PersistentPeers))
	}
	if !strings.HasPrefix(plan.Identity.OperatorAddress, "0aioper1") {
		t.Fatalf("unexpected operator address: %s", plan.Identity.OperatorAddress)
	}
}

func TestWriteNodeInitBundle(t *testing.T) {
	bundle, err := LoadBundle("../..")
	if err != nil {
		t.Fatalf("LoadBundle failed: %v", err)
	}

	plan, err := NodeInitPlan(bundle, "val-1")
	if err != nil {
		t.Fatalf("NodeInitPlan failed: %v", err)
	}

	output := t.TempDir()
	if err := WriteNodeInitBundle(output, plan); err != nil {
		t.Fatalf("WriteNodeInitBundle failed: %v", err)
	}

	for _, relative := range []string{
		"manifest.json",
		filepath.Join("config", "identity.json"),
		filepath.Join("config", "node.json"),
		"README.txt",
	} {
		target := filepath.Join(output, relative)
		if _, err := os.Stat(target); err != nil {
			t.Fatalf("expected %s to exist: %v", target, err)
		}
	}
}
