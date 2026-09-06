#!/usr/bin/env bash
# Evaluate the flake and this machine's profile without changing anything.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

nix flake check
nix run .#dotfiles-local -- check
dir="$(nix run .#dotfiles-local -- ensure)"
nix eval --raw .#homeConfigurations.default.activationPackage.drvPath \
  --override-input local "path:$dir" >/dev/null
echo "profile evaluates"
