#!/usr/bin/env bash
# install.sh — install all GA-status agent-tools binaries to ~/.local/bin
# (or $AGENT_TOOLS_BIN_DIR if set).
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/bmaltais/agent-tools/main/install.sh | bash
#
# Environment variables:
#   AGENT_TOOLS_BIN_DIR      Override install directory (default: ~/.local/bin)
#   AGENT_TOOLS_RELEASE_TAG  Pin a specific release tag (default: latest)

set -euo pipefail

REPO="bmaltais/agent-tools"
BIN_DIR="${AGENT_TOOLS_BIN_DIR:-${HOME}/.local/bin}"
RELEASE_TAG="${AGENT_TOOLS_RELEASE_TAG:-}"

# ---- platform detection ----

case "$(uname -s)" in
  Linux)  OS=linux  ;;
  Darwin) OS=darwin ;;
  *)
    printf 'error: unsupported OS "%s". Supported: Linux, Darwin.\n' "$(uname -s)" >&2
    exit 1
    ;;
esac

case "$(uname -m)" in
  x86_64)        ARCH=amd64 ;;
  aarch64|arm64) ARCH=arm64 ;;
  *)
    printf 'error: unsupported architecture "%s". Supported: x86_64, aarch64/arm64.\n' "$(uname -m)" >&2
    exit 1
    ;;
esac

# ---- sha256 helper ----

sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | cut -d' ' -f1
  else
    printf 'error: no sha256 tool found (need sha256sum or shasum).\n' >&2
    exit 1
  fi
}

# ---- resolve release tag ----

if [ -z "$RELEASE_TAG" ]; then
  printf 'Fetching latest release of %s...\n' "$REPO"
  RELEASE_TAG=$(curl -sf "https://api.github.com/repos/${REPO}/releases/latest" \
    | python3 -c "import sys, json; print(json.load(sys.stdin)['tag_name'])")
fi

printf 'Release: %s  Platform: %s/%s\n' "$RELEASE_TAG" "$OS" "$ARCH"

# ---- download metadata ----

WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

BASE_URL="https://github.com/${REPO}/releases/download/${RELEASE_TAG}"
RAW_BASE="https://raw.githubusercontent.com/${REPO}/${RELEASE_TAG}"

curl -sfL   -o "${WORK}/manifest.json" "${BASE_URL}/release-manifest.json"
curl -sf    -o "${WORK}/registry.json" "${RAW_BASE}/tools/registry.json"
curl -sfL   -o "${WORK}/SHA256SUMS"    "${BASE_URL}/SHA256SUMS"

# ---- resolve GA artifacts for this platform ----

ARTIFACT_NAMES=$(python3 - "$WORK" "$OS" "$ARCH" <<'PY'
import json, sys

workdir, os_val, arch_val = sys.argv[1], sys.argv[2], sys.argv[3]

reg      = json.loads(open(f"{workdir}/registry.json").read())
manifest = json.loads(open(f"{workdir}/manifest.json").read())

ga_ids = {t["id"] for t in reg["t"] if t.get("st") == "ga"}

matches = [
    a["name"]
    for a in manifest["artifacts"]
    if a["tool"] in ga_ids and a["os"] == os_val and a["arch"] == arch_val
]

if not matches:
    sys.stderr.write(
        f"error: no GA artifacts for {os_val}/{arch_val} in release "
        f"{manifest.get('release_tag', '?')}\n"
    )
    sys.exit(1)

print("\n".join(matches))
PY
)

# ---- download, verify, and install each binary ----

mkdir -p "$BIN_DIR"
COUNT=0

while IFS= read -r artifact; do
  [ -z "$artifact" ] && continue

  # Extract tool name: strip "agent-tools_" prefix, strip "-v<version>_<os>_<arch>" suffix
  TOOL=$(printf '%s' "$artifact" | sed 's/^agent-tools_//;s/-v[0-9].*//')

  printf 'Downloading %s...\n' "$artifact"
  curl -sf -L -o "${WORK}/${artifact}" "${BASE_URL}/${artifact}"

  # Verify SHA256 against published SHA256SUMS
  EXPECTED=$(grep "  ${artifact}$" "${WORK}/SHA256SUMS" | cut -d' ' -f1)
  if [ -z "$EXPECTED" ]; then
    printf 'error: no SHA256 entry for "%s" in SHA256SUMS.\n' "$artifact" >&2
    exit 1
  fi

  ACTUAL=$(sha256_of "${WORK}/${artifact}")
  if [ "$ACTUAL" != "$EXPECTED" ]; then
    printf 'error: SHA256 mismatch for %s\n  expected: %s\n  actual:   %s\n' \
      "$artifact" "$EXPECTED" "$ACTUAL" >&2
    exit 1
  fi

  chmod +x "${WORK}/${artifact}"
  cp "${WORK}/${artifact}" "${BIN_DIR}/${TOOL}"
  printf 'Installed: %s  →  %s\n' "$TOOL" "${BIN_DIR}/${TOOL}"
  COUNT=$((COUNT + 1))
done <<< "$ARTIFACT_NAMES"

printf '\nInstalled %d tool(s) to %s\n' "$COUNT" "$BIN_DIR"
printf 'Make sure %s is on your $PATH.\n' "$BIN_DIR"
