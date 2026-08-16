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
  glab
  tmux
  (pkgs.callPackage ./omp.nix { })
]
