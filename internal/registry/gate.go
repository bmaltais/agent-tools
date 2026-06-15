package registry

import "time"

// HighRiskOperations is the definitive list of operation names subject to
// stale-state blocking. Callers must pass one of these strings (or an
// equivalent constant below) to Gate when requesting a sensitive action.
var HighRiskOperations = []string{
	OpInstall,
	OpUpdate,
	OpRegistryMutation,
	OpRemoteAdapterGeneration,
}

// Operation constants for use with GateOptions.Operation.
const (
	OpInstall                 = "install"
	OpUpdate                  = "update"
	OpRegistryMutation        = "registry_mutation"
	OpRemoteAdapterGeneration = "remote_adapter_generation"
)

// Refusal code constants returned in RefusalPayload.Code.
const (
	// CodeStaleStateGate is returned when a high-risk operation is blocked
	// because the registry state is stale and no override was supplied.
	CodeStaleStateGate = "stale_state_gate"
	// CodeOverrideInvalid is returned when an override was supplied but failed
	// validation (missing required fields).
	CodeOverrideInvalid = "override_invalid"
)

// RefusalPayload is the machine-parseable response returned when a high-risk
// operation is blocked by stale-state gating. It is never a plain error string.
type RefusalPayload struct {
	Code      string `json:"code"`
	Operation string `json:"operation"`
	Reason    string `json:"reason"`
}

// OverrideRequest is the caller-supplied bypass for a single stale-gated
// operation. All three fields are required for a valid override.
type OverrideRequest struct {
	Reason         string    `json:"reason"`
	CallerIdentity string    `json:"caller_identity"`
	Timestamp      time.Time `json:"timestamp"`
}

// AuditEntry records a successfully accepted override within the session.
// It is not persisted externally in v1.
type AuditEntry struct {
	Operation string          `json:"operation"`
	Override  OverrideRequest `json:"override"`
	At        time.Time       `json:"at"`
}

// GateOptions describes the operation being requested.
type GateOptions struct {
	// Operation is the name of the operation to check. Use the Op* constants.
	Operation string
	// Override, if non-nil, carries the caller-approved bypass for the gate.
	Override *OverrideRequest
}

// Gate checks whether the given operation should be blocked by stale-state
// gating.
//
//   - If the state is not stale, or the operation is not high-risk, both return
//     values are nil and the caller may proceed.
//   - If the operation is high-risk and the state is stale, Gate either returns
//     a non-nil *RefusalPayload (caller must not proceed) or, when a valid
//     override is supplied, a non-nil *AuditEntry (caller may proceed, override
//     recorded).
//
// A non-nil *RefusalPayload always means the operation is blocked.
// A non-nil *AuditEntry always means the operation was allowed via override.
func Gate(state SessionState, opts GateOptions) (*RefusalPayload, *AuditEntry) {
	if !state.Stale || !isHighRisk(opts.Operation) {
		return nil, nil
	}

	if opts.Override != nil {
		if err := validateOverride(opts.Override); err != nil {
			return &RefusalPayload{
				Code:      CodeOverrideInvalid,
				Operation: opts.Operation,
				Reason:    "override rejected: " + err.Error(),
			}, nil
		}
		return nil, &AuditEntry{
			Operation: opts.Operation,
			Override:  *opts.Override,
			At:        time.Now(),
		}
	}

	return &RefusalPayload{
		Code:      CodeStaleStateGate,
		Operation: opts.Operation,
		Reason:    "operation blocked: registry state is stale",
	}, nil
}

// isHighRisk reports whether op appears in HighRiskOperations.
func isHighRisk(op string) bool {
	for _, r := range HighRiskOperations {
		if op == r {
			return true
		}
	}
	return false
}

// validateOverride returns an error if any required override field is missing.
func validateOverride(o *OverrideRequest) error {
	if o.Reason == "" {
		return &overrideFieldError{"reason"}
	}
	if o.CallerIdentity == "" {
		return &overrideFieldError{"caller_identity"}
	}
	if o.Timestamp.IsZero() {
		return &overrideFieldError{"timestamp"}
	}
	return nil
}

type overrideFieldError struct{ field string }

func (e *overrideFieldError) Error() string { return "missing required field: " + e.field }
