{ pkgs }:
with pkgs; [
  just
  bun
  nodejs
  cargo
  rustc
  rust-analyzer
  rustfmt
  clippy
  neovim
  lazygit
  tmux
]
