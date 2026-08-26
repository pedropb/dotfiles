{ pkgs, lib, config, ... }:
let
  username = builtins.getEnv "USER";
  homeDirectory = builtins.getEnv "HOME";
  localModule = builtins.toPath "${homeDirectory}/src/github.com/pedropb/dotfiles/home/local.nix";
  conditionalIdentities = config.dotfiles.git.conditionalIdentities;
  credentialHelpers = config.dotfiles.git.credentialHelpers;
  clnProviders = config.dotfiles.cln.providers;
in
{

  imports = lib.optional (builtins.pathExists localModule) localModule;

  options.dotfiles.git = {
    conditionalIdentities = lib.mkOption {
      type = lib.types.attrsOf (lib.types.submodule {
        options = {
          gitdir = lib.mkOption {
            type = lib.types.either lib.types.str (lib.types.listOf lib.types.str);
            description = "Git directory prefix (or prefixes) to match.";
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

    credentialHelpers = lib.mkOption {
      type = lib.types.attrsOf lib.types.str;
      default = { };
      example = { "gitlab.example.com" = "!glab auth git-credential"; };
      description = ''
        Git credential helper command per HTTPS host, keyed by the host as the
        remote URL spells it. Each entry clears the inherited helper chain for
        that host first, so platform helpers never answer in the CLI's place.
      '';
    };
  };

  options.dotfiles.cln = {
    defaultProvider = lib.mkOption {
      type = lib.types.str;
      default = "gh";
      description = "Alias of the cln provider used when none is given explicitly.";
    };

    providers = lib.mkOption {
      type = lib.types.attrsOf (lib.types.submodule {
        options = {
          type = lib.mkOption {
            type = lib.types.enum [ "github" "gitlab" ];
            description = "Backend cln uses to build clone URLs and list repositories.";
          };
          host = lib.mkOption {
            type = lib.types.str;
            description = ''Hostname cln clones from, e.g. "github.com".'';
          };
          defaultNamespace = lib.mkOption {
            type = lib.types.nullOr lib.types.str;
            default = null;
            description = "Namespace used when a repository is given without one.";
          };
        };
      });
      default = { };
      example = {
        corp = {
          type = "gitlab";
          host = "git.example.com";
        };
      };
      description = ''
        cln providers, rendered to ~/.config/cln/config.toml. Add private,
        non-public forges in the git-ignored home/local.nix instead of here.
      '';
    };
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
  home.sessionVariables.EDITOR = "nvim";


  # Keep WezTerm's GUI installation independent; Nix manages its configuration.
  home.packages = import ./packages.nix { inherit pkgs; };

  # The gh CLI answers for github.com; private forges add their own helper in
  # home/local.nix.
  dotfiles.git.credentialHelpers."github.com" = lib.mkDefault "!gh auth git-credential";
  dotfiles.cln.providers.gh = lib.mkDefault {
    type = "github";
    host = "github.com";
    defaultNamespace = "pedropb";
  };

  home.file = {
    ".gitconfig".source = ../config/git/config;
  };
  home.activation.createPersonalGitHubDirectory =
    lib.hm.dag.entryAfter [ "writeBoundary" ] ''
      $DRY_RUN_CMD mkdir -p "$HOME/src/github.com/pedropb"
    '';


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
        source "$HOME/.config/zsh/scd.zsh"
      '';
    };
  };

  xdg.enable = true;
  xdg.configFile = {
    "nvim/init.lua".source = ../config/nvim/init.lua;
    "zsh/scd.zsh".source = ../config/zsh/scd.zsh;
    "tmux/tmux.conf".source = ../config/tmux/tmux.conf;
    "starship.toml".source = ../config/starship/starship.toml;
    "stylua/stylua.toml".source = ../config/stylua/stylua.toml;
    "git/local".text = lib.concatStringsSep "\n" (
      lib.mapAttrsToList (host: helper: ''
        [credential "https://${host}"]
          helper =
          helper = ${helper}
      '') credentialHelpers
      ++ lib.concatLists (lib.mapAttrsToList (name: identity:
        map (gitdir: ''
          [includeIf "gitdir:${gitdir}"]
            path = ~/.config/git/identities/${name}
        '') (lib.toList identity.gitdir)
      ) conditionalIdentities)
    );
    "git/personal".source = ../config/git/personal;
    "wezterm/wezterm.lua".source = ../config/wezterm/wezterm.lua;
    "cmux/cmux.json".source = ../config/cmux/cmux.json;
    "ghostty/config".source = ../config/ghostty/config;
    "cln/config.toml".source = (pkgs.formats.toml { }).generate "cln-config.toml" (
      { default_provider = config.dotfiles.cln.defaultProvider; }
      // lib.optionalAttrs (clnProviders != { }) {
        providers = lib.mapAttrs (_: p:
          { inherit (p) type host; }
          // lib.optionalAttrs (p.defaultNamespace != null) {
            default_namespace = p.defaultNamespace;
          }
        ) clnProviders;
      }
    );
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
