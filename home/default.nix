{ pkgs, lib, ... }:
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
    tmux
  ];

  home.file = {
    ".tmux.conf".source = ../config/tmux/tmux.conf;
    ".gitconfig".source = ../config/git/config;
  };

  programs = {
    starship = {
      enable = true;
      enableZshIntegration = true;
    };

    zsh = {
      enable = true;
      dotDir = homeDirectory;
      enableCompletion = true;
      autosuggestion.enable = true;
      syntaxHighlighting.enable = true;

      oh-my-zsh = {
        enable = true;
        plugins = [ "git" ];
      };

      profileExtra = ''
        export PATH="$HOME/.nix-profile/bin:$PATH:$HOME/.local/bin"
        export BAT_THEME="TwoDark"
      '';

      initContent = lib.mkOrder 900 ''
        source ${pkgs.zsh-you-should-use}/share/zsh/plugins/you-should-use/you-should-use.plugin.zsh
      '';
    };
  };

  xdg.enable = true;
  xdg.configFile = {
    "nvim/init.lua".source = ../config/nvim/init.lua;
    "starship.toml".source = ../config/starship/starship.toml;
    "git/personal".source = ../config/git/personal;
    "wezterm/wezterm.lua".source = ../config/wezterm/wezterm.lua;
  };
}
