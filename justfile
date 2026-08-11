
# List available dotfiles commands.
default:
  @just --list

# Evaluate the flake without changing the machine.
check:
  nix flake check --impure

# Build and activate the selected Home Manager profile.
switch:
  nix run --impure .#home-manager -- switch --impure --flake .#default

# First activation preserves conflicting files as *.before-home-manager.
bootstrap:
  nix run --impure .#home-manager -- switch --impure -b before-home-manager --flake .#default
