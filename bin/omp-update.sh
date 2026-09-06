#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 <version>" >&2
  exit 1
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
omp_nix="$repo_root/home/omp.nix"
version="$1"

perl -pi -e "s/version = \"[^\"]*\";/version = \"$version\";/" "$omp_nix"

for asset in omp-darwin-arm64 omp-darwin-x64 omp-linux-arm64 omp-linux-x64; do
  echo "Fetching $asset@$version..." >&2
  base32=$(nix-prefetch-url --type sha256 "https://github.com/can1357/oh-my-pi/releases/download/v${version}/${asset}" 2>/dev/null | tail -1)
  sri=$(nix hash convert --hash-algo sha256 --to sri "$base32" 2>/dev/null)
  perl -0777 -pi -e "s{(asset = \"\Q$asset\E\";\s*\n\s*hash = \")[^\"]*(\";)}{\${1}$sri\${2}}" "$omp_nix"
  echo "  $asset -> $sri" >&2
done

echo "home/omp.nix updated to version $version" >&2
