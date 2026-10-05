#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 <version> [--push]" >&2
  echo "  <version>  release to pin, without the leading 'v' (e.g. 18.6.1)" >&2
  echo "  --push     commit the manifest and push, without prompting" >&2
  exit 1
}

version=""
push=0
for arg in "$@"; do
  case "$arg" in
    --push) push=1 ;;
    -h | --help) usage ;;
    *)
      if [[ -n "$version" ]]; then
        usage
      fi
      version="$arg"
      ;;
  esac
done
if [[ -z "$version" ]]; then
  usage
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
omp_nix="$repo_root/home/omp.nix"
base_url="https://github.com/can1357/oh-my-pi/releases/download/v${version}"

assets=(
  omp-darwin-arm64
  omp-darwin-x64
  omp-linux-arm64
  omp-linux-x64
)

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

# The release publishes SHA256SUMS.txt, so a single ~1 KB request yields every
# hash; the ~1 GB of assets never needs downloading here. `just switch` (via
# fetchurl) verifies these against the real downloads at build time, so a wrong
# or stale digest fails loudly instead of silently pinning bad bytes.
echo "Fetching $base_url" >&2
if ! curl -fsSL --retry 3 "$base_url/SHA256SUMS.txt" -o "$tmpdir/SHA256SUMS.txt"; then
  echo "error: no SHA256SUMS.txt for v$version; $omp_nix left untouched" >&2
  exit 1
fi

# Resolve every asset before editing, so a missing entry cannot half-update the file.
for asset in "${assets[@]}"; do
  awk -v asset="$asset" '{ name = $2; sub(/^[.*\/]+/, "", name); if (name == asset) print $1 }' \
    "$tmpdir/SHA256SUMS.txt" >"$tmpdir/$asset.hex"
  sri=$(nix hash convert --hash-algo sha256 --from base16 --to sri "$(<"$tmpdir/$asset.hex")" 2>/dev/null) || sri=""
  if [[ -z "$sri" ]]; then
    echo "error: no sha256 for $asset@$version; $omp_nix left untouched" >&2
    exit 1
  fi
  printf '%s' "$sri" >"$tmpdir/$asset.sri"
done

perl -pi -e "s/version = \"[^\"]*\";/version = \"$version\";/" "$omp_nix"
for asset in "${assets[@]}"; do
  sri=$(<"$tmpdir/$asset.sri")
  perl -0777 -pi -e "s{(asset = \"\Q$asset\E\";\s*\n\s*hash = \")[^\"]*(\";)}{\${1}$sri\${2}}" "$omp_nix"
  echo "  $asset -> $sri" >&2
done
echo "home/omp.nix updated to version $version" >&2

# Git flow: two prompts when interactive, or --push to commit and push silently.
commit_msg="update omp to $version"
do_commit=0
do_push=0
if ((push)); then
  do_commit=1
  do_push=1
elif [[ -t 0 ]]; then
  read -r -p "Commit home/omp.nix ($commit_msg)? [y/N] " reply
  if [[ "$reply" == [yY]* ]]; then
    do_commit=1
    read -r -p "Push? [y/N] " reply
    if [[ "$reply" == [yY]* ]]; then
      do_push=1
    fi
  fi
fi

if ((do_commit)); then
  if [[ -z "$(git -C "$repo_root" status --porcelain -- "$omp_nix")" ]]; then
    echo "nothing to commit; $omp_nix already at $version" >&2
  else
    git -C "$repo_root" commit -m "$commit_msg" -- "$omp_nix" >&2
    echo "committed: $commit_msg" >&2
  fi
fi

if ((do_push)); then
  git -C "$repo_root" push
fi
