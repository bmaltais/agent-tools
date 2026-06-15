package registry

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type fakeProvider struct {
	manifest    Manifest
	manifestErr error
	reg         []byte
	regErr      error
	overlay     []byte
	overlayErr  error
}

func (f fakeProvider) FetchManifest(context.Context) (Manifest, error) {
	if f.manifestErr != nil {
		return Manifest{}, f.manifestErr
	}
	return f.manifest, nil
}

func (f fakeProvider) FetchRegistry(context.Context) ([]byte, error) {
	if f.regErr != nil {
		return nil, f.regErr
	}
	return f.reg, nil
}

func (f fakeProvider) FetchOverlay(context.Context, string) ([]byte, error) {
	if f.overlayErr != nil {
		return nil, f.overlayErr
	}
	return f.overlay, nil
}

func TestLoadSessionLocalOnly(t *testing.T) {
	root := writeLocalFixture(t)
	state, err := LoadSession(context.Background(), LoadOptions{RootDir: root, Runtime: "copilot"})
	if err != nil {
		t.Fatalf("LoadSession error: %v", err)
	}
	if state.Stale {
		t.Fatalf("expected fresh local state")
	}
	if state.Source != "local" {
		t.Fatalf("source mismatch: got %q", state.Source)
	}
	if state.Registry.SM != 1 || state.Overlay.SchemaMajor != 1 {
		t.Fatalf("schema major mismatch in loaded state")
	}
}

func TestLoadSessionRemoteRefreshSuccess(t *testing.T) {
	root := writeLocalFixture(t)
	remoteRegistry := `{"sv":"1.1.0","sm":1,"ts":"2026-06-15","t":[{"id":"patch-verify","bin":"patch-verify","st":"ga","in":"x","cmd":"c","a":[],"o":[],"p":[],"s":[],"f":[],"x":[]}]}`
	remoteOverlay := `{"rt":"copilot","sm":1,"m":{"discovery":"registry_first"},"t":{"patch-verify":{"risk":"low"}}}`

	provider := fakeProvider{manifest: Manifest{SchemaMajor: 1, RegistrySHA256: sha256Hex([]byte(remoteRegistry))}, reg: []byte(remoteRegistry), overlay: []byte(remoteOverlay)}
	state, err := LoadSession(context.Background(), LoadOptions{RootDir: root, Runtime: "copilot", Provider: provider})
	if err != nil {
		t.Fatalf("LoadSession error: %v", err)
	}
	if state.Stale {
		t.Fatalf("expected fresh state after remote refresh")
	}
	if state.Source != "remote-refresh" {
		t.Fatalf("source mismatch: got %q", state.Source)
	}
	if state.Registry.SV != "1.1.0" {
		t.Fatalf("expected refreshed registry version, got %q", state.Registry.SV)
	}
}

func TestLoadSessionRemoteFailureSetsStale(t *testing.T) {
	root := writeLocalFixture(t)
	provider := fakeProvider{manifestErr: errors.New("network down")}
	state, err := LoadSession(context.Background(), LoadOptions{RootDir: root, Runtime: "copilot", Provider: provider})
	if err != nil {
		t.Fatalf("LoadSession error: %v", err)
	}
	if !state.Stale {
		t.Fatalf("expected stale state on refresh failure")
	}
	if state.Warning == nil || state.Warning.Code != "refresh_manifest_failed" {
		t.Fatalf("expected refresh_manifest_failed warning, got %+v", state.Warning)
	}
	if state.Source != "local" {
		t.Fatalf("expected local fallback, got %q", state.Source)
	}
}

// TestLoadSessionRejectsUnsupportedSchemaMajor verifies the adapter
// compatibility contract: if the overlay declares a different schema major than
// the registry, the loader must fail fast rather than returning a stale or
// partially-loaded session.
func TestLoadSessionRejectsUnsupportedSchemaMajor(t *testing.T) {
	root := t.TempDir()
	// Registry is at sm=2; overlay is still at sm=1 — incompatible adapter.
	mustWriteFile(t, filepath.Join(root, "tools", "registry.json"), `{"sv":"2.0.0","sm":2,"ts":"2026-06-15","t":[{"id":"patch-verify","bin":"patch-verify","st":"ga","in":"x","cmd":"c","a":[],"o":[],"p":[],"s":[],"f":[],"x":[]}]}`)
	mustWriteFile(t, filepath.Join(root, "tools", "overlays", "copilot.json"), `{"rt":"copilot","sm":1,"m":{"discovery":"registry_first"},"t":{"patch-verify":{"risk":"low"}}}`)

	_, err := LoadSession(context.Background(), LoadOptions{RootDir: root, Runtime: "copilot"})
	if err == nil {
		t.Fatal("expected error when registry sm != overlay sm, got nil")
	}
}

// TestLoadSessionAcceptsSupportedSchemaMajor is the positive counterpart:
// when registry sm and overlay sm agree, the session loads without error.
func TestLoadSessionAcceptsSupportedSchemaMajor(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "tools", "registry.json"), `{"sv":"1.0.0","sm":1,"ts":"2026-06-15","t":[{"id":"patch-verify","bin":"patch-verify","st":"ga","in":"x","cmd":"c","a":[],"o":[],"p":[],"s":[],"f":[],"x":[]}]}`)
	mustWriteFile(t, filepath.Join(root, "tools", "overlays", "copilot.json"), `{"rt":"copilot","sm":1,"m":{"discovery":"registry_first"},"t":{"patch-verify":{"risk":"low"}}}`)

	state, err := LoadSession(context.Background(), LoadOptions{RootDir: root, Runtime: "copilot"})
	if err != nil {
		t.Fatalf("expected no error when sm values agree, got: %v", err)
	}
	if state.Stale {
		t.Fatal("expected fresh state when sm values agree")
	}
	if state.Registry.SM != state.Overlay.SchemaMajor {
		t.Fatalf("sm mismatch in loaded state: registry=%d overlay=%d", state.Registry.SM, state.Overlay.SchemaMajor)
	}
}

func writeLocalFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "tools", "registry.json"), `{"sv":"1.0.0","sm":1,"ts":"2026-06-15","t":[{"id":"patch-verify","bin":"patch-verify","st":"ga","in":"x","cmd":"c","a":[],"o":[],"p":[],"s":[],"f":[],"x":[]}]}`)
	mustWriteFile(t, filepath.Join(root, "tools", "overlays", "copilot.json"), `{"rt":"copilot","sm":1,"m":{"discovery":"registry_first"},"t":{"patch-verify":{"risk":"low"}}}`)
	return root
}

func mustWriteFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
