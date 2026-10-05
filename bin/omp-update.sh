#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <version>" >&2
  exit 1
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
omp_nix="$repo_root/home/omp.nix"
version="$1"

assets=(
  omp-darwin-arm64
  omp-darwin-x64
  omp-linux-arm64
  omp-linux-x64
)

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

# Fetch + convert all hashes concurrently; each writes its SRI to $tmpdir/<asset>.
pids=()
for asset in "${assets[@]}"; do
  echo "Fetching $asset@$version..." >&2
  (
    base32=$(nix-prefetch-url --type sha256 \
      "https://github.com/can1357/oh-my-pi/releases/download/v${version}/${asset}" 2>/dev/null | tail -1)
    sri=$(nix hash convert --hash-algo sha256 --to sri "$base32" 2>/dev/null)
    printf '%s' "$sri" >"$tmpdir/$asset"
  ) &
  pids+=("$!")
done

failed=0
for pid in "${pids[@]}"; do
  wait "$pid" || failed=1
done
if ((failed)); then
  echo "error: one or more fetches failed; $omp_nix left untouched" >&2
  exit 1
fi

# Merge only after every fetch succeeded, so the file is never half-updated.
perl -pi -e "s/version = \"[^\"]*\";/version = \"$version\";/" "$omp_nix"

for asset in "${assets[@]}"; do
  sri=$(<"$tmpdir/$asset")
  perl -0777 -pi -e "s{(asset = \"\Q$asset\E\";\s*\n\s*hash = \")[^\"]*(\";)}{\${1}$sri\${2}}" "$omp_nix"
  echo "  $asset -> $sri" >&2
done

echo "home/omp.nix updated to version $version" >&2
