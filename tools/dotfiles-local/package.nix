{ lib, buildGoModule }:
buildGoModule {
  pname = "dotfiles-local";
  version = "0.1.0";
  src = ./.;

  # go-toml is the one dependency: hand-rolling a TOML parser would be fine for
  # writing, but this file is meant to be hand-edited too, and a real parser is
  # what turns a typo into a message with a line number.
  vendorHash = "sha256-2U9hBA+q3nW4YO47PN+eorBLq0z4kj2zCZ6q7wD75PQ=";

  meta = {
    description = "Maintain the machine-specific, private half of the dotfiles Home Manager profile";
    mainProgram = "dotfiles-local";
    license = lib.licenses.mit;
    platforms = lib.platforms.unix;
  };
}
