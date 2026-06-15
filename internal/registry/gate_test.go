package registry

import (
	"testing"
	"time"
)

func TestGateHighRiskBlockedWhenStale(t *testing.T) {
	state := SessionState{Stale: true}
	refusal, audit := Gate(state, GateOptions{Operation: OpInstall})
	if refusal == nil {
		t.Fatal("expected refusal payload, got nil")
	}
	if refusal.Code != CodeStaleStateGate {
		t.Fatalf("unexpected code: %q", refusal.Code)
	}
	if refusal.Operation != OpInstall {
		t.Fatalf("unexpected operation: %q", refusal.Operation)
	}
	if audit != nil {
		t.Fatal("expected nil audit entry when blocked")
	}
}

func TestGateNonHighRiskPassesWhenStale(t *testing.T) {
	state := SessionState{Stale: true}
	refusal, audit := Gate(state, GateOptions{Operation: "list"})
	if refusal != nil {
		t.Fatalf("expected no refusal for non-high-risk op, got %+v", refusal)
	}
	if audit != nil {
		t.Fatal("expected nil audit for non-high-risk op")
	}
}

func TestGateHighRiskPassesWhenFresh(t *testing.T) {
	state := SessionState{Stale: false}
	refusal, audit := Gate(state, GateOptions{Operation: OpUpdate})
	if refusal != nil {
		t.Fatalf("expected no refusal when state is fresh, got %+v", refusal)
	}
	if audit != nil {
		t.Fatal("expected nil audit when state is fresh")
	}
}

func TestGateOverrideAcceptedWithValidFields(t *testing.T) {
	state := SessionState{Stale: true}
	override := &OverrideRequest{
		Reason:         "manual approval by operator",
		CallerIdentity: "ops-team",
		Timestamp:      time.Now(),
	}
	refusal, audit := Gate(state, GateOptions{Operation: OpRegistryMutation, Override: override})
	if refusal != nil {
		t.Fatalf("expected no refusal with valid override, got %+v", refusal)
	}
	if audit == nil {
		t.Fatal("expected audit entry when override accepted")
	}
	if audit.Operation != OpRegistryMutation {
		t.Fatalf("audit operation mismatch: %q", audit.Operation)
	}
	if audit.Override.Reason != override.Reason {
		t.Fatalf("audit reason mismatch: %q", audit.Override.Reason)
	}
	if audit.Override.CallerIdentity != override.CallerIdentity {
		t.Fatalf("audit caller_identity mismatch: %q", audit.Override.CallerIdentity)
	}
	if audit.At.IsZero() {
		t.Fatal("expected non-zero audit timestamp")
	}
}

func TestGateOverrideRejectedMissingReason(t *testing.T) {
	state := SessionState{Stale: true}
	override := &OverrideRequest{
		CallerIdentity: "ops-team",
		Timestamp:      time.Now(),
	}
	refusal, audit := Gate(state, GateOptions{Operation: OpRemoteAdapterGeneration, Override: override})
	if refusal == nil {
		t.Fatal("expected refusal when override missing reason")
	}
	if refusal.Code != CodeOverrideInvalid {
		t.Fatalf("unexpected code: %q", refusal.Code)
	}
	if audit != nil {
		t.Fatal("expected nil audit when override rejected")
	}
}

func TestGateOverrideRejectedMissingCallerIdentity(t *testing.T) {
	state := SessionState{Stale: true}
	override := &OverrideRequest{
		Reason:    "emergency patch",
		Timestamp: time.Now(),
	}
	refusal, audit := Gate(state, GateOptions{Operation: OpUpdate, Override: override})
	if refusal == nil {
		t.Fatal("expected refusal when override missing caller_identity")
	}
	if refusal.Code != CodeOverrideInvalid {
		t.Fatalf("unexpected code: %q", refusal.Code)
	}
	if audit != nil {
		t.Fatal("expected nil audit when override rejected")
	}
}

func TestGateOverrideRejectedMissingTimestamp(t *testing.T) {
	state := SessionState{Stale: true}
	override := &OverrideRequest{
		Reason:         "emergency patch",
		CallerIdentity: "ops-team",
	}
	refusal, audit := Gate(state, GateOptions{Operation: OpInstall, Override: override})
	if refusal == nil {
		t.Fatal("expected refusal when override missing timestamp")
	}
	if refusal.Code != CodeOverrideInvalid {
		t.Fatalf("unexpected code: %q", refusal.Code)
	}
	if audit != nil {
		t.Fatal("expected nil audit when override rejected")
	}
}

func TestGateAllHighRiskOpsBlocked(t *testing.T) {
	state := SessionState{Stale: true}
	for _, op := range HighRiskOperations {
		refusal, _ := Gate(state, GateOptions{Operation: op})
		if refusal == nil {
			t.Errorf("expected refusal for high-risk op %q, got nil", op)
		}
	}
}
