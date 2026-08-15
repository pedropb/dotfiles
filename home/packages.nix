{ pkgs }:
with pkgs; [
  just
  bun
  nodejs
  cargo
  rustc
  rust-analyzer
  rustfmt
  lua
  clippy
  neovim
  lazygit
  tmux
]
