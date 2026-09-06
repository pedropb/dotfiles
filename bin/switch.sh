#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

dir="$(nix run .#dotfiles-local -- ensure)"
nix run .#home-manager -- switch --flake .#default --override-input local "path:$dir"