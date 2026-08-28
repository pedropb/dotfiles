{
  description = "pedropb's cross-platform dotfiles";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
    home-manager = {
      url = "github:nix-community/home-manager/release-26.05";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    # Machine-specific and private configuration, as data. The tracked stub
    # keeps a fresh clone evaluable (and `nix flake check` meaningful); real
    # machines point this at ~/.config/dotfiles with --override-input, which
    # `just switch` does for you. See home/local-config.md.
    local = {
      url = "path:./home/local-stub";
      flake = false;
    };
  };

  outputs = { nixpkgs, home-manager, local, ... }:
    let
      inherit (nixpkgs) lib;

      systems = [ "x86_64-darwin" "aarch64-darwin" "x86_64-linux" "aarch64-linux" ];
      forAllSystems = f: lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});

      localConfig = builtins.fromTOML (builtins.readFile "${local}/local.toml");
    in
    {
      # The profile's system, user, and home directory come from local.toml,
      # so nothing here reads the ambient environment: evaluation is pure.
      homeConfigurations.default = home-manager.lib.homeManagerConfiguration {
        pkgs = nixpkgs.legacyPackages.${localConfig.machine.system};
        extraSpecialArgs = { inherit localConfig; };
        modules = [ ./home/default.nix ];
      };

      packages = forAllSystems (pkgs: {
        inherit (pkgs) home-manager;
        dotfiles-local = pkgs.callPackage ./tools/dotfiles-local/package.nix { };
        cln = pkgs.callPackage ./tools/cln/package.nix { };
      });
    };
}
