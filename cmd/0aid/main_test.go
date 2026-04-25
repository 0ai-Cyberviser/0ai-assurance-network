package main

import (
	"strings"
	"testing"
)

func TestReadActivationAuditPromotionPackagesRejectsEmptyPathElements(t *testing.T) {
	_, err := readActivationAuditPromotionPackages(",build/rotation/promotion.json", ",build/rotation/verification.json")
	if err == nil || !strings.Contains(err.Error(), "paths must not be empty") {
		t.Fatalf("expected empty path element error, got %v", err)
	}
}

func TestReadRetainedInventoryContinuityPackagesRejectsEmptyPathElements(t *testing.T) {
	_, err := readRetainedInventoryContinuityPackages(",build/rotation/inventory.json", ",build/rotation/inventory-verification.json")
	if err == nil || !strings.Contains(err.Error(), "paths must not be empty") {
		t.Fatalf("expected empty path element error, got %v", err)
	}
}
