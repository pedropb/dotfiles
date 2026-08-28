{ lib, buildGoModule }:
buildGoModule {
  pname = "dotfiles-local";
  version = "0.1.0";
  src = ./.;

  # go-toml parses and writes the TOML: a real parser turns a typo into a
  # message with a line number, which matters because the file is meant to
  # be hand-edited too. bubbletea/huh/lipgloss (Charm) drive the `init` and
  # `edit` full-screen editor: interactive pickers, forms, and color instead
  # of hand-rolled line prompts.
  vendorHash = "sha256-BNb3/ecFV3gU+mZMQcbePyc198VJ1sel6VFfW3nJBZ0=";

  meta = {
    description = "Maintain the machine-specific, private half of the dotfiles Home Manager profile";
    mainProgram = "dotfiles-local";
    license = lib.licenses.mit;
    platforms = lib.platforms.unix;
  };
}
