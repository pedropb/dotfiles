{ pkgs, ... }:
let
  username = builtins.getEnv "USER";
  homeDirectory = builtins.getEnv "HOME";
in
{
  assertions = [
    {
      assertion = username != "" && homeDirectory != "";
      message = "Evaluate this Home Manager profile with --impure so USER and HOME are available.";
    }
  ];

  home.username = username;
  home.homeDirectory = homeDirectory;
  home.stateVersion = "24.11";

  # Keep WezTerm's GUI installation independent; Nix manages its configuration.
  home.packages = with pkgs; [
    just
    neovim
    starship
    tmux
  ];

  home.file = {
    ".zprofile".source = ../shell/zprofile;
    ".zshrc".source = ../shell/zshrc;
    ".tmux.conf".source = ../config/tmux/tmux.conf;
    ".gitconfig".source = ../config/git/config;
  };

  xdg.enable = true;
  xdg.configFile = {
    "nvim/init.lua".source = ../config/nvim/init.lua;
    "starship.toml".source = ../config/starship/starship.toml;
    "git/personal".source = ../config/git/personal;
    "wezterm/wezterm.lua".source = ../config/wezterm/wezterm.lua;
  };
}
