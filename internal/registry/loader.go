package registry

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Registry is the agent-neutral tool registry contract.
type Registry struct {
	SV    string `json:"sv"`
	SM    int    `json:"sm"`
	TS    string `json:"ts"`
	Tools []Tool `json:"t"`
}

type Tool struct {
	ID   string   `json:"id"`
	Bin  string   `json:"bin"`
	St   string   `json:"st"`
	In   string   `json:"in"`
	Cmd  string   `json:"cmd"`
	Args []Arg    `json:"a"`
	Out  []string `json:"o"`
	Pre  []string `json:"p"`
	Side []string `json:"s"`
	Fail []string `json:"f"`
	Ex   []string `json:"x"`
}

type Arg struct {
	Key  string `json:"k"`
	Req  bool   `json:"r"`
	Type string `json:"t"`
	Desc string `json:"d"`
}

// Overlay is runtime-specific adapter metadata.
type Overlay struct {
	Runtime      string                     `json:"rt"`
	SchemaMajor  int                        `json:"sm"`
	Mode         map[string]string          `json:"m"`
	ToolAdapters map[string]RuntimeToolHint `json:"t"`
}

type RuntimeToolHint struct {
	PickWhen  []string `json:"pick_when,omitempty"`
	AvoidWhen []string `json:"avoid_when,omitempty"`
	Risk      string   `json:"risk,omitempty"`
}

type Manifest struct {
	SchemaMajor    int    `json:"schema_major"`
	RegistrySHA256 string `json:"registry_sha256"`
}

// RemoteProvider supplies optional refresh data.
type RemoteProvider interface {
	FetchManifest(ctx context.Context) (Manifest, error)
	FetchRegistry(ctx context.Context) ([]byte, error)
	FetchOverlay(ctx context.Context, runtime string) ([]byte, error)
}

type LoadOptions struct {
	RootDir  string
	Runtime  string
	Provider RemoteProvider
}

type RefreshWarning struct {
	Code    string
	Message string
}

type SessionState struct {
	Registry     Registry
	Overlay      Overlay
	RegistryHash string
	Source       string
	Stale        bool
	Warning      *RefreshWarning
}

// LoadSession loads local registry data and optionally refreshes from remote.
func LoadSession(ctx context.Context, opts LoadOptions) (SessionState, error) {
	runtime := opts.Runtime
	if runtime == "" {
		runtime = "copilot"
	}

	localRegBytes, localReg, localHash, localOv, err := loadLocal(opts.RootDir, runtime)
	if err != nil {
		return SessionState{}, err
	}

	state := SessionState{
		Registry:     localReg,
		Overlay:      localOv,
		RegistryHash: localHash,
		Source:       "local",
	}

	if opts.Provider == nil {
		return state, nil
	}

	manifest, err := opts.Provider.FetchManifest(ctx)
	if err != nil {
		state.Stale = true
		state.Warning = &RefreshWarning{Code: "refresh_manifest_failed", Message: err.Error()}
		return state, nil
	}

	if manifest.SchemaMajor != localReg.SM {
		state.Stale = true
		state.Warning = &RefreshWarning{Code: "refresh_schema_major_mismatch", Message: fmt.Sprintf("manifest=%d local=%d", manifest.SchemaMajor, localReg.SM)}
		return state, nil
	}

	if manifest.RegistrySHA256 == localHash {
		return state, nil
	}

	remoteRegBytes, err := opts.Provider.FetchRegistry(ctx)
	if err != nil {
		state.Stale = true
		state.Warning = &RefreshWarning{Code: "refresh_registry_fetch_failed", Message: err.Error()}
		return state, nil
	}
	remoteOvBytes, err := opts.Provider.FetchOverlay(ctx, runtime)
	if err != nil {
		state.Stale = true
		state.Warning = &RefreshWarning{Code: "refresh_overlay_fetch_failed", Message: err.Error()}
		return state, nil
	}

	remoteHash := sha256Hex(remoteRegBytes)
	if remoteHash != manifest.RegistrySHA256 {
		state.Stale = true
		state.Warning = &RefreshWarning{Code: "refresh_hash_mismatch", Message: fmt.Sprintf("manifest=%s fetched=%s", manifest.RegistrySHA256, remoteHash)}
		return state, nil
	}

	var remoteReg Registry
	if err := json.Unmarshal(remoteRegBytes, &remoteReg); err != nil {
		state.Stale = true
		state.Warning = &RefreshWarning{Code: "refresh_registry_parse_failed", Message: err.Error()}
		return state, nil
	}
	var remoteOv Overlay
	if err := json.Unmarshal(remoteOvBytes, &remoteOv); err != nil {
		state.Stale = true
		state.Warning = &RefreshWarning{Code: "refresh_overlay_parse_failed", Message: err.Error()}
		return state, nil
	}
	if remoteReg.SM != remoteOv.SchemaMajor {
		state.Stale = true
		state.Warning = &RefreshWarning{Code: "refresh_schema_overlay_mismatch", Message: fmt.Sprintf("registry=%d overlay=%d", remoteReg.SM, remoteOv.SchemaMajor)}
		return state, nil
	}

	_ = localRegBytes
	state.Registry = remoteReg
	state.Overlay = remoteOv
	state.RegistryHash = remoteHash
	state.Source = "remote-refresh"
	state.Stale = false
	state.Warning = nil
	return state, nil
}

func loadLocal(rootDir string, runtime string) ([]byte, Registry, string, Overlay, error) {
	if rootDir == "" {
		rootDir = "."
	}
	regPath := filepath.Join(rootDir, "tools", "registry.json")
	ovPath := filepath.Join(rootDir, "tools", "overlays", runtime+".json")

	regBytes, err := os.ReadFile(regPath)
	if err != nil {
		return nil, Registry{}, "", Overlay{}, fmt.Errorf("read local registry: %w", err)
	}
	ovBytes, err := os.ReadFile(ovPath)
	if err != nil {
		return nil, Registry{}, "", Overlay{}, fmt.Errorf("read local overlay: %w", err)
	}

	var reg Registry
	if err := json.Unmarshal(regBytes, &reg); err != nil {
		return nil, Registry{}, "", Overlay{}, fmt.Errorf("parse local registry: %w", err)
	}
	var ov Overlay
	if err := json.Unmarshal(ovBytes, &ov); err != nil {
		return nil, Registry{}, "", Overlay{}, fmt.Errorf("parse local overlay: %w", err)
	}

	if reg.SM != ov.SchemaMajor {
		return nil, Registry{}, "", Overlay{}, fmt.Errorf("local schema major mismatch: registry=%d overlay=%d", reg.SM, ov.SchemaMajor)
	}

	return regBytes, reg, sha256Hex(regBytes), ov, nil
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
