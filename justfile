
# List available dotfiles commands.
default:
  @just --list

# Evaluate the flake and this machine's profile without changing anything.
check:
  ./bin/check.sh

# Build and activate the Home Manager profile.
switch:
  ./bin/switch.sh

# First activation: prompt for private configuration, then activate, keeping conflicting files as *.before-home-manager.
bootstrap:
  ./bin/bootstrap.sh

# Print the machine-specific, private configuration this profile is built from.
local:
  nix run .#dotfiles-local -- show

# Report the profile's size, attributed to the dependencies in home/packages.nix.
# Flags pass through to the script: just sizes --json
sizes *args:
  ./bin/profile-size.sh {{args}}

# Run cln's test suite.
test-cln:
  cd tools/cln && go test ./...

# Run dotfiles-local's test suite.
test-local:
  cd tools/dotfiles-local && go test ./...

# Bump the omp package pin: fetch release SHAs and rewrite home/omp.nix.
omp-update version:
  ./bin/omp-update.sh {{version}}
