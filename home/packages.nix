{ pkgs }:
with pkgs;
lib.optional stdenv.isDarwin flameshot
++ [
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
  tree
  (pkgs.callPackage ../tools/cln/package.nix { })
  (pkgs.callPackage ../tools/dotfiles-local/package.nix { })
  (pkgs.callPackage ./omp.nix { })
]
