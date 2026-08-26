{ pkgs }:
with pkgs; [
  just
  bun
  nodejs
  go
  gopls
  cargo
  rustc
  rust-analyzer
  rustfmt
  lua
  clippy
  neovim
  lazygit
  gh
  glab
  tmux
  (pkgs.callPackage ../tools/cln/package.nix { })
  (pkgs.callPackage ./omp.nix { })
]
