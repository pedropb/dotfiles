
# List available dotfiles commands.
default:
  @just --list

# Evaluate the flake and this machine's profile without changing anything.
check:
  #!/usr/bin/env bash
  set -euo pipefail
  nix flake check
  nix run .#dotfiles-local -- check
  dir="$(nix run .#dotfiles-local -- ensure)"
  nix eval --raw .#homeConfigurations.default.activationPackage.drvPath \
    --override-input local "path:$dir" >/dev/null
  echo "profile evaluates"

# Build and activate the Home Manager profile.
switch:
  #!/usr/bin/env bash
  set -euo pipefail
  dir="$(nix run .#dotfiles-local -- ensure)"
  nix run .#home-manager -- switch --flake .#default --override-input local "path:$dir"

# First activation: prompt for private configuration, then activate, keeping conflicting files as *.before-home-manager.
bootstrap:
  #!/usr/bin/env bash
  set -euo pipefail
  nix run .#dotfiles-local -- init
  dir="$(nix run .#dotfiles-local -- ensure)"
  nix run .#home-manager -- switch -b before-home-manager --flake .#default \
    --override-input local "path:$dir"

# Print the machine-specific, private configuration this profile is built from.
local:
  nix run .#dotfiles-local -- show

# Run cln's test suite.
test-cln:
  cd tools/cln && go test ./...

# Run dotfiles-local's test suite.
test-local:
  cd tools/dotfiles-local && go test ./...

# Bump the omp package pin: fetch release SHAs and rewrite home/omp.nix.
omp-update version:
  ./bin/omp-update.sh {{version}}
