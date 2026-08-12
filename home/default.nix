{ pkgs, lib, config, ... }:
let
  username = builtins.getEnv "USER";
  homeDirectory = builtins.getEnv "HOME";
  localModule = builtins.toPath "${homeDirectory}/src/github.com/pedropb/dotfiles/home/local.nix";
  conditionalIdentities = config.dotfiles.git.conditionalIdentities;
in
{

  imports = lib.optional (builtins.pathExists localModule) localModule;

  options.dotfiles.git.conditionalIdentities = lib.mkOption {
    type = lib.types.attrsOf (lib.types.submodule {
      options = {
        gitdir = lib.mkOption {
          type = lib.types.str;
          description = "Git directory prefix to match.";
        };
        name = lib.mkOption {
          type = lib.types.str;
          description = "Author name for matching repositories.";
        };
        email = lib.mkOption {
          type = lib.types.str;
          description = "Author email for matching repositories.";
        };
      };
    });
    default = { };
    description = "Path-specific Git author identities.";
  };

  config = {
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
    shadowenv
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
    "git/local".text = lib.concatStringsSep "\n" (lib.mapAttrsToList (name: identity: ''
      [includeIf "gitdir:${identity.gitdir}"]
        path = ~/.config/git/identities/${name}
    '') conditionalIdentities);
    "git/personal".source = ../config/git/personal;
    "wezterm/wezterm.lua".source = ../config/wezterm/wezterm.lua;
  } // lib.mapAttrs' (name: identity:
    lib.nameValuePair "git/identities/${name}" {
      text = ''
        [user]
          name = ${identity.name}
          email = ${identity.email}
      '';
    }
  ) conditionalIdentities;
  };
}
