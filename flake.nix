{
  description = "pedropb's cross-platform dotfiles";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-26.05";
    home-manager = {
      url = "github:nix-community/home-manager/release-26.05";
      inputs.nixpkgs.follows = "nixpkgs";
    };

    # x86_64-darwin (Intel Mac) is deprecated in Nixpkgs: 26.05 is its last
    # supported release. Everything else tracks the stable input above;
    # Intel Darwin machines pin here instead so they keep evaluating once
    # `nixpkgs` moves past 26.05.
    nixpkgsIntelDarwin.url = "github:NixOS/nixpkgs/nixpkgs-26.05-darwin";
    home-managerIntelDarwin = {
      url = "github:nix-community/home-manager/release-26.05";
      inputs.nixpkgs.follows = "nixpkgsIntelDarwin";
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

  outputs = { nixpkgs, home-manager, nixpkgsIntelDarwin, home-managerIntelDarwin, local, ... }:
    let
      inherit (nixpkgs) lib;

      systems = [ "x86_64-darwin" "aarch64-darwin" "x86_64-linux" "aarch64-linux" ];

      # Intel Darwin resolves through the pinned 26.05 inputs above; every
      # other system tracks the stable, rolling `nixpkgs`/`home-manager`.
      isIntelDarwin = system: system == "x86_64-darwin";
      nixpkgsFor = system: if isIntelDarwin system then nixpkgsIntelDarwin else nixpkgs;
      homeManagerFor = system: if isIntelDarwin system then home-managerIntelDarwin else home-manager;
      pkgsFor = system: (nixpkgsFor system).legacyPackages.${system};

      forAllSystems = f: lib.genAttrs systems (system: f (pkgsFor system));

      localConfig = builtins.fromTOML (builtins.readFile "${local}/local.toml");
    in
    {
      # The profile's system, user, and home directory come from local.toml,
      # so nothing here reads the ambient environment: evaluation is pure.
      homeConfigurations.default = (homeManagerFor localConfig.machine.system).lib.homeManagerConfiguration {
        pkgs = pkgsFor localConfig.machine.system;
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
