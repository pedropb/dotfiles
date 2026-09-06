#!/usr/bin/env bash
# Initialize private configuration and activate this repository for the first time.
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

nix run .#dotfiles-local -- init
dir="$(nix run .#dotfiles-local -- ensure)"
nix run .#home-manager -- switch -b before-home-manager --flake .#default \
  --override-input local "path:$dir"